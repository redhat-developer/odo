package system

import (
	"errors"

	"github.com/mitchellh/go-ps"
)

type Fake struct {
	ProcessId int
	ParentId  int
	// PidTable is a map of pid => executable name of existing processes
	PidTable map[int]string
}

func (o Fake) Pid() int {
	return o.ProcessId
}

func (o Fake) PPid() int {
	return o.ParentId
}

func (o Fake) Executable() string {
	return o.PidTable[o.ProcessId]
}

var _ System = Fake{}

func (o Fake) FindProcess(pid int) (ps.Process, error) {
	if _, found := o.PidTable[pid]; found {
		o.ProcessId = pid
		return o, nil
	}
	return nil, errors.New("no process found")
}
