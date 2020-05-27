/*
Copyright 2019 The Knative Authors

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

package tracker

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

func TestGroupVersionKind(t *testing.T) {
	ref := Reference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespace:  "default",
		Name:       "nginx",
	}

	want := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}
	got := ref.GroupVersionKind()
	if want != got {
		t.Errorf("GroupVersionKind() = %v, wanted = %v", got, want)
	}
}

func TestValidateObjectReference(t *testing.T) {
	tests := []struct {
		name string
		ref  Reference
		want *apis.FieldError
	}{{
		name: "empty reference",
		want: apis.ErrMissingField("apiVersion", "kind", "name", "namespace"),
	}, {
		name: "good reference",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "default",
			Name:       "nginx",
		},
	}, {
		name: "another good reference",
		ref: Reference{
			APIVersion: "v1",
			Kind:       "Service",
			Namespace:  "default",
			Name:       "nginx",
		},
	}, {
		name: "bad apiVersion",
		ref: Reference{
			APIVersion: "a b c d", // Bad!
			Kind:       "Service",
			Namespace:  "default",
			Name:       "nginx",
		},
		want: apis.ErrInvalidValue("name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')", "apiVersion"),
	}, {
		name: "bad kind",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "a b c d", // Bad!
			Namespace:  "default",
			Name:       "nginx",
		},
		want: apis.ErrInvalidValue("a valid C identifier must start with alphabetic character or '_', followed by a string of alphanumeric characters or '_' (e.g. 'my_name',  or 'MY_NAME',  or 'MyName', regex used for validation is '[A-Za-z_][A-Za-z0-9_]*')", "kind"),
	}, {
		name: "bad namespace and name",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "a.b",
			Name:       "c.d",
		},
		want: &apis.FieldError{
			Message: "invalid value: a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')",
			Paths:   []string{"namespace", "name"},
		},
	}, {
		name: "with selector too",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "default",
			Name:       "nginx",
			Selector:   &metav1.LabelSelector{},
		},
		want: apis.ErrDisallowedFields("selector"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ref.ValidateObjectReference(context.Background())
			if (test.want != nil) != (got != nil) {
				t.Errorf("ValidateObjectReference() = %v, wanted %v", got, test.want)
			} else if test.want != nil {
				want, got := test.want.Error(), got.Error()
				if got != want {
					t.Errorf("ValidateObjectReference() = %s, wanted %s", got, want)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		ref  Reference
		want *apis.FieldError
	}{{
		name: "empty reference",
		want: apis.ErrMissingField("apiVersion", "kind", "namespace").Also(
			apis.ErrMissingOneOf("name", "selector")),
	}, {
		name: "good reference",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "default",
			Name:       "nginx",
		},
	}, {
		name: "another good reference",
		ref: Reference{
			APIVersion: "v1",
			Kind:       "Service",
			Namespace:  "default",
			Name:       "nginx",
		},
	}, {
		name: "bad apiVersion",
		ref: Reference{
			APIVersion: "a b c d", // Bad!
			Kind:       "Service",
			Namespace:  "default",
			Name:       "nginx",
		},
		want: apis.ErrInvalidValue("name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')", "apiVersion"),
	}, {
		name: "bad kind",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "a b c d", // Bad!
			Namespace:  "default",
			Name:       "nginx",
		},
		want: apis.ErrInvalidValue("a valid C identifier must start with alphabetic character or '_', followed by a string of alphanumeric characters or '_' (e.g. 'my_name',  or 'MY_NAME',  or 'MyName', regex used for validation is '[A-Za-z_][A-Za-z0-9_]*')", "kind"),
	}, {
		name: "bad namespace and name",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "a.b",
			Name:       "c.d",
		},
		want: &apis.FieldError{
			Message: "invalid value: a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')",
			Paths:   []string{"namespace", "name"},
		},
	}, {
		name: "with selector too",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "default",
			Name:       "nginx",
			Selector:   &metav1.LabelSelector{},
		},
		want: apis.ErrMultipleOneOf("name", "selector"),
	}, {
		name: "with just selector",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "default",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
		},
	}, {
		name: "with invalid selector",
		ref: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "default",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"a b c": "bar",
				},
			},
		},
		want: apis.ErrInvalidValue(`invalid label key "a b c": name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')`, "selector"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ref.Validate(context.Background())
			if (test.want != nil) != (got != nil) {
				t.Errorf("ValidateObjectReference() = %v, wanted %v", got, test.want)
			} else if test.want != nil {
				want, got := test.want.Error(), got.Error()
				if got != want {
					t.Errorf("ValidateObjectReference() = %s, wanted %s", got, want)
				}
			}
		})
	}
}
