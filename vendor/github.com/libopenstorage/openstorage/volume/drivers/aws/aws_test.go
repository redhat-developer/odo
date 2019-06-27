package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/test"
	"github.com/stretchr/testify/require"
)

func testRemoveTags(t *testing.T, driver volume.VolumeDriver) {
	d := driver.(*Driver)
	// Create volume with labels
	sz := int64(1)
	voltype := opsworks.VolumeTypeIo1
	ec2Vol := &ec2.Volume{
		AvailabilityZone: &d.md.zone,
		VolumeType:       &voltype,
		Size:             &sz,
	}
	labelNames := []string{"label1", "label2"}
	labels := make(map[string]string)
	for _, name := range labelNames {
		labels[name] = name
	}
	resp, err := d.ops.Create(ec2Vol, labels)
	require.Nil(t, err, "Failed in CreateVolumeRequest :%v", err)
	vol, ok := resp.(*ec2.Volume)
	require.True(t, ok, "invalid volume returned by create API")
	defer d.ops.Delete(*vol.VolumeId)

	tags, err := d.ops.Tags(*vol.VolumeId)
	require.Nil(t, err, "Failed to apply tags :%v", err)
	require.True(t, len(tags) == len(labelNames), "ApplyTags failed")
	require.Nil(t, d.ops.RemoveTags(*vol.VolumeId, labels), "RemoveTags error")
	tags, err = d.ops.Tags(*vol.VolumeId)
	require.Nil(t, err, "Failed to fetch tags :%v", err)
	require.True(t, len(tags) == 0, "RemoveTags failed")
}

func testFreeDevices(t *testing.T, driver volume.VolumeDriver) {
	d := driver.(*Driver)

	deviceNames := []string{"/dev/sda1", "/dev/sdb", "/dev/xvdf", "/dev/xvdg", "/dev/xvdcg"}
	var blockDeviceMappings []interface{}
	for i, _ := range deviceNames {
		b := &ec2.InstanceBlockDeviceMapping{
			DeviceName: &deviceNames[i],
		}
		blockDeviceMappings = append(blockDeviceMappings, b)
	}
	freeDeviceNames, err := d.ops.FreeDevices(blockDeviceMappings, "/dev/sda1")
	require.NoError(t, err, "Expected no error")
	// Free devices : h -> p
	require.Equal(t, len(freeDeviceNames), 9, "No. of free devices do not match")
	badDeviceName := "/dev/xvdcgh"
	b := &ec2.InstanceBlockDeviceMapping{
		DeviceName: &badDeviceName,
	}

	blockDeviceMappings = append(blockDeviceMappings, b)
	freeDeviceNames, err = d.ops.FreeDevices(blockDeviceMappings, "/dev/sda1")
	require.Error(t, err, "Expected an error")
}

func TestAll(t *testing.T) {
	if _, err := credentials.NewEnvCredentials().Get(); err != nil {
		t.Skip("No AWS credentials, skipping AWS dependent driver tests: ", err)
	}
	driver, err := Init(map[string]string{})
	if err != nil {
		t.Fatalf("Failed to initialize Volume Driver: %v", err)
	}

	testFreeDevices(t, driver)
	ctx := test.NewContext(driver)
	ctx.Filesystem = api.FSType_FS_TYPE_EXT4
	test.RunShort(t, ctx)
	testRemoveTags(t, driver)
}
