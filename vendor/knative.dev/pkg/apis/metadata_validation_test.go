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

package apis

import (
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateObjectMetadata(t *testing.T) {
	cases := []struct {
		name       string
		objectMeta metav1.Object
		expectErr  error
	}{{
		name: "invalid name - dots",
		objectMeta: &metav1.ObjectMeta{
			Name: "do.not.use.dots",
		},
		expectErr: &FieldError{
			Message: "not a DNS 1035 label: [a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')]",
			Paths:   []string{"name"},
		},
	}, {
		name: "invalid name - too long",
		objectMeta: &metav1.ObjectMeta{
			Name: strings.Repeat("a", 64),
		},
		expectErr: &FieldError{
			Message: "not a DNS 1035 label: [must be no more than 63 characters]",
			Paths:   []string{"name"},
		},
	}, {
		name: "invalid name - trailing dash",
		objectMeta: &metav1.ObjectMeta{
			Name: "some-name-",
		},
		expectErr: &FieldError{
			Message: "not a DNS 1035 label: [a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')]",
			Paths:   []string{"name"},
		},
	}, {
		name: "valid generateName",
		objectMeta: &metav1.ObjectMeta{
			GenerateName: "some-name",
		},
		expectErr: (*FieldError)(nil),
	}, {
		name: "valid generateName - trailing dash",
		objectMeta: &metav1.ObjectMeta{
			GenerateName: "some-name-",
		},
		expectErr: (*FieldError)(nil),
	}, {
		name: "invalid generateName - dots",
		objectMeta: &metav1.ObjectMeta{
			GenerateName: "do.not.use.dots",
		},
		expectErr: &FieldError{
			Message: "not a DNS 1035 label prefix: [a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')]",
			Paths:   []string{"generateName"},
		},
	}, {
		name: "invalid generateName - too long",
		objectMeta: &metav1.ObjectMeta{
			GenerateName: strings.Repeat("a", 64),
		},
		expectErr: &FieldError{
			Message: "not a DNS 1035 label prefix: [must be no more than 63 characters]",
			Paths:   []string{"generateName"},
		},
	}, {
		name:       "missing name and generateName",
		objectMeta: &metav1.ObjectMeta{},
		expectErr: &FieldError{
			Message: "name or generateName is required",
			Paths:   []string{"name"},
		},
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			err := ValidateObjectMetadata(c.objectMeta)

			if !reflect.DeepEqual(c.expectErr, err) {
				t.Errorf("Expected: '%#v', Got: '%#v'", c.expectErr, err)
			}
		})
	}
}

type WithPod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              corev1.PodSpec `json:"spec,omitempty"`
}

func getSpec(image string) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{{
			Image: image,
		}},
	}
}

func getAnnotation(groupName, suffix, user string) map[string]string {
	return map[string]string{
		groupName + suffix: user,
	}
}

func TestServiceAnnotationUpdate(t *testing.T) {
	const (
		u1        = "oveja@knative.dev"
		u2        = "cabra@knative.dev"
		groupName = "pkg.knative.dev"
	)
	tests := []struct {
		name          string
		prev          *WithPod
		this          *WithPod
		oldAnnotation map[string]string
		newAnnotation map[string]string
		want          *FieldError
	}{{
		name:          "update creator annotation",
		prev:          nil,
		this:          nil,
		oldAnnotation: getAnnotation(groupName, CreatorAnnotationSuffix, u1),
		newAnnotation: getAnnotation(groupName, CreatorAnnotationSuffix, u2),
		want: &FieldError{Message: "annotation value is immutable",
			Paths: []string{groupName + CreatorAnnotationSuffix}},
	}, {
		name:          "update lastModifier without spec changes",
		prev:          nil,
		this:          nil,
		oldAnnotation: getAnnotation(groupName, UpdaterAnnotationSuffix, u1),
		newAnnotation: getAnnotation(groupName, UpdaterAnnotationSuffix, u2),
		want:          ErrInvalidValue(u2, groupName+UpdaterAnnotationSuffix),
	}, {
		name: "update lastModifier with spec changes",
		prev: &WithPod{
			Spec: getSpec("new-image"),
		},
		this: &WithPod{
			Spec: getSpec("old-image"),
		},
		oldAnnotation: getAnnotation(groupName, UpdaterAnnotationSuffix, u1),
		newAnnotation: getAnnotation(groupName, UpdaterAnnotationSuffix, u2),
		want:          nil,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateCreatorAndModifier(test.prev, test.this, test.oldAnnotation, test.newAnnotation, groupName)
			if !reflect.DeepEqual(test.want.Error(), err.Error()) {
				t.Errorf("Expected: '%#v', Got: '%#v'", test.want, err)
			}
		})
	}
}
