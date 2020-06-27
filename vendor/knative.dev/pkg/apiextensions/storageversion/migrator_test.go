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

package storageversion

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	apix "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apixFake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

var (
	fakeGK = schema.GroupKind{
		Kind:  "Fake",
		Group: "group.dev",
	}

	fakeGR = schema.GroupResource{
		Resource: "fakes",
		Group:    "group.dev",
	}

	fakeCRD = &apix.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fakeGR.String(),
		},
		Spec: apix.CustomResourceDefinitionSpec{
			Group: fakeGK.Group,
			Versions: []apix.CustomResourceDefinitionVersion{
				{Name: "v1alpha1", Served: true, Storage: false},
				{Name: "v1", Served: true, Storage: true},
			},
		},
		Status: apix.CustomResourceDefinitionStatus{
			StoredVersions: []string{
				"v1alpha1",
				"v1",
			},
		},
	}
)

func TestMigrate(t *testing.T) {
	// setup
	resources := []runtime.Object{fake("first"), fake("second")}
	dclient := dynamicFake.NewSimpleDynamicClient(runtime.NewScheme(), resources...)
	cclient := apixFake.NewSimpleClientset(fakeCRD)
	m := NewMigrator(dclient, cclient)

	if err := m.Migrate(context.Background(), fakeGR); err != nil {
		t.Fatalf("Migrate() = %v", err)
	}

	assertPatches(t, dclient.Actions(),
		// patch resource definition dropping non-storage version
		emptyResourcePatch("first", "v1"),
		emptyResourcePatch("second", "v1"),
	)

	assertPatches(t, cclient.Actions(),
		// patch resource definition dropping non-storage version
		crdStorageVersionPatch(fakeCRD.Name, "v1"),
	)
}

// func TestMigrate_Paging(t *testing.T) {
// 	t.Skip("client-go lacks testing pagination " +
// 		"since list options aren't captured in actions")
// }

func TestMigrate_Errors(t *testing.T) {
	tests := []struct {
		name string
		crd  func(*k8stesting.Fake)
		dyn  func(*k8stesting.Fake)
	}{{
		name: "failed to fetch CRD",
		crd: func(fake *k8stesting.Fake) {
			fake.PrependReactor("get", "*",
				func(k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("failed to get crd")
				})
		},
	}, {
		name: "listing fails",
		dyn: func(fake *k8stesting.Fake) {
			fake.PrependReactor("list", "*",
				func(k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("failed to list resources")
				})
		},
	}, {
		name: "patching resource fails",
		dyn: func(fake *k8stesting.Fake) {
			fake.PrependReactor("patch", "*",
				func(k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("failed to patch resources")
				})
		},
	}, {
		name: "patching definition fails",
		crd: func(fake *k8stesting.Fake) {
			fake.PrependReactor("patch", "*",
				func(k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("failed to patch definition")
				})
		},
	},
	// todo paging fails
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resources := []runtime.Object{fake("first"), fake("second")}
			dclient := dynamicFake.NewSimpleDynamicClient(runtime.NewScheme(), resources...)
			cclient := apixFake.NewSimpleClientset(fakeCRD)

			if test.crd != nil {
				test.crd(&cclient.Fake)
			}

			if test.dyn != nil {
				test.dyn(&dclient.Fake)
			}

			m := NewMigrator(dclient, cclient)
			if err := m.Migrate(context.Background(), fakeGR); err == nil {
				t.Error("Migrate should have returned an error")
			}
		})
	}
}

func assertPatches(t *testing.T, actions []k8stesting.Action, want ...k8stesting.PatchAction) {
	t.Helper()

	got := getPatchActions(actions)
	if diff := cmp.Diff(got, want); diff != "" {
		t.Error("Unexpected patches:", diff)
	}
}

func emptyResourcePatch(name, version string) k8stesting.PatchAction {
	return k8stesting.NewPatchAction(
		fakeGR.WithVersion(version),
		"default",
		name,
		types.MergePatchType,
		[]byte("{}"))
}

func crdStorageVersionPatch(name, version string) k8stesting.PatchAction {
	return k8stesting.NewRootPatchSubresourceAction(
		apix.SchemeGroupVersion.WithResource("customresourcedefinitions"),
		name,
		types.StrategicMergePatchType,
		[]byte(`{"status":{"storedVersions":["`+version+`"]}}`),
		"status",
	)
}

func getPatchActions(actions []k8stesting.Action) []k8stesting.PatchAction {
	var patches []k8stesting.PatchAction

	for _, action := range actions {
		if pa, ok := action.(k8stesting.PatchAction); ok {
			patches = append(patches, pa)
		}
	}

	return patches
}

func fake(name string) runtime.Object {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "group.dev/v1",
			"kind":       "Fake",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "default",
			},
		},
	}
}
