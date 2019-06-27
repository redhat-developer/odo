package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/stretchr/testify/require"
)

func RunTest(drivers map[string]storageops.Ops,
	diskTemplates map[string]map[string]interface{},
	t *testing.T) {
	for _, d := range drivers {
		name(t, d)

		for _, template := range diskTemplates[d.Name()] {
			disk := create(t, d, template)
			diskName := id(t, d, disk)
			snapshot(t, d, diskName)
			tags(t, d, diskName)
			enumerate(t, d, diskName)
			inspect(t, d, diskName)
			attach(t, d, diskName)
			devicePath(t, d, diskName)
			teardown(t, d, diskName)
		}
	}
}

func name(t *testing.T, driver storageops.Ops) {
	name := driver.Name()
	require.NotEmpty(t, name, "driver returned empty name")
}

func create(t *testing.T, driver storageops.Ops, template interface{}) interface{} {
	d, err := driver.Create(template, nil)
	require.NoError(t, err, "failed to create disk")
	require.NotNil(t, d, "got nil disk from create api")

	return d
}

func id(t *testing.T, driver storageops.Ops, disk interface{}) string {
	id, err := driver.GetDeviceID(disk)
	require.NoError(t, err, "failed to get disk ID")
	require.NotEmpty(t, id, "got empty disk name/ID")
	return id
}

func snapshot(t *testing.T, driver storageops.Ops, diskName string) {
	snap, err := driver.Snapshot(diskName, true)
	require.NoError(t, err, "failed to create snapshot")
	require.NotEmpty(t, snap, "got empty snapshot from create API")

	snapID, err := driver.GetDeviceID(snap)
	require.NoError(t, err, "failed to get snapshot ID")
	require.NotEmpty(t, snapID, "got empty snapshot name/ID")

	err = driver.SnapshotDelete(snapID)
	require.NoError(t, err, "failed to delete snapshot")
}

func tags(t *testing.T, driver storageops.Ops, diskName string) {
	labels := map[string]string{
		"source": "openstorage-test",
		"foo":    "bar",
	}

	err := driver.ApplyTags(diskName, labels)
	require.NoError(t, err, "failed to apply tags to disk")

	tags, err := driver.Tags(diskName)
	require.NoError(t, err, "failed to get tags for disk")
	require.Len(t, tags, 2, "invalid number of labels found on disk")

	labelsToRemove := map[string]string{"foo": "bar"}
	err = driver.RemoveTags(diskName, labelsToRemove)
	require.NoError(t, err, "failed to remove tags from disk")

	tags, err = driver.Tags(diskName)
	require.NoError(t, err, "failed to get tags for disk")
	require.Len(t, tags, 1, "invalid number of labels found on disk")
}

func enumerate(t *testing.T, driver storageops.Ops, diskName string) {
	disks, err := driver.Enumerate([]*string{&diskName}, nil, storageops.SetIdentifierNone)
	require.NoError(t, err, "failed to create disk")
	require.Len(t, disks, 1, "inspect returned invalid length")
}

func inspect(t *testing.T, driver storageops.Ops, diskName string) {
	disks, err := driver.Inspect([]*string{&diskName})
	require.NoError(t, err, "failed to create disk")
	require.Len(t, disks, 1, fmt.Sprintf("inspect returned invalid length: %d", len(disks)))
}

func attach(t *testing.T, driver storageops.Ops, diskName string) {
	devPath, err := driver.Attach(diskName)
	require.NoError(t, err, "disk attach returned error")
	require.NotEmpty(t, devPath, "disk attach returned empty devicePath")

	mappings, err := driver.DeviceMappings()
	require.NoError(t, err, "get device mappings returned error")
	require.NotEmpty(t, mappings, "received empty device mappings")
}

func devicePath(t *testing.T, driver storageops.Ops, diskName string) {
	devPath, err := driver.DevicePath(diskName)
	require.NoError(t, err, "get device path returned error")
	require.NotEmpty(t, devPath, "received empty devicePath")
}

func teardown(t *testing.T, driver storageops.Ops, diskName string) {
	err := driver.Detach(diskName)
	require.NoError(t, err, "disk detach returned error")

	time.Sleep(3 * time.Second)

	err = driver.Delete(diskName)
	require.NoError(t, err, "failed to delete disk")
}
