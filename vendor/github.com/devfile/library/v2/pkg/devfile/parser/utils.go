//
// Copyright 2022-2023 Red Hat, Inc.
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
	"github.com/hashicorp/go-multierror"
	"os"
	"path"
	"reflect"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
)

type DevfileUtilsClient struct {
}

func NewDevfileUtilsClient() DevfileUtilsClient {
	return DevfileUtilsClient{}
}

type DevfileUtils interface {
	DownloadGitRepoResources(url string, destDir string, token string) error
}

// DownloadGitRepoResources mock implementation of the real method.
func (gc DevfileUtilsClient) DownloadGitRepoResources(url string, destDir string, token string) error {
	var returnedErr error
	if util.IsGitProviderRepo(url) {
		gitUrl, err := util.NewGitURL(url, token)
		if err != nil {
			return err
		}

		if !gitUrl.IsFile || gitUrl.Revision == "" || !strings.Contains(gitUrl.Path, OutputDevfileYamlPath) {
			return fmt.Errorf("error getting devfile from url: failed to retrieve %s", url)
		}

		stackDir, err := os.MkdirTemp("", fmt.Sprintf("git-resources"))
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
		return fmt.Errorf("Failed to download resources from parent devfile.  Unsupported Git Provider for %s ", url)
	}

	return nil
}

// GetDeployComponents gets the default deploy command associated components
func GetDeployComponents(devfileData data.DevfileData) (map[string]string, error) {
	deployCommandFilter := common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandGroupKind: devfilev1.DeployCommandGroupKind,
		},
	}
	deployCommands, err := devfileData.GetCommands(deployCommandFilter)
	if err != nil {
		return nil, err
	}

	deployAssociatedComponents := make(map[string]string)
	var deployAssociatedSubCommands []string

	for _, command := range deployCommands {
		if command.Apply != nil {
			if len(deployCommands) > 1 && command.Apply.Group.IsDefault != nil && !*command.Apply.Group.IsDefault {
				continue
			}
			deployAssociatedComponents[command.Apply.Component] = command.Apply.Component
		} else if command.Composite != nil {
			if len(deployCommands) > 1 && command.Composite.Group.IsDefault != nil && !*command.Composite.Group.IsDefault {
				continue
			}
			deployAssociatedSubCommands = append(deployAssociatedSubCommands, command.Composite.Commands...)
		}
	}

	applyCommandFilter := common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: devfilev1.ApplyCommandType,
		},
	}
	applyCommands, err := devfileData.GetCommands(applyCommandFilter)
	if err != nil {
		return nil, err
	}

	for _, command := range applyCommands {
		if command.Apply != nil {
			for _, deployCommand := range deployAssociatedSubCommands {
				if deployCommand == command.Id {
					deployAssociatedComponents[command.Apply.Component] = command.Apply.Component
				}
			}

		}
	}

	return deployAssociatedComponents, nil
}

// GetImageBuildComponent gets the image build component from the deploy associated components
func GetImageBuildComponent(devfileData data.DevfileData, deployAssociatedComponents map[string]string) (devfilev1.Component, error) {
	imageComponentFilter := common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: devfilev1.ImageComponentType,
		},
	}

	imageComponents, err := devfileData.GetComponents(imageComponentFilter)
	if err != nil {
		return devfilev1.Component{}, err
	}

	var imageBuildComponent devfilev1.Component
	for _, component := range imageComponents {
		if _, ok := deployAssociatedComponents[component.Name]; ok && component.Image != nil {
			if reflect.DeepEqual(imageBuildComponent, devfilev1.Component{}) {
				imageBuildComponent = component
			} else {
				errMsg := "expected to find one devfile image component with a deploy command for build. Currently there is more than one image component"
				return devfilev1.Component{}, fmt.Errorf(errMsg)
			}
		}
	}

	// If there is not one image component defined in the deploy command, err out
	if reflect.DeepEqual(imageBuildComponent, devfilev1.Component{}) {
		errMsg := "expected to find one devfile image component with a deploy command for build. Currently there is no image component"
		return devfilev1.Component{}, fmt.Errorf(errMsg)
	}

	return imageBuildComponent, nil
}
