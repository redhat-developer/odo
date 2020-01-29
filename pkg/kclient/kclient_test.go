package kclient

import (
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktesting "k8s.io/client-go/testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetPVCsFromSelector(t *testing.T) {
	tests := []struct {
		name      string
		pvcName   string
		size      string
		namespace string
		labels    map[string]string
		wantErr   bool
	}{
		{
			name:      "Case: Valid pvc name",
			pvcName:   "mypvc",
			size:      "1Gi",
			namespace: "default",
			labels: map[string]string{
				"mylabel1": "testpvc1",
				"mylabel2": "testpvc2",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = tt.namespace

			quantity, err := resource.ParseQuantity(tt.size)
			if err != nil {
				t.Errorf("resource.ParseQuantity unexpected error %v", err)
			}
			pvcSpec := GeneratePVCSpec(quantity)
			objectMeta := metav1.ObjectMeta{
				Name:        tt.pvcName,
				Namespace:   tt.namespace,
				Labels:      tt.labels,
				Annotations: nil,
			}

			fkclientset.Kubernetes.PrependReactor("create", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				pvc := corev1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						Kind:       PersistentVolumeClaimKind,
						APIVersion: PersistentVolumeClaimAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.pvcName,
					},
				}
				return true, &pvc, nil
			})

			_, err = fkclient.CreatePVC(objectMeta, *pvcSpec)
			if err != nil {
				t.Errorf("TestGetPVCsFromSelector: Error creating PVC %s", tt.pvcName)
			}

			var selector string
			for labelkey, labelvalue := range tt.labels {
				if selector != "" {
					selector = selector + ","
				}
				selector = selector + labelkey + "=" + labelvalue
			}

			listOfPVC := corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   tt.pvcName,
							Labels: tt.labels,
						},
					},
				},
			}

			fkclientset.Kubernetes.PrependReactor("list", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), selector) {
					return true, nil, fmt.Errorf("TestGetPVCsFromSelector: Labels do not match with expected values, expected:%s, got:%s", selector, action.(ktesting.ListAction).GetListRestrictions())
				}
				return true, &listOfPVC, nil
			})

			PVCs, err := fkclient.GetPVCsFromSelector(selector)
			if err != nil {
				t.Errorf("TestGetPVCsFromSelector: Error listing PVCs with selector %s", selector)
			}

			if len(PVCs) == 0 || len(PVCs) > 1 {
				t.Errorf("TestGetPVCsFromSelector: Incorrect amount of PVC found with selector %s", selector)
			} else {
				for _, PVC := range PVCs {
					if PVC.Name != tt.pvcName {
						t.Errorf("TestGetPVCsFromSelector: PVC found with incorrect name, expected: %s actual: %s", tt.pvcName, PVC.Name)
					}
					if !reflect.DeepEqual(PVC.Labels, tt.labels) {
						t.Errorf("TestGetPVCsFromSelector: Labels do not match with expected labels, expected: %s, got %s", tt.labels, PVC.Labels)
					}
				}
			}
		})
	}
}
