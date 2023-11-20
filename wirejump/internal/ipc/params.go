package ipc

// Sad & ugly truth: it's impossible to decode json to interface{}
// and then cast it to a desired struct at run time. Of course,
// such problem should not even exist when "proper" RPC is used,
// since every function will just define desired params and they
// will be guarded by framework/encoding layer. Reflection is also
// of not much help here – while it's definitely possible to create
// a custom type from string definition or map, one still can not
// completely flesh it out during runtime – final .Interface() call
// will be broken, because of missing primitive type.
// So here we go: plan and boring conditional parameter factory...

// This function will return a pointer to the matching
// request function params struct (matched by name)
func GuessParamsType(name string) interface{} {
	switch name {
	case "ListProviders":
		return &ListProvidersRequest{}
	case "SetupProvider":
		return &SetupCommandRequest{}
	case "ManageServers":
		return &ServersCommandRequest{}
	case "ManagePeers":
		return &PeerCommandRequest{}
	case "Status":
		return &StatusCommandRequest{}
	case "Connect":
		return &ConnectCommandRequest{}
	case "Reset":
		return &ResetCommandRequest{}
	case "Version":
		return &VersionCommandRequest{}
	default:
		return nil
	}
}
