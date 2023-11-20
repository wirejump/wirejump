package commands

import (
	"flag"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type SetupCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand

	Provider string
	Username string
	Password string
}

var setupCommandUsage = []string{
	"      --provider\tProvider name to use\t",
	"      --user\tAccountID or username for this provider\t",
	"      --password\tAccount password\t",
}

var setupCommandHelp = []string{
	"This command will setup an account of a particular provider, but ",
	"will not bring connection online. Use 'list' command to get a list of all",
	"registered providers. If provider is already set up, this command",
	"will return its name and exit. Use 'reset' command to reset provider",
	"first, and then run 'setup' again. Use 'connect' and 'disconnect' ",
	"commands to manage upstream connection after setup has been finished.\n",
	"NOTE: all option fields are required; in case some of them are not",
	"provided, interactive mode will be enabled automatically.\n",
}

func NewSetupCommand() *SetupCommand {
	fs, opts := cli.CreateCommand("setup", "Setup upstream provider", setupCommandHelp, setupCommandUsage)
	cmd := SetupCommand{
		fs:   fs,
		opts: opts,
	}

	fs.StringVar(&cmd.Provider, "provider", "", "provider")
	fs.StringVar(&cmd.Username, "username", "", "username")
	fs.StringVar(&cmd.Password, "password", "", "password")

	return &cmd
}

func (c *SetupCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *SetupCommand) Run() error {
	req := ipc.SetupCommandRequest{}
	rep := ipc.SetupCommandReply{}

	// Some ugliness
	if cli.CheckForInteractive(c) {
		req.Provider = cli.GetInputParam("Provider : ", c.Provider)
		req.Username = cli.GetInputParam("Username : ", c.Username)
		req.Password = cli.GetInputParam("Password : ", c.Password)
	} else {
		req.Provider = c.Provider
		req.Username = c.Username
		req.Password = c.Password
	}

	return cli.ExecuteCommand(c.opts, "SetupProvider", req, &rep)
}
