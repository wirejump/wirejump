package commands

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type ResetCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand

	AreYouSure bool
}

var resetCommandHelp = []string{
	"This command will disconnect an existing upstream connection if it's active and",
	"will reset current provider (if it has been initialized with 'setup') command.",
	"It will also remove public key used by upstream interface from an account of that",
	"provider and will reset upstream interface.",
	"This command requires explicit confirmation.\n",
}

var resetCommandUsage = []string{
	"  -f, --force\tDon't ask for confirmation",
}

func NewResetCommand() *ResetCommand {
	fs, opts := cli.CreateCommand("reset", "Reset upstream state", resetCommandHelp, resetCommandUsage)
	cmd := ResetCommand{
		fs:   fs,
		opts: opts,
	}

	fs.BoolVar(&cmd.AreYouSure, "f", false, "force")
	fs.BoolVar(&cmd.AreYouSure, "force", false, "force")

	return &cmd
}

func (c *ResetCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *ResetCommand) Run() error {
	params := ipc.ResetCommandRequest{}
	version := ipc.ResetCommandReply{}

	// Don't enable interactive by default
	if cli.IsInteractive(c.opts) {
		fmt.Println(cli.InteractiveModeBanner)

		maybeYes := cli.GetInputParam("Are you sure (yes/no): ", "")

		if strings.ToLower(maybeYes) == "yes" {
			c.AreYouSure = true
		}
	}

	if c.AreYouSure {
		return cli.ExecuteCommand(c.opts, "Reset", params, &version)
	} else {
		return errors.New("no confirmation provided, command aborted")
	}
}
