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
	return fmt.Errorf("unable to access podman. Do you have podman client installed? Cause: %w", o.err).Error()
}
