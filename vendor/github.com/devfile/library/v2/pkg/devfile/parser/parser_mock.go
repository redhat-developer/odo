//
// Copyright 2023 Red Hat, Inc.
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

package parser

import (
	"fmt"
	"github.com/devfile/library/v2/pkg/util"
	"os"
	"strings"
)

type MockDevfileUtilsClient struct {
	ParentURLAlias string // Specify a valid git URL as an alias if using a localhost HTTP server in order to pass validation.
	MockGitURL     util.MockGitUrl
	GitTestToken   string // Mock Git token.  Specify the string "valid-token" for the mock CloneGitRepo to pass
}

func NewMockDevfileUtilsClient() MockDevfileUtilsClient {
	return MockDevfileUtilsClient{}
}

func (gc MockDevfileUtilsClient) DownloadGitRepoResources(url string, destDir string, token string) error {

	//the url parameter that gets passed in will be the localhost IP of the test server, so it will fail all the validation checks.  We will use the global testURL variable instead
	//skip the Git Provider check since it'll fail
	if util.IsGitProviderRepo(gc.ParentURLAlias) {
		// this converts the test git URL to a mock URL
		mockGitUrl := gc.MockGitURL
		mockGitUrl.Token = gc.GitTestToken

		if !mockGitUrl.IsFile || mockGitUrl.Revision == "" || !strings.Contains(mockGitUrl.Path, OutputDevfileYamlPath) {
			return fmt.Errorf("error getting devfile from url: failed to retrieve %s", url+"/"+mockGitUrl.Path)
		}

		stackDir, err := os.MkdirTemp("", fmt.Sprintf("git-resources"))
		if err != nil {
			return fmt.Errorf("failed to create dir: %s, error: %v", stackDir, err)
		}

		defer func(path string) {
			err := os.RemoveAll(path)
			if err != nil {
				err = fmt.Errorf("failed to create dir: %s, error: %v", stackDir, err)
			}
		}(stackDir)

		err = mockGitUrl.CloneGitRepo(stackDir)
		if err != nil {
			return err
		}

		err = util.CopyAllDirFiles(stackDir, destDir)
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("Failed to download resources from parent devfile.  Unsupported Git Provider for %s ", gc.ParentURLAlias)
	}

	return nil
}
