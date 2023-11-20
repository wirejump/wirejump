package handlers

import (
	"wirejump/internal/ipc"
	"wirejump/internal/state"
)

// Get list of all available providers
func (h *IpcHandler) ListProviders(State *state.AppState, Params *ipc.ListProvidersRequest, Reply *interface{}) error {
	*Reply = ipc.ListProvidersReply{
		Providers: State.AvailableProviders.Names,
	}

	return nil
}
