package handlers

import (
	"errors"
	"fmt"
	"wirejump/internal/ipc"
	"wirejump/internal/network"
	"wirejump/internal/providers"
	"wirejump/internal/state"
)

// Select a particular provider. Will reset existing provider and its connection if present
func (h *IpcHandler) SetupProvider(State *state.AppState, Params *ipc.SetupCommandRequest, Reply *interface{}) error {
	if len(Params.Provider) == 0 {
		return errors.New("provider name cannot be empty")
	}

	// Upstream can be unitialized (setup after reset), so create it if needed
	if State.Network.Upstream == nil {
		State.Network.Upstream = &network.InterfaceConfig{
			Kind: network.InterfaceKindUpstream,
			Name: "",
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
	}

	// Redirect to reset
	if State.UpstreamProvider != nil {
		return fmt.Errorf("provider '%s' is already selected, run 'reset' first", State.UpstreamProvider.Provider.ProviderName)
	}

	// Select new provider from the list and validate it
	if initializer, exists := State.AvailableProviders.Available[Params.Provider]; !exists {
		return fmt.Errorf("provider '%s' does not exist", Params.Provider)
	} else {
		if Params.Provider != "mullvad" && len(Params.Password) == 0 {
			return errors.New("this provider requires a password")
		}

		provider, err := initializer(providers.WireguardProviderAccount{
			AccountID: Params.Username,
			Password:  Params.Password,
		})

		if err != nil {
			return fmt.Errorf("failed to initialize provider: %s", err)
		} else {
			// Try the account right away to ensure its validity
			if acc, err := provider.GetAccountInfo(); err != nil {
				return fmt.Errorf("failed to verify provider account: %s", err)
			} else {
				// Check upstream gateway...
				if !network.IsValidIP(provider.UpstreamGateway) {
					return errors.New("upstream gateway IP address is invalid")
				}

				// ...and write it down
				if err := State.Network.Upstream.UpdateDefaultGateway(provider.UpstreamGateway); err != nil {
					return fmt.Errorf("failed to update upstream gateway: %s", err)
				}

				// Update account validity
				provider.ValidUntil = acc.Expires

				// Finally, update upstream state
				State.UpstreamProvider = &providers.ProviderState{
					Provider: &provider,
				}
			}
		}
	}

	*Reply = nil

	return nil
}
