package handlers

import "wirejump/internal/ipc"

// Redefine IpcHandler so it becomes local type
type IpcHandler ipc.IpcHandler

// Each handler has a form of:
// func (h *IpcHandler) HandlerName(State *state.AppState, Params *ipc.HandlerRequest, Reply *interface{}) error
//
// and SHOULD set Reply pointer value to corresponding result like so:
// Reply = nil
// if there's nothing to return, and
// *Reply = ipc.HandlerReply{}
// if specific return type is defined.
// In this case, Reply should point to an instance of that type,
// since final JSON will be encoded into this very variable.
//
// Each handler should return an error or nil if no error has occured.
//
