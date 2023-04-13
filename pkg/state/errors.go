package state

import "fmt"

type ErrAlreadyRunningOnPlatform struct {
	platform string
	pid      int
}

func NewErrAlreadyRunningOnPlatform(platform string, pid int) ErrAlreadyRunningOnPlatform {
	return ErrAlreadyRunningOnPlatform{platform: platform}
}

func (e ErrAlreadyRunningOnPlatform) Error() string {
	return fmt.Sprintf("a session with PID %d is already running on platform %q", e.pid, e.platform)
}
