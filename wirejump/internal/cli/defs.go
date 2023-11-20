package cli

import "flag"

type Runner interface {
	Info() (*flag.FlagSet, *BasicCommand)
	Run() error
}

// Pointer to all registered commands, should be populated by init()
var allCommands []Runner

// Program (wishful exe) name, shared by all commands. To be set by init()
var programName string

// Program description, shared by all commands. To be set by init()
var programDesc string

// Program usage, shared by all commands. To be set by init()
var programUsage []string

// Will hold command name & desc
type CommandInfo struct {
	name string
	desc string
}

type DefaultOpts struct {
	use_json        bool
	use_interactive bool
}

// Will be used as a base command for subcommands
type BasicCommand struct {
	CommandInfo
	DefaultOpts
}

// Used to filter out custom flags
var defaultFlagValues = []string{"i", "interactive", "j", "json", "help"}

// Default usage string for a command, outlining params from BasicCommand
var defaultCommandUsage = []string{
	"  -i, --interactive\tUse interactive mode for entering data\t",
	"  -j, --json\tUse JSON output\t",
	"  -h, --help\tDisplay this help\t",
}

// Default usage string for the program
var defaultProgramUsage = []string{
	"  -h, --help\tDisplay this help\t",
	"  -v, --version\tDisplay program version\t",
}

type IpcCommand interface{}
type IpcReply interface{}
