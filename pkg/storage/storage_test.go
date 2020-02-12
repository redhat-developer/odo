package storage

import (
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/pkg/version"

	v1 "github.com/openshift/api/apps/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentLabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/storage/labels"
	storageLabels "github.com/openshift/odo/pkg/storage/labels"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func getStorageLabels(storageName, componentName, applicationName string) map[string]string {
	return storageLabels.GetLabels(storageName, componentName, applicationName, true)
}

func Test_GetStorageFromPVC(t *testing.T) {
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

func TestGetMachineReadableFormat(t *testing.T) {
	tests := []struct {
		name         string
		storageName  string
		storageSize  string
		mountedPath  string
		activeStatus bool
		want         Storage
	}{
		{
			name:         "test case 1: with a pvc, valid path and mounted status",
			storageName:  "pvc-example",
			storageSize:  "100Mi",
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
			name:         "test case 2: with a pvc, empty path and unmounted status",
			storageName:  "pvc-example",
			storageSize:  "500Mi",
			mountedPath:  "",
			activeStatus: false,
			want: Storage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc-example",
				},
				TypeMeta: metav1.TypeMeta{Kind: "storage", APIVersion: "odo.openshift.io/v1alpha1"},
				Spec: StorageSpec{
					Size: "500Mi",
				},
				Status: StorageStatus{
					Path: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStorage := GetMachineReadableFormat(tt.storageName, tt.storageSize, tt.mountedPath)
			if !reflect.DeepEqual(tt.want, gotStorage) {
				t.Errorf("the returned storage is different, expected: %v, got: %v", tt.want, gotStorage)
			}
		})
	}
}

func TestGetMachineReadableFormatForList(t *testing.T) {

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
			gotStorage := GetMachineReadableFormatForList(tt.inputStorage)
			if !reflect.DeepEqual(tt.want, gotStorage) {
				t.Errorf("the returned storage is different, expected: %v, got: %v", tt.want, gotStorage)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		name            string
		size            string
		componentName   string
		applicationName string
	}
	tests := []struct {
		name       string
		args       args
		wantLabels map[string]string
		wantErr    bool
	}{
		{
			name: "Case 1: With valid values",
			args: args{
				name:            "storage-0",
				size:            "100Mi",
				componentName:   "nodejs-ex",
				applicationName: "app-ex",
			},
			wantLabels: map[string]string{
				"app":                          "app-ex",
				labels.StorageLabel:            "storage-0",
				applabels.ApplicationLabel:     "app-ex",
				applabels.OdoManagedBy:         "odo",
				applabels.OdoVersion:           version.VERSION,
				componentLabels.ComponentLabel: "nodejs-ex",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := occlient.FakeNew()

			createdPVC, err := Create(fkclient, tt.args.name, tt.args.size, tt.args.componentName, tt.args.applicationName)

			// Check for validating actions performed
			if (len(fkclientset.Kubernetes.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in CreatePVC got: %v", fkclientset.RouteClientset.Actions())
			}

			// Checks for return values in positive cases
			if err == nil {
				quantity, err := resource.ParseQuantity(tt.args.size)
				if err != nil {
					t.Errorf("failed to create quantity by calling resource.ParseQuantity(%v)", tt.args.size)
				}

				// created PVC should be labeled with labels passed to CreatePVC
				if !reflect.DeepEqual(createdPVC.Labels, tt.wantLabels) {
					t.Errorf("labels in created route is not matching expected labels, expected: %v, got: %v", tt.wantLabels, createdPVC.Labels)
				}
				// name, size of createdPVC should be matching to size, name passed to CreatePVC
				if !reflect.DeepEqual(createdPVC.Spec.Resources.Requests["storage"], quantity) {
					t.Errorf("size of PVC is not matching to expected size, expected: %v, got %v", quantity, createdPVC.Spec.Resources.Requests["storage"])
				}
			}
		})
	}
}

func TestList(t *testing.T) {

	mountMap := make(map[string]*corev1.PersistentVolumeClaim)

	pvc1 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-1-app"), "100Mi", getStorageLabels("storage-1", "nodejs", "app"))
	mountMap["data"] = pvc1

	pvc2 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-2-app"), "500Mi", getStorageLabels("storage-2", "nodejs", "app"))
	mountMap["data-1"] = pvc2

	// pvc mounted to different app
	pvc3 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-3-app"), "100Mi", getStorageLabels("storage-3", "wildfly", "app"))

	// unMounted pvc
	pvc4 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-4-app"), "100Mi", getStorageLabels("storage-4", "", "app"))
	delete(pvc4.Labels, componentLabels.ComponentLabel)

	// mounted to the deploymentConfig but not returned and thus doesn't exist on the cluster
	pvc5 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-5-app"), "100Mi", getStorageLabels("storage-5", "nodejs", "app"))

	type args struct {
		componentName   string
		applicationName string
	}
	tests := []struct {
		name              string
		args              args
		componentType     string
		returnedPVCs      corev1.PersistentVolumeClaimList
		mountedMap        map[string]*corev1.PersistentVolumeClaim
		wantedStorageList StorageList
		wantErr           bool
	}{
		{
			name: "case 1: no error and all PVCs mounted",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*pvc1, *pvc2,
				},
			},
			mountedMap: map[string]*corev1.PersistentVolumeClaim{
				"/data":   pvc1,
				"/data-1": pvc2,
			},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{
				GetMachineReadableFormat("storage-1", "100Mi", "/data"),
				GetMachineReadableFormat("storage-2", "500Mi", "/data-1"),
			}),
			wantErr: false,
		},
		{
			name: "case 2: no error and two PVCs mounted and one mounted to a different app",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*pvc1, *pvc2, *pvc3,
				},
			},
			mountedMap: map[string]*corev1.PersistentVolumeClaim{
				"/data":   pvc1,
				"/data-1": pvc2,
			},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{
				GetMachineReadableFormat("storage-1", "100Mi", "/data"),
				GetMachineReadableFormat("storage-2", "500Mi", "/data-1"),
			}),
			wantErr: false,
		},
		{
			name: "case 3: no error and two PVCs mounted and one unmounted",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*pvc1, *pvc2, *pvc4,
				},
			},
			mountedMap: map[string]*corev1.PersistentVolumeClaim{
				"/data":   pvc1,
				"/data-1": pvc2,
			},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{
				GetMachineReadableFormat("storage-1", "100Mi", "/data"),
				GetMachineReadableFormat("storage-2", "500Mi", "/data-1"),
				GetMachineReadableFormat("storage-4", "100Mi", ""),
			}),
			wantErr: false,
		},
		{
			name: "case 4: pvc mounted but doesn't exist on cluster",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*pvc1, *pvc2,
				},
			},
			mountedMap: map[string]*corev1.PersistentVolumeClaim{
				"/data":   pvc1,
				"/data-1": pvc5,
			},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{
				GetMachineReadableFormat("storage-1", "100Mi", "/data"),
				GetMachineReadableFormat("storage-2", "500Mi", "/data-1"),
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := occlient.FakeNew()

			dcTesting := testingutil.OneFakeDeploymentConfigWithMounts(tt.args.componentName, tt.componentType, tt.args.applicationName, tt.mountedMap)

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

			storageList, err := List(fakeClient, tt.args.componentName, tt.args.applicationName)
			if err == nil && !tt.wantErr {
				if !reflect.DeepEqual(storageList, tt.wantedStorageList) {
					t.Errorf("storageList not equal, expected: %v, got: %v", tt.wantedStorageList, storageList.Items)
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})
	}
}

func TestListMounted(t *testing.T) {
	mountMap := make(map[string]*corev1.PersistentVolumeClaim)

	pvc1 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-1-app"), "100Mi", getStorageLabels("storage-1", "nodejs", "app"))
	mountMap["data"] = pvc1

	pvc2 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-2-app"), "500Mi", getStorageLabels("storage-2", "nodejs", "app"))
	mountMap["data-1"] = pvc2

	// unMounted pvc
	pvc3 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-3-app"), "100Mi", getStorageLabels("storage-3", "", "app"))
	delete(pvc3.Labels, componentLabels.ComponentLabel)

	type args struct {
		componentName   string
		applicationName string
	}
	tests := []struct {
		name              string
		args              args
		componentType     string
		returnedPVCs      corev1.PersistentVolumeClaimList
		mountedMap        map[string]*corev1.PersistentVolumeClaim
		wantedStorageList StorageList
		wantErr           bool
	}{
		{
			name: "case 1: no error and all PVCs mounted",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*pvc1, *pvc2,
				},
			},
			mountedMap: map[string]*corev1.PersistentVolumeClaim{
				"/data":   pvc1,
				"/data-1": pvc2,
			},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{
				GetMachineReadableFormat("storage-1", "100Mi", "/data"),
				GetMachineReadableFormat("storage-2", "500Mi", "/data-1"),
			}),
			wantErr: false,
		},
		{
			name: "case 3: no error and two PVCs mounted and one unmounted",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*pvc1, *pvc2, *pvc3,
				},
			},
			mountedMap: map[string]*corev1.PersistentVolumeClaim{
				"/data":   pvc1,
				"/data-1": pvc2,
			},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{
				GetMachineReadableFormat("storage-1", "100Mi", "/data"),
				GetMachineReadableFormat("storage-2", "500Mi", "/data-1"),
			}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := occlient.FakeNew()

			dcTesting := testingutil.OneFakeDeploymentConfigWithMounts(tt.args.componentName, tt.componentType, tt.args.applicationName, tt.mountedMap)

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

			storageList, err := ListMounted(fakeClient, tt.args.componentName, tt.args.applicationName)
			if err == nil && !tt.wantErr {
				if !reflect.DeepEqual(storageList, tt.wantedStorageList) {
					t.Errorf("storageList not equal, expected: %v, got: %v", tt.wantedStorageList, storageList.Items)
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})

	}

}

func TestPush(t *testing.T) {
	mountMap := make(map[string]*corev1.PersistentVolumeClaim)

	pvc1 := testingutil.FakePVC(generatePVCNameFromStorageName("backend-app"), "100Mi", getStorageLabels("backend", "nodejs", "app"))
	mountMap["data"] = pvc1

	pvc2 := testingutil.FakePVC(generatePVCNameFromStorageName("backend-1-app"), "500Mi", getStorageLabels("backend-1", "nodejs", "app"))
	mountMap["data-1"] = pvc2

	pvc3 := testingutil.FakePVC(generatePVCNameFromStorageName("backend-2-app"), "100Mi", getStorageLabels("backend-2", "nodejs", "app"))
	mountMap["data-2"] = pvc3

	type args struct {
		storageList       StorageList
		componentName     string
		applicationName   string
		isComponentExists bool
	}
	tests := []struct {
		name             string
		args             args
		componentType    string
		returnedPVCs     corev1.PersistentVolumeClaimList
		dcMountedMap     map[string]*corev1.PersistentVolumeClaim
		storageToMount   map[string]*corev1.PersistentVolumeClaim
		storageToUnMount map[string]string
		wantErr          bool
	}{
		{
			name: "case 1: component does not exist, no pvc on cluster and two on config",
			args: args{
				storageList: StorageList{
					Items: []Storage{
						GetMachineReadableFormat("backend", "100Mi", "data"),
						GetMachineReadableFormat("backend-1", "500Mi", "data-1"),
					},
				},
				componentName:     "nodejs",
				applicationName:   "app",
				isComponentExists: false,
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}},
			dcMountedMap: map[string]*corev1.PersistentVolumeClaim{},
			storageToMount: map[string]*corev1.PersistentVolumeClaim{
				"data":   pvc1,
				"data-1": pvc2,
			},
			storageToUnMount: map[string]string{},
			wantErr:          false,
		},
		{
			name: "case 2: component exists, two pvc on cluster and none on config",
			args: args{
				storageList:       StorageList{Items: []Storage{}},
				componentName:     "nodejs",
				applicationName:   "app",
				isComponentExists: true,
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{*pvc1, *pvc2}},
			dcMountedMap: map[string]*corev1.PersistentVolumeClaim{
				"data":   pvc1,
				"data-1": pvc2,
			},
			storageToMount: map[string]*corev1.PersistentVolumeClaim{},
			storageToUnMount: map[string]string{
				"data":   getStorageFromPVC(pvc1),
				"data-1": getStorageFromPVC(pvc2),
			},
			wantErr: false,
		},
		{
			name: "case 3: component exists, three PVCs, two in config and cluster but one not in config",
			args: args{
				storageList: StorageList{
					Items: []Storage{
						GetMachineReadableFormat("backend", "100Mi", "data"),
						GetMachineReadableFormat("backend-1", "500Mi", "data-1"),
					},
				},
				componentName:     "nodejs",
				applicationName:   "app",
				isComponentExists: true,
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{*pvc1, *pvc2, *pvc3}},
			dcMountedMap: map[string]*corev1.PersistentVolumeClaim{
				"data":   pvc1,
				"data-1": pvc2,
				"data-2": pvc3,
			},
			storageToMount: map[string]*corev1.PersistentVolumeClaim{},
			storageToUnMount: map[string]string{
				"data-2": getStorageFromPVC(pvc3),
			},
			wantErr: false,
		},
		{
			name: "case 4: component exists, three PVCs, one in config and cluster, one not in cluster and one not in config",
			args: args{
				storageList: StorageList{Items: []Storage{
					GetMachineReadableFormat("backend", "100Mi", "data"),
					GetMachineReadableFormat("backend-1", "500Mi", "data-1"),
				}},
				componentName:     "nodejs",
				applicationName:   "app",
				isComponentExists: true,
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{*pvc1, *pvc3}},
			dcMountedMap: map[string]*corev1.PersistentVolumeClaim{
				"data":   pvc1,
				"data-2": pvc3,
			},
			storageToMount: map[string]*corev1.PersistentVolumeClaim{
				"data-1": pvc2,
			},
			storageToUnMount: map[string]string{
				"data-2": getStorageFromPVC(pvc3),
			},
			wantErr: false,
		},
		{
			name: "case 5: component exists, two PVCs, both on cluster and config, but one with path config mismatch",
			args: args{
				storageList: StorageList{
					Items: []Storage{
						GetMachineReadableFormat("backend", "100Mi", "data"),
						GetMachineReadableFormat("backend-1", "500Mi", "data-100"),
					},
				},
				componentName:     "nodejs",
				applicationName:   "app",
				isComponentExists: true,
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{*pvc1, *pvc2}},
			dcMountedMap: map[string]*corev1.PersistentVolumeClaim{
				"data":   pvc1,
				"data-1": pvc2,
			},
			storageToMount:   map[string]*corev1.PersistentVolumeClaim{},
			storageToUnMount: map[string]string{},
			wantErr:          true,
		},
		{
			name: "case 6: component exists, two PVCs, both on cluster and config, but one with size config mismatch",
			args: args{
				storageList: StorageList{
					Items: []Storage{
						GetMachineReadableFormat("backend", "100Mi", "data"),
						GetMachineReadableFormat("backend-1", "50Mi", "data-1"),
					},
				},
				componentName:     "nodejs",
				applicationName:   "app",
				isComponentExists: true,
			},
			returnedPVCs: corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{*pvc1, *pvc2}},
			dcMountedMap: map[string]*corev1.PersistentVolumeClaim{
				"data":   pvc1,
				"data-1": pvc2,
			},
			storageToMount:   map[string]*corev1.PersistentVolumeClaim{},
			storageToUnMount: map[string]string{},
			wantErr:          true,
		},
		{
			name: "case 7: component exists, one pvc on config and cluster",
			args: args{
				storageList:       StorageList{Items: []Storage{}},
				componentName:     "nodejs",
				applicationName:   "app",
				isComponentExists: true,
			},
			returnedPVCs:     corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}},
			dcMountedMap:     map[string]*corev1.PersistentVolumeClaim{},
			storageToMount:   map[string]*corev1.PersistentVolumeClaim{},
			storageToUnMount: map[string]string{},
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := occlient.FakeNew()

			dcTesting := testingutil.OneFakeDeploymentConfigWithMounts(tt.args.componentName, tt.componentType, tt.args.applicationName, tt.dcMountedMap)

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

			fakeClientSet.Kubernetes.PrependReactor("delete", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				for _, storageName := range tt.storageToUnMount {
					namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(storageName, tt.args.applicationName)
					if err != nil {
						return false, nil, nil
					}
					if generatePVCNameFromStorageName(namespacedOpenShiftObject) == action.(ktesting.DeleteAction).GetName() {
						return true, nil, nil
					}
				}
				return false, nil, nil
			})

			storageToMount, storageToUnmount, err := Push(fakeClient, tt.args.storageList, tt.args.componentName, tt.args.applicationName, tt.args.isComponentExists)

			if err == nil && !tt.wantErr {
				// check if the len of the storageToMount values are the same as the required ones
				if len(storageToMount) != len(tt.storageToMount) {
					t.Errorf("storageToMount value mismatch, expected: %v, got: %v", len(tt.storageToMount), len(storageToMount))
				}

				// check if the PVCs are created with the required specs and will be mounted to the required path
				var createdPVCs []*corev1.PersistentVolumeClaim
				for _, action := range fakeClientSet.Kubernetes.Actions() {
					if _, ok := action.(ktesting.CreateAction); !ok {
						continue
					}
					createdPVCs = append(createdPVCs, action.(ktesting.CreateAction).GetObject().(*corev1.PersistentVolumeClaim))
				}

				for _, pvc := range storageToMount {
					found := false
					for _, createdPVC := range createdPVCs {
						if pvc.Name == createdPVC.Name {
							found = true
							createdPVCSize := createdPVC.Spec.Resources.Requests[corev1.ResourceStorage]
							pvcSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
							if pvcSize.String() != createdPVCSize.String() {
								t.Errorf("pvc with name %v created with wrong size, expected: %v, got %v", pvc.Name, pvcSize, createdPVCSize)
							}
						}
					}

					if !found {
						t.Errorf("pvc with name %v not created", pvc.Name)
					}

				}

				for pathResult, pvcResult := range storageToMount {
					for path, pvc := range tt.storageToMount {
						if pvc.Name == pvcResult.Name {
							if path != pathResult {
								t.Errorf("pvc mounted to wrong path, expected: %v, got: %v", path, pathResult)
							}
						}
					}
				}

				// check if the storageToUnMounted values match the required ones
				if !reflect.DeepEqual(storageToUnmount, tt.storageToUnMount) {
					t.Errorf("storageToUnmount is different, expected: %v, got: %v", tt.storageToUnMount, storageToUnmount)
				}

			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})

	}
}

func TestListStorageWithState(t *testing.T) {
	componentName := "nodejs"
	appName := "app"

	pvc1 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-1-app"), "100Mi", getStorageLabels("storage-1", componentName, appName))

	pvc3 := testingutil.FakePVC(generatePVCNameFromStorageName("storage-3-app"), "100Mi", getStorageLabels("storage-3", componentName, appName))

	storage1 := GetMachineReadableFormat("storage-1", "100Mi", "/tmp1")
	storage1.State = StateTypePushed

	storage2 := GetMachineReadableFormat("storage-2", "100Mi", "/tmp2")
	storage2.State = StateTypeNotPushed

	storage3 := GetMachineReadableFormat("storage-3", "100Mi", "/tmp3")
	storage3.State = StateTypeLocallyDeleted

	tests := []struct {
		name              string
		localConfig       config.LocalConfigInfo
		returnedPVCs      corev1.PersistentVolumeClaimList
		mountedMap        map[string]*corev1.PersistentVolumeClaim
		wantedStorageList StorageList
		wantErr           bool
	}{
		{
			name:        "case 1: Storage is Pushed",
			localConfig: config.GetOneExistingConfigInfoStorage(componentName, appName, "test", "storage-1", "100Mi", "/tmp1"),
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					*pvc1,
				},
			},
			mountedMap: map[string]*corev1.PersistentVolumeClaim{
				"/tmp1": pvc1,
			},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{storage1}),
			wantErr:           false,
		},
		{
			name: "case 1: Storage is Not Pushed",
			// Storage present in local conf
			localConfig: config.GetOneExistingConfigInfoStorage(componentName, appName, "test", "storage-2", "100Mi", "/tmp2"),
			// Return PVC's from cluster
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{},
			},
			mountedMap:        map[string]*corev1.PersistentVolumeClaim{},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{storage2}),
			wantErr:           false,
		},

		{

			name: "case 3: Storage is Locally Deleted",
			// storage-1 present in local conf
			localConfig: config.GetOneExistingConfigInfoStorage(componentName, appName, "test", "storage-1", "100Mi", "/tmp1"),
			// pvc1, pvc3 present in cluster
			returnedPVCs: corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{*pvc1, *pvc3},
			},
			mountedMap:        map[string]*corev1.PersistentVolumeClaim{"/tmp1": pvc1, "/tmp3": pvc3},
			wantedStorageList: GetMachineReadableFormatForList([]Storage{storage1, storage3}),
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

			storageList, err := ListStorageWithState(fakeClient, &tt.localConfig, componentName, appName)
			if err == nil && !tt.wantErr {
				if !reflect.DeepEqual(storageList, tt.wantedStorageList) {
					t.Errorf("storageList not equal, expected: %v, got: %v", tt.wantedStorageList, storageList.Items)
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})
	}
}
