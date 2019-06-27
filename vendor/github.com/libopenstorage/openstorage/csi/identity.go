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
	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.pedge.io/dlog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	csiDriverVersion    = "0.1.0"
	csiDriverNamePrefix = "com.openstorage."
)

var (
	csiVersion = &csi.Version{
		Major: 0,
		Minor: 1,
		Patch: 0,
	}
)

// GetSupportedVersions is a CSI API which returns the supported CSI version
func (s *OsdCsiServer) GetSupportedVersions(
	context.Context,
	*csi.GetSupportedVersionsRequest) (*csi.GetSupportedVersionsResponse, error) {
	return &csi.GetSupportedVersionsResponse{
		SupportedVersions: []*csi.Version{
			csiVersion,
		},
	}, nil
}

// GetPluginInfo is a CSI API which returns the information about the plugin.
// This includes name, version, and any other OSD specific information
func (s *OsdCsiServer) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {

	dlog.Debugf("GetPluginInfo req[%#v]", req)

	// Check arguments
	if req.GetVersion() == nil {
		return nil, status.Error(codes.InvalidArgument, "Version must be provided")
	}

	return &csi.GetPluginInfoResponse{
		Name:          csiDriverNamePrefix + s.driver.Name(),
		VendorVersion: csiDriverVersion,

		// As OSD CSI Driver matures, add here more information
		Manifest: map[string]string{
			"driver": s.driver.Name(),
		},
	}, nil
}
