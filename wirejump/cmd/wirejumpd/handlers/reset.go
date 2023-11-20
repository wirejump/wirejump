package handlers

import (
	"fmt"
	"os"
	"wirejump/internal/ipc"
	"wirejump/internal/state"
)

// Remove current interface key from the account if provider
// is still active and reset interface config; it won't work
// without the key anyway.
func ResetProvider(State *state.AppState) error {
	if err := Disconnect(State); err != nil {
		return fmt.Errorf("failed to disconnect: %s", err)
	}

	if State.Network.Upstream != nil && State.UpstreamProvider != nil {
		file, err := State.Network.Upstream.GetInterfaceConfigPath()

		if err != nil {
			return err
		}

		// Remove current pubkey
		if err := State.UpstreamProvider.Provider.RemovePubkey(State.Network.Upstream.PublicKey); err != nil {
			return fmt.Errorf("cannot remove old pubkey: %s", err)
		}

		// Truncate config file to preserve symlink
		if file, err := os.Create(file); err != nil {
			return fmt.Errorf("cannot delete old interface config: %s", err)
		} else {
			defer file.Close()
		}
	}

	// Reset pointers
	State.Servers = nil
	State.Network.Upstream = nil
	State.UpstreamProvider = nil

	return nil
}

// Reset current upstream & connection
func (h *IpcHandler) Reset(State *state.AppState, Params *ipc.ResetCommandRequest, Reply *interface{}) error {
	return ResetProvider(State)
}
