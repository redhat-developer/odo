/*
Copyright 2019 The Tekton Authors

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

package test

import "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"

// CompareParams can be used with comparison options such as `cmpopts.SortSlices`
// so we can ignore the ordering of params.
func CompareParams(x, y v1alpha1.Param) bool {
	return x.Name < y.Name
}
