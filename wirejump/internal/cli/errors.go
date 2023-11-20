package cli

type CommandError struct {
	Err error
}
type ProgramError struct {
	Err error
}

func (e *CommandError) Error() string {
	return e.Err.Error()
}

func (e *ProgramError) Error() string {
	return e.Err.Error()
}
