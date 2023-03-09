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

package generator

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/pod-security-admission/api"
	psaapi "k8s.io/pod-security-admission/api"
	psapolicy "k8s.io/pod-security-admission/policy"
	"k8s.io/utils/pointer"
)

// ContainerVisitor is called with each container
type ContainerVisitor func(container *corev1.Container)

// visitContainers invokes the visitor function for every container in the given pod template spec
func visitContainers(podTemplateSpec *corev1.PodTemplateSpec, visitor ContainerVisitor) {
	for i := range podTemplateSpec.Spec.InitContainers {
		visitor(&podTemplateSpec.Spec.InitContainers[i])
	}
	for i := range podTemplateSpec.Spec.Containers {
		visitor(&podTemplateSpec.Spec.Containers[i])
	}
	for i := range podTemplateSpec.Spec.EphemeralContainers {
		visitor((*corev1.Container)(&podTemplateSpec.Spec.EphemeralContainers[i].EphemeralContainerCommon))
	}
}

func patchForPolicy(podTemplateSpec *corev1.PodTemplateSpec, policy psaapi.Policy) (*corev1.PodTemplateSpec, error) {
	evaluator, err := psapolicy.NewEvaluator(psapolicy.DefaultChecks())
	if err != nil {
		return nil, err
	}
	results := evaluator.EvaluatePod(policy.Enforce, &podTemplateSpec.ObjectMeta, &podTemplateSpec.Spec)
	for _, result := range results {
		if result.Allowed {
			continue
		}
		switch result.ForbiddenReason {
		case "allowPrivilegeEscalation != false":
			podTemplateSpec = patchAllowPrivilegeEscalation(podTemplateSpec)
		case "unrestricted capabilities":
			podTemplateSpec = patchUnrestrictedCapabilities(podTemplateSpec)
		case "runAsNonRoot != true":
			podTemplateSpec = patchRunAsNonRoot(podTemplateSpec)
		case "seccompProfile":
			podTemplateSpec = patchSeccompProfile(podTemplateSpec, policy.Enforce.Level)
			// Other policies are not implemented as they cannot be encountered with pods created by the library
		}
	}

	newResults := evaluator.EvaluatePod(policy.Enforce, &podTemplateSpec.ObjectMeta, &podTemplateSpec.Spec)
	for _, result := range newResults {
		if !result.Allowed {
			// This is an assertion for future developers, and should never happen in production
			// This could happen during development if some unsecure fields are added from `getPodTemplateSpec` or `GetContainers`/`GetInitContainers`
			return nil, fmt.Errorf("error patching pod for Pod Security Admission. The folowing policy is still not respected: %s (%s)", result.ForbiddenReason, result.ForbiddenDetail)
		}
	}

	return podTemplateSpec, nil
}

func patchAllowPrivilegeEscalation(podTemplateSpec *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
	visitContainers(podTemplateSpec, func(container *corev1.Container) {
		if container.SecurityContext == nil {
			container.SecurityContext = &corev1.SecurityContext{}
		}
		container.SecurityContext.AllowPrivilegeEscalation = pointer.Bool(false)
	})
	return podTemplateSpec
}

func patchUnrestrictedCapabilities(podTemplateSpec *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
	visitContainers(podTemplateSpec, func(container *corev1.Container) {
		if container.SecurityContext == nil {
			container.SecurityContext = &corev1.SecurityContext{}
		}
		if container.SecurityContext.Capabilities == nil {
			container.SecurityContext.Capabilities = &corev1.Capabilities{}
		}
		container.SecurityContext.Capabilities.Drop = append(container.SecurityContext.Capabilities.Drop, "ALL")
	})
	return podTemplateSpec
}

func patchRunAsNonRoot(podTemplateSpec *corev1.PodTemplateSpec) *corev1.PodTemplateSpec {
	if podTemplateSpec.Spec.SecurityContext == nil {
		podTemplateSpec.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}
	podTemplateSpec.Spec.SecurityContext.RunAsNonRoot = pointer.Bool(true)
	// No need to set the value as true for containers, as setting at the Pod level is sufficient
	return podTemplateSpec
}

func patchSeccompProfile(podTemplateSpec *corev1.PodTemplateSpec, level psaapi.Level) *corev1.PodTemplateSpec {
	if level == api.LevelRestricted {
		if podTemplateSpec.Spec.SecurityContext == nil {
			podTemplateSpec.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		if podTemplateSpec.Spec.SecurityContext.SeccompProfile == nil {
			podTemplateSpec.Spec.SecurityContext.SeccompProfile = &corev1.SeccompProfile{}
		}
		podTemplateSpec.Spec.SecurityContext.SeccompProfile.Type = "RuntimeDefault"
	} else if level == api.LevelBaseline {
		visitContainers(podTemplateSpec, func(container *corev1.Container) {
			if container.SecurityContext != nil && container.SecurityContext.SeccompProfile != nil && container.SecurityContext.SeccompProfile.Type == "Unconfined" {
				container.SecurityContext.SeccompProfile = nil
			}
		})
		if podTemplateSpec.Spec.SecurityContext != nil && podTemplateSpec.Spec.SecurityContext.SeccompProfile != nil && podTemplateSpec.Spec.SecurityContext.SeccompProfile.Type == "Unconfined" {
			podTemplateSpec.Spec.SecurityContext.SeccompProfile = nil
		}
	}
	return podTemplateSpec
}
