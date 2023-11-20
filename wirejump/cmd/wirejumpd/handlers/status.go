package handlers

import (
	"wirejump/internal/ipc"
	"wirejump/internal/state"
)

func stringOrNil(value string) *string {
	if value == "" {
		return nil
	} else {
		return &value
	}
}

// Display current server/connection status
func (h *IpcHandler) Status(State *state.AppState, Params *ipc.StatusCommandRequest, Reply *interface{}) error {
	// Upstream is not initialized yet
	if State.UpstreamProvider == nil {
		*Reply = ipc.StatusCommandReply{}

		return nil
	}

	// Create initial provider info
	provider := ipc.ProviderStatus{
		Name: stringOrNil(State.UpstreamProvider.Provider.ProviderName),
	}

	// Fill preferred location
	provider.PreferredLocation = State.UpstreamProvider.PreferredLocation

	// Get account expiration date
	expires := State.UpstreamProvider.Provider.ValidUntil

	// In case expiration date is available
	if expires != 0 {
		provider.AccountExpires = &expires
	}

	// Construct initial upstream status
	upstream := ipc.ConnectionStatus{
		Online:      false,
		ActiveSince: State.UpstreamProvider.ActiveSince,
	}

	if State.Network.Upstream != nil {
		// Interface can be missing since it's a dynamic one
		c, _ := State.Network.Upstream.IsActive()
		upstream.Online = c
	}

	// Fill in connection details if available
	if State.UpstreamProvider.Server != nil {
		upstream.City = stringOrNil(State.UpstreamProvider.Server.City)
		upstream.Country = stringOrNil(State.UpstreamProvider.Server.Country)
	}

	// Create reply
	*Reply = ipc.StatusCommandReply{
		Upstream: upstream,
		Provider: provider,
	}

	return nil
}
