package storage

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name        string
		storageName string
		storageSize string
		mountedPath string
		want        Storage
	}{
		{
			name:        "test case 1: with a pvc, valid path and mounted status",
			storageName: "pvc-example",
			storageSize: "100Mi",
			mountedPath: "data",
			want: Storage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc-example",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Storage", APIVersion: "odo.dev/v1alpha1"},
				Spec: StorageSpec{
					Size: "100Mi",
					Path: "data",
				},
			},
		},
		{
			name:        "test case 2: with a pvc, empty path and unmounted status",
			storageName: "pvc-example",
			storageSize: "500Mi",
			mountedPath: "",
			want: Storage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc-example",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Storage", APIVersion: "odo.dev/v1alpha1"},
				Spec: StorageSpec{
					Size: "500Mi",
					Path: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStorage := NewStorage(tt.storageName, tt.storageSize, tt.mountedPath, nil)
			if !reflect.DeepEqual(tt.want, gotStorage) {
				t.Errorf("the returned storage is different, expected: %v, got: %v", tt.want, gotStorage)
			}
		})
	}
}

func TestNewStorageList(t *testing.T) {

	tests := []struct {
		name         string
		inputStorage []Storage
		want         StorageList
	}{
		{
			name: "test case 1: with a single pvc",
			inputStorage: []Storage{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-pvc-1",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Storage",
						APIVersion: "odo.dev/v1alpha1",
					},
					Spec: StorageSpec{
						Size: "100Mi",
						Path: "data",
					},
				},
			},
			want: StorageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.dev/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []Storage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-pvc-1",
						},
						TypeMeta: metav1.TypeMeta{
							Kind:       "Storage",
							APIVersion: "odo.dev/v1alpha1",
						},
						Spec: StorageSpec{
							Size: "100Mi",
							Path: "data",
						},
					},
				},
			},
		},
		{
			name: "test case 2: with multiple pvcs",
			inputStorage: []Storage{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-pvc-0",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Storage",
						APIVersion: "odo.dev/v1alpha1",
					},
					Spec: StorageSpec{
						Size: "100Mi",
						Path: "data",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-pvc-1",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Storage",
						APIVersion: "odo.dev/v1alpha1",
					},
					Spec: StorageSpec{
						Size: "500Mi",
						Path: "backend",
					},
				},
			},
			want: StorageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.dev/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []Storage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-pvc-0",
						},
						TypeMeta: metav1.TypeMeta{
							Kind:       "Storage",
							APIVersion: "odo.dev/v1alpha1",
						},
						Spec: StorageSpec{
							Size: "100Mi",
							Path: "data",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-pvc-1",
						},
						TypeMeta: metav1.TypeMeta{
							Kind:       "Storage",
							APIVersion: "odo.dev/v1alpha1",
						},
						Spec: StorageSpec{
							Size: "500Mi",
							Path: "backend",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStorage := NewStorageList(tt.inputStorage)
			if !reflect.DeepEqual(tt.want, gotStorage) {
				t.Errorf("the returned storage is different, expected: %v, got: %v", tt.want, gotStorage)
			}
		})
	}
}
