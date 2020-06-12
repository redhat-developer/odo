package storage

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ktesting "k8s.io/client-go/testing"
)

func TestCreateComponentStorage(t *testing.T) {

	testComponentName := "test"
	podName := "testpod"
	fakeUID := types.UID("12345")
	volNames := [...]string{"vol1", "vol2"}
	volSize := "5Gi"

	tests := []struct {
		name     string
		storages []common.Storage
	}{
		{
			name: "storage test",
			storages: []common.Storage{
				{
					Name: "vol1-pvc",
					Volume: common.DevfileVolume{
						Name: volNames[0],
						Size: volSize,
					},
				},
				{
					Name: "vol2-pvc",
					Volume: common.DevfileVolume{
						Name: volNames[1],
						Size: volSize,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := kclient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				deployment := appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       kclient.DeploymentKind,
						APIVersion: kclient.DeploymentAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: podName,
						UID:  fakeUID,
					},
				}
				return true, &deployment, nil
			})

			// Create one of the test volumes
			createdPVC, err := Create(fkclient, tt.storages[0].Volume.Name, tt.storages[0].Volume.Size, testComponentName, tt.storages[0].Name)
			if err != nil {
				t.Errorf("Error creating PVC %v: %v", tt.storages[0].Name, err)
			}

			if createdPVC.Name != tt.storages[0].Name {
				t.Errorf("PVC created name mismatch, expected: %v actual: %v", tt.storages[0].Name, createdPVC.Name)
			}

			fkclientset.Kubernetes.PrependReactor("create", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				labels := map[string]string{
					"component":    testComponentName,
					"storage-name": tt.storages[1].Volume.Name,
				}
				PVC := testingutil.FakePVC(tt.storages[1].Name, tt.storages[1].Volume.Size, labels)
				return true, PVC, nil
			})

			// It should create the remaining PVC and reuse the existing PVC
			err = CreateComponentStorage(fkclient, tt.storages, testComponentName)
			if err != nil {
				t.Errorf("Error creating component storage %v: %v", tt.storages, err)
			}
		})
	}

}

func TestStorageCreate(t *testing.T) {

	testComponentName := "test"
	podName := "testpod"
	fakeUID := types.UID("12345")
	volNames := [...]string{"vol1", "vol2"}
	volSize := "5Gi"
	garbageVolSize := "abc"

	tests := []struct {
		name    string
		storage common.Storage
		wantErr bool
		err     error
	}{
		{
			name: "valid pvc",
			storage: common.Storage{
				Name: "vol1-pvc",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: volSize,
				},
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "pvc with long name",
			storage: common.Storage{
				Name: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: volSize,
				},
			},
			wantErr: true,
			err:     errors.New("Error creating PVC, name is greater than 63"),
		},
		{
			name: "pvc with no name",
			storage: common.Storage{
				Name: "",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: volSize,
				},
			},
			wantErr: true,
			err:     errors.New("Error creating PVC, name is empty"),
		},
		{
			name: "pvc with invalid size",
			storage: common.Storage{
				Name: "vol1-pvc",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: garbageVolSize,
				},
			},
			wantErr: true,
			err:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := kclient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				deployment := appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       kclient.DeploymentKind,
						APIVersion: kclient.DeploymentAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: podName,
						UID:  fakeUID,
					},
				}
				return true, &deployment, nil
			})

			fkclientset.Kubernetes.PrependReactor("create", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				labels := map[string]string{
					"component":    testComponentName,
					"storage-name": tt.storage.Volume.Name,
				}
				if tt.wantErr {
					return true, nil, tt.err
				}
				PVC := testingutil.FakePVC(tt.storage.Name, tt.storage.Volume.Size, labels)
				return true, PVC, nil
			})

			// Create one of the test volumes
			createdPVC, err := Create(fkclient, tt.storage.Volume.Name, tt.storage.Volume.Size, testComponentName, tt.storage.Name)
			if !tt.wantErr && err != nil {
				t.Errorf("Error creating PVC %v: %v", tt.storage.Name, err)
			} else if tt.wantErr && err != nil {
				// don't perform further checks if we want an error
				return
			}

			if createdPVC.Name != tt.storage.Name {
				t.Errorf("PVC created name mismatch, expected: %v actual: %v", tt.storage.Name, createdPVC.Name)
			}
		})
	}

}
