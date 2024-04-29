//
// Copyright Red Hat
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

package util

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/devfile/library/v2/pkg/util"
	"github.com/hashicorp/go-multierror"
)

// Contains common naming conventions for devfiles to look for when downloading resources
var DevfilePossibilities = [...]string{"devfile.yaml", ".devfile.yaml", "devfile.yml", ".devfile.yml"}

type DevfileUtilsClient struct {
}

func NewDevfileUtilsClient() DevfileUtilsClient {
	return DevfileUtilsClient{}
}

// DownloadInMemory is a wrapper to the util.DownloadInMemory() call.
// This is done to help devfile/library clients invoke this function with a client.
func (c DevfileUtilsClient) DownloadInMemory(params util.HTTPRequestParams) ([]byte, error) {
	return util.DownloadInMemory(params)
}

// DownloadGitRepoResources downloads the git repository resources
func (c DevfileUtilsClient) DownloadGitRepoResources(url string, destDir string, token string) error {
	var returnedErr error
	if util.IsGitProviderRepo(url) {
		gitUrl, err := util.NewGitURL(url, token)
		if err != nil {
			return err
		}

		if !gitUrl.IsFile || gitUrl.Revision == "" || !ValidateDevfileExistence((gitUrl.Path)) {
			return fmt.Errorf("error getting devfile from url: failed to retrieve %s", url)
		}

		stackDir, err := os.MkdirTemp("", "git-resources")
		if err != nil {
			return fmt.Errorf("failed to create dir: %s, error: %v", stackDir, err)
		}

		defer func(path string) {
			err := os.RemoveAll(path)
			if err != nil {
				returnedErr = multierror.Append(returnedErr, err)
			}
		}(stackDir)

		gitUrl.Token = token

		err = gitUrl.CloneGitRepo(stackDir)
		if err != nil {
			returnedErr = multierror.Append(returnedErr, err)
			return returnedErr
		}

		dir := path.Dir(path.Join(stackDir, gitUrl.Path))
		err = util.CopyAllDirFiles(dir, destDir)
		if err != nil {
			returnedErr = multierror.Append(returnedErr, err)
			return returnedErr
		}
	} else {
		return fmt.Errorf("failed to download resources from parent devfile.  Unsupported Git Provider for %s ", url)
	}

	return nil
}

// ValidateDevfileExistence verifies if any of the naming possibilities for devfile are present in the url path
func ValidateDevfileExistence(path string) bool {
	for _, devfile := range DevfilePossibilities {
		if strings.Contains(path, devfile) {
			return true
		}
	}
	return false
}
