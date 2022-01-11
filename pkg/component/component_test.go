package component

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"

	"github.com/golang/mock/gomock"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/preference"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getComponentFrom(t *testing.T) {
	type cmpSetting struct {
		componentName   string
		project         string
		applicationName string
		debugPort       int
	}
	tests := []struct {
		name          string
		isEnvInfo     bool
		componentType string
		envURL        []localConfigProvider.LocalURL
		cmpSetting    cmpSetting
		want          Component
		wantErr       bool
	}{
		{
			name:          "Case 1: Get component when env info file exists",
			isEnvInfo:     true,
			componentType: "nodejs",
			envURL: []localConfigProvider.LocalURL{
				{
					Name: "url1",
				},
			},
			cmpSetting: cmpSetting{
				componentName:   "frontend",
				project:         "project1",
				applicationName: "testing",
				debugPort:       1234,
			},
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

		{
			name:          "Case 2: Get component when env info file does not exists",
			isEnvInfo:     false,
			componentType: "nodejs",
			envURL: []localConfigProvider.LocalURL{
				{
					Name: "url2",
				},
			},
			cmpSetting: cmpSetting{
				componentName:   "backend",
				project:         "project2",
				applicationName: "app1",
				debugPort:       5896,
			},
			want: Component{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLocalConfigProvider := localConfigProvider.NewMockLocalConfigProvider(ctrl)

			mockLocalConfigProvider.EXPECT().Exists().Return(tt.isEnvInfo)

			if tt.isEnvInfo {
				mockLocalConfigProvider.EXPECT().GetName().Return(tt.cmpSetting.componentName)

				component := newComponentWithType(tt.cmpSetting.componentName, tt.componentType)

				mockLocalConfigProvider.EXPECT().GetNamespace().Return(tt.cmpSetting.project)

				component.Namespace = tt.cmpSetting.project
				mockLocalConfigProvider.EXPECT().GetApplication().Return(tt.cmpSetting.applicationName)
				mockLocalConfigProvider.EXPECT().GetDebugPort().Return(tt.cmpSetting.debugPort)

				component.Spec = ComponentSpec{
					App:   tt.cmpSetting.applicationName,
					Type:  tt.componentType,
					Ports: []string{fmt.Sprintf("%d", tt.cmpSetting.debugPort)},
				}

				mockLocalConfigProvider.EXPECT().ListURLs().Return(tt.envURL, nil)

				if len(tt.envURL) > 0 {
					for _, url := range tt.envURL {
						component.Spec.URL = append(component.Spec.URL, url.Name)
					}
				}

				tt.want = component

			}

			got, err := getComponentFrom(mockLocalConfigProvider, tt.componentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("getComponentFrom() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getComponentFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestList(t *testing.T) {
// 	namespace := "test"
// 	deploymentList := v1.DeploymentList{Items: []v1.Deployment{
// 		*testingutil.CreateFakeDeployment("comp0"),
// 		*testingutil.CreateFakeDeployment("comp1"),
// 	}}
//
// 	deploymentList.Items[0].Labels[componentlabels.ComponentTypeLabel] = "nodejs"
// 	deploymentList.Items[0].Annotations = map[string]string{
// 		componentlabels.ComponentTypeAnnotation: "nodejs",
// 	}
// 	deploymentList.Items[1].Labels[componentlabels.ComponentTypeLabel] = "wildfly"
// 	deploymentList.Items[1].Annotations = map[string]string{
// 		componentlabels.ComponentTypeAnnotation: "wildfly",
// 	}
// 	tests := []struct {
// 		name           string
// 		deploymentList v1.DeploymentList
// 		projectExists  bool
// 		wantErr        bool
// 		mockError      error
// 		output         ComponentList
// 	}{
// 		{
// 			name:          "Case 1: no component and no config exists",
// 			wantErr:       false,
// 			projectExists: true,
// 			output:        newComponentList([]Component{}),
// 			mockError:     fmt.Errorf("some error"),
// 		},
// 		{
// 			name:           "Case 2: Components are returned from deployments on a kubernetes cluster",
// 			deploymentList: deploymentList,
// 			wantErr:        false,
// 			projectExists:  true,
// 			output: ComponentList{
// 				TypeMeta: metav1.TypeMeta{
// 					Kind:       "List",
// 					APIVersion: "odo.dev/v1alpha1",
// 				},
// 				ListMeta: metav1.ListMeta{},
// 				Items: []Component{
// 					getFakeComponent("comp0", namespace, "app", "nodejs", StateTypePushed),
// 					getFakeComponent("comp1", namespace, "app", "wildfly", StateTypePushed),
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			applicationSelector := applabels.GetSelector("app")
// 			mockKClient := kclient.NewMockClientInterface(ctrl)
// 			compClient := NewClient(mockKClient)
// 			var expectedList []*v1.Deployment
// 			for _, dep := range tt.deploymentList.Items {
// 				expectedList = append(expectedList, &dep)
// 			}
// 			mockKClient.EXPECT().GetDeploymentFromSelector(applicationSelector).Return(tt.deploymentList.Items, nil).AnyTimes()
// 			if tt.deploymentList.Items != nil {
// 				for _, dep := range expectedList {
// 					selector := componentlabels.GetSelector(dep.Name, "app")
// 					mockKClient.EXPECT().GetOneDeployment(dep.Name, "app").Return(dep, nil).AnyTimes()
// 					mockKClient.EXPECT().ListIngresses(selector).Return(&unions.KubernetesIngressList{}, nil).AnyTimes()
// 					mockKClient.EXPECT().ListSecrets(selector).Return([]corev1.Secret{}, nil).AnyTimes()
// 					mockKClient.EXPECT().ListServices(selector).Return([]corev1.Service{}, nil).AnyTimes()
// 				}
// 				mockKClient.EXPECT().ListServices("").Return([]corev1.Service{}, nil).AnyTimes()
// 				mockKClient.EXPECT().IsServiceBindingSupported().Return(false, nil).AnyTimes()
// 				mockKClient.EXPECT().GetCurrentNamespace().Return(namespace).AnyTimes()
// 			} else {
// 				mockKClient.EXPECT().GetOneDeployment("comp0", "app").Return(nil, nil).AnyTimes()
// 				mockKClient.EXPECT().GetOneDeployment("comp1", "app").Return(nil, nil).AnyTimes()
// 			}
// 			results, err := compClient.List(applicationSelector)
//
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("expected err: %v, but err is %v", tt.wantErr, err)
// 			}
//
// 			if !reflect.DeepEqual(tt.output, results) {
// 				t.Errorf("expected output:\n%#v\n\ngot:\n%#v", tt.output, results)
// 			}
// 		})
// 	}
// }

func TestGetDefaultComponentName(t *testing.T) {
	tests := []struct {
		testName           string
		componentType      string
		componentPath      string
		existingComponents ComponentList
		wantErr            bool
		wantRE             string
		mockComponentName  string
		needPrefix         bool
	}{
		{
			testName:           "Case: App prefix configured",
			componentType:      "nodejs",
			componentPath:      "./testing",
			existingComponents: ComponentList{},
			wantErr:            false,
			wantRE:             "nodejs-testing-*",
			mockComponentName:  "nodejs-testing-123",
			needPrefix:         true,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cfg := preference.NewMockClient(ctrl)
			cfg.EXPECT().NamePrefix().Return(nil)

			name, err := GetDefaultComponentName(cfg, tt.componentPath, tt.componentType, tt.existingComponents)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected err: %v, but err is %v", tt.wantErr, err)
			}

			r, _ := regexp.Compile(tt.wantRE)
			match := r.MatchString(name)
			if !match {
				t.Errorf("randomly generated application name %s does not match regexp %s", name, tt.wantRE)
			}
		})
	}
}

func TestGetComponentDir(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		testName string
		args     args
		want     string
		wantErr  bool
	}{
		{
			testName: "Case: Source Path",
			args: args{
				path: "./testing",
			},
			wantErr: false,
			want:    "testing",
		},
		{
			testName: "Case: No clue of any component",
			args: args{
				path: "",
			},
			wantErr: false,
			want:    "component",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name, err := GetComponentDir(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected err: %v, but err is %v", tt.wantErr, err)
			}

			if name != tt.want {
				t.Errorf("received name %s which does not match %s", name, tt.want)
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

func TestComponentClient_Exists(t *testing.T) {

}
func getFakeComponent(compName, namespace, appName, compType string, state State) Component {
	return Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "odo.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      compName,
			Namespace: namespace,
			Labels: map[string]string{
				applabels.App:                      appName,
				applabels.ManagedBy:                "odo",
				applabels.ApplicationLabel:         appName,
				componentlabels.ComponentLabel:     compName,
				componentlabels.ComponentTypeLabel: compType,
			},
			Annotations: map[string]string{
				componentlabels.ComponentTypeAnnotation: compType,
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
