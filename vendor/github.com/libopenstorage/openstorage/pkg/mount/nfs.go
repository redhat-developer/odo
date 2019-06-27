// +build linux

package mount

import (
	"regexp"

	"github.com/docker/docker/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/keylock"
)

// NFSMounter implements Manager and keeps track of active mounts for volume drivers.
type NFSMounter struct {
	servers []string
	Mounter
}

// NewNFSMounter instance
func NewNFSMounter(servers []string, mountImpl MountImpl, allowedDirs []string) (Manager, error) {
	m := &NFSMounter{
		servers: servers,
		Mounter: Mounter{
			mountImpl:   mountImpl,
			mounts:      make(DeviceMap),
			paths:       make(PathMap),
			allowedDirs: allowedDirs,
			kl:          keylock.New(),
		},
	}
	err := m.Load([]string{""})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Reload reloads the mount table for the specified device
func (m *NFSMounter) Reload(device string) error {
	return ErrUnsupported
}

//utility function to test if a server is part of driver config
func (m *NFSMounter) serverExists(server string) bool {
	for _, v := range m.servers {
		if v == server {
			return true
		}
	}
	return false
}

// Load mount table
func (m *NFSMounter) Load(source []string) error {
	info, err := mount.GetMounts()
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`,addr=(.*)`)
MountLoop:
	for _, v := range info {
		if len(m.servers) != 0 {
			if v.Fstype != "nfs" {
				continue
			}
			matches := re.FindStringSubmatch(v.VfsOpts)
			if len(matches) != 2 {
				continue
			}

			if exists := m.serverExists(matches[1]); !exists {
				continue
			}
		}
		mount, ok := m.mounts[v.Source]
		if !ok {
			mount = &Info{
				Device:     v.Source,
				Fs:         v.Fstype,
				Minor:      v.Minor,
				Mountpoint: make([]*PathInfo, 0),
			}
			m.mounts[v.Source] = mount
		}
		// Allow Load to be called multiple times.
		for _, p := range mount.Mountpoint {
			if p.Path == v.Mountpoint {
				continue MountLoop
			}
		}
		mount.Mountpoint = append(mount.Mountpoint,
			&PathInfo{
				Path: normalizeMountPath(v.Mountpoint),
			},
		)
	}
	return nil
}
