package handlers

import (
	"errors"
	"fmt"
	"time"
	"wirejump/internal/ipc"
	"wirejump/internal/providers"
	"wirejump/internal/state"
)

// Cache upstream servers for up to this much seconds
const ServersCacheTime = 3600

// Update available upstream servers
func UpdateUpstreamServers(State *state.AppState) error {
	servers, err := State.UpstreamProvider.Provider.GetAllServers()

	if err != nil {
		return err
	} else {
		new_state := providers.ServersState{}
		available := make(map[string][]providers.WireguardServer)

		for _, serv := range servers {
			available[serv.Country] = append(available[serv.Country], serv)
		}

		new_state.Available = available
		new_state.Locations = []string{}

		for country := range available {
			new_state.Locations = append(new_state.Locations, country)
		}

		// Record last refresh timestamp and provider
		new_state.LastRefresh = time.Now().Unix()
		new_state.ProvidedBy = State.UpstreamProvider.Provider.ProviderName

		// Finally, update the state
		State.Servers = &new_state
	}

	return nil
}

// Determines whether current upstream servers cache can be trusted
func UpstreamCacheIsBad(State *state.AppState) bool {
	if State.UpstreamProvider == nil || State.Servers == nil {
		return true
	}

	status := State.UpstreamProvider.Provider.ProviderName != State.Servers.ProvidedBy ||
		(time.Now().Unix()-State.Servers.LastRefresh > ServersCacheTime)

	return status
}

// Checks if location is valid
func IsValidLocation(State *state.AppState, Location string) bool {
	found := false

	if State.Servers == nil {
		return false
	}

	// Iterate all locations
	for _, location := range State.Servers.Locations {
		if location == Location {
			found = true
		}
	}

	return found
}

// Display available server locations from the list of servers for
// this particular provider or set/reset the preferred location.
// Cache the list for up to ServersCacheTime seconds
func (h *IpcHandler) ManageServers(State *state.AppState, Params *ipc.ServersCommandRequest, Reply *interface{}) error {
	if State.UpstreamProvider == nil {
		return errors.New("no provider selected, setup one first")
	} else {
		if !State.UpstreamProvider.Provider.Initialized {
			return errors.New("provider is set but not initialized")
		}

		// Check if location has been reset
		if Params.Reset {
			State.UpstreamProvider.PreferredLocation = nil
			*Reply = nil

			return nil
		}

		// Get list of providers if needed:
		// - an update is enforced
		// - more than ServersCacheTime has passed
		// - provider has been updated in the meantime
		if Params.ForceUpdate || UpstreamCacheIsBad(State) {
			if err := UpdateUpstreamServers(State); err != nil {
				return err
			}
		}

		// This should never happen, but who knows?..
		if State.Servers == nil {
			return errors.New("servers are still not updated")
		}

		// Check if a location has been provided
		if Params.Preferred != "" {
			if !IsValidLocation(State, Params.Preferred) {
				return fmt.Errorf("location '%s' is not found", Params.Preferred)
			}

			State.UpstreamProvider.PreferredLocation = &Params.Preferred
			*Reply = nil

			return nil
		}

		*Reply = ipc.ServersCommandReply{
			Servers:           State.Servers.Locations,
			LastUpdated:       &State.Servers.LastRefresh,
			PreferredLocation: State.UpstreamProvider.PreferredLocation,
		}

		return nil
	}
}
