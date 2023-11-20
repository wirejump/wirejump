package commands

import (
	"flag"
	"fmt"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type ConnectCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand

	LocationOverride string
	PreserveKeys     bool
}

var connectCommandHelp = []string{
	"This command will manage connection to selected provider.",
	"By default, it will attempt to connect using previously set location;",
	"if no location has been set before, a random one will be picked.",
	"When previous connection is already active, this command will reconnect",
	"current provider using same location (if set) but via different upstream server.",
	"It will also rotate WireGuard keys, unless -p/--preserve-keys is specified.",
	"Use 'setup' command to setup a provider and 'servers' command to set default",
	"location preference.\n",
}

var connectCommandUsage = []string{
	"  -l, --location\tLocation to explicitly use this time",
	"  -p, --preserve-keys\tDon't rotate WireGuard keys during reconnect",
}

func NewConnectCommand() *ConnectCommand {
	fs, opts := cli.CreateCommand("connect", "Manage upstream connection", connectCommandHelp, connectCommandUsage)
	cmd := ConnectCommand{
		fs:   fs,
		opts: opts,
	}

	fs.StringVar(&cmd.LocationOverride, "l", "", "location")
	fs.StringVar(&cmd.LocationOverride, "location", "", "location")

	fs.BoolVar(&cmd.PreserveKeys, "p", false, "preserve")
	fs.BoolVar(&cmd.PreserveKeys, "preserve", false, "preserve")

	return &cmd
}

func (c *ConnectCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *ConnectCommand) Run() error {
	params := ipc.ConnectCommandRequest{}
	reply := ipc.ConnectCommandReply{}

	// Ask for location override explicitly if interactive is enabled
	if cli.IsInteractive(c.opts) {
		fmt.Println(cli.InteractiveModeBanner)

		override := cli.GetInputParam("Override location : ", "")
		params.LocationOverride = &override
	} else {
		if c.LocationOverride != "" {
			params.LocationOverride = &c.LocationOverride
		}
	}

	if c.PreserveKeys {
		params.PreserveKeys = true
	}

	err := cli.ExecuteCommand(c.opts, "Connect", params, &reply)

	return err
}
