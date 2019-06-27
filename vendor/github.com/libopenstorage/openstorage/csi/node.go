/*
Package csi is CSI driver interface for OSD
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
	"os"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/options"
	"github.com/libopenstorage/openstorage/pkg/util"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.pedge.io/dlog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetNodeID is a CSI API which gets the PX NodeId for the local node
func (s *OsdCsiServer) GetNodeID(ctx context.Context, req *csi.GetNodeIDRequest) (*csi.GetNodeIDResponse, error) {
	dlog.Debugf("GetNodeID req[%#v]", req)

	// Check arguments
	if req.GetVersion() == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be provided")
	}

	clus, err := s.cluster.Enumerate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to Enumerate cluster: %s", err)
	}

	result := &csi.GetNodeIDResponse{
		NodeId: clus.NodeId,
	}

	dlog.Infof("NodeId is %s", result.NodeId)

	return result, nil
}

// NodePublishVolume is a CSI API call which mounts the volume on the specified
// target path on the node.
//
// TODO: Support READ ONLY Mounts
//
func (s *OsdCsiServer) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	dlog.Debugf("NodePublishVolume req[%#v]", req)

	// Check arguments
	if req.GetVersion() == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be provided")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id must be provided")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path must be provided")
	}
	if req.GetVolumeCapability() == nil || req.GetVolumeCapability().GetAccessMode() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume access mode must be provided")
	}

	// Get volume information
	v, err := util.VolumeFromName(s.driver, req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Volume id %s not found: %s",
			req.GetVolumeId(),
			err.Error())
	}

	// Gather volume attributes
	spec, _, _, err := s.specHandler.SpecFromOpts(req.GetVolumeAttributes())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Invalid volume attributes: %#v",
			req.GetVolumeAttributes())
	}

	// This seems weird as a way to change opts to map[string]string
	opts := make(map[string]string)
	if len(spec.GetPassphrase()) != 0 {
		opts[options.OptionsSecret] = spec.GetPassphrase()
	}

	// Verify target location is an existing directory
	// See: https://github.com/container-storage-interface/spec/issues/60
	if err := verifyTargetLocation(req.GetTargetPath()); err != nil {
		return nil, status.Errorf(
			codes.Aborted,
			"Failed to use target location %s: %s",
			req.GetTargetPath(),
			err.Error())
	}

	// If this is for a block driver, first attach the volume
	if s.driver.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		if _, err := s.driver.Attach(req.GetVolumeId(), opts); err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Unable to attach volume: %s",
				err.Error())
		}
	}

	// Mount volume onto the path
	if err := s.driver.Mount(req.GetVolumeId(), req.GetTargetPath(), nil); err != nil {
		// Detach on error
		detachErr := s.driver.Detach(v.GetId(), opts)
		if detachErr != nil {
			dlog.Errorf("Unable to detach volume %s: %s",
				v.GetId(),
				detachErr.Error())
		}
		return nil, status.Errorf(
			codes.Internal,
			"Unable to mount volume %s onto %s: %s",
			req.GetVolumeId(),
			req.GetTargetPath(),
			err.Error())
	}

	dlog.Infof("Volume %s mounted on %s",
		req.GetVolumeId(),
		req.GetTargetPath())

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume is a CSI API call which unmounts the volume.
func (s *OsdCsiServer) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	dlog.Debugf("NodeUnPublishVolume req[%#v]", req)

	// Check arguments
	if req.GetVersion() == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be provided")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume id must be provided")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path must be provided")
	}

	// Get volume information
	_, err := util.VolumeFromName(s.driver, req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Volume id %s not found: %s",
			req.GetVolumeId(),
			err.Error())
	}

	// Verify target location is an existing directory
	// See: https://github.com/container-storage-interface/spec/issues/60
	if err = verifyTargetLocation(req.GetTargetPath()); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Mount volume onto the path
	if err = s.driver.Unmount(req.GetVolumeId(), req.GetTargetPath(), nil); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to unmount volume %s onto %s: %s",
			req.GetVolumeId(),
			req.GetTargetPath(),
			err.Error())
	}

	if s.driver.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		if err = s.driver.Detach(req.GetVolumeId(), nil); err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Unable to detach volume: %s",
				err.Error())
		}
	}

	dlog.Infof("Volume %s unmounted", req.GetVolumeId())

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeProbe is a CSI API function which asks the driver to check if the
// node has all the necessary components to run successfully.
func (s *OsdCsiServer) NodeProbe(
	ctx context.Context,
	req *csi.NodeProbeRequest,
) (*csi.NodeProbeResponse, error) {

	dlog.Debugf("NodeProbe req[%#v]", req)

	// Check arguments
	if req.GetVersion() == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be provided")
	}

	// TBD(lpabon) Here we can add support
	// to scan th system in a future patch.

	// For now return 'ok'.
	return &csi.NodeProbeResponse{}, nil
}

// NodeGetCapabilities is a CSI API function which seems to be setup for
// future patches
func (s *OsdCsiServer) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {

	dlog.Debugf("NodeGetCapabilities req[%#v]", req)

	// Check arguments
	if req.GetVersion() == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be provided")
	}

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

func verifyTargetLocation(targetPath string) error {
	fileInfo, err := os.Stat(targetPath)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("Target location %s does not exist", targetPath)
	} else if err != nil {
		return fmt.Errorf(
			"Unknown error while verifying target location %s: %s",
			targetPath,
			err.Error())
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("Target location %s is not a directory", targetPath)
	}

	return nil
}
