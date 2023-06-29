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

package parser

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	apiAttributes "github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"sigs.k8s.io/yaml"

	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// WriteYamlDevfile creates a devfile.yaml file
func (d *DevfileObj) WriteYamlDevfile() error {

	// Check kubernetes components, and restore original uri content
	if d.Ctx.GetConvertUriToInlined() {
		err := restoreK8sCompURI(d)
		if err != nil {
			return errors.Wrapf(err, "failed to restore kubernetes component uri field")
		}
	}
	// Encode data into YAML format
	yamlData, err := yaml.Marshal(d.Data)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal devfile object into yaml")
	}
	// Write to devfile.yaml
	fs := d.Ctx.GetFs()
	if fs == nil {
		fs = filesystem.DefaultFs{}
	}
	err = fs.WriteFile(d.Ctx.GetAbsPath(), yamlData, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to create devfile yaml file")
	}

	// Successful
	klog.V(2).Infof("devfile yaml created at: '%s'", OutputDevfileYamlPath)
	return nil
}

func restoreK8sCompURI(devObj *DevfileObj) error {
	getKubeCompOptions := common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1.KubernetesComponentType,
		},
	}
	getOpenshiftCompOptions := common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1.OpenshiftComponentType,
		},
	}
	kubeComponents, err := devObj.Data.GetComponents(getKubeCompOptions)
	if err != nil {
		return err
	}
	openshiftComponents, err := devObj.Data.GetComponents(getOpenshiftCompOptions)
	if err != nil {
		return err
	}

	for _, kubeComp := range kubeComponents {
		uri := kubeComp.Attributes.GetString(K8sLikeComponentOriginalURIKey, &err)
		if err != nil {
			if _, ok := err.(*apiAttributes.KeyNotFoundError); !ok {
				return err
			}
		}
		if uri != "" {
			kubeComp.Kubernetes.Uri = uri
			kubeComp.Kubernetes.Inlined = ""
			delete(kubeComp.Attributes, K8sLikeComponentOriginalURIKey)
			err = devObj.Data.UpdateComponent(kubeComp)
			if err != nil {
				return err
			}
		}
	}

	for _, openshiftComp := range openshiftComponents {
		uri := openshiftComp.Attributes.GetString(K8sLikeComponentOriginalURIKey, &err)
		if err != nil {
			if _, ok := err.(*apiAttributes.KeyNotFoundError); !ok {
				return err
			}
		}
		if uri != "" {
			openshiftComp.Openshift.Uri = uri
			openshiftComp.Openshift.Inlined = ""
			delete(openshiftComp.Attributes, K8sLikeComponentOriginalURIKey)
			err = devObj.Data.UpdateComponent(openshiftComp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
