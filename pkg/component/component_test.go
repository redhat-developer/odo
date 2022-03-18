package component

import (
	"fmt"
	"reflect"
	"testing"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/kylelemons/godebug/pretty"

	v1 "k8s.io/api/apps/v1"

	"github.com/devfile/library/pkg/util"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/testingutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestList(t *testing.T) {
	deploymentList := v1.DeploymentList{Items: []v1.Deployment{
		*testingutil.CreateFakeDeployment("comp0"),
		*testingutil.CreateFakeDeployment("comp1"),
	}}

	deploymentList.Items[0].Labels[componentlabels.KubernetesNameLabel] = "nodejs"
	deploymentList.Items[0].Annotations = map[string]string{
		componentlabels.OdoProjectTypeAnnotation: "nodejs",
	}
	deploymentList.Items[1].Labels[componentlabels.KubernetesNameLabel] = "wildfly"
	deploymentList.Items[1].Annotations = map[string]string{
		componentlabels.OdoProjectTypeAnnotation: "wildfly",
	}
	tests := []struct {
		name           string
		deploymentList v1.DeploymentList
		projectExists  bool
		wantErr        bool
		output         ComponentList
	}{
		{
			name:          "Case 1: no component and no config exists",
			wantErr:       false,
			projectExists: true,
			output:        newComponentList([]Component{}),
		},
		{
			name:           "Case 2: Components are returned from deployments on a kubernetes cluster",
			deploymentList: deploymentList,
			wantErr:        false,
			projectExists:  true,
			output: ComponentList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.dev/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []Component{
					getFakeComponent("comp0", "test", "app", "nodejs", StateTypePushed),
					getFakeComponent("comp1", "test", "app", "wildfly", StateTypePushed),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fakeClientSet := kclient.FakeNew()
			client.Namespace = "test"

			fakeClientSet.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				listAction, ok := action.(ktesting.ListAction)
				if !ok {
					return false, nil, fmt.Errorf("expected a ListAction, got %v", action)
				}
				if len(tt.deploymentList.Items) <= 0 {
					return true, &tt.deploymentList, nil
				}

				var deploymentLabels0 map[string]string
				var deploymentLabels1 map[string]string
				if len(tt.deploymentList.Items) == 2 {
					deploymentLabels0 = tt.deploymentList.Items[0].Labels
					deploymentLabels1 = tt.deploymentList.Items[1].Labels
				}
				switch listAction.GetListRestrictions().Labels.String() {
				case util.ConvertLabelsToSelector(deploymentLabels0):
					return true, &tt.deploymentList.Items[0], nil
				case util.ConvertLabelsToSelector(deploymentLabels1):
					return true, &tt.deploymentList.Items[1], nil
				default:
					return true, &tt.deploymentList, nil
				}
			})

			results, err := List(client, applabels.GetSelector("app"))

			if (err != nil) != tt.wantErr {
				t.Errorf("expected err: %v, but err is %v", tt.wantErr, err)
			}

			if !reflect.DeepEqual(tt.output, results) {
				t.Errorf("Unexpected output, see the diff in results: %s", pretty.Compare(tt.output, results))
				t.Errorf("expected output:\n%#v\n\ngot:\n%#v", tt.output, results)
			}
		})
	}
}

func Test_getMachineReadableFormat(t *testing.T) {
	type args struct {
		componentName string
		componentType string
	}
	tests := []struct {
		name string
		args args
		want Component
	}{
		{
			name: "Test: Machine Readable Output",
			args: args{componentName: "frontend", componentType: "nodejs"},
			want: Component{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Component",
					APIVersion: "odo.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "frontend",
				},
				Spec: ComponentSpec{
					Type: "nodejs",
				},
				Status: ComponentStatus{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newComponentWithType(tt.args.componentName, tt.args.componentType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMachineReadableFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMachineReadableFormatForList(t *testing.T) {
	type args struct {
		components []Component
	}
	tests := []struct {
		name string
		args args
		want ComponentList
	}{
		{
			name: "Test: machine readable output for list",
			args: args{
				components: []Component{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Component",
							APIVersion: "odo.dev/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "frontend",
						},
						Spec: ComponentSpec{
							Type: "nodejs",
						},
						Status: ComponentStatus{},
					},
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Component",
							APIVersion: "odo.dev/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend",
						},
						Spec: ComponentSpec{
							Type: "wildfly",
						},
						Status: ComponentStatus{},
					},
				},
			},
			want: ComponentList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.dev/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []Component{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Component",
							APIVersion: "odo.dev/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "frontend",
						},
						Spec: ComponentSpec{
							Type: "nodejs",
						},
						Status: ComponentStatus{},
					},
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Component",
							APIVersion: "odo.dev/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend",
						},
						Spec: ComponentSpec{
							Type: "wildfly",
						},
						Status: ComponentStatus{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newComponentList(tt.args.components); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMachineReadableFormatForList() = %v, want %v", got, tt.want)
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

func getFakeComponent(compName, namespace, appName, compType string, state string) Component {
	return Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "odo.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      compName,
			Namespace: namespace,
			Labels: map[string]string{
				applabels.App:                           appName,
				applabels.ManagedBy:                     "odo",
				applabels.ApplicationLabel:              appName,
				componentlabels.KubernetesInstanceLabel: compName,
				componentlabels.KubernetesNameLabel:     compType,
				componentlabels.OdoModeLabel:            componentlabels.ComponentDevName,
			},
			Annotations: map[string]string{
				componentlabels.OdoProjectTypeAnnotation: compType,
			},
		},
		Spec: ComponentSpec{
			Type: compType,
			App:  appName,
		},
		Status: ComponentStatus{
			State: state,
		},
	}

}
