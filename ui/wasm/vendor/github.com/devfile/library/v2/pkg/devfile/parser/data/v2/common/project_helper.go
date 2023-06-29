//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// GetDefaultSource get information about primary source
// returns 3 strings: remote name, remote URL, reference(revision)
func GetDefaultSource(ps v1.GitLikeProjectSource) (remoteName string, remoteURL string, revision string, err error) {
	// get git checkout information
	// if there are multiple remotes we are ignoring them, as we don't need to setup git repository as it is defined here,
	// the only thing that we need is to download the content

	if ps.CheckoutFrom != nil && ps.CheckoutFrom.Revision != "" {
		revision = ps.CheckoutFrom.Revision
	}
	if len(ps.Remotes) > 1 {
		if ps.CheckoutFrom == nil {
			err = fmt.Errorf("there are multiple git remotes but no checkoutFrom information")
			return "", "", "", err
		}
		remoteName = ps.CheckoutFrom.Remote
		if val, ok := ps.Remotes[remoteName]; ok {
			remoteURL = val
		} else {
			err = fmt.Errorf("checkoutFrom.Remote is not defined in Remotes")
			return "", "", "", err

		}
	} else {
		// there is only one remote, using range to get it as there are not indexes
		for name, url := range ps.Remotes {
			remoteName = name
			remoteURL = url
		}

	}

	return remoteName, remoteURL, revision, err

}

// GetProjectSourceType returns the source type of a given project source
func GetProjectSourceType(projectSrc v1.ProjectSource) (v1.ProjectSourceType, error) {
	switch {
	case projectSrc.Git != nil:
		return v1.GitProjectSourceType, nil
	case projectSrc.Zip != nil:
		return v1.ZipProjectSourceType, nil
	case projectSrc.Custom != nil:
		return v1.CustomProjectSourceType, nil

	default:
		return "", fmt.Errorf("unknown project source type")
	}
}
