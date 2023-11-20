package commands

import (
	"flag"
	"fmt"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type ServersCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand

	ForceUpdate bool
	Preferred   string
	Reset       bool
}

var serversCommandHelp = []string{
	"This command will manage desired server location for a particular provider.",
	"If provider has servers in different countries (that's usually the case), these",
	"countries are treated as separate locations. Exact upstream server is selected ",
	"randomly from all servers for a particular location. If location is not set, it",
	"will be selected randomly. By default, there is no location preference.\n",
	"Server locations are cached in memory for 1 hour. To refresh them immediately, ",
	"pass -f/--force flag to force the update.\n",
}

var serversCommandUsage = []string{
	"  -f, --force\tForce servers update",
	"  -p, --preferred\tSet preferred location",
	"  -r, --reset\tRemove location preference",
}

func NewServersCommand() *ServersCommand {
	fs, opts := cli.CreateCommand("servers", "Manage available server locations", serversCommandHelp, serversCommandUsage)
	cmd := ServersCommand{
		fs:   fs,
		opts: opts,
	}
	fs.BoolVar(&cmd.ForceUpdate, "f", false, "force")
	fs.BoolVar(&cmd.ForceUpdate, "force", false, "force")

	fs.StringVar(&cmd.Preferred, "p", "", "preferred")
	fs.StringVar(&cmd.Preferred, "preferred", "", "preferred")

	fs.BoolVar(&cmd.Reset, "r", false, "reset")
	fs.BoolVar(&cmd.Reset, "reset", false, "reset")

	return &cmd
}

func (c *ServersCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *ServersCommand) Run() error {
	params := ipc.ServersCommandRequest{}
	reply := ipc.ServersCommandReply{}

	params.Reset = c.Reset
	params.Preferred = c.Preferred
	params.ForceUpdate = c.ForceUpdate

	// Ask for location explicitly if interactive is enabled
	if cli.IsInteractive(c.opts) {
		fmt.Println(cli.InteractiveModeBanner)

		params.Preferred = cli.GetInputParam("Preferred location : ", params.Preferred)
	}

	return cli.ExecuteCommand(c.opts, "ManageServers", params, &reply)
}
