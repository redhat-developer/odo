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

package defaulting

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	authenticationv1 "k8s.io/api/authentication/v1"
	"knative.dev/pkg/apis"
	. "knative.dev/pkg/logging/testing"
	. "knative.dev/pkg/testing"
	. "knative.dev/pkg/webhook/testing"
)

func TestSetUserInfoAnnotationsWhenWithinCreate(t *testing.T) {
	tests := []struct {
		name                string
		configureContext    func(context.Context) context.Context
		setup               func(context.Context, *Resource)
		expectedAnnotations map[string]string
	}{{
		name: "test create",
		configureContext: func(ctx context.Context) context.Context {
			return apis.WithinCreate(apis.WithUserInfo(ctx, &authenticationv1.UserInfo{Username: user1}))
		},
		setup: func(ctx context.Context, r *Resource) {
			r.Annotations = map[string]string{}
		},
		expectedAnnotations: map[string]string{
			"pkg.knative.dev/creator":      user1,
			"pkg.knative.dev/lastModifier": user1,
		},
	}, {
		name: "test create (should override user info annotations when they are present)",
		configureContext: func(ctx context.Context) context.Context {
			return apis.WithinCreate(apis.WithUserInfo(ctx, &authenticationv1.UserInfo{Username: user1}))
		},
		setup: func(ctx context.Context, r *Resource) {
			r.Annotations = map[string]string{
				"pkg.knative.dev/creator":      user2,
				"pkg.knative.dev/lastModifier": user2,
			}
		},
		expectedAnnotations: map[string]string{
			"pkg.knative.dev/creator":      user1,
			"pkg.knative.dev/lastModifier": user1,
		},
	}, {
		name:                "test create (should not touch annotations when no user info available)",
		configureContext:    apis.WithinCreate,
		setup:               func(ctx context.Context, r *Resource) {},
		expectedAnnotations: nil,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := CreateResource("a name")

			ctx := tc.configureContext(TestContextWithLogger(t))

			tc.setup(ctx, r)

			SetUserInfoAnnotations(r, ctx, "pkg.knative.dev")

			if !reflect.DeepEqual(r.Annotations, tc.expectedAnnotations) {
				t.Logf("Got :  %#v", r.Annotations)
				t.Logf("Want: %#v", tc.expectedAnnotations)
				if diff := cmp.Diff(tc.expectedAnnotations, r.Annotations, cmpopts.EquateEmpty()); diff != "" {
					t.Logf("diff: %v", diff)
				}
				t.Fatalf("Annotations don't match")
			}

		})
	}
}

func TestSetUserInfoAnnotationsWhenWithinUpdate(t *testing.T) {
	tests := []struct {
		name                string
		configureContext    func(context.Context, *Resource) context.Context
		setup               func(context.Context, *Resource)
		expectedAnnotations map[string]string
	}{{
		name: "test update (should add updater annotation when it is not present)",
		configureContext: func(ctx context.Context, r *Resource) context.Context {
			return apis.WithinUpdate(apis.WithUserInfo(ctx, &authenticationv1.UserInfo{Username: user1}), r)
		},
		setup: func(ctx context.Context, r *Resource) {
			r.Annotations = map[string]string{
				"pkg.knative.dev/creator": user2,
			}
			r.Spec.FieldWithDefault = "changing this field"
		},
		expectedAnnotations: map[string]string{
			"pkg.knative.dev/creator":      user2,
			"pkg.knative.dev/lastModifier": user1,
		},
	}, {
		name: "test update (should update updater annotation when it is present)",
		configureContext: func(ctx context.Context, r *Resource) context.Context {
			return apis.WithinUpdate(apis.WithUserInfo(ctx, &authenticationv1.UserInfo{Username: user1}), r)
		},
		setup: func(ctx context.Context, r *Resource) {
			r.Annotations = map[string]string{
				"pkg.knative.dev/creator":      user2,
				"pkg.knative.dev/lastModifier": user2,
			}
			r.Spec.FieldWithDefault = "changing this field"
		},
		expectedAnnotations: map[string]string{
			// should not change
			"pkg.knative.dev/creator":      user2,
			"pkg.knative.dev/lastModifier": user1,
		},
	}, {
		name: "test update (should not touch annotations when no user info available)",
		configureContext: func(ctx context.Context, r *Resource) context.Context {
			return apis.WithinUpdate(ctx, r)
		},
		setup: func(ctx context.Context, r *Resource) {
			// this is not necessary, but let's do this in case the execution flow changes
			r.Spec.FieldWithDefault = "changing this field"
		},
		expectedAnnotations: nil,
	}, {
		name: "test update (should not touch annotations when nothing in spec is changed)",
		configureContext: func(ctx context.Context, r *Resource) context.Context {
			return apis.WithinUpdate(apis.WithUserInfo(ctx, &authenticationv1.UserInfo{Username: user1}), r)
		},
		setup: func(ctx context.Context, r *Resource) {
			// change nothing
		},
		expectedAnnotations: map[string]string{},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := CreateResource("a name")

			ctx := tc.configureContext(TestContextWithLogger(t), r)

			new := r.DeepCopy()

			tc.setup(ctx, new)

			SetUserInfoAnnotations(new, ctx, "pkg.knative.dev")

			if !reflect.DeepEqual(new.Annotations, tc.expectedAnnotations) {
				t.Logf("Got :  %#v", new.Annotations)
				t.Logf("Want: %#v", tc.expectedAnnotations)
				if diff := cmp.Diff(tc.expectedAnnotations, new.Annotations, cmpopts.EquateEmpty()); diff != "" {
					t.Logf("diff: %v", diff)
				}
				t.Fatalf("Annotations don't match")
			}

		})
	}
}
