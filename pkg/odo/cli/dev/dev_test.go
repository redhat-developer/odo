package dev

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/odo/pkg/api"
	"reflect"
	"testing"
)

func Test_validatePortForwardFlagData(t *testing.T) {
	type args struct {
		forwardedPorts           []api.ForwardedPort
		containerEndpointMapping map[string][]v1alpha2.Endpoint
	}
	tests := []struct {
		name          string
		args          args
		wantErr       bool
		wantErrString string
	}{
		// TODO: Add test cases.
		{
			name: "container name is present for all forwarded ports and is also present in the container-endpoint mapping",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{ContainerName: "runtime", LocalPort: 8080, ContainerPort: 8000},
					{ContainerName: "runtime", LocalPort: 9000, ContainerPort: 9090},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9090}},
					"tools":   {{TargetPort: 5050}},
				},
			},
			wantErr: false,
		},
		{
			name: "container name is absent for all forwarded ports",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{LocalPort: 8080, ContainerPort: 8000},
					{LocalPort: 9000, ContainerPort: 9090},
					{LocalPort: 5000, ContainerPort: 5050},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9090}},
					"tools":   {{TargetPort: 5050}},
				},
			},
			wantErr: false,
		},
		{
			name: "container name is present in some of the forwarded ports",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{LocalPort: 8080, ContainerPort: 8000},
					{LocalPort: 9000, ContainerPort: 9090},
					{ContainerName: "tools", LocalPort: 5000, ContainerPort: 5050},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9090}},
					"tools":   {{TargetPort: 5050}},
				},
			},
			wantErr: false,
		},
		{
			name: "container port(without container name) defined by a forwarded port is not found in the container-endpoint mapping",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{LocalPort: 8080, ContainerPort: 8080},
					{LocalPort: 9000, ContainerPort: 9090},
					{LocalPort: 5000, ContainerPort: 5050},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9090}},
					"tools":   {{TargetPort: 5050}},
				},
			},
			wantErr:       true,
			wantErrString: "8080 container port defined by --port-forward not found",
		},
		{
			name: "container port(with container name) defined by a forwarded port is not found in the container-endpoint mapping",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{ContainerName: "runtime", LocalPort: 8080, ContainerPort: 8080},
					{LocalPort: 9000, ContainerPort: 9090},
					{LocalPort: 5000, ContainerPort: 5050},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9090}},
					"tools":   {{TargetPort: 5050}},
				},
			},
			wantErr:       true,
			wantErrString: "8080 container port does not match any endpoints of container:runtime",
		},
		{
			name: "container name defined by a forwarded port is not found in the container-endpoint mapping",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{ContainerName: "invisible", LocalPort: 8080, ContainerPort: 8080},
					{LocalPort: 9000, ContainerPort: 9090},
					{LocalPort: 5000, ContainerPort: 5050},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 8000}, {TargetPort: 9090}},
					"tools":   {{TargetPort: 5050}},
				},
			},
			wantErr:       true,
			wantErrString: "container:invisible defined by --port-forward not found",
		},
		{
			name: "duplicate target port in containers when a port forward does not container container name",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{LocalPort: 9000, ContainerPort: 9090},
					{LocalPort: 9001, ContainerPort: 9090},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 9090}, {TargetPort: 8000}},
					"tools":   {{TargetPort: 9090}},
				},
			},
			wantErr:       true,
			wantErrString: "multiple container component (runtime, tools) found with same container port 9090, port forwarding must be defined with format <localPort>:<containerName>:<containerPort>",
		},
		{
			name: "duplicate local port cannot be used",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{LocalPort: 9000, ContainerPort: 9090},
					{LocalPort: 9000, ContainerPort: 8090},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 9090}, {TargetPort: 8090}},
					"tools":   {{TargetPort: 5000}},
				},
			},
			wantErr:       true,
			wantErrString: "local port 9000 is used more than once, please use unique local ports",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePortForwardFlagData(tt.args.forwardedPorts, tt.args.containerEndpointMapping)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePortForwardFlagData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if diff := cmp.Diff(err.Error(), tt.wantErrString); diff != "" {
					t.Errorf("validatePortForwardFlagData() (error vs. wantErr) diff= %v", diff)
				}
			}
		})
	}
}

func Test_parsePortForwardFlag(t *testing.T) {
	type args struct {
		portForwardFlag []string
	}
	tests := []struct {
		name               string
		args               args
		wantForwardedPorts []api.ForwardedPort
		wantErr            bool
	}{
		// TODO: Add test cases.
		{
			name: "<localPort>:<containerPort>",
			args: args{
				portForwardFlag: []string{"8080:8000", "9090:9000"},
			},
			wantForwardedPorts: []api.ForwardedPort{
				{
					LocalPort:     8080,
					ContainerPort: 8000,
				},
				{LocalPort: 9090, ContainerPort: 9000},
			},
			wantErr: false,
		},
		{
			name: "<localPort>:<containerName>:<containerPort>",
			args: args{
				portForwardFlag: []string{"8080:runtime:8000", "9090:tools:9000"},
			},
			wantForwardedPorts: []api.ForwardedPort{
				{
					ContainerName: "runtime",
					LocalPort:     8080,
					ContainerPort: 8000,
				},
				{ContainerName: "tools", LocalPort: 9090, ContainerPort: 9000},
			},
			wantErr: false,
		},
		{
			name: "<localPort>:<validContainerName>:<containerPort>",
			args: args{
				portForwardFlag: []string{"8080:runtime_123:8000", "9090:tools:9000"},
			},
			wantForwardedPorts: []api.ForwardedPort{
				{
					ContainerName: "runtime_123",
					LocalPort:     8080,
					ContainerPort: 8000,
				},
				{ContainerName: "tools", LocalPort: 9090, ContainerPort: 9000},
			},
			wantErr: false,
		},
		{
			name: "port values are within a given range <localPort>:<validContainerName>:<containerPort>",
			args: args{
				portForwardFlag: []string{"1:runtime_123:65535", "9090:tools:9000"},
			},
			wantForwardedPorts: []api.ForwardedPort{
				{
					ContainerName: "runtime_123",
					LocalPort:     1,
					ContainerPort: 65535,
				},
				{ContainerName: "tools", LocalPort: 9090, ContainerPort: 9000},
			},
			wantErr: false,
		},
		{
			name: "port values are out of range <localPort>:<containerPort>",
			args: args{
				portForwardFlag: []string{"0:65536"},
			},
			wantForwardedPorts: nil,
			wantErr:            true,
		},
		{
			name: "invalid pattern <containerName>:<localPort>:<containerPort>",
			args: args{
				portForwardFlag: []string{"runtime:8080:8000"},
			},
			wantForwardedPorts: nil,
			wantErr:            true,
		},
		{
			name: "invalid container name <localPort>:<invalidContainerName>:<containerPort>",
			args: args{
				portForwardFlag: []string{"8080:runtime-123:8000"},
			},
			wantForwardedPorts: nil,
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotForwardedPorts, err := parsePortForwardFlag(tt.args.portForwardFlag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePortForwardFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotForwardedPorts, tt.wantForwardedPorts) {
				t.Errorf("parsePortForwardFlag() gotForwardedPorts = %v, want %v", gotForwardedPorts, tt.wantForwardedPorts)
			}
		})
	}
}
