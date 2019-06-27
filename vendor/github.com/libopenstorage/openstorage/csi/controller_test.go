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
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestControllerGetCapabilities(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ControllerGetCapabilities(
		context.Background(),
		&csi.ControllerGetCapabilitiesRequest{
			Version: &csi.Version{},
		})
	assert.Nil(t, err)

	// Verify
	expectedValues := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	}
	caps := r.GetCapabilities()
	assert.Len(t, caps, len(expectedValues))
	found := 0
	for _, expectedCap := range expectedValues {
		for _, cap := range caps {
			if cap.GetRpc().GetType() == expectedCap {
				found++
				break
			}
		}
	}
	assert.Equal(t, found, len(expectedValues))
}

func TestControllerPublishVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ControllerPublishVolume(context.Background(), &csi.ControllerPublishVolumeRequest{})
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.Unimplemented)
	assert.Contains(t, serverError.Message(), "not supported")
}

func TestControllerUnPublishVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Make a call
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ControllerUnpublishVolume(context.Background(), &csi.ControllerUnpublishVolumeRequest{})
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.Unimplemented)
	assert.Contains(t, serverError.Message(), "not supported")
}

func TestControllerValidateVolumeCapabilitiesBadArguments(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &csi.ValidateVolumeCapabilitiesRequest{}

	// Missing everything
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Version")

	// Miss capabilities and id
	req.Version = &csi.Version{}
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "capabilities")

	// Miss id and capabilities len is 0
	req.Version = &csi.Version{}
	req.VolumeCapabilities = []*csi.VolumeCapability{}
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "capabilities")

	// Miss id
	req.Version = &csi.Version{}
	req.VolumeCapabilities = []*csi.VolumeCapability{
		&csi.VolumeCapability{},
	}
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "volume_id")
}

func TestControllerValidateVolumeInvalidId(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"
	gomock.InOrder(
		// First time called it will say it is not there
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return(nil, fmt.Errorf("Id not found")),

		// Second time called it will not return an error,
		// but return an empty list
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{}, nil),

		// Third time it is called, it will return
		// a good list with a volume with an id that does
		// not match (even if this probably never could happen)
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: "bad volume id",
				},
			}, nil),

		// Fourth time driver will return a list with more than
		// one volume, which should be unexpected since it only
		// asked for one volume.
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: "bad volume id 1",
				},
				&api.Volume{
					Id: "bad volume id 2",
				},
			}, nil),
	)

	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		VolumeId: id,
	}

	// Missing everything
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "ID not found")

	// Send again, same result
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "ID not found")

	// Now it should be an internal id error
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Driver volume id")

	// Now driver should have returned an unexpected number of volumes
	_, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "unexpected number of volumes")
}

func TestControllerValidateVolumeInvalidCapabilities(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
			},
		}, nil).
		Times(1)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		VolumeId: id,
	}

	// Make request
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)

	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Cannot have both")
}

func TestControllerValidateVolumeAccessModeSNWR(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"

	// RO SH
	// x   x
	// *   x
	// x   *
	// *   *
	gomock.InOrder(
		// not-RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
		},
		VolumeId: id,
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.True(t, r.Supported)

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)
}

func TestControllerValidateVolumeAccessModeSNRO(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"

	// RO SH
	// x   x
	// *   x
	// x   *
	// *   *
	gomock.InOrder(
		// not-RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
				},
			},
		},
		VolumeId: id,
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.True(t, r.Supported)

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)
}

func TestControllerValidateVolumeAccessModeMNRO(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"

	// RO SH
	// x   x
	// *   x
	// x   *
	// *   *
	gomock.InOrder(
		// not-RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
				},
			},
		},
		VolumeId: id,
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.True(t, r.Supported)
}

func TestControllerValidateVolumeAccessModeMNWR(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"

	// RO SH
	// x   x
	// *   x
	// x   *
	// *   *
	gomock.InOrder(
		// not-RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// RO and not-SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: false,
					},
				},
			}, nil),

		// not-RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: false,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),

		// RO and SH
		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id:       id,
					Readonly: true,
					Spec: &api.VolumeSpec{
						Shared: true,
					},
				},
			}, nil),
	)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
				},
			},
		},
		VolumeId: id,
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	r, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect RO and non-SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)

	// Expect non-RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.True(t, r.Supported)

	// Expect RO and SH
	r, err = c.ValidateVolumeCapabilities(context.Background(), req)
	assert.Nil(t, err)
	assert.False(t, r.Supported)
}

func TestControllerValidateVolumeAccessModeUnknown(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup mock
	id := "testvolumeid"
	s.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id:       id,
				Readonly: false,
				Spec: &api.VolumeSpec{
					Shared: false,
				},
			},
		}, nil).
		Times(1)

	// Setup request
	req := &csi.ValidateVolumeCapabilitiesRequest{
		Version: &csi.Version{},
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{
					Mount: &csi.VolumeCapability_MountVolume{},
				},
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: csi.VolumeCapability_AccessMode_UNKNOWN,
				},
			},
		},
		VolumeId: id,
	}

	// Expect non-RO and non-SH
	c := csi.NewControllerClient(s.Conn())
	_, err := c.ValidateVolumeCapabilities(context.Background(), req)
	assert.NotNil(t, err)
}

func TestControllerListVolumesInvalidArguments(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	req := &csi.ListVolumesRequest{}

	// Expect error without version
	_, err := c.ListVolumes(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Version")

	// Expect error with maxentries set
	// To be removed once CSI Spec issue #138 is resolved
	req.Version = &csi.Version{}
	req.MaxEntries = 1
	_, err = c.ListVolumes(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Unimplemented)
	assert.Contains(t, serverError.Message(), "token")
}

func TestControllerListVolumesEnumerateError(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup mock
	s.MockDriver().
		EXPECT().
		Enumerate(gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("TEST")).
		Times(1)

	// Setup request
	req := &csi.ListVolumesRequest{
		Version: &csi.Version{},
	}

	// Expect that the Enumerate call failed
	_, err := c.ListVolumes(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "TEST")
}

func TestControllerListVolumes(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup mock
	mockVolumeList := []*api.Volume{
		&api.Volume{
			Id:            "one",
			AttachedState: api.AttachState_ATTACH_STATE_INTERNAL,
			Readonly:      false,
			State:         api.VolumeState_VOLUME_STATE_ERROR,
			Spec: &api.VolumeSpec{
				Shared: true,
				Size:   uint64(11),
			},
			Error: "TEST",
		},
		&api.Volume{
			Id:            "two",
			AttachedState: api.AttachState_ATTACH_STATE_EXTERNAL,
			Readonly:      true,
			State:         api.VolumeState_VOLUME_STATE_ATTACHED,
			Spec: &api.VolumeSpec{
				Encrypted: true,
				Shared:    true,
				Size:      uint64(22),
			},
			Source: &api.Source{
				Parent: "myparentid",
			},
		},
		&api.Volume{
			Id:            "three",
			AttachedState: api.AttachState_ATTACH_STATE_INTERNAL_SWITCH,
			Readonly:      true,
			State:         api.VolumeState_VOLUME_STATE_TRY_DETACHING,
			Spec: &api.VolumeSpec{
				Shared: false,
				Size:   uint64(33),
			},
		},
	}
	s.MockDriver().
		EXPECT().
		Enumerate(gomock.Any(), gomock.Any()).
		Return(mockVolumeList, nil).
		Times(1)

	// Setup request
	req := &csi.ListVolumesRequest{
		Version: &csi.Version{},
	}

	// Expect error without version
	r, err := c.ListVolumes(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	volumes := r.GetEntries()
	assert.Equal(t, len(mockVolumeList), len(volumes))
	assert.Equal(t, len(r.GetNextToken()), 0)

	found := 0
	for _, mv := range mockVolumeList {
		for _, v := range volumes {
			info := v.GetVolumeInfo()
			assert.NotNil(t, info)

			if mv.GetId() == info.GetId() {
				found++
				assert.Equal(t, info.GetCapacityBytes(), mv.GetSpec().GetSize())

				attributes := info.GetAttributes()
				assert.Equal(t, attributes["readonly"], fmt.Sprintf("%v", mv.GetReadonly()))
				assert.Equal(t, attributes[api.SpecShared], fmt.Sprintf("%v", mv.GetSpec().GetShared()))
				assert.Equal(t, attributes["state"], mv.GetState().String())
				assert.Equal(t, attributes["attached"], mv.GetAttachedState().String())
				assert.Equal(t, attributes["error"], mv.GetError())
				assert.Equal(t, attributes[api.SpecParent], mv.GetSource().GetParent())
				assert.Equal(t, attributes[api.SpecSecure], fmt.Sprintf("%v", mv.GetSpec().GetEncrypted()))
				break
			}
		}
	}
	assert.Equal(t, found, len(mockVolumeList))
}

func TestControllerCreateVolumeInvalidArguments(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	req := &csi.CreateVolumeRequest{}

	// No version
	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Version")

	// No name
	req.Version = &csi.Version{}
	_, err = c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Name")

	// No volume capabilities
	name := "myname"
	req.Name = name
	_, err = c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Volume capabilities")

	// Zero volume capabilities
	req.VolumeCapabilities = []*csi.VolumeCapability{}
	_, err = c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Volume capabilities")

}

func TestControllerCreateVolumeFoundByVolumeFromNameConflict(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	tests := []struct {
		name string
		req  *csi.CreateVolumeRequest
		ret  *api.Volume
	}{
		{
			name: "size",
			req: &csi.CreateVolumeRequest{
				Version: &csi.Version{},
				Name:    "size",
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{},
				},
				CapacityRange: &csi.CapacityRange{

					// Requested size does not match volume size
					RequiredBytes: 1000,
				},
			},
			ret: &api.Volume{
				Id: "size",
				Locator: &api.VolumeLocator{
					Name: "size",
				},
				Spec: &api.VolumeSpec{

					// Size is different
					Size: 10,
				},
			},
		},
		{
			name: "shared",
			req: &csi.CreateVolumeRequest{
				Version: &csi.Version{},
				Name:    "shared",
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{

							// Set as a shared volume
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					},
				},
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 10,
				},
			},
			ret: &api.Volume{
				Id: "shared",
				Locator: &api.VolumeLocator{
					Name: "shared",
				},
				Spec: &api.VolumeSpec{
					Size: 10,

					// Set as non-shared.
					Shared: false,
				},
			},
		},
		{
			name: "parent",
			req: &csi.CreateVolumeRequest{
				Version: &csi.Version{},
				Name:    "parent",
				VolumeCapabilities: []*csi.VolumeCapability{
					&csi.VolumeCapability{},
				},
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 10,
				},
				Parameters: map[string]string{
					"parent": "notmyparent",
				},
			},
			ret: &api.Volume{
				Id: "parent",
				Locator: &api.VolumeLocator{
					Name: "parent",
				},
				Spec: &api.VolumeSpec{
					Size: 10,
				},
				Source: &api.Source{
					Parent: "myparent",
				},
			},
		},
	}

	for _, test := range tests {
		gomock.InOrder(
			s.MockDriver().
				EXPECT().
				Inspect([]string{test.name}).
				Return(nil, fmt.Errorf("not found")).
				Times(1),

			s.MockDriver().
				EXPECT().
				Enumerate(&api.VolumeLocator{Name: test.name}, nil).
				Return([]*api.Volume{test.ret}, nil).
				Times(1),
		)
		_, err := c.CreateVolume(context.Background(), test.req)
		assert.Error(t, err)
		serverError, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, serverError.Code(), codes.AlreadyExists)
	}
}

func TestControllerCreateVolumeNoCapacity(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "myvol"
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
	}

	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(
				locator *api.VolumeLocator,
				Source *api.Source,
				spec *api.VolumeSpec,
			) (string, error) {
				assert.Equal(t, spec.Size, defaultCSIVolumeSize)
				return id, nil
			}).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size:   defaultCSIVolumeSize,
						Shared: true,
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolumeInfo()

	assert.Equal(t, id, volumeInfo.GetId())
	assert.Equal(t, defaultCSIVolumeSize, volumeInfo.GetCapacityBytes())
}

func TestControllerCreateVolumeFoundByVolumeFromName(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
	}

	// Volume is already being created and found by calling VolumeFromName
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolumeInfo()

	assert.Equal(t, name, volumeInfo.GetId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
}

func TestControllerCreateVolumeBadParameters(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecFilesystem: "whatkindoffsisthis?",
		},
	}

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "get parameters")
}

func TestControllerCreateVolumeBadParentId(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "myvol"
	size := uint64(1234)
	parent := "badid"
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecParent: parent,
		},
	}

	// Volume is already being created and found by calling VolumeFromName
	gomock.InOrder(
		// Getting volume information
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		// Getting parent information
		s.MockDriver().
			EXPECT().
			Inspect([]string{parent}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: parent}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
	)

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "get parent volume")
}

func TestControllerCreateVolumeBadSnapshot(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "myvol"
	size := uint64(1234)
	parent := "parent"
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecParent: parent,
		},
	}

	// Volume is already being created and found by calling VolumeFromName
	gomock.InOrder(
		// Getting volume information
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		// Getting parent information
		s.MockDriver().
			EXPECT().
			Inspect([]string{parent}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: parent}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: parent,
					Locator: &api.VolumeLocator{
						Name: parent,
					},
					Spec: &api.VolumeSpec{
						Size: uint64(1234),
					},
				},
			}, nil).
			Times(1),

		// Return an error from snapshot
		s.MockDriver().
			EXPECT().
			Snapshot(parent, false, &api.VolumeLocator{Name: name}).
			Return("", fmt.Errorf("snapshoterr")).
			Times(1),
	)

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "snapshoterr")
}

func TestControllerCreateVolumeWithSharedVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	modes := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
	}

	for _, mode := range modes {
		// Setup request
		name := "myvol"
		size := uint64(1234)
		req := &csi.CreateVolumeRequest{
			Version: &csi.Version{},
			Name:    name,
			VolumeCapabilities: []*csi.VolumeCapability{
				&csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: mode,
					},
				},
			},
			CapacityRange: &csi.CapacityRange{
				RequiredBytes: size,
			},
		}

		// Setup mock functions
		id := "myid"
		gomock.InOrder(
			s.MockDriver().
				EXPECT().
				Inspect([]string{name}).
				Return(nil, fmt.Errorf("not found")).
				Times(1),

			s.MockDriver().
				EXPECT().
				Enumerate(&api.VolumeLocator{Name: name}, nil).
				Return(nil, fmt.Errorf("not found")).
				Times(1),

			s.MockDriver().
				EXPECT().
				Create(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(id, nil).
				Times(1),

			s.MockDriver().
				EXPECT().
				Inspect([]string{id}).
				Return([]*api.Volume{
					&api.Volume{
						Id: id,
						Locator: &api.VolumeLocator{
							Name: name,
						},
						Spec: &api.VolumeSpec{
							Size:   size,
							Shared: true,
						},
					},
				}, nil).
				Times(1),
		)

		r, err := c.CreateVolume(context.Background(), req)
		assert.Nil(t, err)
		assert.NotNil(t, r)
		volumeInfo := r.GetVolumeInfo()

		assert.Equal(t, id, volumeInfo.GetId())
		assert.Equal(t, size, volumeInfo.GetCapacityBytes())
		assert.Equal(t, "true", volumeInfo.GetAttributes()[api.SpecShared])
	}
}

func TestControllerCreateVolumeFails(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
	}

	// Setup mock functions
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("createerror")).
			Times(1),
	)

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "createerror")
}

func TestControllerCreateVolumeNoNewVolumeInfo(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: id}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
	)

	_, err := c.CreateVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestControllerCreateVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolumeInfo()

	assert.Equal(t, id, volumeInfo.GetId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.NotEqual(t, "true", volumeInfo.GetAttributes()[api.SpecShared])
}

func TestControllerCreateVolumeSnapshot(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// Setup request
	mockParentID := "parendId"
	name := "myvol"
	size := uint64(1234)
	req := &csi.CreateVolumeRequest{
		Version: &csi.Version{},
		Name:    name,
		VolumeCapabilities: []*csi.VolumeCapability{
			&csi.VolumeCapability{},
		},
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		Parameters: map[string]string{
			api.SpecParent: mockParentID,
		},
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{mockParentID}).
			Return([]*api.Volume{
				&api.Volume{
					Id: mockParentID,
				},
			}, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Snapshot(mockParentID, false, &api.VolumeLocator{
				Name: name,
			}).
			Return(id, nil).
			Times(1),

		s.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
					Source: &api.Source{
						Parent: mockParentID,
					},
				},
			}, nil).
			Times(1),
	)

	r, err := c.CreateVolume(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	volumeInfo := r.GetVolumeInfo()

	assert.Equal(t, id, volumeInfo.GetId())
	assert.Equal(t, size, volumeInfo.GetCapacityBytes())
	assert.NotEqual(t, "true", volumeInfo.GetAttributes()[api.SpecShared])
	assert.Equal(t, mockParentID, volumeInfo.GetAttributes()[api.SpecParent])
}

func TestControllerDeleteVolumeInvalidArguments(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// No version
	req := &csi.DeleteVolumeRequest{}
	_, err := c.DeleteVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Version")

	// No id
	req.Version = &csi.Version{}
	_, err = c.DeleteVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Volume id")
}

func TestControllerDeleteVolumeError(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// No version
	myid := "myid"
	req := &csi.DeleteVolumeRequest{
		Version:  &csi.Version{},
		VolumeId: myid,
	}

	// Setup mock
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{myid}).
			Return([]*api.Volume{
				&api.Volume{
					Id: myid,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Delete(myid).
			Return(fmt.Errorf("MOCKERRORTEST")).
			Times(1),
	)

	_, err := c.DeleteVolume(context.Background(), req)
	assert.NotNil(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Unable to delete")
	assert.Contains(t, serverError.Message(), "MOCKERRORTEST")
}

func TestControllerDeleteVolume(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	c := csi.NewControllerClient(s.Conn())

	// No version
	myid := "myid"
	req := &csi.DeleteVolumeRequest{
		Version:  &csi.Version{},
		VolumeId: myid,
	}

	// Setup mock
	// According to CSI spec, if the ID is not found, it must return OK
	s.MockDriver().
		EXPECT().
		Inspect([]string{myid}).
		Return(nil, kvdb.ErrNotFound).
		Times(1)

	_, err := c.DeleteVolume(context.Background(), req)
	assert.Nil(t, err)

	// According to CSI spec, if the ID is not found, it must return OK
	// Now return no error, but empty list
	s.MockDriver().
		EXPECT().
		Inspect([]string{myid}).
		Return([]*api.Volume{}, nil).
		Times(1)

	_, err = c.DeleteVolume(context.Background(), req)
	assert.Nil(t, err)

	// Setup mock
	gomock.InOrder(
		s.MockDriver().
			EXPECT().
			Inspect([]string{myid}).
			Return([]*api.Volume{
				&api.Volume{
					Id: myid,
				},
			}, nil).
			Times(1),
		s.MockDriver().
			EXPECT().
			Delete(myid).
			Return(nil).
			Times(1),
	)

	_, err = c.DeleteVolume(context.Background(), req)
	assert.Nil(t, err)
}
