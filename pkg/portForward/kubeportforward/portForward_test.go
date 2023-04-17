package kubeportforward

import (
	"fmt"
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/api"
	"strconv"
	"strings"
	"testing"
)

func Test_getCompleteCustomPortPairs(t *testing.T) {
	const (
		acceptablePortRange = "[20001-30001]"
		acceptableMinPort   = 20001
		acceptableMaxPort   = 30001
	)
	type args struct {
		definedPorts []api.ForwardedPort
		ceMapping    map[string][]v1alpha2.Endpoint
	}
	tests := []struct {
		name string
		args args
		// wantPortPairs format: {containerName: {containerPort: localPort}}
		wantPortPairs map[string]map[string]string
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
			wantPortPairs: map[string]map[string]string{
				"runtime": {"8000": "8080", "9000": acceptablePortRange},
				"tools":   {"5000": acceptablePortRange},
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
			wantPortPairs: map[string]map[string]string{
				"runtime": {"8000": "8080", "9000": acceptablePortRange},
				"tools":   {"5000": "5000"},
			},
		},
		{
			name: "local ports in range [20001-30001] are provided as custom forward ports",
			args: args{
				definedPorts: []api.ForwardedPort{
					{LocalPort: 25001, ContainerPort: 8000},
					{LocalPort: 25002, ContainerPort: 9000},
					{LocalPort: 5000, ContainerPort: 5000},
				},
				ceMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9000}},
					"tools":   {{TargetPort: 5000}, {TargetPort: 8080}},
				},
			},
			wantPortPairs: map[string]map[string]string{
				"runtime": {"8000": "25001", "9000": "25002"},
				"tools":   {"5000": "5000", "8080": acceptablePortRange},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPortPairs := getCustomPortPairs(tt.args.definedPorts, tt.args.ceMapping)

			validatePortPairs := func(gotPortPairs map[string][]string, wantPortPairs map[string]map[string]string) (diff string) {
				for container, portPairs := range gotPortPairs {
					wantPortPair := wantPortPairs[container]
					for _, portPair := range portPairs {
						portMap := strings.Split(portPair, ":")
						lPort, cPort := portMap[0], portMap[1]
						wantLPort := wantPortPair[cPort]
						if wantLPort == acceptablePortRange {
							if intLPort, _ := strconv.Atoi(lPort); intLPort >= acceptableMinPort && intLPort <= acceptableMaxPort {
								continue
							} else {
								diff += fmt.Sprintf("[container %q] %s:%s is not in range %s\n", container, cPort, lPort, acceptablePortRange)
							}
						} else if wantLPort == lPort {
							continue
						} else {
							diff += fmt.Sprintf("[container %q] %s:%s does not match %s:%s\n", container, cPort, lPort, cPort, wantLPort)
						}
					}
				}
				return diff
			}
			if diff := validatePortPairs(gotPortPairs, tt.wantPortPairs); diff != "" {
				t.Errorf("getCompleteCustomPortPairs() (got vs want) diff = %v", diff)
			}
		})
	}
}
