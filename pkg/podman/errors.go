package podman

import (
	"fmt"
)

type PodmanNotFoundError struct {
	err error
}

func NewPodmanNotFoundError(err error) PodmanNotFoundError {
	return PodmanNotFoundError{err: err}
}

func (o PodmanNotFoundError) Error() string {
	msg := "unable to access podman. Do you have podman client installed?"
	if o.err == nil {
		return msg
	}
	return fmt.Errorf("%s Cause: %w", msg, o.err).Error()
}
