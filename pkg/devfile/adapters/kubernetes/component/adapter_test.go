package component

import (
	"testing"

	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/testingutil"

	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
			Name:        testComponentName,
			Labels:      odolabels.Builder().WithComponentName(testComponentName).WithAppName(testAppName).Labels(),
			Annotations: odolabels.Builder().WithProjectType("").Labels(),
		},
	}

	tests := []struct {
		name          string
		componentType devfilev1.ComponentType
		running       bool
		wantErr       bool
	}{
		{
			name:          "Case 1: Invalid devfile",
			componentType: "",
			running:       false,
			wantErr:       true,
		},
		{
			name:          "Case 2: Valid devfile",
			componentType: devfilev1.ContainerComponentType,
			running:       false,
			wantErr:       false,
		},
		{
			name:          "Case 3: Invalid devfile, already running component",
			componentType: "",
			running:       true,
			wantErr:       true,
		},
		{
			name:          "Case 4: Valid devfile, already running component",
			componentType: devfilev1.ContainerComponentType,
			running:       true,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var comp devfilev1.Component
			if tt.componentType != "" {
				odolabels.SetProjectType(deployment.Annotations, string(tt.componentType))
				comp = testingutil.GetFakeContainerComponent("component")
			}
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					metadata := devfileData.GetMetadata()
					metadata.ProjectType = string(tt.componentType)
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

			adapterCtx := AdapterContext{
				ComponentName: testComponentName,
				AppName:       testAppName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := kclient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("patch", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &deployment, nil
			})

			if tt.running {
				fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, &deployment, nil
				})
			}
			ctrl := gomock.NewController(t)
			fakePrefClient := preference.NewMockClient(ctrl)
			fakePrefClient.EXPECT().GetEphemeralSourceVolume()
			componentAdapter := NewKubernetesAdapter(fkclient, fakePrefClient, nil, nil, nil, nil, adapterCtx)
			_, _, err := componentAdapter.createOrUpdateComponent(tt.running, libdevfile.DevfileCommands{}, nil)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
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
		labels      map[string]string
		annotations map[string]string
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
				deployment:    odoTestingUtil.CreateFakeDeployment("nodejs", false),
			},
			args: args{
				labels:      odoTestingUtil.CreateFakeDeployment("nodejs", false).Labels,
				annotations: nil,
			},
			want:    generator.GetObjectMeta("nodejs", "project-0", odoTestingUtil.CreateFakeDeployment("nodejs", false).Labels, nil),
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
				labels:      odoTestingUtil.CreateFakeDeployment("nodejs", false).Labels,
				annotations: nil,
			},
			want:    generator.GetObjectMeta(namespacedKubernetesName, "project-0", odoTestingUtil.CreateFakeDeployment("nodejs", false).Labels, nil),
			wantErr: false,
		},
		{
			name: "case 3: deployment exists and there is annotations successfully passed in",
			fields: fields{
				componentName: "nodejs",
				appName:       "app",
				deployment:    odoTestingUtil.CreateFakeDeployment("nodejs", false),
			},
			args: args{
				labels:      odoTestingUtil.CreateFakeDeployment("nodejs", false).Labels,
				annotations: odolabels.Builder().WithMode(odolabels.ComponentDevMode).Labels(),
			},
			want:    generator.GetObjectMeta("nodejs", "project-0", odoTestingUtil.CreateFakeDeployment("nodejs", false).Labels, odolabels.Builder().WithMode(odolabels.ComponentDevMode).Labels()),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, _ := kclient.FakeNew()
			fakeClient.Namespace = "project-0"

			a := Adapter{
				kubeClient: fakeClient,
				AdapterContext: AdapterContext{
					ComponentName: tt.fields.componentName,
					AppName:       tt.fields.appName,
				},
			}
			got, err := a.generateDeploymentObjectMeta(tt.fields.deployment, tt.args.labels, tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateDeploymentObjectMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Adapter.generateDeploymentObjectMeta() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
