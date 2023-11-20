package ipc

// Every command consists of Request-Reply pair. Error message
// is encoded separately, as well as default status message, thus
// Reply parts only specify information which should be printed
// separately

// ConnectionStatus represents current connection status
type ConnectionStatus struct {
	Online      bool    `json:"online"`
	ActiveSince *int64  `json:"active_since" pretty:"Active since" timefield:""`
	Country     *string `json:"country"`
	City        *string `json:"city"`
}

// AccountStatus contains some account information
type ProviderStatus struct {
	Name              *string `json:"name"`
	PreferredLocation *string `json:"preferred" pretty:"Preferred location"`
	AccountExpires    *int64  `json:"expires" pretty:"Account expires" timefield:""`
}

// Command with no params
type EmptyCommandRequest struct {
	Empty int
}

// Command with no reply
type EmptyCommandReply struct {
	Empty int
}

// Version command
type VersionCommandRequest EmptyCommandRequest

// Version reply
type VersionCommandReply struct {
	Version string `json:"version"`
}

// Status command
type StatusCommandRequest EmptyCommandRequest

// Status reply
type StatusCommandReply struct {
	Upstream ConnectionStatus `json:"upstream" pretty:"Upstream connection"`
	Provider ProviderStatus   `json:"provider"`
}

// List command
type ListProvidersRequest EmptyCommandRequest

// List reply
type ListProvidersReply struct {
	Providers []string `json:"providers"`
}

// Servers command
type ServersCommandRequest struct {
	ForceUpdate bool
	Preferred   string
	Reset       bool
}

// Server reply
type ServersCommandReply struct {
	Servers           []string `json:"servers"`
	LastUpdated       *int64   `json:"updated" pretty:"Last updated" timefield:""`
	PreferredLocation *string  `json:"preferred" pretty:"Preferred location"`
}

// Setup command
type SetupCommandRequest struct {
	Provider string `json:"provider"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Setup reply
type SetupCommandReply EmptyCommandReply

// Connect command
type ConnectCommandRequest struct {
	LocationOverride *string
	PreserveKeys     bool
	Disconnect       bool
}

// Connect reply
type ConnectCommandReply EmptyCommandReply

const PeerCommandAddPeer = 1
const PeerCommandDeletePeer = 2

// Peer command
type PeerCommandRequest struct {
	Operation int
	Pubkey    string
	Isolated  bool
}

// Peer reply
type PeerCommandReply struct {
	Peer struct {
		IPv4Address string `json:"ipv4_address" pretty:"IPv4 Address"`
		Isolated    bool   `json:"isolated"`
	} `json:"peer"`
}

// Reset command
type ResetCommandRequest EmptyCommandRequest

// Reset reply
type ResetCommandReply EmptyCommandReply
