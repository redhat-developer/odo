package gce_test

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/libopenstorage/openstorage/pkg/storageops/gce"
	"github.com/libopenstorage/openstorage/pkg/storageops/test"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
)

const (
	newDiskSizeInGB    = 10
	newDiskPrefix      = "openstorage-test"
	newDiskDescription = "Disk created by Openstorage tests"
)

var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.NewV4())

func initGCE(t *testing.T) (storageops.Ops, map[string]interface{}) {
	driver, err := gce.NewClient()
	require.NoError(t, err, "failed to instantiate storage ops driver")

	template := &compute.Disk{
		Description: newDiskDescription,
		Name:        diskName,
		SizeGb:      newDiskSizeInGB,
	}

	return driver, map[string]interface{}{
		diskName: template,
	}
}

func TestAll(t *testing.T) {
	if gce.IsDevMode() {
		drivers := make(map[string]storageops.Ops)
		diskTemplates := make(map[string]map[string]interface{})

		d, disks := initGCE(t)
		drivers[d.Name()] = d
		diskTemplates[d.Name()] = disks

		test.RunTest(drivers, diskTemplates, t)
	} else {
		fmt.Printf("skipping GCE tests as environment is not set...\n")
		t.Skip("skipping GCE tests as environment is not set...")
	}

}
