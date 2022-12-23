package component

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	_delete "github.com/redhat-developer/odo/pkg/component/delete"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/labels"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestComponentOptions_deleteNamedComponent(t *testing.T) {

	pod1 := corev1.Pod{}
	pod1.SetName("a-name-app")

	type fields struct {
		name                  string
		namespace             string
		forceFlag             bool
		kubernetesClient      func(ctrl *gomock.Controller) kclient.ClientInterface
		deleteComponentClient func(ctrl *gomock.Controller) _delete.Client
		podmanClient          func(ctrl *gomock.Controller) podman.Client
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "No cluster resource found",
			fields: fields{
				name:      "my-component",
				namespace: "my-namespace",
				forceFlag: false,
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					return client
				},
				podmanClient: func(ctrl *gomock.Controller) podman.Client {
					return nil
				},
				deleteComponentClient: func(ctrl *gomock.Controller) _delete.Client {
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().ListClusterResourcesToDelete(gomock.Any(), "my-component", "my-namespace").Return(nil, nil)
					client.EXPECT().DeleteResources(gomock.Any(), false).Times(0)
					return client
				},
			},
			wantErr: false,
		},
		{
			name: "2 cluster resources to delete",
			fields: fields{
				name:      "my-component",
				namespace: "my-namespace",
				forceFlag: true,
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					return client
				},
				podmanClient: func(ctrl *gomock.Controller) podman.Client {
					return nil
				},
				deleteComponentClient: func(ctrl *gomock.Controller) _delete.Client {
					var resources []unstructured.Unstructured
					res1 := getUnstructured("dep1", "deployment", "v1")
					res2 := getUnstructured("svc1", "service", "v1")
					resources = append(resources, res1, res2)
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().ListClusterResourcesToDelete(gomock.Any(), "my-component", "my-namespace").Return(resources, nil)
					client.EXPECT().DeleteResources([]unstructured.Unstructured{res1, res2}, false).Times(1)
					return client
				},
			},
		},
		{
			name: "1 podman resource to delete",
			fields: fields{
				name:      "my-component",
				forceFlag: true,
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					return nil
				},
				podmanClient: func(ctrl *gomock.Controller) podman.Client {
					client := podman.NewMockClient(ctrl)
					client.EXPECT().CleanupPodResources(&pod1).Times(1)
					return client
				},
				deleteComponentClient: func(ctrl *gomock.Controller) _delete.Client {
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().ListPodmanResourcesToDelete("app", "my-component").Return(true, []*corev1.Pod{&pod1}, nil).Times(1)
					return client
				},
			},
		},
		{
			name: "2 cluster and 1 podman resources to delete",
			fields: fields{
				name:      "my-component",
				namespace: "my-namespace",
				forceFlag: true,
				kubernetesClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					return client
				},
				podmanClient: func(ctrl *gomock.Controller) podman.Client {
					client := podman.NewMockClient(ctrl)
					client.EXPECT().CleanupPodResources(&pod1).Times(1)
					return client
				},
				deleteComponentClient: func(ctrl *gomock.Controller) _delete.Client {
					var resources []unstructured.Unstructured
					res1 := getUnstructured("dep1", "deployment", "v1")
					res2 := getUnstructured("svc1", "service", "v1")
					resources = append(resources, res1, res2)
					client := _delete.NewMockClient(ctrl)
					client.EXPECT().ListClusterResourcesToDelete(gomock.Any(), "my-component", "my-namespace").Return(resources, nil)
					client.EXPECT().ListPodmanResourcesToDelete("app", "my-component").Return(true, []*corev1.Pod{&pod1}, nil).Times(1)
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
					PodmanClient:     tt.fields.podmanClient(ctrl),
				},
			}
			ctx := odocontext.WithApplication(context.TODO(), "app")
			ctx = odocontext.WithComponentName(ctx, "a-name")
			if err := o.deleteNamedComponent(ctx); (err != nil) != tt.wantErr {
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
				deleteClient.EXPECT().ListClusterResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(true, resources, nil)
				deleteClient.EXPECT().ListClusterResourcesToDelete(gomock.Any(), compName, projectName).Return(resources, nil)
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
				deleteClient.EXPECT().ListClusterResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(true, resources, nil)
				deleteClient.EXPECT().ListClusterResourcesToDelete(gomock.Any(), compName, projectName).Return(resources, nil)
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
				deleteClient.EXPECT().ListClusterResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(false, nil, errors.New("some error"))
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
				deleteClient.EXPECT().ListClusterResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(true, resources, nil)
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
				deleteClient.EXPECT().ListClusterResourcesToDeleteFromDevfile(gomock.Any(), appName, gomock.Any(), labels.ComponentAnyMode).Return(false, nil, nil)
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
				clientset: &clientset.Clientset{
					KubernetesClient: kubeClient,
					DeleteClient:     deleteClient,
				},
			}
			ctx := odocontext.WithNamespace(context.Background(), projectName)
			ctx = odocontext.WithApplication(ctx, "app")
			ctx = odocontext.WithWorkingDirectory(ctx, workingDir)
			ctx = odocontext.WithComponentName(ctx, compName)
			ctx = odocontext.WithDevfileObj(ctx, &info)
			if err = o.deleteDevfileComponent(ctx); (err != nil) != tt.wantErr {
				t.Errorf("deleteDevfileComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// populateWorkingDir populates the working directory with .odo and devfile.yaml, and returns envinfo
func populateWorkingDir(fs filesystem.Filesystem, workingDir, compName, projectName string) parser.DevfileObj {
	_ = fs.MkdirAll(filepath.Join(workingDir), 0755)
	devfileObj := testingutil.GetTestDevfileObjFromFile("devfile-deploy.yaml")
	devfileYAML, err := yaml.Marshal(devfileObj.Data)
	if err != nil {
		return parser.DevfileObj{}
	}
	_ = fs.WriteFile(filepath.Join(workingDir, "devfile.yaml"), devfileYAML, 0644)
	return devfileObj
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
			got := listResourcesMissingFromDevfilePresentOnCluster(tt.args.componentName, tt.args.devfileResources, tt.args.clusterResources)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("listResourcesMissingFromDevfilePresentOnCluster() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_messageWithPlatforms(t *testing.T) {
	type args struct {
		cluster   bool
		podman    bool
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "cluster only",
			args: args{
				cluster:   true,
				name:      "componame",
				namespace: "def",
			},
			want: `No resource found for component "componame" in namespace "def"
`,
		},
		{
			name: "podman only",
			args: args{
				podman: true,
				name:   "componame",
			},
			want: `No resource found for component "componame" on podman
`,
		},
		{
			name: "cluster and podman",
			args: args{
				cluster:   true,
				podman:    true,
				name:      "componame",
				namespace: "def",
			},
			want: `No resource found for component "componame" in namespace "def" or on podman
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := messageWithPlatforms(tt.args.cluster, tt.args.podman, tt.args.name, tt.args.namespace); got != tt.want {
				t.Errorf("messageWithPlatforms() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_infoMsg(t *testing.T) {
	type args struct {
		cluster       bool
		podman        bool
		componentName string
		namespace     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "cluster only",
			args: args{
				cluster:       true,
				componentName: "compo",
				namespace:     "def",
			},
			want: `This will delete "compo" from the namespace "def".`,
		},
		{
			name: "podman only",
			args: args{
				podman:        true,
				componentName: "compo",
			},
			want: `This will delete "compo" from podman.`,
		},
		{
			name: "cluster and podman",
			args: args{
				cluster:       true,
				podman:        true,
				componentName: "compo",
				namespace:     "def",
			},
			want: `This will delete "compo" from the namespace "def" and from podman.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := infoMsg(tt.args.cluster, tt.args.podman, tt.args.componentName, tt.args.namespace); got != tt.want {
				t.Errorf("infoMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
