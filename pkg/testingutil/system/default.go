package system

import "github.com/mitchellh/go-ps"

type Default struct{}

var _ System = Default{}

func (o Default) FindProcess(pid int) (ps.Process, error) {
	return ps.FindProcess(pid)
}

func (o Default) PidExists(pid int) (bool, error) {
	processes, err := ps.Processes()
	if err != nil {
		return false, err
	}
	for _, process := range processes {
		if process.Pid() == pid {
			return true, nil
		}
	}
	return false, nil
}
