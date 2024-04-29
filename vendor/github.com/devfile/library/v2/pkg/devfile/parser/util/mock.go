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
	"net/http"
	"os"

	"github.com/devfile/library/v2/pkg/util"
)

// Default filenames for create devfile to be used in mocks
const (
	OutputDevfileYamlPath = "devfile.yaml"
)

type MockDevfileUtilsClient struct {
	// Specify a valid git URL as an alias if using a localhost HTTP server in order to pass validation.
	ParentURLAlias string

	// MockGitUrl struct for mocking git related ops
	MockGitURL util.MockGitUrl

	// Mock Git token.  Specify the string "valid-token" for the mock CloneGitRepo to pass
	GitTestToken string

	// Options to specify what file download needs to be mocked
	DownloadOptions util.MockDownloadOptions
}

func NewMockDevfileUtilsClient() MockDevfileUtilsClient {
	return MockDevfileUtilsClient{}
}

func (gc *MockDevfileUtilsClient) DownloadInMemory(params util.HTTPRequestParams) ([]byte, error) {
	var httpClient = &http.Client{Transport: &http.Transport{
		ResponseHeaderTimeout: util.HTTPRequestResponseTimeout,
	}, Timeout: util.HTTPRequestResponseTimeout}

	if gc.MockGitURL.Host != "" {
		if util.IsGitProviderRepo(gc.MockGitURL.Host) {
			gc.MockGitURL.Token = gc.GitTestToken
		}
	} else if params.URL != "" {
		// Not all clients have the ability to pass in mock data
		// So we should be adaptable and use the function params
		// and mock the output
		if util.IsGitProviderRepo(params.URL) {
			gc.MockGitURL.Host = params.URL
			gc.MockGitURL.Token = params.Token
		}
	}

	if gc.DownloadOptions.MockParent == nil {
		gc.DownloadOptions.MockParent = &util.MockParent{}
	}

	file, err := gc.MockGitURL.DownloadInMemoryWithClient(params, httpClient, gc.DownloadOptions)

	if gc.DownloadOptions.MockParent != nil && gc.DownloadOptions.MockParent.IsMainDevfileDownloaded && gc.DownloadOptions.MockParent.IsParentDevfileDownloaded {
		// Since gc is a pointer, if both the main and parent devfiles are downloaded, reset the flag.
		// So that other tests can use the Mock Parent Devfile download if required.
		gc.DownloadOptions.MockParent.IsMainDevfileDownloaded = false
		gc.DownloadOptions.MockParent.IsParentDevfileDownloaded = false
	}

	if gc.MockGitURL.Host != "" && params.URL != "" {
		// Since gc is a pointer, reset the mock data if both the URL and Host are present
		gc.MockGitURL.Host = ""
		gc.MockGitURL.Token = ""
	}

	return file, err
}

func (gc MockDevfileUtilsClient) DownloadGitRepoResources(url string, destDir string, token string) error {

	// if mock data is unavailable as certain clients cant provide mock data
	// then adapt and create mock data from actual params
	if gc.ParentURLAlias == "" {
		gc.ParentURLAlias = url
		gc.MockGitURL.IsFile = true
		gc.MockGitURL.Revision = "main"
		gc.MockGitURL.Path = OutputDevfileYamlPath
		gc.MockGitURL.Host = "github.com"
		gc.MockGitURL.Protocol = "https"
		gc.MockGitURL.Owner = "devfile"
		gc.MockGitURL.Repo = "library"
	}

	if gc.GitTestToken == "" {
		gc.GitTestToken = token
	}

	//the url parameter that gets passed in will be the localhost IP of the test server, so it will fail all the validation checks.  We will use the global testURL variable instead
	//skip the Git Provider check since it'll fail
	if util.IsGitProviderRepo(gc.ParentURLAlias) {
		// this converts the test git URL to a mock URL
		mockGitUrl := gc.MockGitURL
		mockGitUrl.Token = gc.GitTestToken

		if !mockGitUrl.IsFile || mockGitUrl.Revision == "" || !ValidateDevfileExistence((mockGitUrl.Path)) {
			return fmt.Errorf("error getting devfile from url: failed to retrieve %s", url+"/"+mockGitUrl.Path)
		}

		stackDir, err := os.MkdirTemp("", "git-resources")
		if err != nil {
			return fmt.Errorf("failed to create dir: %s, error: %v", stackDir, err)
		}

		defer func(path string) {
			err = os.RemoveAll(path)
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
		return fmt.Errorf("failed to download resources from parent devfile.  Unsupported Git Provider for %s ", gc.ParentURLAlias)
	}

	return nil
}
