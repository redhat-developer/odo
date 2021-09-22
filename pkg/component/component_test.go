package component

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"

	v1 "k8s.io/api/apps/v1"

	"github.com/devfile/library/pkg/util"
	"github.com/golang/mock/gomock"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestGetComponentFrom(t *testing.T) {
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

func TestGetS2IPaths(t *testing.T) {

	tests := []struct {
		name    string
		podEnvs []corev1.EnvVar
		want    []string
	}{
		{
			name: "Case 1: odo expected s2i envs available",
			podEnvs: []corev1.EnvVar{
				{
					Name:  occlient.EnvS2IDeploymentDir,
					Value: "abc",
				},
				{
					Name:  occlient.EnvS2ISrcOrBinPath,
					Value: "def",
				},
				{
					Name:  occlient.EnvS2IWorkingDir,
					Value: "ghi",
				},
				{
					Name:  occlient.EnvS2ISrcBackupDir,
					Value: "ijk",
				},
			},
			want: []string{
				filepath.FromSlash("abc/src"),
				filepath.FromSlash("def/src"),
				filepath.FromSlash("ghi/src"),
				filepath.FromSlash("ijk/src"),
			},
		},
		{
			name: "Case 2: some of the odo expected s2i envs not available",
			podEnvs: []corev1.EnvVar{
				{
					Name:  occlient.EnvS2IDeploymentDir,
					Value: "abc",
				},
				{
					Name:  occlient.EnvS2ISrcOrBinPath,
					Value: "def",
				},
				{
					Name:  occlient.EnvS2ISrcBackupDir,
					Value: "ijk",
				},
			},
			want: []string{
				filepath.FromSlash("abc/src"),
				filepath.FromSlash("def/src"),
				filepath.FromSlash("ijk/src"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getS2IPaths(tt.podEnvs)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("got: %+v, want: %+v", got, tt.want)
			}
		})
	}
}
func TestGetComponentPorts(t *testing.T) {
	type args struct {
		componentName   string
		applicationName string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		output  []string
	}{
		{
			name: "Case 1: Invalid/Unexisting component name",
			args: args{
				componentName:   "r",
				applicationName: "app",
			},
			wantErr: true,
			output:  []string{},
		},
		{
			name: "Case 2: Valid params with multiple containers each with multiple ports",
			args: args{
				componentName:   "python",
				applicationName: "app",
			},
			output:  []string{"10080/TCP", "8080/TCP", "9090/UDP", "10090/UDP"},
			wantErr: false,
		},
		{
			name: "Case 3: Valid params with single container and single port",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			output:  []string{"8080/TCP"},
			wantErr: false,
		},
		{
			name: "Case 4: Valid params with single container and multiple port",
			args: args{
				componentName:   "wildfly",
				applicationName: "app",
			},
			output:  []string{"8090/TCP", "8080/TCP"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()
			fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, testingutil.FakeDeploymentConfigs(), nil
			})

			// The function we are testing
			output, err := GetComponentPorts(client, tt.args.componentName, tt.args.applicationName)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component List() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Sort the output and expected o/p in-order to avoid issues due to order as its not important
			sort.Strings(output)
			sort.Strings(tt.output)

			// Check if the output is the same as what's expected (tags)
			// and only if output is more than 0 (something is actually returned)
			if len(output) > 0 && !(reflect.DeepEqual(output, tt.output)) {
				t.Errorf("expected tags: %s, got: %s", tt.output, output)
			}
		})
	}
}

func TestGetComponentLinkedSecretNames(t *testing.T) {
	type args struct {
		componentName   string
		applicationName string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		output  []string
	}{
		{
			name: "Case 1: Invalid/Unexisting component name",
			args: args{
				componentName:   "r",
				applicationName: "app",
			},
			wantErr: true,
			output:  []string{},
		},
		{
			name: "Case 2: Valid params nil env source",
			args: args{
				componentName:   "python",
				applicationName: "app",
			},
			output:  []string{},
			wantErr: false,
		},
		{
			name: "Case 3: Valid params multiple secrets",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			output:  []string{"s1", "s2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()

			fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, testingutil.FakeDeploymentConfigs(), nil
			})

			// The function we are testing
			output, err := GetComponentLinkedSecretNames(client, tt.args.componentName, tt.args.applicationName)

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component List() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Sort the output and expected o/p in-order to avoid issues due to order as its not important
			sort.Strings(output)
			sort.Strings(tt.output)

			// Check if the output is the same as what's expected (tags)
			// and only if output is more than 0 (something is actually returned)
			if len(output) > 0 && !(reflect.DeepEqual(output, tt.output)) {
				t.Errorf("expected tags: %s, got: %s", tt.output, output)
			}
		})
	}
}

func TestList(t *testing.T) {
	mockConfig := config.GetOneExistingConfigInfo("comp", "app", "test")
	componentConfig, err := GetComponentFromConfig(&mockConfig)
	if err != nil {
		t.Errorf("error occured while calling GetComponentFromConfig, error: %v", err)
	}
	componentConfig.Status.State = StateTypeNotPushed
	componentConfig2, err := GetComponentFromConfig(&mockConfig)
	if err != nil {
		t.Errorf("error occured while calling GetComponentFromConfig, error: %v", err)
	}
	componentConfig2.Status.State = StateTypeUnknown

	deploymentList := v1.DeploymentList{Items: []v1.Deployment{
		*testingutil.CreateFakeDeployment("comp0"),
		*testingutil.CreateFakeDeployment("comp1"),
	}}

	deploymentList.Items[0].Labels[componentlabels.ComponentTypeLabel] = "nodejs"
	deploymentList.Items[0].Annotations = map[string]string{
		ComponentSourceTypeAnnotation:           "local",
		componentlabels.ComponentTypeAnnotation: "nodejs",
	}
	deploymentList.Items[1].Labels[componentlabels.ComponentTypeLabel] = "wildfly"
	deploymentList.Items[1].Annotations = map[string]string{
		ComponentSourceTypeAnnotation:           "local",
		componentlabels.ComponentTypeAnnotation: "wildfly",
	}

	tests := []struct {
		name                      string
		deploymentConfigSupported bool
		deploymentList            v1.DeploymentList
		projectExists             bool
		wantErr                   bool
		existingLocalConfigInfo   *config.LocalConfigInfo
		output                    ComponentList
	}{
		{
			name:          "Case 1: no component and no config exists",
			wantErr:       false,
			projectExists: true,
			output:        newComponentList([]Component{}),
		},
		{
			name:                      "Case 2: Components are returned from deployments on a kubernetes cluster",
			deploymentList:            deploymentList,
			wantErr:                   false,
			projectExists:             true,
			deploymentConfigSupported: false,
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
			if !tt.deploymentConfigSupported {
				os.Setenv("KUBERNETES", "true")
				defer os.Unsetenv("KUBERNETES")
			}

			client, fakeClientSet := occlient.FakeNew()
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

			results, err := List(client, applabels.GetSelector("app"), tt.existingLocalConfigInfo)

			if (err != nil) != tt.wantErr {
				t.Errorf("expected err: %v, but err is %v", tt.wantErr, err)
			}

			if !reflect.DeepEqual(tt.output, results) {
				t.Errorf("expected output:\n%#v\n\ngot:\n%#v", tt.output, results)
			}
		})
	}
}

func TestGetDefaultComponentName(t *testing.T) {
	tests := []struct {
		testName           string
		componentType      string
		componentPath      string
		componentPathType  config.SrcType
		existingComponents ComponentList
		wantErr            bool
		wantRE             string
		needPrefix         bool
	}{
		{
			testName:           "Case: App prefix not configured",
			componentType:      "nodejs",
			componentPathType:  config.GIT,
			componentPath:      "https://github.com/openshift/nodejs.git",
			existingComponents: ComponentList{},
			wantErr:            false,
			wantRE:             "nodejs-*",
			needPrefix:         false,
		},
		{
			testName:           "Case: App prefix configured",
			componentType:      "nodejs",
			componentPathType:  config.LOCAL,
			componentPath:      "./testing",
			existingComponents: ComponentList{},
			wantErr:            false,
			wantRE:             "nodejs-testing-*",
			needPrefix:         true,
		},
		{
			testName:           "Case: App prefix configured",
			componentType:      "wildfly",
			componentPathType:  config.BINARY,
			componentPath:      "./testing.war",
			existingComponents: ComponentList{},
			wantErr:            false,
			wantRE:             "wildfly-testing-*",
			needPrefix:         true,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			odoConfigFile, kubeConfigFile, err := testingutil.SetUp(
				testingutil.ConfigDetails{
					FileName:      "odo-test-config",
					Config:        testingutil.FakeOdoConfig("odo-test-config", false, ""),
					ConfigPathEnv: "GLOBALODOCONFIG",
				}, testingutil.ConfigDetails{
					FileName:      "kube-test-config",
					Config:        testingutil.FakeKubeClientConfig(),
					ConfigPathEnv: "KUBECONFIG",
				},
			)
			defer testingutil.CleanupEnv([]*os.File{odoConfigFile, kubeConfigFile}, t)
			if err != nil {
				t.Errorf("failed to setup test env. Error %v", err)
			}

			name, err := GetDefaultComponentName(tt.componentPath, tt.componentPathType, tt.componentType, tt.existingComponents)
			if err != nil {
				t.Errorf("failed to setup mock environment. Error: %v", err)
			}

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
		path      string
		paramType config.SrcType
	}
	tests := []struct {
		testName string
		args     args
		want     string
		wantErr  bool
	}{
		{
			testName: "Case: Git URL",
			args: args{
				paramType: config.GIT,
				path:      "https://github.com/openshift/nodejs-ex.git",
			},
			want:    "nodejs-ex",
			wantErr: false,
		},
		{
			testName: "Case: Source Path",
			args: args{
				paramType: config.LOCAL,
				path:      "./testing",
			},
			wantErr: false,
			want:    "testing",
		},
		{
			testName: "Case: Binary path",
			args: args{
				paramType: config.BINARY,
				path:      "./testing.war",
			},
			wantErr: false,
			want:    "testing",
		},
		{
			testName: "Case: No clue of any component",
			args: args{
				paramType: config.NONE,
				path:      "",
			},
			wantErr: false,
			want:    "component",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name, err := GetComponentDir(tt.args.path, tt.args.paramType)
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

func TestGetComponentFromConfig(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv("LOCALODOCONFIG", tempConfigFile.Name())

	localExistingConfigInfoValue := config.GetOneExistingConfigInfo("comp", "app", "project")
	localExistingConfigInfoUrls, err := localExistingConfigInfoValue.LocalConfig.ListURLs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	localExistingConfigInfoStorage, err := localExistingConfigInfoValue.ListStorage()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	localExistingConfigInfoPorts, err := localExistingConfigInfoValue.GetComponentPorts()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	localNonExistingConfigInfoValue := config.GetOneNonExistingConfigInfo()
	gitExistingConfigInfoValue := config.GetOneGitExistingConfigInfo("comp", "app", "project")
	gitExistingConfigInfoUrls, err := gitExistingConfigInfoValue.LocalConfig.ListURLs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	gitExistingConfigInfoStorage, err := gitExistingConfigInfoValue.ListStorage()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	gitExistingConfigInfoPorts, err := gitExistingConfigInfoValue.GetComponentPorts()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tests := []struct {
		name           string
		isConfigExists bool
		existingConfig config.LocalConfigInfo
		wantSpec       Component
	}{
		{
			name:           "case 1: config file exists",
			isConfigExists: true,
			existingConfig: localExistingConfigInfoValue,
			wantSpec: Component{
				Spec: ComponentSpec{
					App:    localExistingConfigInfoValue.GetApplication(),
					Type:   localExistingConfigInfoValue.GetType(),
					Source: localExistingConfigInfoValue.GetSourceLocation(),
					URL: []string{
						localExistingConfigInfoUrls[0].Name, localExistingConfigInfoUrls[1].Name,
					},
					Storage: []string{
						localExistingConfigInfoStorage[0].Name, localExistingConfigInfoStorage[1].Name,
					},
					Env: []corev1.EnvVar{
						{
							Name:  localExistingConfigInfoValue.LocalConfig.GetEnvs()[0].Name,
							Value: localExistingConfigInfoValue.LocalConfig.GetEnvs()[0].Value,
						},
						{
							Name:  localExistingConfigInfoValue.LocalConfig.GetEnvs()[1].Name,
							Value: localExistingConfigInfoValue.LocalConfig.GetEnvs()[1].Value,
						},
					},
					SourceType: "local",
					Ports:      localExistingConfigInfoPorts,
				},
			},
		},
		{
			name:           "case 2: config file doesn't exists",
			isConfigExists: false,
			existingConfig: localNonExistingConfigInfoValue,
			wantSpec:       Component{},
		},
		{
			name:           "case 3: config file exists",
			isConfigExists: true,
			existingConfig: gitExistingConfigInfoValue,
			wantSpec: Component{
				Spec: ComponentSpec{
					App:    gitExistingConfigInfoValue.GetApplication(),
					Type:   gitExistingConfigInfoValue.GetType(),
					Source: gitExistingConfigInfoValue.GetSourceLocation(),
					URL: []string{
						gitExistingConfigInfoUrls[0].Name, gitExistingConfigInfoUrls[1].Name,
					},
					Storage: []string{
						gitExistingConfigInfoStorage[0].Name, gitExistingConfigInfoStorage[1].Name,
					},
					Env: []corev1.EnvVar{
						{
							Name:  gitExistingConfigInfoValue.LocalConfig.GetEnvs()[0].Name,
							Value: gitExistingConfigInfoValue.LocalConfig.GetEnvs()[0].Value,
						},
						{
							Name:  gitExistingConfigInfoValue.LocalConfig.GetEnvs()[1].Name,
							Value: gitExistingConfigInfoValue.LocalConfig.GetEnvs()[1].Value,
						},
					},
					SourceType: "git",
					Ports:      gitExistingConfigInfoPorts,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.NewLocalConfigInfo("")
			if err != nil {
				t.Error(err)
			}
			cfg := &tt.existingConfig

			got, _ := GetComponentFromConfig(cfg)

			if !reflect.DeepEqual(got.Spec, tt.wantSpec.Spec) {
				t.Errorf("the component spec is different, want: %v,\n got: %v", tt.wantSpec.Spec, got.Spec)
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

func getFakeComponent(compName, namespace, appName, compType string, state State) Component {
	sourceType := "local"
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
				ComponentSourceTypeAnnotation:           sourceType,
			},
		},
		Spec: ComponentSpec{
			Type:       compType,
			App:        appName,
			SourceType: sourceType,
		},
		Status: ComponentStatus{
			State: state,
		},
	}

}
