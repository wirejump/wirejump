package cli

import (
	"flag"
	"fmt"
	"os"
)

// Initialize default parameters
func init() {
	programName = "unknown_program"
	programDesc = "unknown_program description string"
}

func GetAllCommands() []Runner {
	return allCommands
}

// Dumb getter to get json option
func IsJSON(cmd *BasicCommand) bool {
	return cmd.DefaultOpts.use_json
}

// Dumb getter to get interactive option
func IsInteractive(cmd *BasicCommand) bool {
	return cmd.DefaultOpts.use_interactive
}

// Registers command to be available to the program
func RegisterCommands(cmds []Runner) {
	allCommands = cmds
}

func SetProgramInfo(name string, desc string, usage []string) {
	programName = name
	programDesc = desc
	programUsage = usage
}

// Check if any of non-standard flags has been omitted
func IncompleteFlags(fs *flag.FlagSet) bool {
	empty := false

	fs.VisitAll(func(f *flag.Flag) {
		isDefault := false

		for _, def := range defaultFlagValues {
			if f.Name == def {
				isDefault = true
			}
		}

		if !isDefault && f.Value.String() == "" {
			empty = true
		}
	})

	return empty
}

// Create named FlagSet with default flags already populated
func CreateBasicCommand(name string) (*BasicCommand, *flag.FlagSet) {
	// Create command
	cmd := BasicCommand{}

	// Create FlagSet
	fs := flag.NewFlagSet(name, flag.ExitOnError)

	// Add default help option
	// -h is being handled by flag already
	fs.BoolFunc("help", "help", func(s string) error {
		fs.Usage()

		os.Exit(0)

		return nil
	})

	// Add -j, --json flag handling
	fs.BoolVar(&cmd.use_json, "j", false, "json")
	fs.BoolVar(&cmd.use_json, "json", false, "json")

	// Add -i, --interactive flag handling
	fs.BoolVar(&cmd.use_interactive, "i", false, "interactive")
	fs.BoolVar(&cmd.use_interactive, "interactive", false, "interactive")

	// Construct usage
	return &cmd, fs
}

func CreateCommandUsageFunc(name string, desc string, helpString []string, usageString []string) func() {
	basicUsage := []string{
		fmt.Sprintf("Usage:  %s %s [OPTIONS]\n", programName, name),
		fmt.Sprintf("%s\n", desc),
	}

	allUsage := append(basicUsage, helpString...)

	if len(usageString) > 0 {
		allUsage = append(allUsage, "Options:")
		allUsage = append(allUsage, usageString...)
		allUsage = append(allUsage, "")
	}

	allUsage = append(allUsage, "Default options:")
	allUsage = append(allUsage, defaultCommandUsage...)

	return func() {
		tabwriterPrint(allUsage)
	}
}

func CreateCommand(name string, desc string, helpstring []string, usage []string) (*flag.FlagSet, *BasicCommand) {
	opts, fs := CreateBasicCommand(name)

	fs.Usage = CreateCommandUsageFunc(name, desc, helpstring, usage)

	// Copy name & desc to info struct as well
	opts.CommandInfo.name = name
	opts.CommandInfo.desc = desc

	return fs, opts
}
