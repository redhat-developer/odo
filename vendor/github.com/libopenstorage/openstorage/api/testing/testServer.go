package testing

import (
	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"
	"go.pedge.io/dlog"

	client "github.com/libopenstorage/openstorage/api/client"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

// testServer is a simple struct used abstract
// the creation and setup of mock server
type testServer struct {
	client *client.Client
	m      *mockdriver.MockVolumeDriver
	c      *mockcluster.MockCluster
	mc     *gomock.Controller
}

func newTestServer(driver string) *testServer {

	var ts = &testServer{}

	// Add driver to registry
	ts.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	ts.m = mockdriver.NewMockVolumeDriver(ts.mc)
	ts.c = mockcluster.NewMockCluster(ts.mc)

	err := volumedrivers.Add(driver, func(map[string]string) (volume.VolumeDriver, error) {
		return ts.m, nil
	})

	if err != nil {
		dlog.Errorf("Failed to add the driver [%s] for tests", driver)
	}

	// Register the mock driver
	err = volumedrivers.Register(driver, nil)

	return ts
}

// MockDriver helper method.
func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

// MockCluster helper method.
func (s *testServer) MockCluster() *mockcluster.MockCluster {
	return s.c
}

// Stop method to to remove the driver and check mocks.
func (s *testServer) Stop() {
	// Remove from registry
	volumedrivers.Remove("mock")
	// Check mocks
	s.mc.Finish()
}
