// +build linux

package mount

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/libopenstorage/openstorage/pkg/chattr"
	"github.com/libopenstorage/openstorage/pkg/keylock"
	"github.com/libopenstorage/openstorage/pkg/options"
	"github.com/libopenstorage/openstorage/pkg/sched"
	"go.pedge.io/dlog"
)

// Manager defines the interface for keep track of volume driver mounts.
type Manager interface {
	// String representation of the mount table
	String() string
	// Reload mount table for specified device.
	Reload(source string) error
	// Load mount table for all devices that match the list of identifiers
	Load(source []string) error
	// Inspect mount table for specified source. ErrEnoent may be returned.
	Inspect(source string) []*PathInfo
	// Mounts returns paths for specified source.
	Mounts(source string) []string
	// HasMounts determines returns the number of mounts for the source.
	HasMounts(source string) int
	// HasTarget determines returns the number of mounts for the target.
	HasTarget(target string) (string, bool)
	// Exists returns true if the device is mounted at specified path.
	// returned if the device does not exists.
	Exists(source, path string) (bool, error)
	// GetSourcePath scans mount for a specified mountPath and returns the
	// sourcePath if found or returnes an ErrEnoent
	GetSourcePath(mountPath string) (string, error)
	// GetSourcePaths returns all source paths from the mount table
	GetSourcePaths() []string
	// Mount device at mountpoint
	Mount(
		minor int,
		device string,
		path string,
		fs string,
		flags uintptr,
		data string,
		timeout int,
		opts map[string]string) error
	// Unmount device at mountpoint and remove from the matrix.
	// ErrEnoent is returned if the device or mountpoint for the device
	// is not found.
	Unmount(source, path string, flags int, timeout int, opts map[string]string) error
	// RemoveMountPath removes the given path
	RemoveMountPath(path string, opts map[string]string) error
	// EmptyTrashDir removes all directories from the mounter trash directory
	EmptyTrashDir() error
}

// MountImpl backend implementation for Mount/Unmount calls
type MountImpl interface {
	Mount(source, target, fstype string, flags uintptr, data string, timeout int) error
	Unmount(target string, flags int, timeout int) error
}

// MountType indicates different mount types supported
type MountType int

const (
	// DeviceMount indicates a device mount type
	DeviceMount MountType = 1 << iota
	// NFSMount indicates a NFS mount point
	NFSMount
	// CustomMount indicates a custom mount type with its
	// own defined way of handling mount table
	CustomMount
)

const mountPathRemoveDelay = 30 * time.Second

var (
	// ErrExist is returned if path is already mounted to a different device.
	ErrExist = errors.New("Mountpath already exists")
	// ErrEnoent is returned for a non existent mount point
	ErrEnoent = errors.New("Mountpath is not mounted")
	// ErrEinval is returned is fields for an entry do no match
	// existing fields
	ErrEinval = errors.New("Invalid arguments for mount entry")
	// ErrUnsupported is returned for an unsupported operation or a mount type.
	ErrUnsupported = errors.New("Not supported")
	// ErrMountpathNotAllowed is returned when the requested mountpath is not
	// a part of the provided allowed mount paths
	ErrMountpathNotAllowed = errors.New("Mountpath is not allowed")
)

// DeviceMap map device name to Info
type DeviceMap map[string]*Info

// PathMap map path name to device
type PathMap map[string]string

// PathInfo is a reference counted path
type PathInfo struct {
	Path string
}

// Info per device
type Info struct {
	sync.Mutex
	Device     string
	Minor      int
	Mountpoint []*PathInfo
	Fs         string
}

// Mounter implements Ops and keeps track of active mounts for volume drivers.
type Mounter struct {
	sync.Mutex
	mountImpl     MountImpl
	mounts        DeviceMap
	paths         PathMap
	allowedDirs   []string
	kl            keylock.KeyLock
	trashLocation string
}

// DefaultMounter defaults to syscall implementation.
type DefaultMounter struct {
}

// Mount default mount implementation is syscall.
func (m *DefaultMounter) Mount(
	source string,
	target string,
	fstype string,
	flags uintptr,
	data string,
	timeout int,
) error {
	return syscall.Mount(source, target, fstype, flags, data)
}

// Unmount default unmount implementation is syscall.
func (m *DefaultMounter) Unmount(target string, flags int, timeout int) error {
	return syscall.Unmount(target, flags)
}

// String representation of Mounter
func (m *Mounter) String() string {
	s := struct {
		mounts        DeviceMap
		paths         PathMap
		allowedDirs   []string
		trashLocation string
	}{
		mounts:        m.mounts,
		paths:         m.paths,
		allowedDirs:   m.allowedDirs,
		trashLocation: m.trashLocation,
	}

	return fmt.Sprintf("%#v", s)
}

// Inspect mount table for device
func (m *Mounter) Inspect(sourcePath string) []*PathInfo {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[sourcePath]
	if !ok {
		return nil
	}
	return v.Mountpoint
}

// Mounts returns  mount table for device
func (m *Mounter) Mounts(sourcePath string) []string {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[sourcePath]
	if !ok {
		return nil
	}

	mounts := make([]string, len(v.Mountpoint))
	for i, v := range v.Mountpoint {
		mounts[i] = v.Path
	}

	return mounts
}

// GetSourcePaths returns all source paths from the mount table
func (m *Mounter) GetSourcePaths() []string {
	m.Lock()
	defer m.Unlock()

	sourcePaths := make([]string, len(m.mounts))
	i := 0
	for path := range m.mounts {
		sourcePaths[i] = path
		i++
	}
	return sourcePaths
}

// HasMounts determines returns the number of mounts for the device.
func (m *Mounter) HasMounts(sourcePath string) int {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[sourcePath]
	if !ok {
		return 0
	}
	return len(v.Mountpoint)
}

// HasTarget returns true/false based on the target provided
func (m *Mounter) HasTarget(targetPath string) (string, bool) {
	m.Lock()
	defer m.Unlock()

	for k, v := range m.mounts {
		for _, p := range v.Mountpoint {
			if p.Path == targetPath {
				return k, true
			}
		}
	}
	return "", false
}

// Exists scans mountpaths for specified device and returns true if path is one of the
// mountpaths. ErrEnoent may be retuned if the device is not found
func (m *Mounter) Exists(sourcePath string, path string) (bool, error) {
	m.Lock()
	defer m.Unlock()

	v, ok := m.mounts[sourcePath]
	if !ok {
		return false, ErrEnoent
	}
	for _, p := range v.Mountpoint {
		if p.Path == path {
			return true, nil
		}
	}
	return false, nil
}

// GetSourcePath scans mount for a specified mountPath and returns the sourcePath
// if found or returnes an ErrEnoent
func (m *Mounter) GetSourcePath(mountPath string) (string, error) {
	m.Lock()
	defer m.Unlock()

	for k, v := range m.mounts {
		for _, p := range v.Mountpoint {
			if p.Path == mountPath {
				return k, nil
			}
		}
	}
	return "", ErrEnoent
}

func normalizeMountPath(mountPath string) string {
	if len(mountPath) > 1 && strings.HasSuffix(mountPath, "/") {
		return mountPath[:len(mountPath)-1]
	}
	return mountPath
}

func (m *Mounter) maybeRemoveDevice(device string) {
	m.Lock()
	defer m.Unlock()
	if info, ok := m.mounts[device]; ok {
		// If the device has no more mountpoints, remove it from the map
		if len(info.Mountpoint) == 0 {
			delete(m.mounts, device)
		}
	}
}

func (m *Mounter) hasPath(path string) (string, bool) {
	m.Lock()
	defer m.Unlock()
	p, ok := m.paths[path]
	return p, ok
}

func (m *Mounter) addPath(path, device string) {
	m.Lock()
	defer m.Unlock()
	m.paths[path] = device
}

func (m *Mounter) deletePath(path string) bool {
	m.Lock()
	defer m.Unlock()
	if _, pathExists := m.paths[path]; pathExists {
		delete(m.paths, path)
		return true
	}
	return false
}

// Mount new mountpoint for specified device.
func (m *Mounter) Mount(
	minor int,
	devPath, path, fs string,
	flags uintptr,
	data string,
	timeout int,
	opts map[string]string,
) error {
	// device gets overwritten if opts specifies fuse mount with
	// options.OptionsDeviceFuseMount.
	device := devPath
	if value, ok := opts[options.OptionsDeviceFuseMount]; ok {
		// fuse mounts show-up with this key as device.
		device = value
	}

	path = normalizeMountPath(path)
	if len(m.allowedDirs) > 0 {
		foundPrefix := false
		for _, allowedDir := range m.allowedDirs {
			if strings.Contains(path, allowedDir) {
				foundPrefix = true
				break
			}
		}
		if !foundPrefix {
			return ErrMountpathNotAllowed
		}
	}
	dev, ok := m.hasPath(path)
	if ok && dev != device {
		dlog.Warnf("cannot mount %q,  device %q is mounted at %q", device, dev, path)
		return ErrExist
	}
	m.Lock()
	info, ok := m.mounts[device]
	if !ok {
		info = &Info{
			Device:     device,
			Mountpoint: make([]*PathInfo, 0),
			Minor:      minor,
			Fs:         fs,
		}
	}
	m.mounts[device] = info
	m.Unlock()
	info.Lock()
	defer info.Unlock()

	// Validate input params
	if fs != info.Fs {
		dlog.Warnf("%s Existing mountpoint has fs %q cannot change to %q",
			device, info.Fs, fs)
		return ErrEinval
	}

	// Try to find the mountpoint. If it already exists, do nothing
	for _, p := range info.Mountpoint {
		if p.Path == path {
			return nil
		}
	}

	h := m.kl.Acquire(path)
	defer m.kl.Release(&h)

	// Record previous state of the path
	pathWasReadOnly := m.isPathSetImmutable(path)
	if err := m.makeMountpathReadOnly(path); err != nil {
		return fmt.Errorf("failed to make %s readonly Err: %v", path, err)
	}

	// The device is not mounted at path, mount it and add to its mountpoints.
	if err := m.mountImpl.Mount(devPath, path, fs, flags, data, timeout); err != nil {
		// Rollback only if was writeable
		if !pathWasReadOnly {
			if e := m.makeMountpathWriteable(path); e != nil {
				return fmt.Errorf("failed to make %v writeable during rollback. Err: %v Mount err: %v",
					path, e, err)
			}
		}

		return err
	}

	info.Mountpoint = append(info.Mountpoint, &PathInfo{Path: path})
	m.addPath(path, device)

	return nil
}

// Unmount device at mountpoint and from the matrix.
// ErrEnoent is returned if the device or mountpoint for the device is not found.
func (m *Mounter) Unmount(
	devPath string,
	path string,
	flags int,
	timeout int,
	opts map[string]string,
) error {
	m.Lock()
	// device gets overwritten if opts specifies fuse mount with
	// options.OptionsDeviceFuseMount.
	device := devPath
	path = normalizeMountPath(path)
	if value, ok := opts[options.OptionsDeviceFuseMount]; ok {
		// fuse mounts show-up with this key as device.
		device = value
	}
	info, ok := m.mounts[device]
	if !ok {
		m.Unlock()
		return ErrEnoent
	}
	m.Unlock()
	info.Lock()
	defer info.Unlock()
	for i, p := range info.Mountpoint {
		if p.Path != path {
			continue
		}
		err := m.mountImpl.Unmount(path, flags, timeout)
		if err != nil {
			return err
		}
		if pathExists := m.deletePath(path); !pathExists {
			dlog.Warnf("Path %q for device %q does not exist in pathMap",
				path, device)
		}
		// Blow away this mountpoint.
		info.Mountpoint[i] = info.Mountpoint[len(info.Mountpoint)-1]
		info.Mountpoint = info.Mountpoint[0 : len(info.Mountpoint)-1]
		m.maybeRemoveDevice(device)
		if options.IsBoolOptionSet(opts, options.OptionsDeleteAfterUnmount) {
			m.RemoveMountPath(path, opts)
		}

		return nil
	}
	dlog.Warnf("Device %q is not mounted at path %q", path, device)
	return nil
}

func (m *Mounter) removeMountPath(path string) error {
	h := m.kl.Acquire(path)
	defer m.kl.Release(&h)

	if devicePath, mounted := m.HasTarget(path); !mounted {
		if err := m.makeMountpathWriteable(path); err != nil {
			dlog.Warnf("Failed to make path: %v writeable. Err: %v", path, err)
			return err
		}
	} else {
		dlog.Infof("Not making %v writeable as %v is mounted on it", path, devicePath)
		return nil
	}

	if _, err := os.Stat(path); err == nil {
		dlog.Infof("Removing mount path directory: %v", path)
		if err = os.Remove(path); err != nil {
			dlog.Warnf("Failed to remove path: %v Err: %v", path, err)
			return err
		}
	}
	return nil
}

// RemoveMountPath makes the path writeable and removes it after a fixed delay
func (m *Mounter) RemoveMountPath(mountPath string, opts map[string]string) error {
	if _, err := os.Stat(mountPath); err == nil {
		if options.IsBoolOptionSet(opts, options.OptionsWaitBeforeDelete) {
			hasher := md5.New()
			hasher.Write([]byte(mountPath))
			symlinkName := hex.EncodeToString(hasher.Sum(nil))
			symlinkPath := path.Join(m.trashLocation, symlinkName)

			if err = os.Symlink(mountPath, symlinkPath); err != nil {
				if !os.IsExist(err) {
					dlog.Errorf("Error creating sym link %s => %s. Err: %v", symlinkPath, mountPath, err)
				}
			}

			if _, err = sched.Instance().Schedule(
				func(sched.Interval) {
					if err = m.removeMountPath(mountPath); err != nil {
						return
					}

					if err = os.Remove(symlinkPath); err != nil {
						return
					}
				},
				sched.Periodic(time.Second),
				time.Now().Add(mountPathRemoveDelay),
				true /* run only once */); err != nil {
				dlog.Errorf("Failed to schedule task to remove path:%v. Err: %v", mountPath, err)
				return err
			}
		} else {
			return m.removeMountPath(mountPath)
		}
	}

	return nil
}

func (m *Mounter) EmptyTrashDir() error {
	files, err := ioutil.ReadDir(m.trashLocation)
	if err != nil {
		dlog.Errorf("failed to read trash dir: %s. Err: %v", m.trashLocation, err)
		return err
	}

	for _, file := range files {
		if _, err := sched.Instance().Schedule(
			func(sched.Interval) {
				e := m.removeSoftlinkAndTarget(path.Join(m.trashLocation, file.Name()))
				if e != nil {
					dlog.Errorf("failed to remove link: %s. Err: %v", path.Join(m.trashLocation, file.Name()), e)
					err = e
					// continue with other directories
				}
			},
			sched.Periodic(time.Second),
			time.Now().Add(mountPathRemoveDelay),
			true /* run only once */); err != nil {
			dlog.Errorf("Failed to cleanup of trash dir. Err: %v", err)
			return err
		}
	}

	return nil
}

func (m *Mounter) removeSoftlinkAndTarget(link string) error {
	if _, err := os.Stat(link); err == nil {
		target, err := os.Readlink(link)
		if err != nil {
			return err
		}

		if err = m.removeMountPath(target); err != nil {
			return err
		}
	}

	if err := os.Remove(link); err != nil {
		return err
	}

	return nil
}

// isPathSetImmutable returns true on error in getting path info or if path
// is immutable .
func (m *Mounter) isPathSetImmutable(mountpath string) bool {
	return chattr.IsImmutable(mountpath)
}

// makeMountpathReadOnly makes given mountpath read-only
func (m *Mounter) makeMountpathReadOnly(mountpath string) error {
	return chattr.AddImmutable(mountpath)
}

// makeMountpathWriteable makes given mountpath writeable
func (m *Mounter) makeMountpathWriteable(mountpath string) error {
	return chattr.RemoveImmutable(mountpath)
}

// New returns a new Mount Manager
func New(
	mounterType MountType,
	mountImpl MountImpl,
	identifiers []string,
	customMounter CustomMounter,
	allowedDirs []string,
	trashLocation string,
) (Manager, error) {

	if mountImpl == nil {
		mountImpl = &DefaultMounter{}
	}

	switch mounterType {
	case DeviceMount:
		return NewDeviceMounter(identifiers, mountImpl, allowedDirs, trashLocation)
	case NFSMount:
		return NewNFSMounter(identifiers, mountImpl, allowedDirs)
	case CustomMount:
		return NewCustomMounter(identifiers, mountImpl, customMounter, allowedDirs)
	}
	return nil, ErrUnsupported
}
