package state

import (
	"sync"
	"wirejump/internal/network"
	"wirejump/internal/providers"
)

var once sync.Once
var stateInstance *ProtectedState

type ConfigurationState struct {
	UpstreamName   string
	DownstreamName string
}

type AppState struct {
	Config             *ConfigurationState
	Network            network.NetworkState
	UpstreamProvider   *providers.ProviderState
	Servers            *providers.ServersState
	AvailableProviders providers.ProvidersState
}

type ProtectedState struct {
	Mutex sync.Mutex
	State AppState
}

// Create initial state if needed
func GetStateInstance() *ProtectedState {
	if stateInstance == nil {
		state := AppState{}
		once.Do(
			func() {
				stateInstance = &ProtectedState{
					State: state,
					Mutex: sync.Mutex{},
				}
			})
	}

	return stateInstance
}
