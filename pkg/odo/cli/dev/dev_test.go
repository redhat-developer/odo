package dev

import (
	"fmt"
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/redhat-developer/odo/pkg/api"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_validatePortForwardFlagData(t *testing.T) {
	type serverCloser interface {
		Close()
	}
	type args struct {
		forwardedPorts           []api.ForwardedPort
		containerEndpointMapping map[string][]v1alpha2.Endpoint
		address                  string
		setupServerFunc          func(address string) (serverCloser, error)
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		wantErrStrings []string
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
			wantErr:        true,
			wantErrStrings: []string{"container port 8080 not found in the devfile container endpoints"},
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
			wantErr:        true,
			wantErrStrings: []string{"container port 8080 does not match any endpoints of container \"runtime\" in the devfile"},
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
			wantErr:        true,
			wantErrStrings: []string{"container \"invisible\" not found in the devfile"},
		},
		{
			name: "duplicate container ports when a port mapping does not container container name",
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
			wantErr:        true,
			wantErrStrings: []string{"multiple container components (runtime, tools) found with same container port 9090 in the devfile, port forwarding must be defined with format <localPort>:<containerName>:<containerPort>"},
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
			wantErr:        true,
			wantErrStrings: []string{"local port 9000 is used more than once, please use unique local ports"},
		},
		{
			name: "port mapping contains multiple invalids",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{LocalPort: 9000, ContainerPort: 9090},
					{LocalPort: 9000, ContainerPort: 5858},
					{LocalPort: 9090, ContainerPort: 5858},
					{LocalPort: 8000, ContainerPort: 3000, ContainerName: "invisible"},
					{LocalPort: 5000, ContainerPort: 3000, ContainerName: "runtime"},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 9090}, {TargetPort: 5858}},
					"tools":   {{TargetPort: 5858}, {TargetPort: 3000}},
				},
			},
			wantErr:        true,
			wantErrStrings: []string{"container port 3000 does not match any endpoints of container \"runtime\" in the devfile", "multiple container components (runtime, tools) found with same container port 5858 in the devfile, port forwarding must be defined with format <localPort>:<containerName>:<containerPort>", "container \"invisible\" not found in the devfile", "local port 9000 is used more than once, please use unique local ports"},
		},
		{
			name: "local port is busy",
			args: args{
				forwardedPorts: []api.ForwardedPort{
					{LocalPort: 9000, ContainerPort: 9000},
				},
				containerEndpointMapping: map[string][]v1alpha2.Endpoint{
					"runtime": {{TargetPort: 9000}},
				},
				setupServerFunc: func(address string) (serverCloser, error) {
					l, err := net.Listen("tcp", fmt.Sprintf("%s:9000", address))
					if err != nil {
						return nil, err
					}
					s := &httptest.Server{
						Listener: l,
						Config:   &http.Server{},
					}
					s.Start()

					return s, nil
				},
				address: "localhost",
			},
			wantErr:        true,
			wantErrStrings: []string{"local port 9000 is already in use on address localhost"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.setupServerFunc != nil {
				sCloser, err := tt.args.setupServerFunc(tt.args.address)
				if err != nil {
					t.Errorf("failed to setup server: %s", err.Error())
					return
				}
				defer sCloser.Close()
			}
			errStrings, err := validatePortForwardFlagData(tt.args.forwardedPorts, tt.args.containerEndpointMapping, tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePortForwardFlagData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if diff := cmp.Diff(errStrings, tt.wantErrStrings, cmpopts.SortSlices(func(x, y string) bool { return x < y })); diff != "" {
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
			if diff := cmp.Diff(gotForwardedPorts, tt.wantForwardedPorts); diff != "" {
				t.Errorf("parsePortForwardFlag() (got vs want) diff = %v", diff)
			}
		})
	}
}

func Test_validateCustomAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "--address=localhost",
			args: args{
				address: "localhost",
			},
			wantErr: false,
		},
		{
			name: "--address is a valid IPv4",
			args: args{
				address: "192.168.0.1",
			},
			wantErr: false,
		},
		{
			name: "--address=0.0.0.0",
			args: args{
				address: "0.0.0.0",
			},
			wantErr: false,
		},
		{
			name: "--address is a valid IPv6 address",
			args: args{
				address: "ABCD:EF01:2345:6789:ABCD:EF01:2345:6789",
			},
			wantErr: false,
		},
		{
			name: "--address=::1",
			args: args{
				address: "::1",
			},
			wantErr: false,
		},
		{
			name: "--address is not a valid address",
			args: args{
				address: "e9:9e:06:ee:8c:4c",
			},
			wantErr: true,
		},
		{
			name: "--address=something",
			args: args{
				address: "something",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateCustomAddress(tt.args.address); (err != nil) != tt.wantErr {
				t.Errorf("validateCustomAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
