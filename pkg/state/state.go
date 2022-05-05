package state

import (
	"encoding/json"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type State struct {
	content Content
	fs      filesystem.Filesystem
}

func NewStateClient(fs filesystem.Filesystem) *State {
	return &State{
		fs: fs,
	}
}

func (o *State) SetForwardedPorts(fwPorts []api.ForwardedPort) error {
	// TODO(feloy) When other data is persisted into the state file, it will be needed to read the file first
	o.content.ForwardedPorts = fwPorts
	return o.save()
}

func (o *State) SaveExit() error {
	o.content.ForwardedPorts = nil
	return o.save()
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
