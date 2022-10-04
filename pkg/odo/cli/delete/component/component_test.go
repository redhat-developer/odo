package component

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	_delete "github.com/redhat-developer/odo/pkg/component/delete"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/testingutil"
)

func TestComponentOptions_deleteNamedComponent(t *testing.T) {
	type fields struct {
		name                  string
		namespace             string
		forceFlag             bool
		kubernetesClient      func(ctrl *gomock.Controller) kclient.ClientInterface
		deleteComponentClient func(ctrl *gomock.Controller) _delete.Client
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "No resource found",
			fields: fields{
				name:      "my-component",
				namespace: "my-namespace",
				forceFlag: false,
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					return client
				},
				deleteComponentClient: func(ctrl *gomock.Controller) _delete.Client {
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().ListClusterResourcesToDelete("my-component", "my-namespace").Return(nil, nil)
					client.EXPECT().DeleteResources(gomock.Any(), false).Times(0)
					return client
				},
			},
			wantErr: false,
		},
		{
			name: "2 resources to delete",
			fields: fields{
				name:      "my-component",
				namespace: "my-namespace",
				forceFlag: true,
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					return client
				},
				deleteComponentClient: func(ctrl *gomock.Controller) _delete.Client {
					var resources []unstructured.Unstructured
					res1 := getUnstructured("dep1", "deployment", "v1")
					res2 := getUnstructured("svc1", "service", "v1")
					resources = append(resources, res1, res2)
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().ListClusterResourcesToDelete("my-component", "my-namespace").Return(resources, nil)
					client.EXPECT().DeleteResources([]unstructured.Unstructured{res1, res2}, false).Times(1)
					return client
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &ComponentOptions{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				forceFlag: tt.fields.forceFlag,
				clientset: &clientset.Clientset{
					KubernetesClient: tt.fields.kubernetesClient(ctrl),
					DeleteClient:     tt.fields.deleteComponentClient(ctrl),
				},
			}
			if err := o.deleteNamedComponent(); (err != nil) != tt.wantErr {
				t.Errorf("ComponentOptions.deleteNamedComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComponentOptions_deleteDevfileComponent(t *testing.T) {
	const (
		compName    = "nodejs-prj1-api-abhz"
		appName     = "app"
		projectName = "a-project"
	)
	prefixDir, err := os.MkdirTemp(os.TempDir(), "unittests-")
	if err != nil {
		t.Errorf("Error creating temp directory for tests")
		return
	}
	workingDir := filepath.Join(prefixDir, "myapp")
	resources := []unstructured.Unstructured{
		getUnstructured("my-component", "Deployment", "apps/v1"),
		getUnstructured(compName, "Deployment", "apps/v1"),
	}
	type fields struct {
		name      string
		forceFlag bool
	}
	tests := []struct {
		name         string
		fields       fields
		wantErr      bool
		deleteClient func(ctrl *gomock.Controller) _delete.Client
	}{
		{
			name: "deleting a component with access to devfile",
			deleteClient: func(ctrl *gomock.Controller) _delete.Client {
				deleteClient := _delete.NewMockClient(ctrl)
				deleteClient.EXPECT().ListResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(true, resources, nil)
				deleteClient.EXPECT().ListClusterResourcesToDelete(compName, projectName).Return(resources, nil)
				deleteClient.EXPECT().ExecutePreStopEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				deleteClient.EXPECT().DeleteResources(resources, false).Return([]unstructured.Unstructured{})
				return deleteClient
			},
			fields: fields{
				forceFlag: true,
			},
			wantErr: false,
		},
		{
			name: "deleting a component should not fail even if ExecutePreStopEvents fails",
			deleteClient: func(ctrl *gomock.Controller) _delete.Client {
				deleteClient := _delete.NewMockClient(ctrl)
				deleteClient.EXPECT().ListResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(true, resources, nil)
				deleteClient.EXPECT().ListClusterResourcesToDelete(compName, projectName).Return(resources, nil)
				deleteClient.EXPECT().ExecutePreStopEvents(gomock.Any(), appName, gomock.Any()).Return(errors.New("some error"))
				deleteClient.EXPECT().DeleteResources(resources, false).Return(nil)
				return deleteClient
			},
			fields: fields{
				forceFlag: true,
			},
			wantErr: false,
		},
		{
			name: "deleting a component should fail if ListResourcesToDeleteFromDevfile fails",
			deleteClient: func(ctrl *gomock.Controller) _delete.Client {
				deleteClient := _delete.NewMockClient(ctrl)
				deleteClient.EXPECT().ListResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(false, nil, errors.New("some error"))
				return deleteClient
			},
			fields: fields{
				forceFlag: true,
			},
			wantErr: true,
		},
		{
			name: "deleting a component should be aborted if forceFlag is not passed",
			deleteClient: func(ctrl *gomock.Controller) _delete.Client {
				deleteClient := _delete.NewMockClient(ctrl)
				deleteClient.EXPECT().ListResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(true, resources, nil)
				return deleteClient
			},
			fields: fields{
				forceFlag: false,
			},
			wantErr: false,
		},
		{
			name: "nothing to delete",
			deleteClient: func(ctrl *gomock.Controller) _delete.Client {
				deleteClient := _delete.NewMockClient(ctrl)
				deleteClient.EXPECT().ListResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(false, nil, nil)
				return deleteClient
			},
			fields: fields{
				forceFlag: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// the first one is to cleanup the directory before execution (in case there are remaining files from a previous execution)
			os.RemoveAll(prefixDir)
			// the second one to cleanup after execution
			defer os.RemoveAll(prefixDir)
			info := populateWorkingDir(filesystem.DefaultFs{}, workingDir, compName, projectName)
			ctrl := gomock.NewController(t)
			kubeClient := prepareKubeClient(ctrl, projectName)
			deleteClient := tt.deleteClient(ctrl)
			o := &ComponentOptions{
				name:      tt.fields.name,
				forceFlag: tt.fields.forceFlag,
				Context:   prepareContext(ctrl, kubeClient, info, workingDir),
				clientset: &clientset.Clientset{
					KubernetesClient: kubeClient,
					DeleteClient:     deleteClient,
				},
			}
			ctx := odocontext.WithNamespace(context.Background(), projectName)
			if err = o.deleteDevfileComponent(ctx); (err != nil) != tt.wantErr {
				t.Errorf("deleteDevfileComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// populateWorkingDir populates the working directory with .odo and devfile.yaml, and returns envinfo
func populateWorkingDir(fs filesystem.Filesystem, workingDir, compName, projectName string) *envinfo.EnvSpecificInfo {
	_ = fs.MkdirAll(filepath.Join(workingDir), 0755)
	env, err := envinfo.NewEnvSpecificInfo(workingDir)
	if err != nil {
		return env
	}
	devfileObj := testingutil.GetTestDevfileObjFromFile("devfile-deploy.yaml")
	devfileYAML, err := yaml.Marshal(devfileObj.Data)
	if err != nil {
		return env
	}
	_ = fs.WriteFile(filepath.Join(workingDir, "devfile.yaml"), devfileYAML, 0644)
	env.SetDevfileObj(devfileObj)
	return env
}

// prepareKubeClient prepares the mock kclient.ClientInterface3 and returns it
func prepareKubeClient(ctrl *gomock.Controller, projectName string) kclient.ClientInterface {
	kubeClient := kclient.NewMockClientInterface(ctrl)
	kubeClient.EXPECT().GetNamespaceNormal(projectName).Return(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: projectName,
			},
		}, nil).AnyTimes()
	kubeClient.EXPECT().SetNamespace(projectName).AnyTimes()
	return kubeClient
}

// prepareContext prepares the mock genericclioptions.Context and returns it
func prepareContext(ctrl *gomock.Controller, kubeClient kclient.ClientInterface, info *envinfo.EnvSpecificInfo, workingDir string) *genericclioptions.Context {
	cmdline := cmdline.NewMockCmdline(ctrl)
	cmdline.EXPECT().GetWorkingDirectory().Return(workingDir, nil).AnyTimes()
	cmdline.EXPECT().FlagValueIfSet("project").Return("").AnyTimes()
	cmdline.EXPECT().FlagValueIfSet("app").Return("").AnyTimes()
	cmdline.EXPECT().FlagValueIfSet("component").Return("").AnyTimes()
	cmdline.EXPECT().FlagValueIfSet("o").Return("").AnyTimes()
	cmdline.EXPECT().GetKubeClient().Return(kubeClient, nil).AnyTimes()
	createParameters := genericclioptions.NewCreateParameters(cmdline).NeedDevfile(workingDir)
	context, err := genericclioptions.New(createParameters)
	if err != nil {
		return nil
	}
	context.EnvSpecificInfo = info
	return context
}

// getUnstructured returns an unstructured.Unstructured object
func getUnstructured(name, kind, apiVersion string) (u unstructured.Unstructured) {
	u.SetName(name)
	u.SetKind(kind)
	u.SetAPIVersion(apiVersion)
	return
}

func Test_listResourcesMissingFromDevfilePresentOnCluster(t *testing.T) {
	const componentName = "mynodejs"
	deployment := getUnstructured("my-deploy", "Deployment", "apps/v1")
	svc := getUnstructured("my-service", "Service", "v1")
	deployment2 := getUnstructured("my-deploy-2", "Deployment", "apps/v1")
	endpoint := getUnstructured(fmt.Sprintf("my-endpoint-%s", componentName), "Endpoints", "apps/v1")
	type args struct {
		componentName    string
		devfileResources []unstructured.Unstructured
		clusterResources []unstructured.Unstructured
	}
	tests := []struct {
		name string
		args args
		want []unstructured.Unstructured
	}{
		// TODO: Add test cases.
		{
			name: "devfile and cluster has same resources",
			args: args{
				componentName:    componentName,
				devfileResources: []unstructured.Unstructured{deployment, svc},
				clusterResources: []unstructured.Unstructured{deployment, svc},
			},
			want: nil,
		},
		{
			name: "devfile and cluster have different resources",
			args: args{
				componentName:    componentName,
				devfileResources: []unstructured.Unstructured{deployment, svc},
				clusterResources: []unstructured.Unstructured{deployment2, svc},
			},
			want: []unstructured.Unstructured{deployment2},
		},
		{
			name: "component's endpoint is one of the cluster resources",
			args: args{
				componentName:    componentName,
				devfileResources: []unstructured.Unstructured{deployment, svc},
				clusterResources: []unstructured.Unstructured{deployment2, svc, endpoint},
			},
			want: []unstructured.Unstructured{deployment2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listResourcesMissingFromDevfilePresentOnCluster(tt.args.componentName, tt.args.devfileResources, tt.args.clusterResources); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listResourcesMissingFromDevfilePresentOnCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}
