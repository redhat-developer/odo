package storage

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"
	v1 "github.com/openshift/api/apps/v1"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func Test_s2iClient_List(t *testing.T) {
	componentName := "nodejs"
	appName := "app"

	pvc1 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-1-app"), "100Mi", getStorageLabels("storage-1", componentName, appName))

	pvc3 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-3-app"), "100Mi", getStorageLabels("storage-3", componentName, appName))

	storage1 := NewStorage("storage-1", "100Mi", "/tmp1")
	storage1.Status = StateTypePushed

	storage2 := NewStorage("storage-2", "100Mi", "/tmp2")
	storage2.Status = StateTypeNotPushed

	storage3 := NewStorage("storage-3", "100Mi", "/tmp3")
	storage3.Status = StateTypeLocallyDeleted

	type fields struct {
		generic generic
	}
	tests := []struct {
		name                 string
		fields               fields
		wantErr              bool
		returnedLocalStorage []localConfigProvider.LocalStorage
		returnedPVCs         corev1.PersistentVolumeClaimList
		mountedMap           map[string]*corev1.PersistentVolumeClaim
		wantedStorageList    StorageList
	}{
		{
			name: "case 1: Storage is Pushed",

			returnedLocalStorage: []localConfigProvider.LocalStorage{
				{
					Name: "storage-1",
					Size: "100Mi",
					Path: "/tmp1",
				},
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*pvc1,
				},
			},
			mountedMap: map[string]*corev1.PersistentVolumeClaim{
				"/tmp1": pvc1,
			},
			wantedStorageList: NewStorageList([]Storage{storage1}),
			wantErr:           false,
		},
		{
			name: "case 2: Storage is Not Pushed",
			// Storage present in local conf
			returnedLocalStorage: []localConfigProvider.LocalStorage{
				{
					Name: "storage-2",
					Size: "100Mi",
					Path: "/tmp2",
				},
			},
			// Return PVC's from cluster
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{},
			},
			mountedMap:        map[string]*corev1.PersistentVolumeClaim{},
			wantedStorageList: NewStorageList([]Storage{storage2}),
			wantErr:           false,
		},

		{

			name: "case 3: Storage is Locally Deleted",
			// storage-1 present in local conf
			returnedLocalStorage: []localConfigProvider.LocalStorage{
				{
					Name: "storage-1",
					Size: "100Mi",
					Path: "/tmp1",
				},
			},
			// pvc1, pvc3 present in cluster
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{*pvc1, *pvc3},
			},
			mountedMap:        map[string]*corev1.PersistentVolumeClaim{"/tmp1": pvc1, "/tmp3": pvc3},
			wantedStorageList: NewStorageList([]Storage{storage1, storage3}),
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := occlient.FakeNew()

			dcTesting := testingutil.OneFakeDeploymentConfigWithMounts(componentName, "", appName, tt.mountedMap)

			fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &v1.DeploymentConfigList{
					Items: []v1.DeploymentConfig{
						*dcTesting,
					},
				}, nil
			})

			fakeClientSet.Kubernetes.PrependReactor("list", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.returnedPVCs, nil
			})

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLocalConfig := localConfigProvider.NewMockLocalConfigProvider(ctrl)
			mockLocalConfig.EXPECT().GetName().Return(componentName).AnyTimes()
			mockLocalConfig.EXPECT().GetApplication().Return(appName).AnyTimes()
			mockLocalConfig.EXPECT().ListStorage().Return(tt.returnedLocalStorage, nil)

			tt.fields.generic.localConfig = mockLocalConfig
			s := s2iClient{
				generic: tt.fields.generic,
				client:  *fakeClient,
			}
			got, err := s.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantedStorageList) {
				t.Errorf("List() error = %v", pretty.Compare(got, tt.wantedStorageList))
			}
		})
	}
}
