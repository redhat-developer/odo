package aws_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/libopenstorage/openstorage/pkg/storageops/aws"
	"github.com/libopenstorage/openstorage/pkg/storageops/test"
	uuid "github.com/satori/go.uuid"
)

const (
	newDiskSizeInGB = 10
	newDiskPrefix   = "openstorage-test"
)

var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.NewV4())

func TestAll(t *testing.T) {
	drivers := make(map[string]storageops.Ops)
	diskTemplates := make(map[string]map[string]interface{})

	if d, err := aws.NewEnvClient(); err != aws.ErrAWSEnvNotAvailable {
		volType := opsworks.VolumeTypeGp2
		volSize := int64(newDiskSizeInGB)
		zone := os.Getenv("AWS_ZONE")
		ebsVol := &ec2.Volume{
			AvailabilityZone: &zone,
			VolumeType:       &volType,
			Size:             &volSize,
		}
		drivers[d.Name()] = d
		diskTemplates[d.Name()] = map[string]interface{}{
			diskName: ebsVol,
		}
	} else {
		fmt.Printf("skipping AWS tests as environment is not set...\n")
	}

	test.RunTest(drivers, diskTemplates, t)
}
