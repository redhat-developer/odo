package nfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"syscall"
	"time"

	"go.pedge.io/dlog"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/pkg/mount"
	"github.com/libopenstorage/openstorage/pkg/seed"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/common"
	"github.com/portworx/kvdb"
	"math/rand"
	"strings"
)

const (
	Name         = "nfs"
	Type         = api.DriverType_DRIVER_TYPE_FILE
	NfsDBKey     = "OpenStorageNFSKey"
	nfsMountPath = "/var/lib/openstorage/nfs/"
	nfsBlockFile = ".blockdevice"
)

// Implements the open storage volume interface.
type driver struct {
	volume.IODriver
	volume.StoreEnumerator
	volume.StatsDriver
	volume.QuiesceDriver
	volume.CredsDriver
	volume.CloudBackupDriver
	nfsServers []string
	nfsPath    string
	mounter    mount.Manager
}

func Init(params map[string]string) (volume.VolumeDriver, error) {
	path, ok := params["path"]
	if !ok {
		return nil, errors.New("No NFS path provided")
	}
	server, ok := params["server"]
	if !ok {
		dlog.Printf("No NFS server provided, will attempt to bind mount %s", path)
	} else {
		dlog.Printf("NFS driver initializing with %s:%s ", server, path)
	}

	//support more than one server using CSV
	//TB-FIXME: modify driver params flow to support map[string]struct/array
	servers := strings.Split(server, ",")

	// Create a mount manager for this NFS server. Blank sever is OK.
	mounter, err := mount.New(mount.NFSMount, nil, servers, nil, []string{}, "")
	if err != nil {
		dlog.Warnf("Failed to create mount manager for server: %v (%v)", server, err)
		return nil, err
	}
	inst := &driver{
		IODriver:          volume.IONotSupported,
		StoreEnumerator:   common.NewDefaultStoreEnumerator(Name, kvdb.Instance()),
		StatsDriver:       volume.StatsNotSupported,
		QuiesceDriver:     volume.QuiesceNotSupported,
		nfsServers:        servers,
		CredsDriver:       volume.CredsNotSupported,
		nfsPath:           path,
		mounter:           mounter,
		CloudBackupDriver: volume.CloudBackupNotSupported,
	}

	//make directory for each nfs server
	for _, v := range servers {
		dlog.Infof("Calling mkdirAll: %s", nfsMountPath+v)
		if err := os.MkdirAll(nfsMountPath+v, 0744); err != nil {
			return nil, err
		}
	}
	src := inst.nfsPath
	if server != "" {
		src = ":" + inst.nfsPath
	}

	//mount each nfs server
	for _, v := range inst.nfsServers {
		// If src is already mounted at dest, leave it be.
		mountExists, err := mounter.Exists(src, nfsMountPath+v)
		if !mountExists {
			// Mount the nfs server locally on a unique path.
			syscall.Unmount(nfsMountPath+v, 0)
			if server != "" {
				err = syscall.Mount(
					src,
					nfsMountPath+v,
					"nfs",
					0,
					"nolock,addr="+v,
				)
			} else {
				err = syscall.Mount(src, nfsMountPath+v, "", syscall.MS_BIND, "")
			}
			if err != nil {
				dlog.Printf("Unable to mount %s:%s at %s (%+v)",
					v, inst.nfsPath, nfsMountPath+v, err)
				return nil, err
			}
		}
	}

	volumeInfo, err := inst.StoreEnumerator.Enumerate(&api.VolumeLocator{}, nil)
	if err == nil {
		for _, info := range volumeInfo {
			if info.Status == api.VolumeStatus_VOLUME_STATUS_NONE {
				info.Status = api.VolumeStatus_VOLUME_STATUS_UP
				inst.UpdateVol(info)
			}
		}
	}

	dlog.Println("NFS initialized and driver mounted at: ", nfsMountPath)
	return inst, nil
}

func (d *driver) Name() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

// Status diagnostic information
func (d *driver) Status() [][2]string {
	return [][2]string{}
}

//
//Utility functions
//
func (d *driver) getNewVolumeServer() (string, error) {
	//randomly select one
	if d.nfsServers != nil && len(d.nfsServers) > 0 {
		return d.nfsServers[rand.Intn(len(d.nfsServers))], nil
	}

	return "", errors.New("No NFS servers found")
}

//get nfsPath for specified volume
func (d *driver) getNFSPath(v *api.Volume) (string, error) {
	locator := v.GetLocator()
	server, ok := locator.VolumeLabels["server"]
	if !ok {
		dlog.Warnf("No server label found on volume")
		return "", errors.New("No server label found on volume: " + v.Id)
	}

	return path.Join(nfsMountPath, server), nil
}

//get nfsPath for specified volume
func (d *driver) getNFSPathById(volumeID string) (string, error) {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return "", err
	}

	return d.getNFSPath(v)
}

//get nfsPath plus volume name for specified volume
func (d *driver) getNFSVolumePath(v *api.Volume) (string, error) {
	parentPath, err := d.getNFSPath(v)
	if err != nil {
		return "", err
	}

	return path.Join(parentPath, v.Id), nil
}

//get nfsPath plus volume name for specified volume
func (d *driver) getNFSVolumePathById(volumeID string) (string, error) {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return "", err
	}

	return d.getNFSVolumePath(v)
}

//append unix time to volumeID
func (d *driver) getNewSnapVolID(volumeID string) string {
	return volumeID + "-" + strconv.FormatUint(uint64(time.Now().Unix()), 10)
}

//
// These functions below implement the volume driver interface.
//

func (d *driver) Create(
	locator *api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (string, error) {

	volumeID := locator.Name
	if volumeID == "" && source.Parent != "" {
		volumeID = d.getNewSnapVolID(source.Parent)
		dlog.Infof("Creating snap vol id: %s", volumeID)
	}

	if _, err := d.GetVol(volumeID); err == nil {
		return "", errors.New("Volume with that name already exists")
	}

	//snapshot passes nil volumelabels
	if locator.VolumeLabels == nil {
		locator.VolumeLabels = make(map[string]string)
	}

	//check if user passed server as option
	labels := locator.GetVolumeLabels()
	_, ok := labels["server"]
	if !ok {
		server, err := d.getNewVolumeServer()
		if err != nil {
			dlog.Infof("no nfs servers found...")
			return "", err
		} else {
			dlog.Infof("Assigning random nfs server: %s to volume: %s", server, volumeID)
		}

		labels["server"] = server
	}

	// Create a directory on the NFS server with this UUID.
	volPathParent := path.Join(nfsMountPath, labels["server"])
	volPath := path.Join(volPathParent, volumeID)
	err := os.MkdirAll(volPath, 0744)
	if err != nil {
		dlog.Println(err)
		return "", err
	}
	if source != nil {
		if len(source.Seed) != 0 {
			seed, err := seed.New(source.Seed, spec.VolumeLabels)
			if err != nil {
				dlog.Warnf("Failed to initailize seed from %q : %v",
					source.Seed, err)
				return "", err
			}
			err = seed.Load(path.Join(volPath, config.DataDir))
			if err != nil {
				dlog.Warnf("Failed to  seed from %q to %q: %v",
					source.Seed, volPathParent, err)
				return "", err
			}
		}
	}

	f, err := os.Create(path.Join(volPathParent, volumeID+nfsBlockFile))
	if err != nil {
		dlog.Println(err)
		return "", err
	}
	defer f.Close()

	if err := f.Truncate(int64(spec.Size)); err != nil {
		dlog.Println(err)
		return "", err
	}

	v := common.NewVolume(
		volumeID,
		api.FSType_FS_TYPE_NFS,
		locator,
		source,
		spec,
	)
	v.DevicePath = path.Join(volPathParent, volumeID+nfsBlockFile)

	if err := d.CreateVol(v); err != nil {
		return "", err
	}
	return v.Id, err
}

func (d *driver) Delete(volumeID string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		dlog.Println(err)
		return err
	}

	// Delete the simulated block volume
	os.Remove(v.DevicePath)

	nfsVolPath, err := d.getNFSVolumePath(v)
	if err != nil {
		return err
	}

	// Delete the directory on the nfs server.
	os.RemoveAll(nfsVolPath)

	err = d.DeleteVol(volumeID)
	if err != nil {
		dlog.Println(err)
		return err
	}

	return nil
}

func (d *driver) MountedAt(mountpath string) string {
	return ""
}

func (d *driver) Mount(volumeID string, mountpath string, options map[string]string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		dlog.Println(err)
		return err
	}

	nfsPath, err := d.getNFSPath(v)
	if err != nil {
		dlog.Printf("Could not find server for volume: %s", volumeID)
		return err
	}

	srcPath := path.Join(":", nfsPath, volumeID)
	mountExists, err := d.mounter.Exists(srcPath, mountpath)
	if !mountExists {
		d.mounter.Unmount(path.Join(nfsPath, volumeID), mountpath,
			syscall.MNT_DETACH, 0, nil)
		if err := d.mounter.Mount(
			0, path.Join(nfsPath, volumeID),
			mountpath,
			string(v.Spec.Format),
			syscall.MS_BIND,
			"",
			0,
			nil,
		); err != nil {
			dlog.Printf("Cannot mount %s at %s because %+v",
				path.Join(nfsPath, volumeID), mountpath, err)
			return err
		}
	}
	if v.AttachPath == nil {
		v.AttachPath = make([]string, 0)
	}
	v.AttachPath = append(v.AttachPath, mountpath)
	return d.UpdateVol(v)
}

func (d *driver) Unmount(volumeID string, mountpath string, options map[string]string) error {
	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}
	if len(v.AttachPath) == 0 {
		return fmt.Errorf("Device %v not mounted", volumeID)
	}

	nfsVolPath, err := d.getNFSVolumePath(v)
	if err != nil {
		return err
	}

	err = d.mounter.Unmount(nfsVolPath, mountpath,
		syscall.MNT_DETACH, 0, nil)
	if err != nil {
		return err
	}
	v.AttachPath = d.mounter.Mounts(nfsVolPath)
	return d.UpdateVol(v)
}

func (d *driver) Snapshot(volumeID string, readonly bool, locator *api.VolumeLocator) (string, error) {
	volIDs := []string{volumeID}
	vols, err := d.Inspect(volIDs)
	if err != nil {
		return "", nil
	}
	source := &api.Source{Parent: volumeID}
	newVolumeID, err := d.Create(locator, source, vols[0].Spec)
	if err != nil {
		return "", nil
	}

	nfsVolPath, err := d.getNFSVolumePathById(volumeID)
	if err != nil {
		return "", err
	}

	newNfsVolPath, err := d.getNFSVolumePathById(newVolumeID)
	if err != nil {
		return "", err
	}

	// NFS does not support snapshots, so just copy the files.
	if err := copyDir(nfsVolPath, newNfsVolPath); err != nil {
		d.Delete(newVolumeID)
		return "", nil
	}
	return newVolumeID, nil
}

func (d *driver) Restore(volumeID string, snapID string) error {
	if _, err := d.Inspect([]string{volumeID, snapID}); err != nil {
		return err
	}

	nfsVolPath, err := d.getNFSVolumePathById(volumeID)
	if err != nil {
		return err
	}

	snapNfsVolPath, err := d.getNFSVolumePathById(snapID)
	if err != nil {
		return err
	}

	// NFS does not support restore, so just copy the files.
	if err := copyDir(snapNfsVolPath, nfsVolPath); err != nil {
		return err
	}
	return nil
}

func (d *driver) Attach(volumeID string, attachOptions map[string]string) (string, error) {

	nfsPath, err := d.getNFSPathById(volumeID)
	if err != nil {
		return "", err
	}

	return path.Join(nfsPath, volumeID+nfsBlockFile), nil
}

func (d *driver) Detach(volumeID string, options map[string]string) error {
	return nil
}

func (d *driver) Set(volumeID string, locator *api.VolumeLocator, spec *api.VolumeSpec) error {
	if spec != nil {
		return volume.ErrNotSupported
	}
	v, err := d.GetVol(volumeID)
	if err != nil {
		return err
	}
	if locator != nil {
		v.Locator = locator
	}
	return d.UpdateVol(v)
}

func (d *driver) Shutdown() {
	dlog.Printf("%s Shutting down", Name)

	for _, v := range d.nfsServers {
		dlog.Infof("Umounting: %s", nfsMountPath+v)
		syscall.Unmount(path.Join(nfsMountPath, v), 0)
	}
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}

func copyDir(source string, dest string) (err error) {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}
