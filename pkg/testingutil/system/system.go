package system

import "github.com/mitchellh/go-ps"

type System interface {
	FindProcess(pid int) (ps.Process, error)
	PidExists(pid int) (bool, error)
}
