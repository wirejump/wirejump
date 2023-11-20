package handlers

import (
	"wirejump/internal/ipc"
	"wirejump/internal/state"
	"wirejump/internal/version"
)

// Get current running server version
func (h *IpcHandler) Version(State *state.AppState, Params *ipc.VersionCommandRequest, Reply *interface{}) error {
	*Reply = ipc.VersionCommandReply{
		Version: version.VersionString(),
	}

	return nil
}
