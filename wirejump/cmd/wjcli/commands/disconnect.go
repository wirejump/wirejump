package commands

import (
	"flag"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type DisconnectCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand
}

func NewDisconnectCommand() *DisconnectCommand {
	fs, opts := cli.CreateCommand("disconnect", "Disconnect upstream", []string{}, []string{})

	return &DisconnectCommand{fs, opts}
}

func (c *DisconnectCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *DisconnectCommand) Run() error {
	params := ipc.ConnectCommandRequest{}
	version := ipc.ConnectCommandReply{}

	params.Disconnect = true

	return cli.ExecuteCommand(c.opts, "Connect", params, &version)
}
