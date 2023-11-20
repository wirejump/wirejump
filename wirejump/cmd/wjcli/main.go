package main

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"wirejump/cmd/wjcli/commands"
	"wirejump/internal/cli"
	"wirejump/internal/ipc"
)

const programName = "wjcli"
const programDesc = "WireJump server management CLI"

// Assumed workflow:
// peer -> list -> setup -> [servers] -> connect -> [status] -> reset

func init() {
	cli.SetProgramInfo(programName, programDesc, []string{})

	// Register commands
	cli.RegisterCommands([]cli.Runner{
		commands.NewPeerCommand(),
		commands.NewListCommand(),
		commands.NewSetupCommand(),
		commands.NewServersCommand(),
		commands.NewConnectCommand(),
		commands.NewStatusCommand(),
		commands.NewDisconnectCommand(),
		commands.NewResetCommand(),
		commands.NewVersionCommand(),
	})
}

// Setup local RPC connection and return error on failure
func SetupRPC() (*rpc.Client, error) {
	if _, err := os.Stat(ipc.SocketFile); err != nil {
		return nil, &cli.ProgramError{Err: err}
	}

	client, err := rpc.Dial("unix", ipc.SocketFile)

	if err != nil {
		return nil, &cli.ProgramError{Err: err}
	}

	return client, nil
}

func main() {
	var programError *cli.ProgramError

	// If no args are supplied
	if len(os.Args) == 1 {
		cli.ProgramUsage()

		os.Exit(0)
	}

	// Parse flags early, in case it's something like --help or --version
	cli.HandleDefaultProgramOptions()

	client, err := SetupRPC()

	if err != nil {
		fmt.Printf("Failed to start RPC: %s\nIs server running?\n", err)

		os.Exit(1)
	}

	ipc.SetRPCClient(client)

	defer client.Close()

	// Execute a subcommand as is and print to stdout on success
	// If JSON format is requested, error will be outlined there
	if err := cli.ExecuteSubcommand(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)

		// ProgramError requires usage hint
		if errors.As(err, &programError) {
			fmt.Fprintf(os.Stderr, "See '%s --help'.\n", programName)
		}

		os.Exit(1)
	}
}
