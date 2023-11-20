package network

import "wirejump/internal/utils"

// Base directory which should hold scripts & configs
const BasePath = "/opt/wirejump"

// Where to write upstream gateway changes
const UpstreamGatewayConfig = "upstream_gateway"

// Interface config file extension
const InterfaceConfigSuffix = "conf"

// Interface is either upstream or downstream
const (
	InterfaceKindUpstream   = "upstream"
	InterfaceKindDownstream = "downstream"
)

// Interface script action can be either up or down
const (
	InterfaceScriptKindUp   = "up"
	InterfaceScriptKindDown = "down"
)

// Type of InterfaceKind* settings
type InterfaceKindType string

// Wireguard interface struct
type InterfaceConfig struct {
	// Name of the interface in the system. Will be passed to `wg-quick` command
	Name string

	// Interface kind, to pick a right config file. Either `upstream` or `downstream`
	Kind InterfaceKindType

	// Interface address
	Address string

	// Interface private key
	PrivateKey string

	// Interface public key
	PublicKey string
}

// Wireguard interface methods
type InterfaceControl interface {
	GetInterfaceConfigPath() (string, error)
	GetInterfaceScriptPath(string) (string, error)
	GeneratePrivateKey() error
	GeneratePublicKey() error
	BringUp() error
	BringDown() error
	UpdatePeerConfig(string, string, []string) error
	IsActive() (bool, error)
	UpdateDefaultGateway(string) error
	ReadConfig() (utils.INIFile, error)
	WriteConfig(utils.INIFile) error
}

// Current interface state
type NetworkState struct {
	Upstream   *InterfaceConfig
	Downstream *InterfaceConfig
}
