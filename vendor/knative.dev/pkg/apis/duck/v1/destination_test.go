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

package v1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

const (
	kind       = "SomeKind"
	apiVersion = "v1mega1"
	name       = "a-name"
	namespace  = "b-namespace"
)

func TestValidateDestination(t *testing.T) {
	ctx := context.Background()

	validRef := KReference{
		Kind:       kind,
		APIVersion: apiVersion,
		Name:       name,
		Namespace:  namespace,
	}

	validURL := apis.URL{
		Scheme: "http",
		Host:   "host",
	}

	tests := map[string]struct {
		dest *Destination
		want string
	}{"nil valid": {
		dest: nil,
	}, "valid ref": {
		dest: &Destination{
			Ref: &validRef,
		},
	}, "invalid ref, missing name": {
		dest: &Destination{
			Ref: &KReference{
				Namespace:  namespace,
				Kind:       kind,
				APIVersion: apiVersion,
			},
		},
		want: "missing field(s): ref.name",
	}, "invalid ref, missing api version": {
		dest: &Destination{
			Ref: &KReference{
				Namespace: namespace,
				Kind:      kind,
				Name:      name,
			},
		},
		want: "missing field(s): ref.apiVersion",
	}, "invalid ref, missing kind": {
		dest: &Destination{
			Ref: &KReference{
				Namespace:  namespace,
				APIVersion: apiVersion,
				Name:       name,
			},
		},
		want: "missing field(s): ref.kind",
	}, "valid uri": {
		dest: &Destination{
			URI: &validURL,
		},
	}, "invalid, uri has no host": {
		dest: &Destination{
			URI: &apis.URL{
				Scheme: "http",
			},
		},
		want: "invalid value: Relative URI is not allowed when Ref and [apiVersion, kind, name] is absent: uri",
	}, "invalid, uri is not absolute url": {
		dest: &Destination{
			URI: &apis.URL{
				Host: "host",
			},
		},
		want: "invalid value: Relative URI is not allowed when Ref and [apiVersion, kind, name] is absent: uri",
	}, "invalid, both uri and ref, uri is absolute URL": {
		dest: &Destination{
			URI: &validURL,
			Ref: &validRef,
		},
		want: "Absolute URI is not allowed when Ref or [apiVersion, kind, name] is present: [apiVersion, kind, name], ref, uri",
	}, "invalid, both ref, [apiVersion, kind, name] and uri  are nil": {
		dest: &Destination{},
		want: "expected at least one, got none: ref, uri",
	}, "valid, both uri and ref, uri is not a absolute URL": {
		dest: &Destination{
			URI: &apis.URL{
				Path: "/handler",
			},
			Ref: &validRef,
		},
	}}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotErr := tc.dest.Validate(ctx)

			if tc.want != "" {
				if got, want := gotErr.Error(), tc.want; got != want {
					t.Errorf("%s: Error() = %v, wanted %v", name, got, want)
				}
			} else if gotErr != nil {
				t.Errorf("%s: Validate() = %v, wanted nil", name, gotErr)
			}
		})
	}
}

func TestDestinationGetRef(t *testing.T) {
	ref := &KReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
	}
	tests := map[string]struct {
		dest *Destination
		want *KReference
	}{"nil destination": {
		dest: nil,
		want: nil,
	}, "uri": {
		dest: &Destination{
			URI: &apis.URL{
				Host: "foo",
			},
		},
		want: nil,
	}, "ref": {
		dest: &Destination{
			Ref: ref,
		},
		want: ref,
	}}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			got := tc.dest.GetRef()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Unexpected result (-want +got): %s", diff)
			}
		})
	}
}

func TestDestinationSetDefaults(t *testing.T) {
	ctx := context.Background()

	const parentNamespace = "parentNamespace"

	tests := map[string]struct {
		d    *Destination
		ctx  context.Context
		want string
	}{"uri set, nothing in ref, not modified ": {
		d:   &Destination{URI: apis.HTTP("example.com")},
		ctx: ctx,
	}, "namespace set, nothing in context, not modified ": {
		d:    &Destination{Ref: &KReference{Namespace: namespace}},
		ctx:  ctx,
		want: namespace,
	}, "namespace set, context set, not modified ": {
		d:    &Destination{Ref: &KReference{Namespace: namespace}},
		ctx:  apis.WithinParent(ctx, metav1.ObjectMeta{Namespace: parentNamespace}),
		want: namespace,
	}, "namespace set, uri set, context set, not modified ": {
		d:    &Destination{Ref: &KReference{Namespace: namespace}, URI: apis.HTTP("example.com")},
		ctx:  apis.WithinParent(ctx, metav1.ObjectMeta{Namespace: parentNamespace}),
		want: namespace,
	}, "namespace not set, context set, defaulted": {
		d:    &Destination{Ref: &KReference{}},
		ctx:  apis.WithinParent(ctx, metav1.ObjectMeta{Namespace: parentNamespace}),
		want: parentNamespace,
	}, "namespace not set, uri set, context set, defaulted": {
		d:    &Destination{Ref: &KReference{}, URI: apis.HTTP("example.com")},
		ctx:  apis.WithinParent(ctx, metav1.ObjectMeta{Namespace: parentNamespace}),
		want: parentNamespace,
	}}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.d.SetDefaults(tc.ctx)
			if tc.d.Ref != nil && tc.d.Ref.Namespace != tc.want {
				t.Errorf("Got: %s wanted %s", tc.d.Ref.Namespace, tc.want)
			}
			if tc.d.Ref == nil && tc.want != "" {
				t.Errorf("Got: nil Ref wanted %s", tc.want)
			}
		})
	}
}
