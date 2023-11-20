package ipc

// Socket file location to be used for IPC between CLI and server.
// Default location is /var/run/wirejump/wirejumpd.socket
const SocketFile = "/var/run/wirejumpd/wirejumpd.sock"

// IpcWrappedHandler allows server to register RPC entrypoint
type IpcWrappedHandler int

// IpcHandler will be used by actual (unwrapped) RPC methods
type IpcHandler struct{}

// IpcCommand is a command request struct sent by client wrapper
type IpcCommand struct {
	Function   string
	ParamsJSON []byte
}

// IpcReply is a command reply struct sent by server
type IpcReply struct {
	Empty     bool
	ReplyJSON []byte
}

// IpcGeneric is a wraper type to better document
// use of pointers across the code. It's used by ipc.Exec,
// ipc.Handlers and cli.Exec functions to name a few.
// type IpcGeneric interface{}

// WirejumpAPI describes all functions server must implement
type WirejumpAPI interface {
	Version() error
	Status() error
	ListProviders() error
	SetupProvider() error
	ListServers() error
	Reset() error
	ResetKeys() error
	Connect() error
}
