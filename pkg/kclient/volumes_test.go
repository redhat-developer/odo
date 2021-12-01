package kclient

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/testingutil"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktesting "k8s.io/client-go/testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/redhat-developer/odo/pkg/util"
)

func TestCreatePVC(t *testing.T) {

	tests := []struct {
		name      string
		pvcName   string
		size      string
		namespace string
		labels    map[string]string
		wantErr   bool
	}{
		{
			name:      "Case 1: Valid pvc name",
			pvcName:   "mypvc",
			size:      "1Gi",
			namespace: "default",
			labels: map[string]string{
				"testpvc": "testpvc",
			},
			wantErr: false,
		},
		{
			name:      "Case 2: Invalid pvc name",
			pvcName:   "",
			size:      "1Gi",
			namespace: "default",
			labels: map[string]string{
				"testpvc": "testpvc",
			},
			wantErr: true,
		},
		{
			name:      "Case 3: Invalid pvc size",
			pvcName:   "mypvc",
			size:      "garbage",
			namespace: "default",
			labels: map[string]string{
				"testpvc": "testpvc",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = tt.namespace

			quantity, err := resource.ParseQuantity(tt.size)
			if err != nil && tt.size != "garbage" {
				t.Errorf("resource.ParseQuantity unexpected error %v", err)
			} else if err != nil && tt.size == "garbage" {
				return
			}

			objectMeta := generator.GetObjectMeta(tt.pvcName, tt.namespace, tt.labels, nil)

			fkclientset.Kubernetes.PrependReactor("create", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.pvcName == "" {
					return true, nil, errors.Errorf("pvc name is empty")
				}
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

			pvcParams := generator.PVCParams{
				ObjectMeta: objectMeta,
				Quantity:   quantity,
			}
			pvc := generator.GetPVC(pvcParams)

			createdPVC, err := fkclient.CreatePVC(*pvc)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreatePVC unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action in StartPVC got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if createdPVC.Name != tt.pvcName {
						t.Errorf("deployment name does not match the expected name, expected: %s, got %s", tt.pvcName, createdPVC.Name)
					}
				}
			}
		})
	}
}

func TestDeletePVC(t *testing.T) {
	tests := []struct {
		name    string
		pvcName string
		wantErr bool
	}{
		{
			name:    "storage 10Gi",
			pvcName: "postgresql",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("delete", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			err := fakeClient.DeletePVC(tt.pvcName)

			//Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.DeletePVC(name) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			if (len(fakeClientSet.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in DeletePVC got: %v", fakeClientSet.Kubernetes.Actions())
			}

			// Check for value with which the function has called
			DeletedPVC := fakeClientSet.Kubernetes.Actions()[0].(ktesting.DeleteAction).GetName()
			if DeletedPVC != tt.pvcName {
				t.Errorf("Delete action is performed with wrong pvcName, expected: %s, got %s", tt.pvcName, DeletedPVC)

			}
		})
	}
}

func TestListPVCs(t *testing.T) {
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
		{
			name:      "Case: Wrong Label Selector",
			pvcName:   "mypvc",
			size:      "1Gi",
			namespace: "default",
			labels: map[string]string{
				"mylabel1": "testpvc1",
				"mylabel2": "testpvc2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = tt.namespace

			selector := util.ConvertLabelsToSelector(tt.labels)

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
				if tt.name == "Case: Wrong Label Selector" {
					return true, nil, fmt.Errorf("TestGetPVCsFromSelector: Labels do not match with expected values, expected:%s, got:%s", selector, selector+",garbage=true")
				}
				return true, &listOfPVC, nil
			})

			PVCs, err := fkclient.ListPVCs(selector)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGetPVCsFromSelector: Error listing PVCs with selector: %v", err)
			}

			if len(PVCs) == 0 || len(PVCs) > 1 {
				if !tt.wantErr {
					t.Errorf("TestGetPVCsFromSelector: Incorrect amount of PVC found with selector %s", selector)
				}
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

func TestGetPVCFromName(t *testing.T) {
	tests := []struct {
		name    string
		pvcName string
		wantPVC *corev1.PersistentVolumeClaim
		wantErr bool
	}{
		{
			name:    "storage 10Gi",
			pvcName: "postgresql",
			wantPVC: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "postgresql",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("get", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.wantPVC, nil
			})

			returnPVC, err := fakeClient.GetPVCFromName(tt.pvcName)

			//Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.GetPVCFromName(name) unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			// Check for validating actions performed
			if (len(fakeClientSet.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in GetPVCFromName got: %v", fakeClientSet.Kubernetes.Actions())
			}
			// Check for value with which the function has called
			PVCname := fakeClientSet.Kubernetes.Actions()[0].(ktesting.GetAction).GetName()
			if PVCname != tt.pvcName {
				t.Errorf("Get action is performed with wrong pvcName, expected: %s, got %s", tt.pvcName, PVCname)

			}
			// Check for returnPVC and tt.wantPVC is same
			if returnPVC != tt.wantPVC {
				t.Errorf("Get action has returned pvc with wrong name, expected: %s, got %s", tt.wantPVC, returnPVC)
			}
		})
	}
}

func TestListPVCNames(t *testing.T) {
	type args struct {
		selector string
	}
	tests := []struct {
		name         string
		args         args
		returnedPVCs *corev1.PersistentVolumeClaimList
		want         []string
		wantErr      bool
	}{
		{
			name: "case 1: two pvcs returned",
			args: args{
				"component-name=nodejs",
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("storage-1", "1Gi", map[string]string{"component-name": "nodejs"}),
					*testingutil.FakePVC("storage-2", "1Gi", map[string]string{"component-name": "nodejs"}),
				},
			},
			want: []string{"storage-1", "storage-2"},
		},
		{
			name: "case 2: no pvcs returned",
			args: args{
				"component-name=nodejs",
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()

			fkclientset.Kubernetes.PrependReactor("list", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedPVCs, nil
			})

			got, err := fkclient.ListPVCNames(tt.args.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPVCNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListPVCNames() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateStorageOwnerReference(t *testing.T) {
	fakeDeployment := testingutil.CreateFakeDeployment("nodejs")
	type args struct {
		pvc            *corev1.PersistentVolumeClaim
		ownerReference []metav1.OwnerReference
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: valid pvc",
			args: args{
				pvc: testingutil.FakePVC("pvc-1", "1Gi", map[string]string{}),
				ownerReference: []metav1.OwnerReference{
					generator.GetOwnerReference(fakeDeployment),
				},
			},
			wantErr: false,
		},
		{
			name: "case 2: empty owner reference for pvc",
			args: args{
				pvc:            testingutil.FakePVC("pvc-1", "1Gi", map[string]string{}),
				ownerReference: []metav1.OwnerReference{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("get", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				returnedPVC := *tt.args.pvc
				return true, &returnedPVC, nil
			})

			fakeClientSet.Kubernetes.PrependReactor("update", "persistentvolumeclaims", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				pvc := action.(ktesting.UpdateAction).GetObject().(*corev1.PersistentVolumeClaim)
				if pvc.OwnerReferences == nil || pvc.OwnerReferences[0].Name != fakeDeployment.Name {
					t.Errorf("owner reference not set for dc %s", tt.args.pvc.Name)
				}
				return true, pvc, nil
			})

			if err := fakeClient.GetAndUpdateStorageOwnerReference(tt.args.pvc, tt.args.ownerReference...); (err != nil) != tt.wantErr {
				t.Errorf("updateStorageOwnerReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
