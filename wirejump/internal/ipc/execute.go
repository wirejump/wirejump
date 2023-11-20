package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/rpc"
	"reflect"
	"wirejump/internal/state"
)

// RPC client handle
var rpcClient *rpc.Client

// Used to check if server is already busy
var isLocked = make(chan bool, 1)

func init() {
	isLocked <- false
}

func GetRPCClient() *rpc.Client {
	return rpcClient
}

func SetRPCClient(c *rpc.Client) {
	rpcClient = c
}

// RemoteExec will execute RPC function on remote end with given params.
// Since it calls a single RPC entrypoint function and forwards its results,
// it's sole purpose is to ensure custom data (wrapped params & reply) can
// be coded into structs and shuffled around by the means of net/rpc, which
// will treat them as strings and will successfully encode other params.
// This function is supposed to be executed by IPC client.
func RemoteExec(client *rpc.Client, name string, params interface{}, reply interface{}, empty *bool) error {
	// Encode provided params
	as_json, err := json.Marshal(params)

	if err != nil {
		return fmt.Errorf("ipc.RemoteExec: failed to encode params: %s", err)
	}

	// Create request & reply
	req := IpcCommand{Function: name, ParamsJSON: as_json}
	rep := IpcReply{}

	// Execute command
	// TODO: use reflect to get this name at runtime
	err = client.Call("IpcWrappedHandler.ExecuteRPC", req, &rep)

	if err != nil {
		// This should be used in debug log
		// return fmt.Errorf("failed to exec '%s': %s", name, err)
		return err
	}

	// Encode reply if needed
	if !rep.Empty {
		err = json.Unmarshal(rep.ReplyJSON, reply)

		if err != nil {
			return fmt.Errorf("ipc.RemoteExec: failed to decode reply: %s", err)
		}
	}

	*empty = rep.Empty

	return nil
}

// LocalExec will call desired RPC method applying all necessary
// middleware-related duties. This function is supposed to be executed
// by IPC server. Essentially it uses the same mechanism as net/rpc,
// requiring RPC methods to be exported for handler type.
func LocalExec(handler interface{}, request IpcCommand, reply *IpcReply) error {
	// Some input checks
	if handler == nil {
		return errors.New("ipc.LocalExec: handler is nil")
	}
	if request.Function == "" {
		return errors.New("ipc.LocalExec: function name is required")
	}
	if reply == nil {
		return errors.New("ipc.LocalExec: reply is nil")
	}

	// Force single user mode: terminate any other operation
	// with an error, if another handler is already running
	select {
	case <-isLocked:
	default:
		return errors.New("server is currently locked by another operation. Please try again later")
	}

	defer func() {
		isLocked <- false
	}()

	// Lookup desired function
	name := request.Function
	method := reflect.ValueOf(handler).MethodByName(name)

	if reflect.Value.IsValid(method) {
		// Create matching param receiver struct
		decodedParams := GuessParamsType(name)

		if decodedParams == nil {
			return fmt.Errorf("ipc.LocalExec: GuessParamType is missing params for %s", name)
		}

		// Will hold output
		var commandResult interface{}

		// Decode params
		if err := json.Unmarshal(request.ParamsJSON, decodedParams); err != nil {
			return fmt.Errorf("ipc.LocalExec: failed to decode params: %s", err)
		}

		// Place for pre-middleware

		// Get app state and lock on it
		appState := state.GetStateInstance()

		appState.Mutex.Lock()
		defer appState.Mutex.Unlock()

		// Create params
		args := []reflect.Value{
			reflect.ValueOf(&appState.State),
			reflect.ValueOf(decodedParams),
			reflect.ValueOf(&commandResult),
		}

		// Call
		result := method.Call(args)[0].Interface()

		// Place for post-middleware here

		// Create reply
		output := IpcReply{}

		// Check whether result encoding is required;
		// if result is nil, struct is not required
		if commandResult == nil {
			output.Empty = true
		} else {
			msg, err := json.Marshal(commandResult)

			if err != nil {
				return fmt.Errorf("ipc.LocalExec: failed to encode result: %s", err)
			} else {
				output.ReplyJSON = msg
			}
		}

		// Set result
		*reply = output

		// Cast output error if needed
		if reflect.TypeOf(result) == nil {
			return nil
		} else {
			return result.(error)
		}
	} else {
		return fmt.Errorf("ipc.LocalExec: method not found: %s", name)
	}
}
