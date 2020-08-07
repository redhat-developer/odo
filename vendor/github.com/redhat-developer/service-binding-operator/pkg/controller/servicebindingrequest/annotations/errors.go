package annotations

import (
	"errors"
	"fmt"
)

type InvalidArgumentErr string

func (e InvalidArgumentErr) Error() string {
	return fmt.Sprintf("invalid argument value for path %q", string(e))
}

var ResourceNameFieldNotFoundErr = errors.New("secret name field not found")

type UnknownBindingTypeErr string

func (e UnknownBindingTypeErr) Error() string {
	return string(e) + " is not supported"
}

type ErrInvalidBindingValue string

func (e ErrInvalidBindingValue) Error() string {
	return fmt.Sprintf("invalid binding value %q", string(e))
}
