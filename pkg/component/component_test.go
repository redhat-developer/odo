package component

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestGetComponentPorts(t *testing.T) {
	type args struct {
		componentName   string
		applicationName string
		namespace       string
		componentType   string
		containerPort   int32
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
				componentName:   "nodejs",
				applicationName: "nodejs",
				namespace:       "myproject",
				componentType:   "nodejs",
				containerPort:   8080,
			},
			wantErr: true,
			output:  []string{},
		},
		{
			name: "Case 2: Valid params",
			args: args{
				componentName:   "python",
				applicationName: "app",
				namespace:       "myproject",
				componentType:   "python",
				containerPort:   8080,
			},
			output:  []string{"8080/TCP"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Fake the client with the appropriate arguments
			client, fakeClientSet := occlient.FakeNew()
			fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, testingutil.FakeDeploymentConfigs(tt.args.namespace, tt.args.componentName, tt.args.componentType, tt.args.applicationName, tt.args.containerPort), nil
			})

			// The function we are testing
			output, err := GetComponentPorts(client, tt.args.componentName, tt.args.applicationName)

			//Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component List() unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check if the output is the same as what's expected (tags)
			// and only if output is more than 0 (something is actually returned)
			if len(output) > 0 && !(reflect.DeepEqual(output, tt.output)) {
				t.Errorf("expected tags: %s, got: %s", tt.output, output)
			}
		})
	}
}
