package kubeportforward

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/api"
	"reflect"
	"testing"
)

func Test_getCompleteCustomPortPairs(t *testing.T) {
	type args struct {
		definedPorts []api.ForwardedPort
		ceMapping    map[string][]v1alpha2.Endpoint
	}
	tests := []struct {
		name          string
		args          args
		wantPortPairs map[string][]string
	}{
		// TODO: Add test cases.
		{
			name: "ports are provided with container name",
			args: args{
				definedPorts: []api.ForwardedPort{
					{ContainerName: "runtime", LocalPort: 8080, ContainerPort: 8000},
				},
				ceMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9000}},
					"tools":   {{TargetPort: 5000}},
				},
			},
			wantPortPairs: map[string][]string{
				"runtime": {"8080:8000", "20001:9000"},
				"tools":   {"20002:5000"},
			},
		},
		{
			name: "ports are provided without container name",
			args: args{
				definedPorts: []api.ForwardedPort{
					{LocalPort: 8080, ContainerPort: 8000},
					{LocalPort: 5000, ContainerPort: 5000},
				},
				ceMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9000}},
					"tools":   {{TargetPort: 5000}},
				},
			},
			wantPortPairs: map[string][]string{
				"runtime": {"8080:8000", "20001:9000"},
				"tools":   {"5000:5000"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPortPairs := getCustomPortPairs(tt.args.definedPorts, tt.args.ceMapping); !reflect.DeepEqual(gotPortPairs, tt.wantPortPairs) {
				t.Errorf("getCompleteCustomPortPairs() = %v, want %v", gotPortPairs, tt.wantPortPairs)
			}
		})
	}
}
