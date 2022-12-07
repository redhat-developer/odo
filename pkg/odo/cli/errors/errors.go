package errors

import (
	"errors"
	"fmt"
)

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

type Warning struct {
	msg string
	err error
}

func NewWarning(msg string, err error) Warning {
	return Warning{
		msg: msg,
		err: err,
	}
}

func (o Warning) Error() string {
	return fmt.Errorf("%s: %w", o.msg, o.err).Error()
}

func IsWarning(err error) bool {
	_, ok := err.(Warning)
	return ok
}

func AsWarning(err error) bool {
	return errors.As(err, &Warning{})
}
