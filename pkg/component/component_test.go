package component

import (
	"errors"
	"reflect"
	"testing"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/labels"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/redhat-developer/odo/pkg/api"
)

func TestListAllClusterComponents(t *testing.T) {
	res1 := getUnstructured("dep1", "deployment", "v1", "Unknown", "Unknown", "my-ns")
	res2 := getUnstructured("svc1", "service", "v1", "odo", "nodejs", "my-ns")
	res3 := getUnstructured("dep1", "deployment", "v1", "Unknown", "Unknown", "my-ns")
	res3.SetLabels(map[string]string{})
	commonLabels := labels.Builder().WithComponentName("comp1").WithManager("odo")

	resDev := getUnstructured("depDev", "deployment", "v1", "odo", "nodejs", "my-ns")
	labelsDev := commonLabels.WithMode("Dev").Labels()
	resDev.SetLabels(labelsDev)

	resDeploy := getUnstructured("depDeploy", "deployment", "v1", "odo", "nodejs", "my-ns")
	labelsDeploy := commonLabels.WithMode("Deploy").Labels()
	resDeploy.SetLabels(labelsDeploy)

	type fields struct {
		kubeClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []api.ComponentAbstract
		wantErr bool
	}{
		{
			name: "1 non-odo resource returned with Unknown",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, res1)
					client := kclient.NewMockClientInterface(ctrl)
					selector := ""
					client.EXPECT().GetAllResourcesFromSelector(selector, "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want: []api.ComponentAbstract{{
				Name:      "dep1",
				ManagedBy: "Unknown",
				RunningIn: nil,
				Type:      "Unknown",
			}},
			wantErr: false,
		},
		{
			name: "0 non-odo resource without instance label is not returned",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, res3)
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetAllResourcesFromSelector(gomock.Any(), "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "1 non-odo resource returned with Unknown, and 1 odo resource returned with odo",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, res1, res2)
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetAllResourcesFromSelector(gomock.Any(), "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want: []api.ComponentAbstract{{
				Name:      "dep1",
				ManagedBy: "Unknown",
				RunningIn: nil,
				Type:      "Unknown",
			}, {
				Name:      "svc1",
				ManagedBy: "odo",
				RunningIn: nil,
				Type:      "nodejs",
			}},
			wantErr: false,
		},
		{
			name: "one resource in Dev and Deploy modes",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, resDev, resDeploy)
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetAllResourcesFromSelector(gomock.Any(), "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want: []api.ComponentAbstract{{
				Name:      "comp1",
				ManagedBy: "odo",
				RunningIn: api.RunningModeList{api.RunningModeDev, api.RunningModeDeploy},
				Type:      "nodejs",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			got, err := ListAllClusterComponents(tt.fields.kubeClient(ctrl), tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAllClusterComponents error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListAllClusterComponents got = %+v\nwant = %+v\ncomparison:\n %v", got, tt.want, pretty.Compare(got, tt.want))
			}
		})
	}
}

func TestGetComponentTypeFromDevfileMetadata(t *testing.T) {
	tests := []devfilepkg.DevfileMetadata{
		{
			Name:        "ReturnProject",
			ProjectType: "Maven",
			Language:    "Java",
		},
		{
			Name:     "ReturnLanguage",
			Language: "Java",
		},
		{
			Name: "ReturnNA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var want string
			got := GetComponentTypeFromDevfileMetadata(tt)
			switch tt.Name {
			case "ReturnProject":
				want = tt.ProjectType
			case "ReturnLanguage":
				want = tt.Language
			case "ReturnNA":
				want = NotAvailable
			}
			if got != want {
				t.Errorf("Incorrect component type returned; got: %q, want: %q", got, want)
			}
		})
	}
}

// getUnstructured returns an unstructured.Unstructured object
func getUnstructured(name, kind, apiVersion, managed, componentType, namespace string) (u unstructured.Unstructured) {
	u.SetName(name)
	u.SetKind(kind)
	u.SetAPIVersion(apiVersion)
	u.SetNamespace(namespace)
	u.SetLabels(labels.Builder().
		WithComponentName(name).
		WithManager(managed).
		Labels())
	u.SetAnnotations(labels.Builder().
		WithProjectType(componentType).
		Labels())
	return
}

func TestGetRunningModes(t *testing.T) {

	resourceDev1 := unstructured.Unstructured{}
	resourceDev1.SetLabels(labels.Builder().WithMode(labels.ComponentDevMode).Labels())

	resourceDev2 := unstructured.Unstructured{}
	resourceDev2.SetLabels(labels.Builder().WithMode(labels.ComponentDevMode).Labels())

	resourceDeploy1 := unstructured.Unstructured{}
	resourceDeploy1.SetLabels(labels.Builder().WithMode(labels.ComponentDeployMode).Labels())

	resourceDeploy2 := unstructured.Unstructured{}
	resourceDeploy2.SetLabels(labels.Builder().WithMode(labels.ComponentDeployMode).Labels())

	otherResource := unstructured.Unstructured{}

	packageManifestResource := unstructured.Unstructured{}
	packageManifestResource.SetKind("PackageManifest")
	packageManifestResource.SetLabels(labels.Builder().WithMode(labels.ComponentDevMode).Labels())

	type args struct {
		client func(ctrl *gomock.Controller) kclient.ClientInterface
		name   string
	}
	tests := []struct {
		name    string
		args    args
		want    []api.RunningMode
		wantErr bool
	}{
		{
			name: "No resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{}, nil)
					return c
				},
				name: "aname",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Only PackageManifest resource",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource}, nil)
					return c
				},
				name: "aname",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No dev/deploy resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource}, nil)
					return c
				},
				name: "aname",
			},
			want: []api.RunningMode{},
		},
		{
			name: "Only Dev resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDev1, resourceDev2}, nil)
					return c
				},
				name: "aname",
			},
			want: []api.RunningMode{api.RunningModeDev},
		},
		{
			name: "Only Deploy resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDeploy1, resourceDeploy2}, nil)
					return c
				},
				name: "aname",
			},
			want: []api.RunningMode{api.RunningModeDeploy},
		},
		{
			name: "Dev and Deploy resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDev1, resourceDev2, resourceDeploy1, resourceDeploy2}, nil)
					return c
				},
				name: "aname",
			},
			want: []api.RunningMode{api.RunningModeDev, api.RunningModeDeploy},
		},
		{
			name: "Unknown",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
					return c
				},
				name: "aname",
			},
			want: []api.RunningMode{api.RunningModeUnknown},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			got, err := GetRunningModes(tt.args.client(ctrl), tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRunningModes() = %v, want %v", got, tt.want)
			}
		})
	}
}
