package state

import "fmt"

type ErrAlreadyRunningOnPlatform struct {
	platform string
}

func NewErrAlreadyRunningOnPlatform(platform string) ErrAlreadyRunningOnPlatform {
	return ErrAlreadyRunningOnPlatform{platform: platform}
}

func (e ErrAlreadyRunningOnPlatform) Error() string {
	return fmt.Sprintf("a session is already running on platform %q", e.platform)
}
