package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func ExecuteSubcommand(args []string) error {
	if len(args) < 1 {
		return &ProgramError{Err: errors.New("command name is required")}
	}

	subcommand := os.Args[1]

	for _, cmd := range GetAllCommands() {
		fs, _ := cmd.Info()

		if fs.Name() == subcommand {
			fs.Parse(os.Args[2:])

			return cmd.Run()
		}
	}

	return &ProgramError{Err: fmt.Errorf("unknown command: %s", subcommand)}
}

// Print generic help, outlining all registered commands
func ProgramUsage() {
	cmds := GetAllCommands()
	defCommand := ""

	if len(cmds) > 0 {
		defCommand = " COMMAND"
	}

	basicUsage := []string{
		fmt.Sprintf("Usage:  %s [OPTIONS]%s\n", programName, defCommand),
		fmt.Sprintf("%s\n", programDesc),
		"Options:",
	}

	if len(programUsage) > 0 {
		basicUsage = append(basicUsage, programUsage...)
	}

	basicAndOpts := append(basicUsage, defaultProgramUsage...)
	allUsage := basicAndOpts

	if len(cmds) > 0 {
		commandsUsage := []string{
			"\nCommands:",
		}

		// Iterate all commands
		for _, cmd := range GetAllCommands() {
			_, opts := cmd.Info()

			commandsUsage = append(commandsUsage, fmt.Sprintf("  %s\t%s\t", opts.CommandInfo.name, opts.CommandInfo.desc))
		}

		allUsage = append(allUsage, commandsUsage...)
	}

	tabwriterPrint(allUsage)
}

// Create default program options and their handlers
func MakeDefaultProgramOptions() *flag.FlagSet {
	fs := flag.NewFlagSet(programName, flag.ExitOnError)

	// Add --help, -h is already handled by flag package itself
	fs.BoolFunc("help", "help", defaultUsageHandler)

	// Add -v, --version
	fs.BoolFunc("v", "version", defaultVersionHandler)
	fs.BoolFunc("version", "version", defaultVersionHandler)

	// Add default usage func for -h
	fs.Usage = func() {
		ProgramUsage()
	}

	return fs
}

// Handle default cmdline options
func HandleDefaultProgramOptions() {
	fs := MakeDefaultProgramOptions()

	fs.Parse(os.Args[1:])
}
