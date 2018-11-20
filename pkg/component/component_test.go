package component

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/redhat-developer/odo/pkg/models"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

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
			name: "Case 1: Invalid/Non-existant component name",
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

func TestGetDefaultComponentName(t *testing.T) {
	tests := []struct {
		testName           string
		componentType      string
		componentPath      string
		componentPathType  models.CreateType
		existingComponents []ComponentInfo
		wantErr            bool
		wantRE             string
		needPrefix         bool
	}{
		{
			testName:           "Case: App prefix not configured",
			componentType:      "nodejs",
			componentPathType:  models.GIT,
			componentPath:      "https://github.com/openshift/nodejs.git",
			existingComponents: []ComponentInfo{},
			wantErr:            false,
			wantRE:             "nodejs-*",
			needPrefix:         false,
		},
		{
			testName:           "Case: App prefix configured",
			componentType:      "nodejs",
			componentPathType:  models.LOCAL,
			componentPath:      "./testing",
			existingComponents: []ComponentInfo{},
			wantErr:            false,
			wantRE:             "testing-nodejs-*",
			needPrefix:         true,
		},
		{
			testName:           "Case: App prefix configured",
			componentType:      "wildfly",
			componentPathType:  models.BINARY,
			componentPath:      "./testing.war",
			existingComponents: []ComponentInfo{},
			wantErr:            false,
			wantRE:             "testing-wildfly-*",
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
					ConfigPathEnv: "ODOCONFIG",
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
		paramType models.CreateType
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
				paramType: models.GIT,
				path:      "https://github.com/openshift/nodejs-ex.git",
			},
			want:    "nodejs-ex",
			wantErr: false,
		},
		{
			testName: "Case: Source Path",
			args: args{
				paramType: models.LOCAL,
				path:      "./testing",
			},
			wantErr: false,
			want:    "testing",
		},
		{
			testName: "Case: Binary path",
			args: args{
				paramType: models.BINARY,
				path:      "./testing.war",
			},
			wantErr: false,
			want:    "testing",
		},
		{
			testName: "Case: No clue of any component",
			args: args{
				paramType: models.NONE,
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
