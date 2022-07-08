package adapters

import "fmt"

type ErrPortForward struct {
	cause error
}

func NewErrPortForward(cause error) ErrPortForward {
	return ErrPortForward{cause: cause}
}

func (e ErrPortForward) Error() string {
	return fmt.Sprintf("fail starting the port forwarding: %s", e.cause)
}
