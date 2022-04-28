package state

import (
	"encoding/json"
	"os"
	"time"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type State struct {
	content             Content
	fs                  filesystem.Filesystem
	getSecondsFromEpoch func() int64
	getpid              func() int
}

func NewStateClient(fs filesystem.Filesystem) *State {
	return &State{
		fs:                  fs,
		getSecondsFromEpoch: getSecondsFromEpoch,
		getpid:              os.Getpid,
	}
}

func (o *State) SetForwardedPorts(fwPorts []api.ForwardedPort) error {
	// TODO(feloy) When other data is persisted into the state file, it will be needed to read the file first
	o.content.ForwardedPorts = fwPorts
	o.setMetadata()
	return o.save()
}

func (o *State) SaveExit() error {
	o.content.ForwardedPorts = nil
	o.content.PID = 0
	o.content.Timestamp = o.getSecondsFromEpoch()
	return o.save()
}

// setMetadata sets the metadata in the state with current PID and epoch
func (o *State) setMetadata() {
	o.content.PID = o.getpid()
	o.content.Timestamp = o.getSecondsFromEpoch()
}

// save writes the content structure in json format in file
func (o *State) save() error {
	jsonContent, err := json.MarshalIndent(o.content, "", " ")
	if err != nil {
		return err
	}
	// .odo directory is supposed to exist, don't create it
	return o.fs.WriteFile(_filepath, jsonContent, 0644)
}

func getSecondsFromEpoch() int64 {
	return time.Now().Unix()
}
