package commands

import (
	"flag"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type VersionCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand
}

func NewVersionCommand() *VersionCommand {
	fs, opts := cli.CreateCommand("version", "Get server daemon version", []string{}, []string{})

	return &VersionCommand{fs, opts}
}

func (c *VersionCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *VersionCommand) Run() error {
	params := ipc.VersionCommandRequest{}
	version := ipc.VersionCommandReply{}

	return cli.ExecuteCommand(c.opts, "Version", params, &version)
}
