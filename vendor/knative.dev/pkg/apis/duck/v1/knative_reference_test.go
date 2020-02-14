/*
Copyright 2020 The Knative Authors

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

package v1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestValidate(t *testing.T) {
	ctx := context.Background()

	validRef := KReference{
		Kind:       kind,
		APIVersion: apiVersion,
		Name:       name,
		Namespace:  namespace,
	}

	tests := map[string]struct {
		ref  *KReference
		ctx  context.Context
		want *apis.FieldError
	}{
		"nil valid": {
			ref:  nil,
			ctx:  ctx,
			want: apis.ErrMissingField("name", "kind", "apiVersion"),
		},
		"valid ref": {
			ref:  &validRef,
			ctx:  ctx,
			want: nil,
		},
		"invalid ref, empty": {
			ref:  &KReference{},
			ctx:  ctx,
			want: apis.ErrMissingField("name", "kind", "apiVersion"),
		},
		"invalid ref, missing kind": {
			ref: &KReference{
				Namespace:  namespace,
				Name:       name,
				APIVersion: apiVersion,
			},
			ctx:  ctx,
			want: apis.ErrMissingField("kind"),
		},
		"invalid ref, missing api version": {
			ref: &KReference{
				Namespace: namespace,
				Name:      name,
				Kind:      kind,
			},
			ctx:  ctx,
			want: apis.ErrMissingField("apiVersion"),
		},
		"invalid ref, mismatched namespaces": {
			ref: &KReference{
				Namespace:  namespace,
				Name:       name,
				Kind:       kind,
				APIVersion: apiVersion,
			},
			ctx: apis.WithinParent(ctx, metav1.ObjectMeta{Namespace: "diffns"}),
			want: &apis.FieldError{
				Message: "mismatched namespaces",
				Paths:   []string{"namespace"},
				Details: `parent namespace: "diffns" does not match ref: "b-namespace"`,
			},
		},
		"valid ref, mismatched namespaces, but overridden": {
			ref: &KReference{
				Namespace:  namespace,
				Name:       name,
				Kind:       kind,
				APIVersion: apiVersion,
			},
			ctx: apis.AllowDifferentNamespace(apis.WithinParent(ctx, metav1.ObjectMeta{Namespace: "diffns"})),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotErr := tc.ref.Validate(tc.ctx)

			if tc.want != nil {
				if diff := cmp.Diff(tc.want.Error(), gotErr.Error()); diff != "" {
					t.Errorf("Got: %v wanted %v", gotErr, tc.want)
				}
			} else if gotErr != nil {
				t.Errorf("Validate() = %v, wanted nil", gotErr)
			}
		})
	}
}

func TestKReferenceSetDefaults(t *testing.T) {
	ctx := context.Background()

	const parentNamespace = "parentNamespace"

	tests := map[string]struct {
		ref  *KReference
		ctx  context.Context
		want string
	}{
		"namespace set, nothing in context, not modified ": {
			ref:  &KReference{Namespace: namespace},
			ctx:  ctx,
			want: namespace,
		},
		"namespace set, context set, not modified ": {
			ref:  &KReference{Namespace: namespace},
			ctx:  apis.WithinParent(ctx, metav1.ObjectMeta{Namespace: parentNamespace}),
			want: namespace,
		},
		"namespace not set, context set, defaulted": {
			ref:  &KReference{},
			ctx:  apis.WithinParent(ctx, metav1.ObjectMeta{Namespace: parentNamespace}),
			want: parentNamespace,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.ref.SetDefaults(tc.ctx)
			if tc.ref.Namespace != tc.want {
				t.Errorf("Namespace = %s; want: %s", tc.ref.Namespace, tc.want)
			}
		})
	}
}
