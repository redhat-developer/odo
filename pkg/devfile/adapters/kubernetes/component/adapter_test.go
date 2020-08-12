package component

import (
	"testing"

	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	"github.com/google/go-cmp/cmp"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"

	"github.com/spf13/afero"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"

	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimeUnstructured "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	dynamicfakeclient "k8s.io/client-go/dynamic/fake"
)

func NewFilesystem() afero.Fs {
	return afero.NewOsFs()
}

func TestCreateOrUpdateComponent(t *testing.T) {

	testComponentName := "test"
	deployment := v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       kclient.DeploymentKind,
			APIVersion: kclient.DeploymentAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: testComponentName,
		},
	}

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		envInfo       envinfo.EnvSpecificInfo
		running       bool
		wantErr       bool
	}{
		{
			name:          "Case 1: Invalid devfile",
			componentType: "",
			envInfo:       envinfo.EnvSpecificInfo{},
			running:       false,
			wantErr:       true,
		},
		{
			name:          "Case 2: Valid devfile",
			componentType: versionsCommon.ContainerComponentType,
			envInfo:       envinfo.EnvSpecificInfo{},
			running:       false,
			wantErr:       false,
		},
		{
			name:          "Case 3: Invalid devfile, already running component",
			componentType: "",
			envInfo:       envinfo.EnvSpecificInfo{},
			running:       true,
			wantErr:       true,
		},
		{
			name:          "Case 4: Valid devfile, already running component",
			componentType: versionsCommon.ContainerComponentType,
			envInfo:       envinfo.EnvSpecificInfo{},
			running:       true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var comp versionsCommon.DevfileComponent
			if tt.componentType != "" {
				comp = testingutil.GetFakeContainerComponent("component")
			}
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components:   []versionsCommon.DevfileComponent{comp},
					ExecCommands: []versionsCommon.Exec{getExecCommand("run", versionsCommon.RunCommandGroupType)},
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := kclient.FakeNew()

			if tt.running {
				fkclientset.Kubernetes.PrependReactor("update", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, &deployment, nil
				})

				fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, &deployment, nil
				})
			}

			componentAdapter := New(adapterCtx, *fkclient)
			err := componentAdapter.createOrUpdateComponent(tt.running, tt.envInfo)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestGetFirstContainerWithSourceVolume(t *testing.T) {
	tests := []struct {
		name           string
		containers     []corev1.Container
		want           string
		wantSourcePath string
		wantErr        bool
	}{
		{
			name: "Case: One container, no volumes",
			containers: []corev1.Container{
				{
					Name: "test",
				},
			},
			want:           "",
			wantSourcePath: "",
			wantErr:        true,
		},
		{
			name: "Case: One container, no source volume",
			containers: []corev1.Container{
				{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
					},
				},
			},
			want:           "",
			wantSourcePath: "",
			wantErr:        true,
		},
		{
			name: "Case: One container, source volume",
			containers: []corev1.Container{
				{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      kclient.OdoSourceVolume,
							MountPath: kclient.OdoSourceVolumeMount,
						},
					},
				},
			},
			want:           "test",
			wantSourcePath: kclient.OdoSourceVolumeMount,
			wantErr:        false,
		},
		{
			name: "Case: One container, multiple volumes",
			containers: []corev1.Container{
				{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
						{
							Name:      kclient.OdoSourceVolume,
							MountPath: kclient.OdoSourceVolumeMount,
						},
					},
				},
			},
			want:           "test",
			wantSourcePath: kclient.OdoSourceVolumeMount,
			wantErr:        false,
		},
		{
			name: "Case: Multiple containers, no source volumes",
			containers: []corev1.Container{
				{
					Name: "test",
				},
				{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
					},
				},
			},
			want:           "",
			wantSourcePath: "",
			wantErr:        true,
		},
		{
			name: "Case: Multiple containers, multiple volumes",
			containers: []corev1.Container{
				{
					Name: "test",
				},
				{
					Name: "container-two",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
						{
							Name:      kclient.OdoSourceVolume,
							MountPath: kclient.OdoSourceVolumeMount,
						},
					},
				},
			},
			want:           "container-two",
			wantSourcePath: kclient.OdoSourceVolumeMount,
			wantErr:        false,
		},
		{
			name: "Case: Multiple volumes, different source volume path",
			containers: []corev1.Container{
				{
					Name: "test",
				},
				{
					Name: "container-two",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
						{
							Name:      kclient.OdoSourceVolume,
							MountPath: "/some/path",
						},
					},
				},
			},
			want:           "container-two",
			wantSourcePath: "/some/path",
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		container, sourcePath, err := getFirstContainerWithSourceVolume(tt.containers)
		if container != tt.want {
			t.Errorf("expected %s, actual %s", tt.want, container)
		}

		if sourcePath != tt.wantSourcePath {
			t.Errorf("expected %s, actual %s", tt.wantSourcePath, sourcePath)
		}
		if !tt.wantErr == (err != nil) {
			t.Errorf("expected %v, actual %v", tt.wantErr, err)
		}
	}
}

func TestDoesComponentExist(t *testing.T) {

	tests := []struct {
		name             string
		componentType    versionsCommon.DevfileComponentType
		componentName    string
		getComponentName string
		envInfo          envinfo.EnvSpecificInfo
		want             bool
		wantErr          bool
	}{
		{
			name:             "Case 1: Valid component name",
			componentName:    "test-name",
			getComponentName: "test-name",
			envInfo:          envinfo.EnvSpecificInfo{},
			want:             true,
			wantErr:          false,
		},
		{
			name:             "Case 2: Non-existent component name",
			componentName:    "test-name",
			getComponentName: "fake-component",
			envInfo:          envinfo.EnvSpecificInfo{},
			want:             false,
			wantErr:          false,
		},
		{
			name:             "Case 3: Error condition",
			componentName:    "test-name",
			getComponentName: "test-name",
			envInfo:          envinfo.EnvSpecificInfo{},
			want:             false,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components:   []versionsCommon.DevfileComponent{testingutil.GetFakeContainerComponent("component")},
					ExecCommands: []versionsCommon.Exec{getExecCommand("run", versionsCommon.RunCommandGroupType)},
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := kclient.FakeNew()
			fkWatch := watch.NewFake()

			fkclientset.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			// DoesComponentExist requires an already started component, so start it.
			componentAdapter := New(adapterCtx, *fkclient)
			err := componentAdapter.createOrUpdateComponent(false, tt.envInfo)

			// Checks for unexpected error cases
			if err != nil {
				t.Errorf("component adapter start unexpected error %v", err)
			}

			fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				emptyDeployment := testingutil.CreateFakeDeployment("")
				deployment := testingutil.CreateFakeDeployment(tt.getComponentName)

				if tt.wantErr {
					return true, emptyDeployment, errors.Errorf("deployment get error")
				} else if tt.getComponentName == tt.componentName {
					return true, deployment, nil
				}

				return true, emptyDeployment, kerrors.NewNotFound(schema.GroupResource{}, "")
			})

			// Verify that a component with the specified name exists
			componentExists, err := componentAdapter.DoesComponentExist(tt.getComponentName)
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if !tt.wantErr && componentExists != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, componentExists)
			}

		})
	}

}

func TestWaitAndGetComponentPod(t *testing.T) {

	testComponentName := "test"

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		status        corev1.PodPhase
		wantErr       bool
	}{
		{
			name:    "Case 1: Running",
			status:  corev1.PodRunning,
			wantErr: false,
		},
		{
			name:    "Case 2: Failed pod",
			status:  corev1.PodFailed,
			wantErr: true,
		},
		{
			name:    "Case 3: Unknown pod",
			status:  corev1.PodUnknown,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components: []versionsCommon.DevfileComponent{testingutil.GetFakeContainerComponent("component")},
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := kclient.FakeNew()
			fkWatch := watch.NewFake()

			// Change the status
			go func() {
				fkWatch.Modify(kclient.FakePodStatus(tt.status, testComponentName))
			}()

			fkclientset.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			componentAdapter := New(adapterCtx, *fkclient)
			_, err := componentAdapter.waitAndGetComponentPod(false)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestAdapterDelete(t *testing.T) {
	type args struct {
		labels map[string]string
	}

	emptyPods := &corev1.PodList{
		Items: []corev1.Pod{},
	}

	tests := []struct {
		name            string
		args            args
		existingPod     *corev1.PodList
		componentName   string
		componentExists bool
		wantErr         bool
	}{
		{
			name: "case 1: component exists and given labels are valid",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*testingutil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "component",
			componentExists: true,
			wantErr:         false,
		},
		{
			name: "case 2: component exists and given labels are not valid",
			args: args{labels: nil},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*testingutil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "component",
			componentExists: true,
			wantErr:         true,
		},
		{
			name: "case 3: component doesn't exists",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*testingutil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "nocomponent",
			componentExists: false,
			wantErr:         false,
		},
		{
			name: "case 4: resource forbidden",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*testingutil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "resourceforbidden",
			componentExists: false,
			wantErr:         false,
		},
		{
			name: "case 5: component check error",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			existingPod: &corev1.PodList{
				Items: []corev1.Pod{
					*testingutil.CreateFakePod("component", "component"),
				},
			},
			componentName:   "componenterror",
			componentExists: true,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					// ComponentType: "nodejs",
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			if !tt.componentExists {
				adapterCtx.ComponentName = "doesNotExists"
			}

			fkclient, fkclientset := kclient.FakeNew()

			a := New(adapterCtx, *fkclient)

			fkclientset.Kubernetes.PrependReactor("delete-collection", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				if util.ConvertLabelsToSelector(tt.args.labels) != action.(ktesting.DeleteCollectionAction).GetListRestrictions().Labels.String() {
					return true, nil, errors.Errorf("collection labels are not matching, wanted: %v, got: %v", util.ConvertLabelsToSelector(tt.args.labels), action.(ktesting.DeleteCollectionAction).GetListRestrictions().Labels.String())
				}
				return true, nil, nil
			})

			fkclientset.Kubernetes.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				if tt.componentName == "nocomponent" {
					return true, emptyPods, &kclient.PodNotFoundError{Selector: "somegarbage"}
				} else if tt.componentName == "resourceforbidden" {
					return true, emptyPods, kerrors.NewForbidden(schema.GroupResource{}, "", nil)
				} else if tt.componentName == "componenterror" {
					return true, emptyPods, errors.Errorf("pod check error")
				}
				return true, tt.existingPod, nil
			})

			if err := a.Delete(tt.args.labels, false); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateDockerConfigSecret(t *testing.T) {
	testConfigJsonString := "{ \"auths\" : { \"https://index.docker.io/v1/\": { \"auth\": \"test-auth-token\", \"email\": \"test-email\"} },\"HttpHeaders\": {	\"User-Agent\": \"Docker-Client/19.03.8 (darwin)\"},\"experimental\": \"disabled\"}"
	testDockerConfigData := []byte(testConfigJsonString)
	testFilename := "test-data/test-config-json"
	testSecretName := "test-secret"
	testNs := "test-namespace"

	testNamespacedName := types.NamespacedName{
		Name:      testSecretName,
		Namespace: testNs,
	}

	testSecret := &corev1.Secret{

		TypeMeta:   TypeMeta("Secret", "v1"),
		ObjectMeta: SecretObjectMeta(testNamespacedName),
		Type:       corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: testDockerConfigData,
		},
	}

	testSecretData, err := runtimeUnstructured.DefaultUnstructuredConverter.ToUnstructured(testSecret)
	if err != nil {
		t.Error(err)
	}

	testSecretBytes, err := json.Marshal(testSecretData)
	if err != nil {
		t.Error(err)
		t.Errorf("error while marshalling")
	}

	var testSecretUnstructured *unstructured.Unstructured
	if err := json.Unmarshal(testSecretBytes, &testSecretUnstructured); err != nil {
		t.Error(err)
		t.Errorf("error unmarshalling into unstructured")
	}

	want := testSecretUnstructured

	fkclient, _ := kclient.FakeNew()

	scheme := runtime.NewScheme()
	fkclient.DynamicClient = dynamicfakeclient.NewSimpleDynamicClient(scheme)

	if err != nil {
		t.Errorf("error generating fake dynamic client")
	}

	testAdapter := Adapter{
		Client: *fkclient,
	}

	err = testAdapter.createDockerConfigSecret(testFilename, testSecretName, testNs)
	if err != nil {
		t.Error(err)
	}
	got, err := testAdapter.Client.DynamicClient.Resource(secretGroupVersionResource).
		Namespace(testNs).
		Get(testSecretName, metav1.GetOptions{})

	if err != nil {
		t.Error(err)
		t.Errorf("failed to get secret")
	}

	diff := cmp.Diff(got, want)
	if diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}
}

func TestIsInternalRegistry(t *testing.T) {

	imageTagInternal := "image-registry.openshift-image-registry.svc:5000/test-namespace/test-name"
	imageTagExternal := "docker.io/test-username/test-imae-repository-name"
	imageTagInvalid := "namespace/test-image-name"

	tests := []struct {
		name               string
		imageTag           string
		isInternalRegistry bool
		wantErr            bool
	}{
		{
			name:               "Case: ImageTag is for internal registry",
			imageTag:           imageTagInternal,
			isInternalRegistry: true,
			wantErr:            false,
		},
		{
			name:               "Case: ImageTag is for external registry",
			imageTag:           imageTagExternal,
			isInternalRegistry: false,
			wantErr:            false,
		},
		{
			name:               "Case: ImageTag is invalid",
			imageTag:           imageTagInvalid,
			isInternalRegistry: false,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isInternalRegistryResult, err := isInternalRegistry(tt.imageTag)
			if err != nil && !tt.wantErr {
				t.Errorf("returned error when it should not have")
			} else if err == nil && tt.wantErr {
				t.Errorf("should have returned an error but did not")
			}

			if isInternalRegistryResult != tt.isInternalRegistry {
				t.Errorf("returned the wrong result")
			}
		})
	}
}

func getExecCommand(id string, group common.DevfileCommandGroupType) versionsCommon.Exec {

	commands := [...]string{"ls -la", "pwd"}
	component := "component"
	workDir := [...]string{"/", "/root"}

	return versionsCommon.Exec{
		Id:          id,
		CommandLine: commands[0],
		Component:   component,
		WorkingDir:  workDir[0],
		Group:       &common.Group{Kind: group},
	}

}
