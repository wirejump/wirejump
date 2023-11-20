package network

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"wirejump/internal/utils"
)

var Base64Regex = regexp.MustCompile(`^[A-Za-z0-9+\/]+={0,3}$`)

// Curve25519 keys are always 32 bytes long
func IsValidKey(key string) bool {
	data, err := base64.StdEncoding.DecodeString(key)

	// If key can't be decoded, it's not a valid one
	if err != nil {
		return false
	}

	return Base64Regex.Match([]byte(key)) && err == nil && len(data) == 32
}

func IsValidCIDR(cidr string) bool {
	//netip.
	_, _, e := net.ParseCIDR(cidr)

	return e == nil
}

func IsValidIP(ip string) bool {
	_, err := netip.ParseAddr(ip)

	return err == nil
}

// Get next free IP address for the given prefix with respect to already taken addresses.
// Totally inefficient, but fun. Returns error if there are no free addresses left.
func GetFreeIP(occupied []netip.Addr, prefix netip.Prefix) (netip.Addr, error) {
	pool := []netip.Addr{}
	source := prefix.Addr()

	// Return early for huge ranges
	if prefix.Bits() < 16 {
		return netip.Addr{}, errors.New("really? you are using more than /16 pool here?")
	}

	// Get all possible addresses for this prefix, excluding occupied ones
	for {
		already := false
		next := source.Next()

		if prefix.Contains(next) {
			for _, taken := range occupied {
				if taken == next {
					already = true
					break
				}
			}

			// Not occupied yet
			if !already {
				pool = append(pool, next)
			}
		} else {
			break
		}

		// Occupied, advance
		source = source.Next()
	}

	total := len(pool)

	if total == 0 {
		return netip.Addr{}, errors.New("IP pool is exhausted")
	}

	index := rand.Intn(total)

	return pool[index], nil
}

// Create initial interface state and generate interface keys
func CreateInterface(name string, kind InterfaceKindType) (InterfaceConfig, error) {
	if kind != InterfaceKindUpstream && kind != InterfaceKindDownstream {
		return InterfaceConfig{}, errors.New("invalid interface kind")
	}

	Interface := InterfaceConfig{
		Name: name,
		Kind: kind,
	}

	// Generate private key
	if err := Interface.GeneratePrivateKey(); err != nil {
		return InterfaceConfig{}, fmt.Errorf("failed to generate private key: %s", err)
	}

	// Generate public key
	if err := Interface.GeneratePublicKey(); err != nil {
		return InterfaceConfig{}, fmt.Errorf("failed to generate public key: %s", err)
	}

	return Interface, nil
}

// Create interface state from the config file.
func CreateInterfaceFromConfig(name string, kind InterfaceKindType) (InterfaceConfig, error) {
	created, err := CreateInterface(name, kind)

	if err != nil {
		return InterfaceConfig{}, err
	}

	p, err := created.GetInterfaceConfigPath()

	if err != nil {
		return InterfaceConfig{}, err
	}

	pconfig, err := utils.ReadINI(p)

	if err != nil {
		return InterfaceConfig{}, err
	}

	parsed, ok := pconfig["Interface"]

	// Config file is broken/empty, overwrite it with current interface state
	if !ok {
		config := utils.INIFile{
			"Interface": {
				utils.INIPair{
					"PrivateKey": created.PrivateKey,
				},
			},
		}

		return created, created.WriteConfig(config)
	}

	// Update private key
	if privkey, ok := parsed[0]["PrivateKey"]; ok {
		created.PrivateKey = privkey
	}

	// Regenerate private key if needed
	if !IsValidKey(created.PrivateKey) {
		if err := created.GeneratePrivateKey(); err != nil {
			return InterfaceConfig{}, err
		}
	}

	// Update public key
	if pubkey, ok := parsed[0]["PublicKey"]; ok {
		created.PublicKey = pubkey
	}

	// Regenerate public key if needed
	if !IsValidKey(created.PublicKey) {
		if err := created.GeneratePublicKey(); err != nil {
			return InterfaceConfig{}, err
		}
	}

	// Update address
	if addr, ok := parsed[0]["Address"]; ok {
		created.Address = addr
	}

	// Reset address if it's invalid
	if !IsValidCIDR(created.Address) {
		created.Address = ""
	}

	return created, nil
}

// Get interface config path according to its kind.
// Will return error if interface kind is not specified
func (i *InterfaceConfig) GetInterfaceConfigPath() (string, error) {
	if i == nil {
		return "", errors.New("interface ptr is nil")
	}

	if i.Kind != InterfaceKindUpstream && i.Kind != InterfaceKindDownstream {
		return "", errors.New("interface kind is undefined")
	}

	return path.Join(BasePath, "config", fmt.Sprintf("%s.%s", i.Kind, InterfaceConfigSuffix)), nil
}

// Since connection is going to be managed by wg-quick tool, these scripts
// should be added to interface config file to execute additional actions
func (i *InterfaceConfig) GetInterfaceScriptPath(action InterfaceKindType) (string, error) {
	if i == nil {
		return "", errors.New("interface ptr is nil")
	}

	if action != InterfaceScriptKindUp && action != InterfaceScriptKindDown {
		return "", errors.New("unknown interface script action")
	}

	if i.Kind != InterfaceKindUpstream && i.Kind != InterfaceKindDownstream {
		return "", errors.New("interface kind is undefined")
	}

	return path.Join(BasePath, "scripts", fmt.Sprintf("%s.sh \"%%i\" %s", i.Kind, action)), nil
}

// Generate private key
func (i *InterfaceConfig) GeneratePrivateKey() error {
	if i == nil {
		return errors.New("interface ptr is nil")
	}

	out, err := exec.Command("wg", "genkey").Output()

	if err != nil {
		return err
	}

	i.PrivateKey = string(bytes.Trim(out, "\r\n"))

	return nil
}

// Generate public key from a private one
func (i *InterfaceConfig) GeneratePublicKey() error {
	if i == nil {
		return errors.New("interface ptr is nil")
	}

	var out bytes.Buffer
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = bytes.NewBuffer([]byte(i.PrivateKey))
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return err
	}

	i.PublicKey = strings.Trim(out.String(), "\r\n")

	return nil
}

// Update interface configuration for the particular peer
func (i *InterfaceConfig) UpdatePeerConfig(operation string, pubkey string, allowed []string) error {
	if i == nil {
		return errors.New("interface ptr is nil")
	}

	if operation != "add" && operation != "remove" {
		return errors.New("unknown operation requested")
	}

	if !IsValidKey(pubkey) {
		return errors.New("invalid pubkey")
	}

	command := []string{"wg", "set", i.Name, "peer", pubkey}

	if operation == "add" {
		joined := strings.Join(allowed, ",")
		command = append(command, "allowed-ips", joined)
	} else {
		command = append(command, "remove")
	}

	stderr := new(strings.Builder)
	cmd := exec.Command("sudo", command...)
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return errors.New(strings.Trim(stderr.String(), "\r\n"))
	}

	return nil
}

// Bring interface up
func (i *InterfaceConfig) BringUp() error {
	if i == nil {
		return errors.New("interface ptr is nil")
	}

	stderr := new(strings.Builder)
	cmd := exec.Command("sudo", "wg-quick", "up", i.Name)
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return errors.New(strings.Trim(stderr.String(), "\r\n"))
	}

	return nil
}

// Bring interface down
func (i *InterfaceConfig) BringDown() error {
	if i == nil {
		return errors.New("interface ptr is nil")
	}

	stderr := new(strings.Builder)
	cmd := exec.Command("sudo", "wg-quick", "down", i.Name)
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return errors.New(strings.Trim(stderr.String(), "\r\n"))
	}

	return nil
}

// Check whether interface is up or not
func (i *InterfaceConfig) IsActive() (bool, error) {
	if i == nil {
		return false, errors.New("interface ptr is nil")
	}

	interfaces, err := net.Interfaces()

	if err != nil {
		return false, err
	} else {
		for _, iface := range interfaces {
			if i.Name == iface.Name {
				return iface.Flags&net.FlagUp != 0, nil
			}
		}
	}

	return false, errors.New("interface not found")
}

// Overwrite default interface gateway file
func (i *InterfaceConfig) UpdateDefaultGateway(addr string) error {
	gatewayPath := path.Join(BasePath, "config", UpstreamGatewayConfig)
	file, err := os.Create(gatewayPath)

	if err != nil {
		return err
	}

	defer file.Close()

	file.Truncate(0)
	file.Seek(0, 0)

	_, e := fmt.Fprintln(file, addr)

	return e
}

// Read interface config from file
func (i *InterfaceConfig) ReadConfig() (utils.INIFile, error) {
	if i == nil {
		return utils.INIFile{}, errors.New("interface ptr is nil")
	}

	p, err := i.GetInterfaceConfigPath()

	if err != nil {
		return utils.INIFile{}, err
	}

	cfg, err := utils.ReadINI(p)

	if err != nil {
		return utils.INIFile{}, err
	}

	return cfg, nil
}

// Write interface config to the config file
func (i *InterfaceConfig) WriteConfig(c utils.INIFile) error {
	if i == nil {
		return errors.New("interface ptr is nil")
	}

	p, err := i.GetInterfaceConfigPath()

	if err != nil {
		return err
	}

	if err := utils.WriteINI(p, c); err != nil {
		return err
	}

	return nil
}
