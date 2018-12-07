package storage

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/storage/labels"
	corev1 "k8s.io/api/core/v1"
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
