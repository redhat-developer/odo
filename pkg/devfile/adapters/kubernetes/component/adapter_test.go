package component

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentLabels "github.com/openshift/odo/pkg/component/labels"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	odoTestingUtil "github.com/openshift/odo/pkg/testingutil"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

func TestCreateOrUpdateComponent(t *testing.T) {

	testComponentName := "test"
	testAppName := "app"
	deployment := v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       kclient.DeploymentKind,
			APIVersion: kclient.DeploymentAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: testComponentName,
			Labels: map[string]string{
				applabels.ApplicationLabel:     testAppName,
				componentLabels.ComponentLabel: testComponentName,
			},
		},
	}

	tests := []struct {
		name          string
		componentType devfilev1.ComponentType
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
			componentType: devfilev1.ContainerComponentType,
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
			componentType: devfilev1.ContainerComponentType,
			envInfo:       envinfo.EnvSpecificInfo{},
			running:       true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var comp devfilev1.Component
			if tt.componentType != "" {
				comp = testingutil.GetFakeContainerComponent("component")
			}
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APIVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{comp})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands([]devfilev1.Command{getExecCommand("run", devfilev1.RunCommandGroupKind)})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				AppName:       testAppName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := occlient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("patch", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &deployment, nil
			})

			if tt.running {
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
			name: "Case: One container, Project Source Env",
			containers: []corev1.Container{
				{
					Name: "test",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath2",
						},
						{
							Name:  generator.EnvProjectsSrc,
							Value: "/mypath",
						},
					},
				},
			},
			want:           "test",
			wantSourcePath: "/mypath",
			wantErr:        false,
		},
		{
			name: "Case: Multiple containers, multiple Project Source Env",
			containers: []corev1.Container{
				{
					Name: "test1",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath1",
						},
						{
							Name:  generator.EnvProjectsSrc,
							Value: "/mypath1",
						},
					},
				},
				{
					Name: "test2",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath2",
						},
						{
							Name:  generator.EnvProjectsSrc,
							Value: "/mypath2",
						},
					},
				},
			},
			want:           "test1",
			wantSourcePath: "/mypath1",
			wantErr:        false,
		},
		{
			name: "Case: Multiple containers, no Project Source Env",
			containers: []corev1.Container{
				{
					Name: "test1",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath1",
						},
					},
				},
				{
					Name: "test2",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath2",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, syncFolder, err := getFirstContainerWithSourceVolume(tt.containers)
			if container != tt.want {
				t.Errorf("expected %s, actual %s", tt.want, container)
			}
			if syncFolder != tt.wantSourcePath {
				t.Errorf("expected %s, actual %s", tt.wantSourcePath, syncFolder)
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, actual %v", tt.wantErr, err)
			}
		})
	}
}

func TestDoesComponentExist(t *testing.T) {

	tests := []struct {
		name             string
		componentType    devfilev1.ComponentType
		componentName    string
		appName          string
		getComponentName string
		envInfo          envinfo.EnvSpecificInfo
		want             bool
		wantErr          bool
	}{
		{
			name:             "Case 1: Valid component name",
			componentName:    "test-name",
			appName:          "app",
			getComponentName: "test-name",
			envInfo:          envinfo.EnvSpecificInfo{},
			want:             true,
			wantErr:          false,
		},
		{
			name:             "Case 2: Non-existent component name",
			componentName:    "test-name",
			appName:          "app",
			getComponentName: "fake-component",
			envInfo:          envinfo.EnvSpecificInfo{},
			want:             false,
			wantErr:          false,
		},
		{
			name:             "Case 3: Error condition",
			componentName:    "test-name",
			appName:          "app",
			getComponentName: "test-name",
			envInfo:          envinfo.EnvSpecificInfo{},
			want:             false,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APIVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent("component")})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands([]devfilev1.Command{getExecCommand("run", devfilev1.RunCommandGroupKind)})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				AppName:       tt.appName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := occlient.FakeNew()
			fkWatch := watch.NewFake()

			fkclientset.Kubernetes.PrependReactor("patch", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				if patchAction, is := action.(ktesting.PatchAction); is {
					patch := patchAction.GetPatch()
					var deployment v1.Deployment
					err := json.Unmarshal(patch, &deployment)
					if err != nil {
						t.Errorf("unable to parse deployment %q\n", err)
						return false, nil, err
					}
					return true, &deployment, nil
				}
				return false, nil, nil
			})

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

			fkclientset.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				emptyDeployment := odoTestingUtil.CreateFakeDeployment("")
				deployment := odoTestingUtil.CreateFakeDeployment(tt.getComponentName)

				if tt.wantErr {
					return true, &v1.DeploymentList{Items: []v1.Deployment{*emptyDeployment}}, errors.Errorf("deployment get error")
				} else if tt.getComponentName == tt.componentName {
					return true, &v1.DeploymentList{Items: []v1.Deployment{*deployment}}, nil
				}

				return true, &v1.DeploymentList{Items: []v1.Deployment{}}, nil
			})

			// Verify that a component with the specified name exists
			componentExists, err := componentAdapter.DoesComponentExist(tt.getComponentName, "")
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
		componentType devfilev1.ComponentType
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
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APIVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent("component")})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := occlient.FakeNew()
			fkWatch := watch.NewFake()

			// Change the status
			go func() {
				fkWatch.Modify(kclient.FakePodStatus(tt.status, testComponentName))
			}()

			fkclientset.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			componentAdapter := New(adapterCtx, *fkclient)
			_, err := componentAdapter.getPod(false)

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
					*odoTestingUtil.CreateFakePod("component", "component"),
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
					*odoTestingUtil.CreateFakePod("component", "component"),
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
					*odoTestingUtil.CreateFakePod("component", "component"),
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
					*odoTestingUtil.CreateFakePod("component", "component"),
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
					*odoTestingUtil.CreateFakePod("component", "component"),
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
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APIVersion200))
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			if !tt.componentExists {
				adapterCtx.ComponentName = "doesNotExists"
			}

			fkclient, fkclientset := occlient.FakeNew()

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

			if err := a.Delete(tt.args.labels, false, false); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getExecCommand(id string, group devfilev1.CommandGroupKind) devfilev1.Command {

	commands := [...]string{"ls -la", "pwd"}
	component := "component"
	workDir := [...]string{"/", "/root"}

	return devfilev1.Command{
		Id: id,
		CommandUnion: devfilev1.CommandUnion{
			Exec: &devfilev1.ExecCommand{
				LabeledCommand: devfilev1.LabeledCommand{
					BaseCommand: devfilev1.BaseCommand{
						Group: &devfilev1.CommandGroup{Kind: group},
					},
				},
				CommandLine: commands[0],
				Component:   component,
				WorkingDir:  workDir[0],
			},
		},
	}

}

func TestAdapter_generateDeploymentObjectMeta(t *testing.T) {
	namespacedKubernetesName, err := util.NamespaceKubernetesObject("nodejs", "app")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	type fields struct {
		componentName string
		appName       string
		deployment    *v1.Deployment
	}
	type args struct {
		labels map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    metav1.ObjectMeta
		wantErr bool
	}{
		{
			name: "case 1: deployment exists",
			fields: fields{
				componentName: "nodejs",
				appName:       "app",
				deployment:    odoTestingUtil.CreateFakeDeployment("nodejs"),
			},
			args: args{
				labels: odoTestingUtil.CreateFakeDeployment("nodejs").Labels,
			},
			want:    generator.GetObjectMeta("nodejs", "project-0", odoTestingUtil.CreateFakeDeployment("nodejs").Labels, nil),
			wantErr: false,
		},
		{
			name: "case 2: deployment doesn't exists",
			fields: fields{
				componentName: "nodejs",
				appName:       "app",
				deployment:    nil,
			},
			args: args{
				labels: odoTestingUtil.CreateFakeDeployment("nodejs").Labels,
			},
			want:    generator.GetObjectMeta(namespacedKubernetesName, "project-0", odoTestingUtil.CreateFakeDeployment("nodejs").Labels, nil),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, _ := occlient.FakeNew()
			fakeClient.Namespace = "project-0"

			a := Adapter{
				Client: *fakeClient,
				GenericAdapter: &adaptersCommon.GenericAdapter{
					AdapterContext: adaptersCommon.AdapterContext{
						ComponentName: tt.fields.componentName,
						AppName:       tt.fields.appName,
					},
				},
				deployment: tt.fields.deployment,
			}
			got, err := a.generateDeploymentObjectMeta(tt.args.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateDeploymentObjectMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateDeploymentObjectMeta() got = %v, want %v", got, tt.want)
			}
		})
	}
}
