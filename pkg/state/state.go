package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/api"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type State struct {
	content Content
	fs      filesystem.Filesystem
}

var _ Client = (*State)(nil)

func NewStateClient(fs filesystem.Filesystem) *State {
	return &State{
		fs: fs,
	}
}

func (o *State) SetForwardedPorts(ctx context.Context, fwPorts []api.ForwardedPort) error {
	var (
		pid = odocontext.GetPID(ctx)
	)
	// TODO(feloy) When other data is persisted into the state file, it will be needed to read the file first
	o.content.ForwardedPorts = fwPorts
	o.content.PID = pid
	return o.save(pid)
}

func (o *State) GetForwardedPorts(ctx context.Context) ([]api.ForwardedPort, error) {
	var (
		pid = odocontext.GetPID(ctx)
	)
	err := o.read(pid)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil // if the state file does not exist, no ports are forwarded
		}
		return nil, err
	}
	return o.content.ForwardedPorts, err
}

func (o *State) SaveExit(ctx context.Context) error {
	var (
		pid = odocontext.GetPID(ctx)
	)
	o.content.ForwardedPorts = nil
	o.content.PID = 0
	err := o.delete(pid)
	if err != nil {
		return err
	}
	return o.saveCommonIfOwner(pid)
}

// save writes the content structure in json format in file
func (o *State) save(pid int) error {

	err := o.saveCommonIfOwner(pid)
	if err != nil {
		return err
	}

	jsonContent, err := json.MarshalIndent(o.content, "", " ")
	if err != nil {
		return err
	}
	// .odo directory is supposed to exist, don't create it
	dir := filepath.Dir(getFilename(pid))
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}
	return o.fs.WriteFile(getFilename(pid), jsonContent, 0644)
}

func (o *State) read(pid int) error {
	jsonContent, err := o.fs.ReadFile(getFilename(pid))
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonContent, &o.content)
}

func (o *State) delete(pid int) error {
	return o.fs.Remove(getFilename(pid))
}

func getFilename(pid int) string {
	if pid == 0 {
		return _filepath
	}
	return fmt.Sprintf(_filepathPid, pid)
}

func (o *State) saveCommonIfOwner(pid int) error {

	ok, err := o.isFreeOrOwnedBy(pid)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	jsonContent, err := json.MarshalIndent(o.content, "", " ")
	if err != nil {
		return err
	}
	// .odo directory is supposed to exist, don't create it
	dir := filepath.Dir(_filepath)
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}
	return o.fs.WriteFile(_filepath, jsonContent, 0644)
}

func (o *State) isFreeOrOwnedBy(pid int) (bool, error) {
	jsonContent, err := o.fs.ReadFile(_filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		// File not found, it is free
		return false, err
	}
	var savedContent Content
	err = json.Unmarshal(jsonContent, &savedContent)
	if err != nil {
		return false, err
	}
	if savedContent.PID == 0 {
		// PID is 0 in file, it is free
		return true, nil
	}
	if savedContent.PID == pid {
		// File is owned by process
		return true, nil
	}
	return false, nil
}
