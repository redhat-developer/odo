package storage

import (
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	storageLabels "github.com/redhat-developer/odo/pkg/storage/labels"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/redhat-developer/odo/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func Test_kubernetesClient_ListFromCluster(t *testing.T) {
	type fields struct {
		generic generic
	}
	tests := []struct {
		name                string
		fields              fields
		returnedDeployments *appsv1.DeploymentList
		returnedPVCs        *corev1.PersistentVolumeClaimList
		want                StorageList
		wantErr             bool
	}{
		{
			name: "case 1: should error out for multiple pods returned",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeployment("nodejs"),
					*testingutil.CreateFakeDeployment("nodejs"),
				},
			},
			wantErr: true,
		},
		{
			name: "case 2: pod not found",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{},
			},
			want:    StorageList{},
			wantErr: false,
		},
		{
			name: "case 3: no volume mounts on pod",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeployment("nodejs"),
				},
			},
			want:    StorageList{},
			wantErr: false,
		},
		{
			name: "case 4: two volumes mounted on a single container",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"},
							{Name: "volume-1-vol", MountPath: "/path"},
						}),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
					*testingutil.FakePVC("volume-1", "10Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-1"}),
				},
			},
			want: StorageList{
				Items: []Storage{
					generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), "", "container-0"),
					generateStorage(NewStorage("volume-1", "10Gi", "/path", nil), "", "container-0"),
				},
			},
			wantErr: false,
		},
		{
			name: "case 5: one volume is mounted on a single container and another on both",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"},
							{Name: "volume-1-vol", MountPath: "/path"},
						}),
						testingutil.CreateFakeContainerWithVolumeMounts("container-1", []corev1.VolumeMount{
							{Name: "volume-1-vol", MountPath: "/path"},
						}),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
					*testingutil.FakePVC("volume-1", "10Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-1"}),
				},
			},
			want: StorageList{
				Items: []Storage{
					generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), "", "container-0"),
					generateStorage(NewStorage("volume-1", "10Gi", "/path", nil), "", "container-0"),
					generateStorage(NewStorage("volume-1", "10Gi", "/path", nil), "", "container-1"),
				},
			},
			wantErr: false,
		},
		{
			name: "case 6: pvc for volumeMount not found",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"},
						}),
						testingutil.CreateFakeContainer("container-1"),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
					*testingutil.FakePVC("volume-1", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-1"}),
				},
			},
			wantErr: true,
		},
		{
			name: "case 7: the storage label should be used as the name of the storage",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-nodejs-vol", MountPath: "/data"},
						}),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0-nodejs", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
				},
			},
			want: StorageList{
				Items: []Storage{
					generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), "", "container-0"),
				},
			},
			wantErr: false,
		},
		{
			name: "case 8: no pvc found for mount path",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"},
							{Name: "volume-1-vol", MountPath: "/path"},
						}),
						testingutil.CreateFakeContainerWithVolumeMounts("container-1", []corev1.VolumeMount{
							{Name: "volume-1-vol", MountPath: "/path"},
							{Name: "volume-vol", MountPath: "/path1"},
						}),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
					*testingutil.FakePVC("volume-1", "10Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-1"}),
				},
			},
			want:    StorageList{},
			wantErr: true,
		},
		{
			name: "case 9: avoid the source volume's mount",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"},
							{Name: "volume-1-vol", MountPath: "/path"},
						}),
						testingutil.CreateFakeContainerWithVolumeMounts("container-1", []corev1.VolumeMount{
							{Name: "volume-1-vol", MountPath: "/path"},
							{Name: OdoSourceVolume, MountPath: "/path1"},
						}),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
					*testingutil.FakePVC("volume-1", "10Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-1"}),
				},
			},
			want: StorageList{
				Items: []Storage{
					generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), "", "container-0"),
					generateStorage(NewStorage("volume-1", "10Gi", "/path", nil), "", "container-0"),
					generateStorage(NewStorage("volume-1", "10Gi", "/path", nil), "", "container-1"),
				},
			},
			wantErr: false,
		},
		{
			name: "case 10: avoid the volume mounts used by the init containers",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "nodejs",
							Labels: map[string]string{
								"component": "nodejs",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									InitContainers: []corev1.Container{
										{
											Name: "supervisord",
											VolumeMounts: []corev1.VolumeMount{
												{
													Name:      "odo-shared-project",
													MountPath: "/opt/",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{},
			},
			want:    StorageList{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := kclient.FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("list", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedPVCs, nil
			})

			fakeClientSet.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedDeployments, nil
			})

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLocalConfig := localConfigProvider.NewMockLocalConfigProvider(ctrl)
			mockLocalConfig.EXPECT().GetName().Return(tt.fields.generic.componentName).AnyTimes()
			mockLocalConfig.EXPECT().GetApplication().Return(tt.fields.generic.appName).AnyTimes()

			tt.fields.generic.localConfigProvider = mockLocalConfig

			k := kubernetesClient{
				generic: tt.fields.generic,
				client:  fakeClient,
			}
			got, err := k.ListFromCluster()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFromCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListFromCluster() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_kubernetesClient_List(t *testing.T) {
	type fields struct {
		generic generic
	}
	tests := []struct {
		name                 string
		fields               fields
		want                 StorageList
		wantErr              bool
		returnedLocalStorage []localConfigProvider.LocalStorage
		returnedDeployments  *appsv1.DeploymentList
		returnedPVCs         *corev1.PersistentVolumeClaimList
	}{
		{
			name: "case 1: no volume on devfile and no pod on cluster",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedLocalStorage: []localConfigProvider.LocalStorage{},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{},
			},
			want:    NewStorageList(nil),
			wantErr: false,
		},
		{
			name: "case 2: no volume on devfile and pod",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedLocalStorage: []localConfigProvider.LocalStorage{},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{testingutil.CreateFakeContainer("container-0")},
						[]corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{},
			},
			want:    NewStorageList(nil),
			wantErr: false,
		},
		{
			name: "case 3: same two volumes on cluster and devFile",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedLocalStorage: []localConfigProvider.LocalStorage{
				{
					Name:      "volume-0",
					Size:      "5Gi",
					Path:      "/data",
					Container: "container-0",
				},
				{
					Name:      "volume-1",
					Size:      "10Gi",
					Path:      "/path",
					Container: "container-0",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"},
							{Name: "volume-1-vol", MountPath: "/path"},
						}),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
					*testingutil.FakePVC("volume-1", "10Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-1"}),
				},
			},
			want: NewStorageList([]Storage{
				generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), StateTypePushed, "container-0"),
				generateStorage(NewStorage("volume-1", "10Gi", "/path", nil), StateTypePushed, "container-0"),
			}),
			wantErr: false,
		},
		{
			name: "case 4: both volumes, present on the cluster and devFile, are different",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedLocalStorage: []localConfigProvider.LocalStorage{
				{
					Name:      "volume-0",
					Size:      "5Gi",
					Path:      "/data",
					Container: "container-0",
				},
				{
					Name:      "volume-1",
					Size:      "10Gi",
					Path:      "/path",
					Container: "container-0",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-00-vol", MountPath: "/data"},
							{Name: "volume-11-vol", MountPath: "/path"},
						}),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-00", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-00"}),
					*testingutil.FakePVC("volume-11", "10Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-11"}),
				},
			},
			want: NewStorageList([]Storage{
				generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), StateTypeNotPushed, "container-0"),
				generateStorage(NewStorage("volume-1", "10Gi", "/path", nil), StateTypeNotPushed, "container-0"),
				generateStorage(NewStorage("volume-00", "5Gi", "/data", nil), StateTypeLocallyDeleted, "container-0"),
				generateStorage(NewStorage("volume-11", "10Gi", "/path", nil), StateTypeLocallyDeleted, "container-0"),
			}),
			wantErr: false,
		},
		{
			name: "case 5: two containers with different volumes but one container is not pushed",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedLocalStorage: []localConfigProvider.LocalStorage{
				{
					Name:      "volume-0",
					Size:      "5Gi",
					Path:      "/data",
					Container: "container-0",
				},
				{
					Name:      "volume-1",
					Size:      "10Gi",
					Path:      "/data",
					Container: "container-1",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"},
						}),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
				},
			},
			want: NewStorageList([]Storage{
				generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), StateTypePushed, "container-0"),
				generateStorage(NewStorage("volume-1", "10Gi", "/data", nil), StateTypeNotPushed, "container-1"),
			}),
			wantErr: false,
		},
		{
			name: "case 6: two containers with different volumes on the cluster but one container is deleted locally",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedLocalStorage: []localConfigProvider.LocalStorage{
				{
					Name:      "volume-0",
					Size:      "5Gi",
					Path:      "/data",
					Container: "container-0",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeploymentsWithContainers("nodejs", []corev1.Container{
						testingutil.CreateFakeContainerWithVolumeMounts("container-0", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"},
						}),
						testingutil.CreateFakeContainerWithVolumeMounts("container-1", []corev1.VolumeMount{
							{Name: "volume-0-vol", MountPath: "/data"}},
						),
					}, []corev1.Container{}),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*testingutil.FakePVC("volume-0", "5Gi", map[string]string{"component": "nodejs", storageLabels.DevfileStorageLabel: "volume-0"}),
				},
			},
			want: NewStorageList([]Storage{
				generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), StateTypePushed, "container-0"),
				generateStorage(NewStorage("volume-0", "5Gi", "/data", nil), StateTypeLocallyDeleted, "container-1"),
			}),
			wantErr: false,
		},
		{
			name: "case 7: multiple pods are present on the cluster",
			fields: fields{
				generic: generic{
					componentName: "nodejs",
					appName:       "app",
				},
			},
			returnedLocalStorage: []localConfigProvider.LocalStorage{
				{
					Name:      "volume-0",
					Size:      "5Gi",
					Path:      "/data",
					Container: "container-0",
				},
			},
			returnedDeployments: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					*testingutil.CreateFakeDeployment("nodejs"),
					*testingutil.CreateFakeDeployment("nodejs"),
				},
			},
			returnedPVCs: &corev1.PersistentVolumeClaimList{},
			want:         StorageList{},
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := kclient.FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("list", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedPVCs, nil
			})

			fakeClientSet.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedDeployments, nil
			})

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLocalConfig := localConfigProvider.NewMockLocalConfigProvider(ctrl)
			mockLocalConfig.EXPECT().GetName().Return(tt.fields.generic.componentName).AnyTimes()
			mockLocalConfig.EXPECT().GetApplication().Return(tt.fields.generic.appName).AnyTimes()
			mockLocalConfig.EXPECT().ListStorage().Return(tt.returnedLocalStorage, nil)
			mockLocalConfig.EXPECT().Exists().Return(true)
			tt.fields.generic.localConfigProvider = mockLocalConfig

			k := kubernetesClient{
				generic: tt.fields.generic,
				client:  fakeClient,
			}
			got, err := k.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_kubernetesClient_Create(t *testing.T) {
	type fields struct {
		generic generic
	}
	type args struct {
		storage Storage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "case 1: valid storage",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			args: args{
				storage: NewStorageWithContainer("storage-0", "5Gi", "/data", "runtime", util.GetBoolPtr(false)),
			},
		},
		{
			name: "case 2: invalid storage size",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			args: args{
				storage: NewStorageWithContainer("storage-0", "example", "/data", "runtime", util.GetBoolPtr(false)),
			},
			wantErr: true,
		},
		{
			name: "case 3: valid odo project related storage",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			args: args{
				storage: NewStorageWithContainer("odo-projects-vol", "5Gi", "/data", "runtime", util.GetBoolPtr(false)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := kclient.FakeNew()

			k := kubernetesClient{
				generic: tt.fields.generic,
				client:  fkclient,
			}
			if err := k.Create(tt.args.storage); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == true {
				return
			}

			// Check for validating actions performed
			if (len(fkclientset.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in CreatePVC got: %v", fkclientset.Kubernetes.Actions())
				return
			}

			createdPVC := fkclientset.Kubernetes.Actions()[0].(ktesting.CreateAction).GetObject().(*corev1.PersistentVolumeClaim)
			quantity, err := resource.ParseQuantity(tt.args.storage.Spec.Size)
			if err != nil {
				t.Errorf("failed to create quantity by calling resource.ParseQuantity(%v)", tt.args.storage.Spec.Size)
			}

			wantLabels := storageLabels.GetLabels(tt.args.storage.Name, tt.fields.generic.componentName, tt.fields.generic.appName, true)
			wantLabels["component"] = tt.fields.generic.componentName
			wantLabels[storageLabels.DevfileStorageLabel] = tt.args.storage.Name

			if strings.Contains(tt.args.storage.Name, OdoSourceVolume) {
				// Add label for source pvc
				wantLabels[storageLabels.SourcePVCLabel] = tt.args.storage.Name
			}

			// created PVC should be labeled with labels passed to CreatePVC
			if !reflect.DeepEqual(createdPVC.Labels, wantLabels) {
				t.Errorf("labels in created pvc is not matching expected labels, expected: %v, got: %v", wantLabels, createdPVC.Labels)
			}
			// name, size of createdPVC should be matching to size, name passed to CreatePVC
			if !reflect.DeepEqual(createdPVC.Spec.Resources.Requests["storage"], quantity) {
				t.Errorf("size of PVC is not matching to expected size, expected: %v, got %v", quantity, createdPVC.Spec.Resources.Requests["storage"])
			}

			wantedPVCName, err := generatePVCName(tt.args.storage.Name, tt.fields.generic.componentName, tt.fields.generic.appName)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(createdPVC.Name, wantedPVCName) {
				t.Errorf("name of the PVC is not matching to expected name, expected: %v, got %v", wantedPVCName, createdPVC.Name)
			}
		})
	}
}

func Test_kubernetesClient_Delete(t *testing.T) {
	pvcName := "pvc-0"
	returnedPVCs := corev1.PersistentVolumeClaimList{
		Items: []corev1.PersistentVolumeClaim{
			*testingutil.FakePVC(pvcName, "5Gi", getStorageLabels("storage-0", "nodejs", "app")),
		},
	}

	type fields struct {
		generic generic
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "case 1: delete successful",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			args: args{
				"storage-0",
			},
			wantErr: false,
		},
		{
			name: "case 2: pvc not found",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: "nodejs",
				},
			},
			args: args{
				"storage-example",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := kclient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("list", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &returnedPVCs, nil
			})

			fkclientset.Kubernetes.PrependReactor("delete", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				if action.(ktesting.DeleteAction).GetName() != pvcName {
					t.Errorf("delete called with = %v, want %v", action.(ktesting.DeleteAction).GetName(), pvcName)
				}
				return true, nil, nil
			})

			k := kubernetesClient{
				generic: tt.fields.generic,
				client:  fkclient,
			}
			if err := k.Delete(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(fkclientset.Kubernetes.Actions()) != 2 {
				t.Errorf("expected 2 action, got %v", len(fkclientset.Kubernetes.Actions()))
			}
		})
	}
}
