package kubeportforward

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/api"
)

func Test_getForwardedPort(t *testing.T) {
	type args struct {
		mapping map[string][]v1alpha2.Endpoint
		s       string
		address string
	}
	tests := []struct {
		name    string
		args    args
		want    api.ForwardedPort
		wantErr bool
	}{
		{
			name: "find port in container",
			args: args{
				mapping: map[string][]v1alpha2.Endpoint{
					"container1": {
						v1alpha2.Endpoint{Name: "port-11", TargetPort: 3000},
						v1alpha2.Endpoint{Name: "debug-11", TargetPort: 4200},
					},
					"container2": {
						v1alpha2.Endpoint{Name: "port-21", TargetPort: 80},
						v1alpha2.Endpoint{Name: "port-22", TargetPort: 8080},
					},
				},
				s: "Forwarding from 127.0.0.1:40407 -> 3000",
			},
			want: api.ForwardedPort{
				ContainerName: "container1",
				PortName:      "port-11",
				LocalAddress:  "127.0.0.1",
				IsDebug:       false,
				LocalPort:     40407,
				ContainerPort: 3000,
			},
			wantErr: false,
		},
		{
			name: "find port in container and use custom address",
			args: args{
				mapping: map[string][]v1alpha2.Endpoint{
					"container1": {
						v1alpha2.Endpoint{Name: "port-11", TargetPort: 3000},
						v1alpha2.Endpoint{Name: "debug-11", TargetPort: 4200},
					},
					"container2": {
						v1alpha2.Endpoint{Name: "port-21", TargetPort: 80},
						v1alpha2.Endpoint{Name: "port-22", TargetPort: 8080},
					},
				},
				s:       "Forwarding from 192.168.0.1:40407 -> 3000",
				address: "192.168.0.1",
			},
			want: api.ForwardedPort{
				ContainerName: "container1",
				PortName:      "port-11",
				LocalAddress:  "192.168.0.1",
				IsDebug:       false,
				LocalPort:     40407,
				ContainerPort: 3000,
			},
			wantErr: false,
		},
		{
			name: "string error",
			args: args{
				mapping: map[string][]v1alpha2.Endpoint{
					"container1": {
						v1alpha2.Endpoint{Name: "port-11", TargetPort: 3000},
						v1alpha2.Endpoint{Name: "debug-11", TargetPort: 4200},
					},
					"container2": {
						v1alpha2.Endpoint{Name: "port-21", TargetPort: 80},
						v1alpha2.Endpoint{Name: "port-22", TargetPort: 8080},
					},
				},
				s: "Forwarding from 127.0.0.1:40407 => 3000",
			},
			want:    api.ForwardedPort{},
			wantErr: true,
		},
		{
			name: "find debug port in container",
			args: args{
				mapping: map[string][]v1alpha2.Endpoint{
					"container1": {
						v1alpha2.Endpoint{Name: "port-11", TargetPort: 3000},
						v1alpha2.Endpoint{Name: "debug-11", TargetPort: 4200},
					},
					"container2": {
						v1alpha2.Endpoint{Name: "port-21", TargetPort: 80},
						v1alpha2.Endpoint{Name: "port-22", TargetPort: 8080},
					},
				},
				s: "Forwarding from 127.0.0.1:40407 -> 4200",
			},
			want: api.ForwardedPort{
				ContainerName: "container1",
				PortName:      "debug-11",
				IsDebug:       true,
				LocalAddress:  "127.0.0.1",
				LocalPort:     40407,
				ContainerPort: 4200,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getForwardedPort(tt.args.mapping, tt.args.s, tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("getForwardedPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getForwardedPort() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
