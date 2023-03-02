package kclient

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/redhat-developer/odo/pkg/api"
)

func TestClient_GetRegistryList(t *testing.T) {
	type fields struct {
		Namespace     string
		DynamicClient func() (dynamic.Interface, error)
	}
	tests := []struct {
		name    string
		fields  fields
		want    []api.Registry
		wantErr bool
	}{
		{
			name: "generic error when listing namespaced devfile registries",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					client := fake.NewSimpleDynamicClient(scheme)
					client.PrependReactor("list", "devfileregistrieslists", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("some error")
					})

					return client, nil
				},
			},
			wantErr: true,
		},
		{
			name: "forbidden error when listing namespaced devfile registries, but generic error when listing cluster-scoped registries",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					client := fake.NewSimpleDynamicClient(scheme)
					client.PrependReactor("list", "devfileregistrieslists", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, kerrors.NewForbidden(schema.GroupResource{}, "some-name", errors.New("forbidden"))
					})
					client.PrependReactor("list", "clusterdevfileregistrieslists", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("some-error")
					})

					return client, nil
				},
			},
			wantErr: true,
		},
		{
			name: "forbidden errors when listing both namespaced and cluster-scoped registries",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					client := fake.NewSimpleDynamicClient(scheme)
					client.PrependReactor("list", "devfileregistrieslists", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, kerrors.NewForbidden(schema.GroupResource{}, "some-name", errors.New("forbidden"))
					})
					client.PrependReactor("list", "clusterdevfileregistrieslists", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, kerrors.NewForbidden(schema.GroupResource{}, "some-name", errors.New("forbidden"))
					})

					return client, nil
				},
			},
			wantErr: false,
		},
		{
			name: "unauthorized errors when listing both namespaced and cluster-scoped registries",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					client := fake.NewSimpleDynamicClient(scheme)
					client.PrependReactor("list", "devfileregistrieslists", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, kerrors.NewUnauthorized("unauthorized")
					})
					client.PrependReactor("list", "clusterdevfileregistrieslists", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, kerrors.NewUnauthorized("unauthorized")
					})

					return client, nil
				},
			},
			wantErr: false,
		},
		{
			name: "no registries in cluster",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					client := fake.NewSimpleDynamicClient(scheme)
					return client, nil
				},
			},
			wantErr: false,
		},
		{
			name: "with invalid content in devfileRegistries spec",
			fields: fields{
				Namespace: "my-ns",
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var invalidDevfileRegistry unstructured.Unstructured
					err := invalidDevfileRegistry.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "invalid-devfile-reg1",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "invalidField": [
					     {
					       "name": "devfile-reg01",
					       "url": "https://devfile-reg01.example.com"
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					var invalidClusterDevfileRegistry unstructured.Unstructured
					err = invalidClusterDevfileRegistry.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "ClusterDevfileRegistriesList",
					 "metadata": {
					   "name": "invalid-cluster-devfile-reg1"
					 },
					 "spec": {
					   "invalidField": [
					     {
					       "name": "cluster-devfile-reg01",
					       "url": "https://cluster-devfile-reg01.example.com"
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &invalidDevfileRegistry, &invalidClusterDevfileRegistry)
					return client, nil
				},
			},
			wantErr: false,
		},
		{
			name: "with invalid structure in devfileRegistries spec of namespaced resource",
			fields: fields{
				Namespace: "my-ns",
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var invalidDevfileRegistry unstructured.Unstructured
					err := invalidDevfileRegistry.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "invalid-devfile-reg1",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries":
					     {
					       "name": "devfile-reg01",
					       "url": "https://devfile-reg01.example.com"
					     }
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &invalidDevfileRegistry)
					return client, nil
				},
			},
			wantErr: true,
		},
		{
			name: "with invalid structure in devfileRegistries spec of cluster resource",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var invalidClusterDevfileRegistry unstructured.Unstructured
					err := invalidClusterDevfileRegistry.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "ClusterDevfileRegistriesList",
					 "metadata": {
					   "name": "invalid-cluster-devfile-reg1"
					 },
					 "spec": {
					   "devfileRegistries":
					     {
					       "name": "cluster-devfile-reg01",
					       "url": "https://cluster-devfile-reg01.example.com"
					     }
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &invalidClusterDevfileRegistry)
					return client, nil
				},
			},
			wantErr: true,
		},
		{
			name: "with invalid content in devfileRegistries spec of namespaced resource",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var invalidDevfileRegistry unstructured.Unstructured
					err := invalidDevfileRegistry.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "invalid-devfile-reg1",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     "not-a-map"
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &invalidDevfileRegistry)
					return client, nil
				},
			},
			wantErr: true,
		},
		{
			name: "with invalid name type in devfileRegistries spec of namespaced resource",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var invalidDevfileRegistry unstructured.Unstructured
					err := invalidDevfileRegistry.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "invalid-devfile-reg1",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": true,
					       "url": "https://devfile-reg01.example.com"
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &invalidDevfileRegistry)
					return client, nil
				},
			},
			wantErr: true,
		},
		{
			name: "with invalid url type in devfileRegistries spec of namespaced resource",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var invalidDevfileRegistry unstructured.Unstructured
					err := invalidDevfileRegistry.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "invalid-devfile-reg1",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "my-reg",
					       "url": 1
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &invalidDevfileRegistry)
					return client, nil
				},
			},
			wantErr: true,
		},
		{
			name: "with invalid skipTLSVerify type in devfileRegistries spec of namespaced resource",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var invalidDevfileRegistry unstructured.Unstructured
					err := invalidDevfileRegistry.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "invalid-devfile-reg1",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "my-reg",
					       "url": "https://devfile-reg01.example.com",
						   "skipTLSVerify": "not-an-int"
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &invalidDevfileRegistry)
					return client, nil
				},
			},
			wantErr: true,
		},
		{
			name: "with only namespaced registries",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var nsRegistry1 unstructured.Unstructured
					err := nsRegistry1.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "devfile-reg1",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "devfile-reg11",
					       "url": "https://devfile-reg11.example.com"
					     },
						 {
					       "name": "devfile-reg12",
					       "url": "https://devfile-reg12.example.com",
 						   "skipTLSVerify": true
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					var nsRegistry2 unstructured.Unstructured
					err = nsRegistry2.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "devfile-reg2",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "devfile-reg21",
					       "url": "https://devfile-reg21.example.com"
					     },
						 {
					       "name": "devfile-reg22",
					       "url": "https://devfile-reg22.example.com",
 						   "skipTLSVerify": true
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &nsRegistry1, &nsRegistry2)
					return client, nil
				},
			},
			wantErr: false,
			want: []api.Registry{
				{Name: "devfile-reg11", URL: "https://devfile-reg11.example.com", Secure: true},
				{Name: "devfile-reg12", URL: "https://devfile-reg12.example.com"},
				{Name: "devfile-reg21", URL: "https://devfile-reg21.example.com", Secure: true},
				{Name: "devfile-reg22", URL: "https://devfile-reg22.example.com"},
			},
		},
		{
			name: "with only cluster-scoped registries",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var clusterRegistry1 unstructured.Unstructured
					err := clusterRegistry1.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "ClusterDevfileRegistriesList",
					 "metadata": {
					   "name": "cluster-devfile-reg1"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "cluster-devfile-reg11",
					       "url": "https://cluster-devfile-reg11.example.com"
					     },
						 {
					       "name": "cluster-devfile-reg12",
					       "url": "https://cluster-devfile-reg12.example.com",
 						   "skipTLSVerify": true
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					var clusterRegistry2 unstructured.Unstructured
					err = clusterRegistry2.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "ClusterDevfileRegistriesList",
					 "metadata": {
					   "name": "cluster-devfile-reg2"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "cluster-devfile-reg21",
					       "url": "https://cluster-devfile-reg21.example.com"
					     },
						 {
					       "name": "cluster-devfile-reg22",
					       "url": "https://cluster-devfile-reg22.example.com",
 						   "skipTLSVerify": true
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(scheme, &clusterRegistry1, &clusterRegistry2)
					return client, nil
				},
			},
			wantErr: false,
			want: []api.Registry{
				{
					Name:   "cluster-devfile-reg11",
					URL:    "https://cluster-devfile-reg11.example.com",
					Secure: true,
				},
				{Name: "cluster-devfile-reg12", URL: "https://cluster-devfile-reg12.example.com"},
				{
					Name:   "cluster-devfile-reg21",
					URL:    "https://cluster-devfile-reg21.example.com",
					Secure: true,
				},
				{Name: "cluster-devfile-reg22", URL: "https://cluster-devfile-reg22.example.com"},
			},
		},
		{
			name: "namespaced registries have higher priority over cluster-scoped ones",
			fields: fields{
				DynamicClient: func() (dynamic.Interface, error) {
					scheme := runtime.NewScheme()
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "ClusterDevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})
					scheme.AddKnownTypeWithName(schema.GroupVersionKind{
						Group:   "registry.devfile.io",
						Version: "v1alpha1",
						Kind:    "DevfileRegistriesListList",
					}, &unstructured.UnstructuredList{})

					var nsRegistryNoUrl unstructured.Unstructured
					err := nsRegistryNoUrl.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "nsRegistryNoUrl",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "devfile-reg00",
						   "url": ""
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					var nsRegistry1 unstructured.Unstructured
					err = nsRegistry1.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "devfile-reg1",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "devfile-reg11",
					       "url": "https://devfile-reg11.example.com"
					     },
						 {
					       "name": "devfile-reg12",
					       "url": "https://devfile-reg12.example.com",
 						   "skipTLSVerify": true
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					var nsRegistry2 unstructured.Unstructured
					err = nsRegistry2.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "DevfileRegistriesList",
					 "metadata": {
					   "name": "devfile-reg2",
						"namespace": "my-ns"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "devfile-reg21",
					       "url": "https://devfile-reg21.example.com"
					     },
						 {
					       "name": "devfile-reg22",
					       "url": "https://devfile-reg22.example.com",
 						   "skipTLSVerify": true
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					var clusterRegistry1 unstructured.Unstructured
					err = clusterRegistry1.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "ClusterDevfileRegistriesList",
					 "metadata": {
					   "name": "cluster-devfile-reg1"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "cluster-devfile-reg11",
					       "url": "https://cluster-devfile-reg11.example.com"
					     },
						 {
					       "name": "cluster-devfile-reg12",
					       "url": "https://cluster-devfile-reg12.example.com",
 						   "skipTLSVerify": true
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					var clusterRegistry2 unstructured.Unstructured
					err = clusterRegistry2.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "ClusterDevfileRegistriesList",
					 "metadata": {
					   "name": "cluster-devfile-reg2"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "cluster-devfile-reg21",
					       "url": "https://cluster-devfile-reg21.example.com"
					     },
						 {
					       "name": "cluster-devfile-reg22",
					       "url": "https://cluster-devfile-reg22.example.com",
 						   "skipTLSVerify": true
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					var clusterRegistryNoUrl unstructured.Unstructured
					err = clusterRegistryNoUrl.UnmarshalJSON([]byte(`
					{
					 "apiVersion": "registry.devfile.io/v1alpha1",
					 "kind": "ClusterDevfileRegistriesList",
					 "metadata": {
					   "name": "clusterRegistryNoUrl"
					 },
					 "spec": {
					   "devfileRegistries": [
					     {
					       "name": "cluster-devfile-reg00",
					       "url": ""
					     }
					   ]
					 }
					}`))
					if err != nil {
						return nil, err
					}

					client := fake.NewSimpleDynamicClient(
						scheme,
						&clusterRegistryNoUrl,
						&nsRegistry1,
						&clusterRegistry1,
						&nsRegistryNoUrl,
						&nsRegistry2,
						&clusterRegistry2,
					)

					return client, nil
				},
			},
			wantErr: false,
			want: []api.Registry{
				{Name: "devfile-reg11", URL: "https://devfile-reg11.example.com", Secure: true},
				{Name: "devfile-reg12", URL: "https://devfile-reg12.example.com"},
				{Name: "devfile-reg21", URL: "https://devfile-reg21.example.com", Secure: true},
				{Name: "devfile-reg22", URL: "https://devfile-reg22.example.com"},
				{
					Name:   "cluster-devfile-reg11",
					URL:    "https://cluster-devfile-reg11.example.com",
					Secure: true,
				},
				{Name: "cluster-devfile-reg12", URL: "https://cluster-devfile-reg12.example.com"},
				{
					Name:   "cluster-devfile-reg21",
					URL:    "https://cluster-devfile-reg21.example.com",
					Secure: true,
				},
				{Name: "cluster-devfile-reg22", URL: "https://cluster-devfile-reg22.example.com"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynamicClient, err := tt.fields.DynamicClient()
			if err != nil {
				t.Errorf("unable to create dynamic client: %v", err)
				return
			}

			c := &Client{
				Namespace:     tt.fields.Namespace,
				DynamicClient: dynamicClient,
			}
			got, err := c.GetRegistryList()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetRegistryList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Client.GetRegistryList() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
