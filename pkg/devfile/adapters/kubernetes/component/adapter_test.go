package component

import (
	"context"
	"errors"
	"testing"

	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/testingutil"

	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

			fkclient, fkclientset := kclient.FakeNew()

			fkclientset.Kubernetes.PrependReactor("patch", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &deployment, nil
			})

			if tt.running {
				fkclientset.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, &deployment, nil
				})
			}

			fkclientset.Kubernetes.PrependReactor("get", "namespaces", func(action ktesting.Action) (bool, runtime.Object, error) {
				ns := &corev1.Namespace{}
				ns.SetName("my-ns")
				return true, ns, nil
			})
			ctrl := gomock.NewController(t)
			fakePrefClient := preference.NewMockClient(ctrl)
			fakePrefClient.EXPECT().GetEphemeralSourceVolume().AnyTimes()
			fakeConfigAutomount := configAutomount.NewMockClient(ctrl)
			fakeConfigAutomount.EXPECT().GetAutomountingVolumes().AnyTimes()
			componentAdapter := NewKubernetesAdapter(fkclient, fakePrefClient, nil, nil, nil, nil, fakeConfigAutomount, nil, devObj)
			ctx := context.Background()
			ctx = odocontext.WithApplication(ctx, "app")
			ctx = odocontext.WithComponentName(ctx, "my-component")
			ctx = odocontext.WithDevfilePath(ctx, "/path/to/devfile")
			_, _, err := componentAdapter.createOrUpdateComponent(ctx, tt.running, libdevfile.DevfileCommands{}, nil)

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
			}
			ctx := context.Background()
			ctx = odocontext.WithApplication(ctx, "app")
			ctx = odocontext.WithComponentName(ctx, "nodejs")
			ctx = odocontext.WithDevfilePath(ctx, "/path/to/devfile")
			got, err := a.generateDeploymentObjectMeta(ctx, tt.fields.deployment, tt.args.labels, tt.args.annotations)
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

func TestAdapter_deleteRemoteResources(t *testing.T) {
	type fields struct {
		kubeClientCustomizer func(kubeClient *kclient.MockClientInterface)
	}
	type args struct {
		objectsToRemove []unstructured.Unstructured
	}

	var u1 unstructured.Unstructured
	u1.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "metrics.k8s.io",
		Version: "v1beta1",
		Kind:    "PodMetrics",
	})
	u1.SetName("my-pod-metrics")

	var u2 unstructured.Unstructured
	u2.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "postgresql.k8s.enterprisedb.io",
		Version: "v1",
		Kind:    "Cluster",
	})
	u2.SetName("my-pg-cluster")
	toRemove := []unstructured.Unstructured{u1, u2}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nothing to delete - nil list",
			args: args{
				objectsToRemove: nil,
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				},
			},
			wantErr: false,
		},
		{
			name: "nothing to delete - empty list",
			args: args{
				objectsToRemove: []unstructured.Unstructured{},
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				},
			},
			wantErr: false,
		},
		{
			name: "error getting information about resource",
			args: args{
				objectsToRemove: toRemove,
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u1.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u2.GroupVersionKind())).Return(schema.GroupVersionResource{}, errors.New("error on GetGVRFromGVK(u2)"))
					// Only u1 should be deleted
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u1.GetName()), gomock.Any(), gomock.Any()).Return(nil).Times(1)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u2.GetName()), gomock.Any(), gomock.Any()).Times(0)
				},
			},
			wantErr: true,
		},
		{
			name: "generic error deleting resource",
			args: args{
				objectsToRemove: toRemove,
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u1.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u2.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u1.GetName()), gomock.Any(), gomock.Any()).Return(nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u2.GetName()), gomock.Any(), gomock.Any()).Return(errors.New("generic error while deleting u1"))
				},
			},
			wantErr: true,
		},
		{
			name: "generic error deleting all resources",
			args: args{
				objectsToRemove: toRemove,
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u1.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u2.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u1.GetName()), gomock.Any(), gomock.Any()).
						Return(errors.New("generic error while deleting u1"))
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u2.GetName()), gomock.Any(), gomock.Any()).
						Return(errors.New("generic error while deleting u2"))
				},
			},
			wantErr: true,
		},
		{
			name: "not found error deleting resource",
			args: args{
				objectsToRemove: toRemove,
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u1.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u2.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u1.GetName()), gomock.Any(), gomock.Any()).Return(nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u2.GetName()), gomock.Any(), gomock.Any()).
						Return(kerrors.NewNotFound(schema.GroupResource{}, "u2"))
				},
			},
			wantErr: false,
		},
		{
			name: "method not allowed error deleting resource",
			args: args{
				objectsToRemove: toRemove,
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u1.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u2.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u1.GetName()), gomock.Any(), gomock.Any()).Return(nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u2.GetName()), gomock.Any(), gomock.Any()).
						Return(kerrors.NewMethodNotSupported(schema.GroupResource{Resource: "PodMetrics"}, "DELETE"))
				},
			},
			wantErr: false,
		},
		{
			name: "not found error deleting all resources",
			args: args{
				objectsToRemove: toRemove,
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u1.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u2.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u1.GetName()), gomock.Any(), gomock.Any()).
						Return(kerrors.NewNotFound(schema.GroupResource{}, "u1"))
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u2.GetName()), gomock.Any(), gomock.Any()).
						Return(kerrors.NewNotFound(schema.GroupResource{}, "u2"))
				},
			},
			wantErr: false,
		},
		{
			name: "method not allowed error deleting all resources",
			args: args{
				objectsToRemove: toRemove,
			},
			fields: fields{
				kubeClientCustomizer: func(kubeClient *kclient.MockClientInterface) {
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u1.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().GetGVRFromGVK(gomock.Eq(u2.GroupVersionKind())).Return(schema.GroupVersionResource{}, nil)
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u1.GetName()), gomock.Any(), gomock.Any()).
						Return(kerrors.NewMethodNotSupported(schema.GroupResource{}, "DELETE"))
					kubeClient.EXPECT().DeleteDynamicResource(gomock.Eq(u2.GetName()), gomock.Any(), gomock.Any()).
						Return(kerrors.NewMethodNotSupported(schema.GroupResource{}, "DELETE"))
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			kubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.fields.kubeClientCustomizer != nil {
				tt.fields.kubeClientCustomizer(kubeClient)
			}
			a := Adapter{
				kubeClient: kubeClient,
			}
			if err := a.deleteRemoteResources(tt.args.objectsToRemove); (err != nil) != tt.wantErr {
				t.Errorf("deleteRemoteResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
