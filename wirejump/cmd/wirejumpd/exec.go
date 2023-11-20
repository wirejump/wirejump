package main

import (
	"wirejump/cmd/wirejumpd/handlers"
	"wirejump/internal/ipc"
)

// Redefine remote types so that local function ExecuteRPC will be picked up by net/rpc

type IpcReply ipc.IpcReply
type IpcCommand ipc.IpcCommand
type IpcWrappedHandler ipc.IpcWrappedHandler

// ExecuteRPC is local RPC entrypoint. It wraps exec function from ipc package
// and references a type to be used for handler queries (handlers.IpcHandler)
func (t *IpcWrappedHandler) ExecuteRPC(ipcargs IpcCommand, ipcreply *IpcReply) error {
	// Wrap params
	request := ipc.IpcCommand{
		Function:   ipcargs.Function,
		ParamsJSON: ipcargs.ParamsJSON,
	}

	// Create reply
	reply := ipc.IpcReply{}

	// Execute func
	err := ipc.LocalExec(
		&handlers.IpcHandler{},
		request,
		&reply,
	)

	// Set reply
	ipcreply.Empty = reply.Empty
	ipcreply.ReplyJSON = reply.ReplyJSON

	// Return status
	return err
}
