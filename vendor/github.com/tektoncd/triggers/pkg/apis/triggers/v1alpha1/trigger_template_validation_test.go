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

package v1alpha1_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	b "github.com/tektoncd/triggers/test/builder"

	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
)

var simpleResourceTemplate = runtime.RawExtension{
	Raw: []byte(`{"kind":"PipelineRun","apiVersion":"tekton.dev/v1alpha1","metadata":{"creationTimestamp":null},"spec":{},"status":{}}`),
}
var v1beta1ResourceTemplate = runtime.RawExtension{
	Raw: []byte(`{"kind":"PipelineRun","apiVersion":"tekton.dev/v1beta1","metadata":{"creationTimestamp":null},"spec":{},"status":{}}`),
}
var paramResourceTemplate = runtime.RawExtension{
	Raw: []byte(`{"kind":"PipelineRun","apiVersion":"tekton.dev/v1alpha1","metadata":{"creationTimestamp":null},"spec": "$(params.foo)","status":{}}`),
}

func TestTriggerTemplate_Validate(t *testing.T) {
	tcs := []struct {
		name     string
		template *v1alpha1.TriggerTemplate
		want     *apis.FieldError
	}{
		{
			name: "invalid objectmetadata, name with dot",
			template: b.TriggerTemplate("t.t", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"),
				b.TriggerResourceTemplate(simpleResourceTemplate))),
			want: &apis.FieldError{
				Message: "Invalid resource name: special character . must not be present",
				Paths:   []string{"metadata.name"},
			},
		},
		{
			name: "invalid objectmetadata, name too long",
			template: b.TriggerTemplate(
				"ttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt",
				"foo", b.TriggerTemplateSpec(
					b.TriggerTemplateParam("foo", "desc", "val"),
					b.TriggerResourceTemplate(simpleResourceTemplate))),
			want: &apis.FieldError{
				Message: "Invalid resource name: length must be no more than 63 characters",
				Paths:   []string{"metadata.name"},
			},
		},
		{
			name: "valid template",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"),
				b.TriggerResourceTemplate(simpleResourceTemplate))),
			want: nil,
		}, {
			name: "valid v1beta1 template",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"),
				b.TriggerResourceTemplate(v1beta1ResourceTemplate))),
			want: nil,
		}, {
			name: "missing resource template",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"))),
			want: &apis.FieldError{
				Message: "missing field(s)",
				Paths:   []string{"spec.resourcetemplates"},
			},
		}, {
			name: "resource template missing kind",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"),
				b.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"apiVersion": "foo"}`)}))),
			want: &apis.FieldError{
				Message: "missing field(s)",
				Paths:   []string{"spec.resourcetemplates[0].kind"},
			},
		}, {
			name: "resource template missing apiVersion",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"),
				b.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"kind": "foo"}`)}))),
			want: &apis.FieldError{
				Message: "missing field(s)",
				Paths:   []string{"spec.resourcetemplates[0].apiVersion"},
			},
		}, {
			name: "resource template invalid apiVersion",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"),
				b.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"kind": "pipelinerun", "apiVersion": "foobar"}`)}))),
			want: &apis.FieldError{
				Message: `invalid value: no kind "pipelinerun" is registered for version "foobar"`,
				Paths:   []string{"spec.resourcetemplates[0]"},
			},
		}, {
			name: "resource template invalid kind",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"),
				b.TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"kind": "tekton.dev/v1alpha1", "apiVersion": "foo"}`)}))),
			want: &apis.FieldError{
				Message: `invalid value: no kind "tekton.dev/v1alpha1" is registered for version "foo"`,
				Paths:   []string{"spec.resourcetemplates[0]"},
			},
		}, {
			name: "params used in resource template are declared",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerTemplateParam("foo", "desc", "val"),
				b.TriggerResourceTemplate(paramResourceTemplate))),
			want: nil,
		}, {
			name: "params used in resource template are not declared",
			template: b.TriggerTemplate("tt", "foo", b.TriggerTemplateSpec(
				b.TriggerResourceTemplate(paramResourceTemplate))),
			want: &apis.FieldError{
				Message: "invalid value: undeclared param '$(params.foo)'",
				Paths:   []string{"spec.resourcetemplates[0]"},
				Details: "'$(params.foo)' must be declared in spec.params",
			},
		}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.template.Validate(context.Background())
			if d := cmp.Diff(got, tc.want, cmpopts.IgnoreUnexported(apis.FieldError{})); d != "" {
				t.Errorf("TriggerTemplate Validation failed: %s", d)
			}
		})
	}
}
