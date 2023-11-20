package handlers

import (
	"errors"
	"fmt"
	"log"
	"time"
	"wirejump/internal/ipc"
	"wirejump/internal/network"
	"wirejump/internal/providers"
	"wirejump/internal/state"
	"wirejump/internal/utils"
)

// Shutdown existing connection. Will be reused by Reset command
func Disconnect(State *state.AppState) error {
	// Upstream does not exist at all
	if State.Network.Upstream == nil {
		return nil
	}

	active, err := State.Network.Upstream.IsActive()

	// err means that interface is neither found nor exists == it's down,
	// so it's safe to ignore it until actual bringdown fails here:
	if active && err == nil {
		err = State.Network.Upstream.BringDown()

		if err != nil {
			return err
		}
	}

	// Reset current provider info
	State.UpstreamProvider.Server = nil
	State.UpstreamProvider.ActiveSince = nil

	return nil
}

// Connection handler. Will rotate keys on reconnect.
// Will connect using preset location or reconnect existing location.
// Detailed (re)connection logic:
// - shut down previous connection if it's active
// - remove current interface key from upstream
// - create new interace key and update it in the account
// - select new upstream which != previous upstream
// - bring connection back up
func (h *IpcHandler) Connect(State *state.AppState, Params *ipc.ConnectCommandRequest, Reply *interface{}) error {
	new_location := ""
	new_upstream := providers.WireguardServer{}

	if State.UpstreamProvider == nil {
		return errors.New("setup a provider first")
	}

	if !State.UpstreamProvider.Provider.Initialized {
		return errors.New("provider is not initialized")
	}

	// Create upstream interface if it does not exist
	if State.Network.Upstream == nil {
		iface, err := network.CreateInterface(State.Config.UpstreamName, network.InterfaceKindUpstream)

		if err != nil {
			return err
		}

		State.Network.Upstream = &iface
	}

	// Go for upstream guessing first if there was no explicit disconnect request.
	// This will clarify location and save existing connection if something is wrong.
	if !Params.Disconnect {
		// Nudge upstream server cache
		if UpstreamCacheIsBad(State) {
			if err := UpdateUpstreamServers(State); err != nil {
				return fmt.Errorf("connect needs fresh servers, but update has failed: %s", err)
			}
		}

		// No location preference provided, go for random one; TODO: debug log it
		if State.UpstreamProvider.PreferredLocation == nil {
			randomLocation := providers.GetRandomElement(State.Servers.Locations)

			if randomLocation != nil {
				new_location = *randomLocation
			} else {
				return errors.New("no server locations available, check provider settings")
			}
		} else {
			new_location = *State.UpstreamProvider.PreferredLocation
		}

		// If location override is specified, try to use it
		if Params.LocationOverride != nil {
			override := *Params.LocationOverride

			if !IsValidLocation(State, override) {
				return fmt.Errorf("location '%s' is not found", override)
			}

			new_location = override
		}

		// Determine upstream server
		upstream, err := providers.GuessNewUpstream(State.Servers, State.UpstreamProvider.Server, new_location)

		// Fail early and preserve current connection if there's no new upstream available
		if err != nil {
			return fmt.Errorf("unable to guess upstream: %s", err)
		}

		new_upstream = upstream
	}

	// Shut down existing connection
	if err := Disconnect(State); err != nil {
		return fmt.Errorf("failed to shutdown existing connection: %s", err)
	}

	// Disconnect requested
	if Params.Disconnect {
		return nil
	}

	// Rotate keys
	if !Params.PreserveKeys {
		// Remove current key from the account
		if err := State.UpstreamProvider.Provider.RemovePubkey(State.Network.Upstream.PublicKey); err != nil {
			log.Println("failed to remove old pubkey:", err)
		}

		// Create new private key or quit. That's pretty important,
		// since new connection can't be made without a key
		if err := State.Network.Upstream.GeneratePrivateKey(); err != nil {
			return fmt.Errorf("failed to create new private key: %s", err)
		}

		// Create new public key or quit, same restrictions apply
		if err := State.Network.Upstream.GeneratePublicKey(); err != nil {
			return fmt.Errorf("failed to create new public key: %s", err)
		}

		// Add generated pubkey to the account
		if err := State.UpstreamProvider.Provider.AddPubkey(State.Network.Upstream.PublicKey); err != nil {
			return fmt.Errorf("failed to add key to the account: %s", err)
		}
	}

	// Get interface address
	if addr, err := State.UpstreamProvider.Provider.GetAddress(State.Network.Upstream.PublicKey); err != nil {
		return fmt.Errorf("failed to get upstream IP address: %s", err)
	} else {
		State.Network.Upstream.Address = addr
	}

	// Get interface scripts and ignore errors, as interface and script actions
	// are certainly defined at this point
	upscript, _ := State.Network.Upstream.GetInterfaceScriptPath("up")
	downscript, _ := State.Network.Upstream.GetInterfaceScriptPath("down")

	// Assemble final upstream config
	config := utils.INIFile{
		"Interface": {
			utils.INIPair{
				"Address":    State.Network.Upstream.Address,
				"PrivateKey": State.Network.Upstream.PrivateKey,
				"Table":      "off", // This is needed because custom routing table will be used
				"PostUp":     upscript,
				"PreDown":    downscript,
			},
		},
		"Peer": {
			utils.INIPair{
				"PublicKey":  new_upstream.Pubkey,
				"AllowedIPs": "0.0.0.0/0",
				"Endpoint":   fmt.Sprintf("%s:%d", new_upstream.IPv4, new_upstream.Port),
			},
		},
	}

	// Write new interface config, since both upstream and address can be new at this point
	if err := State.Network.Upstream.WriteConfig(config); err != nil {
		return fmt.Errorf("failed to write interface config: %s", err)
	}

	// Finally bring interface back up
	if err := State.Network.Upstream.BringUp(); err != nil {
		return fmt.Errorf("failed to bring interface up: %s", err)
	}

	// Record current time
	t := time.Now().Unix()
	State.UpstreamProvider.ActiveSince = &t

	// Finally, new_upstream has proven itself good, so it can be updated
	State.UpstreamProvider.Server = &new_upstream

	return nil
}
