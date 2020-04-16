// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

import "sigs.k8s.io/kustomize/pkg/gvk"

// PatchTarget represents the kubernetes object that the patch is applied to
type PatchTarget struct {
	gvk.Gvk   `json:",inline,omitempty" yaml:",inline,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name      string `json:"name" yaml:"name"`
}
