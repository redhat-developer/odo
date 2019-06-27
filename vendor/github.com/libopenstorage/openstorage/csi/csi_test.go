/*
CSI Interface for OSD
Copyright 2017 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package csi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"
	"golang.org/x/net/context"

	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

const (
	mockDriverName = "mock"
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service
type testServer struct {
	conn   *grpc.ClientConn
	server Server
	m      *mockdriver.MockVolumeDriver
	c      *mockcluster.MockCluster
	mc     *gomock.Controller
}

func setupMockDriver(tester *testServer, t *testing.T) {
	volumedrivers.Add(mockDriverName, func(map[string]string) (volume.VolumeDriver, error) {
		return tester.m, nil
	})

	var err error

	// Register mock driver
	err = volumedrivers.Register(mockDriverName, nil)
	assert.Nil(t, err)
}

func newTestServer(t *testing.T) *testServer {
	tester := &testServer{}

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)
	tester.c = mockcluster.NewMockCluster(tester.mc)

	setupMockDriver(tester, t)

	var err error
	// Setup simple driver
	tester.server, err = NewOsdCsiServer(&OsdCsiServerConfig{
		DriverName: mockDriverName,
		Net:        "tcp",
		Address:    "127.0.0.1:0",
		Cluster:    tester.c,
	})
	assert.Nil(t, err)
	err = tester.server.Start()
	assert.Nil(t, err)

	// Setup a connection to the driver
	tester.conn, err = grpc.Dial(tester.server.Address(), grpc.WithInsecure())
	assert.Nil(t, err)

	return tester
}

func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

func (s *testServer) MockCluster() *mockcluster.MockCluster {
	return s.c
}

func (s *testServer) Stop() {
	// Remove from registry
	volumedrivers.Remove("mock")

	// Shutdown servers
	s.conn.Close()
	s.server.Stop()

	// Check mocks
	s.mc.Finish()
}

func (s *testServer) Conn() *grpc.ClientConn {
	return s.conn
}

func (s *testServer) Server() Server {
	return s.server
}

func TestCSIServerStart(t *testing.T) {
	s := newTestServer(t)
	assert.True(t, s.Server().IsRunning())
	defer s.Stop()

	// Check if we can still talk to the server
	// after starting multiple times.
	err := s.Server().Start()
	assert.True(t, s.Server().IsRunning())
	assert.NotNil(t, err)
	err = s.Server().Start()
	assert.True(t, s.Server().IsRunning())
	assert.NotNil(t, err)
	err = s.Server().Start()
	assert.True(t, s.Server().IsRunning())
	assert.NotNil(t, err)

	// Make a call
	s.MockDriver().EXPECT().Name().Return("mock").Times(2)
	c := csi.NewIdentityClient(s.Conn())
	r, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{
		Version: &csi.Version{},
	})
	assert.Nil(t, err)

	// Verify
	name := r.GetName()
	version := r.GetVendorVersion()
	assert.Equal(t, name, csiDriverNamePrefix+"mock")
	assert.Equal(t, version, csiDriverVersion)
}

func TestCSIServerStop(t *testing.T) {
	s := newTestServer(t)
	assert.True(t, s.Server().IsRunning())
	s.Stop()
	assert.False(t, s.Server().IsRunning())

	assert.NotPanics(t, s.Stop)
	assert.False(t, s.Server().IsRunning())
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.Server().IsRunning())
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.Server().IsRunning())
	assert.NotPanics(t, s.Stop)
	assert.False(t, s.Server().IsRunning())
}

func TestNewCSIServerBadParameters(t *testing.T) {
	setupMockDriver(&testServer{}, t)
	s, err := NewOsdCsiServer(nil)
	assert.Nil(t, s)
	assert.NotNil(t, err)

	s, err = NewOsdCsiServer(&OsdCsiServerConfig{})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "must be provided")

	s, err = NewOsdCsiServer(&OsdCsiServerConfig{
		Net: "test",
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "must be provided")

	s, err = NewOsdCsiServer(&OsdCsiServerConfig{
		Net:     "test",
		Address: "blah",
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "must be provided")

	s, err = NewOsdCsiServer(&OsdCsiServerConfig{
		Net:        "test",
		Address:    "blah",
		DriverName: "name",
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unable to get driver")

	// Add driver to registry
	mc := gomock.NewController(t)
	defer mc.Finish()
	m := mockdriver.NewMockVolumeDriver(mc)
	volumedrivers.Add("mock", func(map[string]string) (volume.VolumeDriver, error) {
		return m, nil
	})
	defer volumedrivers.Remove("mock")
	s, err = NewOsdCsiServer(&OsdCsiServerConfig{
		Net:        "test",
		Address:    "blah",
		DriverName: "mock",
	})
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unable to setup server")
}
