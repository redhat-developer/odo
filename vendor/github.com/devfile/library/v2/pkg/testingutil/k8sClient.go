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

package testingutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeK8sClient struct {
	client.Client         // To satisfy interface; override all used methods
	DevWorkspaceResources map[string]v1alpha2.DevWorkspaceTemplate
	Errors                map[string]string
}

func (client *FakeK8sClient) Get(_ context.Context, namespacedName client.ObjectKey, obj client.Object) error {
	template, ok := obj.(*v1alpha2.DevWorkspaceTemplate)
	if !ok {
		return fmt.Errorf("called Get() in fake client with non-DevWorkspaceTemplate")
	}
	if element, ok := client.DevWorkspaceResources[namespacedName.Name]; ok {
		*template = element
		return nil
	}

	if err, ok := client.Errors[namespacedName.Name]; ok {
		return errors.New(err)
	}
	return fmt.Errorf("test does not define an entry for %s", namespacedName.Name)
}
