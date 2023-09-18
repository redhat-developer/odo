package system

import "github.com/mitchellh/go-ps"

type Default struct{}

var _ System = Default{}

func (o Default) FindProcess(pid int) (ps.Process, error) {
	return ps.FindProcess(pid)
}
