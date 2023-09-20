package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mitchellh/go-ps"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/testingutil/system"
)

type State struct {
	content Content
	fs      filesystem.Filesystem
	system  system.System
}

var _ Client = (*State)(nil)

func NewStateClient(fs filesystem.Filesystem, system system.System) *State {
	return &State{
		fs:     fs,
		system: system,
	}
}

func (o *State) Init(ctx context.Context) error {
	var (
		pid      = odocontext.GetPID(ctx)
		platform = fcontext.GetPlatform(ctx, commonflags.PlatformCluster)
	)
	o.content.PID = pid
	o.content.Platform = platform
	return o.save(ctx, pid)

}

func (o *State) SetForwardedPorts(ctx context.Context, fwPorts []api.ForwardedPort) error {
	var (
		pid      = odocontext.GetPID(ctx)
		platform = fcontext.GetPlatform(ctx, commonflags.PlatformCluster)
	)
	// TODO(feloy) When other data is persisted into the state file, it will be needed to read the file first
	o.content.ForwardedPorts = fwPorts
	o.content.PID = pid
	o.content.Platform = platform
	return o.save(ctx, pid)
}

func (o *State) GetForwardedPorts(ctx context.Context) ([]api.ForwardedPort, error) {
	var (
		result    []api.ForwardedPort
		platforms []string
		platform  = fcontext.GetPlatform(ctx, "")
	)
	if platform == "" {
		platforms = []string{commonflags.PlatformCluster, commonflags.PlatformPodman}
	} else {
		platforms = []string{platform}
	}

	for _, platform = range platforms {
		content, err := o.read(platform)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // if the state file does not exist, no ports are forwarded
			}
			return nil, err
		}
		result = append(result, content.ForwardedPorts...)
	}
	return result, nil
}

func (o *State) SaveExit(ctx context.Context) error {
	var (
		pid = odocontext.GetPID(ctx)
	)
	o.content.ForwardedPorts = nil
	o.content.PID = 0
	o.content.Platform = ""
	o.content.APIServerPort = 0
	err := o.delete(pid)
	if err != nil {
		return err
	}
	return o.saveCommonIfOwner(pid)
}

func (o *State) SetAPIServerPort(ctx context.Context, port int) error {
	var (
		pid      = odocontext.GetPID(ctx)
		platform = fcontext.GetPlatform(ctx, commonflags.PlatformCluster)
	)

	o.content.APIServerPort = port
	o.content.Platform = platform
	return o.save(ctx, pid)
}

func (o *State) GetAPIServerPorts(ctx context.Context) ([]api.DevControlPlane, error) {
	var (
		result    []api.DevControlPlane
		platforms []string
		platform  = fcontext.GetPlatform(ctx, "")
	)
	if platform == "" {
		platforms = []string{commonflags.PlatformCluster, commonflags.PlatformPodman}
	} else {
		platforms = []string{platform}
	}

	for _, platform = range platforms {
		content, err := o.read(platform)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // if the state file does not exist, no API Servers are listening
			}
			return nil, err
		}
		if content.APIServerPort == 0 {
			continue
		}
		controlPlane := api.DevControlPlane{
			Platform:      platform,
			LocalPort:     content.APIServerPort,
			APIServerPath: "/api/v1/",
		}
		if feature.IsEnabled(ctx, feature.UIServer) {
			controlPlane.WebInterfacePath = "/"
		}
		result = append(result, controlPlane)
	}
	return result, nil
}

// save writes the content structure in json format in file
func (o *State) save(ctx context.Context, pid int) error {

	err := o.checkFirstInPlatform(ctx)
	if err != nil {
		return err
	}

	err = o.saveCommonIfOwner(pid)
	if err != nil {
		return err
	}

	return o.writeStateFile(getFilename(pid))
}

func (o *State) writeStateFile(path string) error {
	jsonContent, err := json.MarshalIndent(o.content, "", " ")
	if err != nil {
		return err
	}
	// .odo directory is supposed to exist, don't create it
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}
	return o.fs.WriteFile(path, jsonContent, 0644)
}

// read returns the content of the devstate.${PID}.json file for the given platform
func (o *State) read(platform string) (Content, error) {

	var content Content

	// We could use Glob, but it is not implemented by the Filesystem abstraction
	entries, err := o.fs.ReadDir(_dirpath)
	if err != nil {
		return Content{}, nil
	}
	re := regexp.MustCompile(`^devstate\.[0-9]*\.json$`)
	for _, entry := range entries {
		if !re.MatchString(entry.Name()) {
			continue
		}
		jsonContent, err := o.fs.ReadFile(filepath.Join(_dirpath, entry.Name()))
		if err != nil {
			return Content{}, err
		}
		// Ignore error, to handle empty file
		_ = json.Unmarshal(jsonContent, &content)
		if content.Platform == platform {
			break
		} else {
			content = Content{}
		}
	}
	if content.Platform == "" {
		return Content{}, fs.ErrNotExist
	}
	return content, nil
}

func (o *State) delete(pid int) error {
	return o.fs.Remove(getFilename(pid))
}

func getFilename(pid int) string {
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

	return o.writeStateFile(_filepath)
}

func (o *State) isFreeOrOwnedBy(pid int) (bool, error) {
	jsonContent, err := o.fs.ReadFile(_filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File not found, it is free
			return true, nil
		}
		return false, err
	}
	var savedContent Content
	// Ignore error, to handle empty file
	_ = json.Unmarshal(jsonContent, &savedContent)
	if savedContent.PID == 0 {
		// PID is 0 in file, it is free
		return true, nil
	}
	if savedContent.PID == pid {
		// File is owned by process
		return true, nil
	}

	exists, err := o.system.PidExists(savedContent.PID)
	if err != nil {
		return false, err
	}
	if !exists {
		// Process already finished
		return true, nil
	}

	return false, nil
}

func (o *State) checkFirstInPlatform(ctx context.Context) error {
	var (
		pid      = odocontext.GetPID(ctx)
		platform = fcontext.GetPlatform(ctx, "cluster")
	)

	re := regexp.MustCompile(`^devstate\.[0-9]*\.json$`)
	entries, err := o.fs.ReadDir(_dirpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// No file found => no problem
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !re.MatchString(entry.Name()) {
			continue
		}
		jsonContent, err := o.fs.ReadFile(filepath.Join(_dirpath, entry.Name()))
		if err != nil {
			return err
		}
		var content Content
		// Ignore error, to handle empty file
		_ = json.Unmarshal(jsonContent, &content)

		if content.Platform != platform {
			continue
		}

		if content.PID == pid {
			continue
		}
		exists, err := o.system.PidExists(content.PID)
		if err != nil {
			return err
		}
		if exists {
			var process ps.Process
			process, err = o.system.FindProcess(content.PID)
			if err != nil {
				klog.V(4).Infof("process %d exists but is not accessible, ignoring", content.PID)
				continue
			}
			if process.Executable() != "odo" && process.Executable() != "odo.exe" {
				klog.V(4).Infof("process %d exists but is not odo, ignoring", content.PID)
				continue
			}
			// Process exists => problem
			return NewErrAlreadyRunningOnPlatform(platform, content.PID)
		}
	}
	return nil
}

func (o *State) GetOrphanFiles(ctx context.Context) ([]string, error) {
	var (
		pid    = odocontext.GetPID(ctx)
		result []string
	)

	re := regexp.MustCompile(`^devstate\.?[0-9]*\.json$`)
	entries, err := o.fs.ReadDir(_dirpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// No file found => no orphan files
			return nil, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if !re.MatchString(entry.Name()) {
			continue
		}
		filename, err := getFullFilename(entry)
		if err != nil {
			return nil, err
		}

		jsonContent, err := o.fs.ReadFile(filepath.Join(_dirpath, entry.Name()))
		if err != nil {
			return nil, err
		}
		var content Content
		// Ignore error, to handle empty file
		_ = json.Unmarshal(jsonContent, &content)

		if content.PID == pid {
			continue
		}
		if content.PID == 0 {
			// This is devstate.json with pid=0
			continue
		}
		exists, err := o.system.PidExists(content.PID)
		if err != nil {
			return nil, err
		}
		if exists {
			var process ps.Process
			process, err = o.system.FindProcess(content.PID)
			if err != nil {
				klog.V(4).Infof("process %d exists but is not accessible => orphan", content.PID)
				result = append(result, filename)
				continue
			}
			if process == nil {
				klog.V(4).Infof("process %d does not exist => orphan", content.PID)
				result = append(result, filename)
				continue
			}
			if process.Executable() != "odo" && process.Executable() != "odo.exe" {
				klog.V(4).Infof("process %d exists but is not odo => orphan", content.PID)
				result = append(result, filename)
				continue
			}
			// Process exists => not orphan
			klog.V(4).Infof("process %d exists and is odo => not orphan", content.PID)
			continue
		}
		klog.V(4).Infof("process %d does not exist => orphan", content.PID)
		result = append(result, filename)
	}
	return result, nil
}

func getFullFilename(entry fs.FileInfo) (string, error) {
	return filepath.Abs(filepath.Join(_dirpath, entry.Name()))
}
