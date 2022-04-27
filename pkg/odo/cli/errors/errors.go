package errors

import "fmt"

type NoCommandInDevfileError struct {
	command string
}

func NewNoCommandInDevfileError(command string) NoCommandInDevfileError {
	return NoCommandInDevfileError{
		command: command,
	}
}

func (o NoCommandInDevfileError) Error() string {
	return fmt.Sprintf("no command of kind %q found in the devfile", o.command)
}
