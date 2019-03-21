package storage

import (
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/storage/labels"
	storagelabels "github.com/openshift/odo/pkg/storage/labels"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getStorageFromPVC(t *testing.T) {
	type args struct {
		pvc *corev1.PersistentVolumeClaim
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test case 1: with pvc containing labels",
			args: args{
				pvc: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							labels.StorageLabel: "example-pvc",
						},
					},
				},
			},
			want: "example-pvc",
		},
		{
			name: "test case 2: with no pvc containing labels",
			args: args{
				pvc: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStorageName := getStorageFromPVC(tt.args.pvc)
			if !reflect.DeepEqual(tt.want, gotStorageName) {
				t.Errorf("the returned storage is different, expected: %v, got: %v", tt.want, gotStorageName)
			}
		})
	}
}

func Test_getMachineReadableFormat(t *testing.T) {
	quantity, err := resource.ParseQuantity("100Mi")
	if err != nil {
		t.Errorf("unable to parse size")
	}
	tests := []struct {
		name         string
		inputPVC     *corev1.PersistentVolumeClaim
		mountedPath  string
		activeStatus bool
		want         Storage
	}{
		{
			name: "test case 1: with a pvc, valid path and mounted status",
			inputPVC: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc-example",
					Labels: map[string]string{
						storagelabels.StorageLabel: "pvc-example",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: quantity,
						},
					},
				},
			},
			mountedPath:  "data",
			activeStatus: true,
			want: Storage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc-example",
				},
				TypeMeta: metav1.TypeMeta{Kind: "storage", APIVersion: "odo.openshift.io/v1alpha1"},
				Spec: StorageSpec{
					Size: "100Mi",
				},
				Status: StorageStatus{
					Path: "data",
				},
			},
		},
		{
			name: "test case 2: with a pvc, empty path and unmounted status",
			inputPVC: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc-example",
					Labels: map[string]string{
						storagelabels.StorageLabel: "pvc-example",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: quantity,
						},
					},
				},
			},
			mountedPath:  "",
			activeStatus: false,
			want: Storage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc-example",
				},
				TypeMeta: metav1.TypeMeta{Kind: "storage", APIVersion: "odo.openshift.io/v1alpha1"},
				Spec: StorageSpec{
					Size: "100Mi",
				},
				Status: StorageStatus{
					Path: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStorage := getMachineReadableFormat(*tt.inputPVC, tt.mountedPath)
			if !reflect.DeepEqual(tt.want, gotStorage) {
				t.Errorf("the returned storage is different, expected: %v, got: %v", tt.want, gotStorage)
			}
		})
	}
}

func Test_getMachineReadableFormatForList(t *testing.T) {

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
						Kind:       "List",
						APIVersion: "odo.openshift.io/v1alpha1",
					},
					Spec: StorageSpec{
						Size: "100Mi",
					},
					Status: StorageStatus{
						Path: "data",
					},
				},
			},
			want: StorageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.openshift.io/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []Storage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-pvc-1",
						},
						TypeMeta: metav1.TypeMeta{
							Kind:       "List",
							APIVersion: "odo.openshift.io/v1alpha1",
						},
						Spec: StorageSpec{
							Size: "100Mi",
						},
						Status: StorageStatus{
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
						Kind:       "List",
						APIVersion: "odo.openshift.io/v1alpha1",
					},
					Spec: StorageSpec{
						Size: "100Mi",
					},
					Status: StorageStatus{
						Path: "data",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-pvc-1",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "List",
						APIVersion: "odo.openshift.io/v1alpha1",
					},
					Spec: StorageSpec{
						Size: "500Mi",
					},
					Status: StorageStatus{
						Path: "backend",
					},
				},
			},
			want: StorageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.openshift.io/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []Storage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-pvc-0",
						},
						TypeMeta: metav1.TypeMeta{
							Kind:       "List",
							APIVersion: "odo.openshift.io/v1alpha1",
						},
						Spec: StorageSpec{
							Size: "100Mi",
						},
						Status: StorageStatus{
							Path: "data",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-pvc-1",
						},
						TypeMeta: metav1.TypeMeta{
							Kind:       "List",
							APIVersion: "odo.openshift.io/v1alpha1",
						},
						Spec: StorageSpec{
							Size: "500Mi",
						},
						Status: StorageStatus{
							Path: "backend",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStorage := getMachineReadableFormatForList(tt.inputStorage)
			if !reflect.DeepEqual(tt.want, gotStorage) {
				t.Errorf("the returned storage is different, expected: %v, got: %v", tt.want, gotStorage)
			}
		})
	}
}
