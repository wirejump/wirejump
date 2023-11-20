package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"wirejump/internal/ipc"
)

var InteractiveModeBanner = "Interactive mode enabled"

func CheckForInteractive(cmd Runner) bool {
	fs, opts := cmd.Info()

	if IsInteractive(opts) {
		fmt.Println(InteractiveModeBanner)

		return true
	}

	if IncompleteFlags(fs) {
		fmt.Println("Incomplete options provided, forcing interactive mode")

		return true
	}

	return false
}

// Fall back to prompt if defaultValue is empty
func GetInputParam(promptValue string, defaultValue string) string {
	if defaultValue == "" {
		scanner := bufio.NewScanner(os.Stdin)

		fmt.Print(promptValue)
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		return strings.Trim(scanner.Text(), " ")
	} else {
		return defaultValue
	}
}

// Input confirmation function
func GetYesOrNo(promptValue string) bool {
	var value string

	fmt.Println(promptValue)

	for {
		n, err := fmt.Scanln(&value)

		if n == 0 || err != nil {
			fmt.Println("Please enter 'yes' or 'no'")
		}

		if strings.ToLower(value) == "yes" {
			return true
		}
	}
}

// Execute remote command with given params and print the result
// Return error on failure or when decoding has failed
func ExecuteCommand(cmd *BasicCommand, name string, params IpcCommand, reply interface{}) error {
	var empty bool
	var ferror error
	var buffer interface{}

	failed := false
	output := new(bytes.Buffer)
	handle := ipc.GetRPCClient()

	if cmd == nil {
		return &ProgramError{Err: errors.New("command is nil")}
	}

	if params == nil {
		return &ProgramError{Err: errors.New("params are nil")}
	}

	if reply == nil {
		return &ProgramError{Err: errors.New("reply is nil")}
	}

	if handle == nil {
		return &ProgramError{Err: errors.New("IPC is not initialized")}
	}

	// Execute command and get result
	err := ipc.RemoteExec(handle, name, params, reply, &empty)

	// Use different buffers for JSON output in case of error
	if err != nil {
		failed = true
		buffer = err.Error()
	} else {
		// some formatting middleware here

		// No error and no reply, make default message
		// Note: original pointer is set by cli command
		// and points to an instance of command-specific
		// reply struct, which's json.Unmarshall'ed into
		// ipc.RemoteExec (if output is not empty).
		// Thus following code will change this pointer to
		// the static string.
		if empty {
			success := "Command executed successfully"
			reply = &success
		}

		buffer = reply
	}

	if IsJSON(cmd) {
		ferror = JSONFormatter(output, failed, buffer)
	} else {
		ferror = PrettyFormatter(output, reply)
	}

	// Output encoding error means something is really bad,
	// So print it directly; JSON output would be broken anyway
	if ferror != nil {
		fmt.Fprintln(os.Stderr, ferror)
	}

	// In case command has failed AND JSON is requested,
	// Format error nicely
	if err != nil {
		wrapped := &CommandError{Err: err}

		if IsJSON(cmd) {
			wrapped.Err = fmt.Errorf(output.String())
		}

		return wrapped
	}

	// No errors, print output to stdout
	fmt.Print(output)

	return nil
}
