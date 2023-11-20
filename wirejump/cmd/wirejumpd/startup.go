package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"wirejump/internal/cli"
	"wirejump/internal/network"
	"wirejump/internal/providers"
	"wirejump/internal/state"
	"wirejump/internal/utils"
)

const programName = "wirejumpd"
const programDesc = "WireJump connection manager daemon"

var programUsage = []string{
	"  -c, --config PATH\tUse specified config file (required, no default)",
}

func init() {
	cli.SetProgramInfo(programName, programDesc, programUsage)

	// Remove log timestamp as systemd will add its own
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func ErrorExit(msg ...any) {
	fmt.Println(msg...)
	os.Exit(1)
}

// Handle default program args
func HandleProgramArgs() string {
	var configFile string

	// If no args are supplied
	if len(os.Args) == 1 {
		cli.ProgramUsage()

		os.Exit(0)
	}

	// Create default flag set
	fs := cli.MakeDefaultProgramOptions()

	// Add required options
	fs.StringVar(&configFile, "c", "", "config")
	fs.StringVar(&configFile, "config", "", "config")

	// Parse
	fs.Parse(os.Args[1:])

	// Check if config path is provided
	if configFile == "" {
		ErrorExit("Config file is required")
	}

	return configFile
}

// Parse and validate config file
func ParseConfig(configFile string) (state.ConfigurationState, error) {
	config := state.ConfigurationState{}

	// Try to parse config
	cfg, err := utils.ReadINI(configFile)

	if err != nil {
		return state.ConfigurationState{}, err
	}

	if cfg["Config"][0]["Upstream"] == "" {
		return state.ConfigurationState{}, errors.New("'Upstream' interface name is required")
	} else {
		config.UpstreamName = cfg["Config"][0]["Upstream"]
	}

	if cfg["Config"][0]["Downstream"] == "" {
		return state.ConfigurationState{}, errors.New("'Downstream' interface name is required")
	} else {
		config.DownstreamName = cfg["Config"][0]["Downstream"]
	}

	return config, nil
}

// Update application state based on app config
func UpdateAppState(State *state.ProtectedState, Conf *state.ConfigurationState) error {
	if State == nil {
		return errors.New("app state is nil")
	}

	if Conf == nil {
		return errors.New("configuration state is nil")
	}

	// Create upstream interface and parse its config if available
	upstream, err := network.CreateInterfaceFromConfig(Conf.UpstreamName, network.InterfaceKindUpstream)

	if err != nil {
		return fmt.Errorf("failed to create initial upstream interface: %s", err)
	}

	// Create downstream and parse its config if available
	downstream, err := network.CreateInterfaceFromConfig(Conf.DownstreamName, network.InterfaceKindDownstream)

	if err != nil {
		return fmt.Errorf("failed to create initial downstream interface: %s", err)
	}

	// Downstream MUST have a valid address
	if downstream.Address == "" {
		return fmt.Errorf("downstream interface MUST have a valid address")
	}

	// Load and check available providers
	providers := providers.LoadAvailableProviders()

	if providers.Available == nil || providers.Names == nil {
		return errors.New("no upstream providers are available")
	}

	// Lock and update the state
	State.Mutex.Lock()

	State.State.Network = network.NetworkState{
		Upstream:   &upstream,
		Downstream: &downstream,
	}

	State.State.Config = Conf
	State.State.AvailableProviders = providers

	State.Mutex.Unlock()

	return nil
}
