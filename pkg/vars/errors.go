package vars

import "fmt"

type ErrBadKey struct {
	msg string
}

func NewErrBadKey(msg string) ErrBadKey {
	return ErrBadKey{msg: msg}
}

func (e ErrBadKey) Error() string {
	return fmt.Sprintf("poorly formatted environment: %s", e.msg)
}
