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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// FakeResourceRequirements creates a fake resource requirements from cpu and memory
func FakeResourceRequirements(cpu, memory string) (corev1.ResourceRequirements, error) {
	var resReq corev1.ResourceRequirements

	limits := make(corev1.ResourceList)
	var err error
	limits[corev1.ResourceCPU], err = resource.ParseQuantity(cpu)
	if err != nil {
		return resReq, err
	}
	limits[corev1.ResourceMemory], err = resource.ParseQuantity(memory)
	if err != nil {
		return resReq, err
	}
	resReq.Limits = limits

	requests := make(corev1.ResourceList)
	requests[corev1.ResourceCPU], err = resource.ParseQuantity(cpu)
	if err != nil {
		return resReq, err
	}
	requests[corev1.ResourceMemory], err = resource.ParseQuantity(memory)
	if err != nil {
		return resReq, err
	}

	resReq.Requests = requests

	return resReq, nil
}
