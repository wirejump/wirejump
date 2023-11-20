package commands

import (
	"flag"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type ListCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand
}

func NewListCommand() *ListCommand {
	fs, opts := cli.CreateCommand("list", "List available providers", []string{}, []string{})

	return &ListCommand{fs, opts}
}

func (c *ListCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *ListCommand) Run() error {
	params := ipc.ListProvidersRequest{}
	reply := ipc.ListProvidersReply{}

	return cli.ExecuteCommand(c.opts, "ListProviders", params, &reply)
}
