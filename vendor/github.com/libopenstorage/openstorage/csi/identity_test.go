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
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewCSIServerGetPluginInfo(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	s.MockDriver().EXPECT().Name().Return("mock").Times(2)

	// Setup client
	c := csi.NewIdentityClient(s.Conn())

	// No version added
	_, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Version")

	// No version added
	r, err := c.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{
		Version: &csi.Version{},
	})
	assert.NoError(t, err)

	// Verify
	name := r.GetName()
	version := r.GetVendorVersion()
	assert.Equal(t, name, csiDriverNamePrefix+"mock")
	assert.Equal(t, version, csiDriverVersion)

	manifest := r.GetManifest()
	assert.Len(t, manifest, 1)
	assert.Equal(t, manifest["driver"], "mock")
}

func TestNewCSIServerGetSupportedVersions(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewIdentityClient(s.Conn())
	r, err := c.GetSupportedVersions(context.Background(), &csi.GetSupportedVersionsRequest{})
	assert.Nil(t, err)

	// Verify
	versions := r.GetSupportedVersions()
	assert.Equal(t, len(versions), 1)
	assert.True(t, reflect.DeepEqual(versions[0], csiVersion))
}
