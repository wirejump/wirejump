package commands

import (
	"flag"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type StatusCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand
}

func NewStatusCommand() *StatusCommand {
	fs, opts := cli.CreateCommand("status", "Get current connection status", []string{}, []string{})

	return &StatusCommand{fs, opts}
}

func (c *StatusCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *StatusCommand) Run() error {
	params := ipc.StatusCommandRequest{}
	reply := ipc.StatusCommandReply{}

	return cli.ExecuteCommand(c.opts, "Status", params, &reply)
}
