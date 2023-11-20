package commands

import (
	"errors"
	"flag"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

type PeerCommand struct {
	fs   *flag.FlagSet
	opts *cli.BasicCommand

	Add      bool
	Remove   bool
	Pubkey   string
	Isolated bool
}

var peerCommandUsage = []string{
	"      --add\tAdd peer\t",
	"      --remove\tRemove peer\t",
	"      --pubkey\tPeer public key\t",
	"      --isolated\tIsolate this peer from other peers on the network\t",
}

var peerCommandHelp = []string{
	"This command will manage downstream peers. While it's completely possible",
	"to do that manually via editing interface file and then restarting the interface,",
	"this command allows to get a randomly allocated IP for the peer taking already",
	"used addresses into account. Another benefit of using this over manual editing",
	"is that interface configuration will be updated without restart, so other peers",
	"connections will not be affected.\n",
	"Use --add to add the peer, and --remove to remove the peer. Both commands require",
	"a valid public key.\n",
	"By default, all peers are put into one shared network without any restrictions;",
	"this allows them to communicate directly should the need arise. If this behaviour",
	"is undesired, pass --isolated flag.\n",
}

func NewPeerCommand() *PeerCommand {
	fs, opts := cli.CreateCommand("peer", "Manage downstream peers", peerCommandHelp, peerCommandUsage)
	cmd := PeerCommand{
		fs:   fs,
		opts: opts,
	}

	fs.StringVar(&cmd.Pubkey, "pubkey", "", "pubkey")
	fs.BoolVar(&cmd.Add, "add", false, "add")
	fs.BoolVar(&cmd.Remove, "remove", false, "remove")
	fs.BoolVar(&cmd.Isolated, "isolated", false, "isolated")

	return &cmd
}

func (c *PeerCommand) Info() (*flag.FlagSet, *cli.BasicCommand) {
	return c.fs, c.opts
}

func (c *PeerCommand) Run() error {
	req := ipc.PeerCommandRequest{}
	rep := ipc.PeerCommandReply{}

	if !c.Add && !c.Remove {
		return errors.New("either --add or --remove is required")
	}

	if c.Add && c.Remove {
		return errors.New("--add and --remove cannot be used together")
	}

	if c.Add {
		req.Operation = ipc.PeerCommandAddPeer
	}

	if c.Remove {
		req.Operation = ipc.PeerCommandDeletePeer
	}

	// Get pubkey
	if cli.CheckForInteractive(c) {
		req.Pubkey = cli.GetInputParam("Public key : ", "")
	} else {
		req.Pubkey = c.Pubkey
	}

	req.Isolated = c.Isolated

	return cli.ExecuteCommand(c.opts, "ManagePeers", req, &rep)
}
