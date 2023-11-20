package providers

// Server entity available from upstream
type WireguardServer struct {
	Country string
	City    string
	IPv4    string
	Port    int
	Pubkey  string
}

// Upstream account details
type WireguardAccount struct {
	// Account expiration date as UNIX timestamp
	Expires int64
}

// Upstream provider credentials
type WireguardProviderAccount struct {
	AccountID string
	Password  string
}

// Upstream provider info
type WireguardProvider struct {
	URL             func(...interface{}) string
	Account         WireguardProviderAccount
	ValidUntil      int64
	Initialized     bool
	ProviderName    string
	UpstreamGateway string
}

// Upstream provider factory
type WireguardProviderInitializer func(WireguardProviderAccount) (WireguardProvider, error)

// Required methods for a provider
type UpstreamAPI interface {
	// Ultimately API request needs some provider internal data,
	// which are being populated on init, so this goes here as well
	APIRequest(string, string, bool, interface{}, interface{}) error

	// GetAccountInfo returns various account related settings, such as validity time.
	GetAccountInfo() (WireguardAccount, error)

	// GetAllServers returns all available WireGuard servers for this provider/account.
	GetAllServers() ([]WireguardServer, error)

	// AddPubkey adds WireGuard public key to an account. It should ignore already existing keys.
	AddPubkey(string) error

	// RemovePubkey removes WireGuard public key from an account. It should ignore missing keys.
	RemovePubkey(string) error

	// GetAddress will return IPv4 address of upstream interface with a certain public key.
	GetAddress(string) (string, error)
}

// Holds available providers, will be populated on startup
type ProvidersState struct {
	Available map[string]WireguardProviderInitializer
	Names     []string
}

// Current upstream state
type ProviderState struct {
	Provider          *WireguardProvider
	ActiveSince       *int64
	PreferredLocation *string
	Server            *WireguardServer
}

// Current servers availability state
type ServersState struct {
	Available   map[string][]WireguardServer
	Locations   []string
	ProvidedBy  string
	LastRefresh int64
}
