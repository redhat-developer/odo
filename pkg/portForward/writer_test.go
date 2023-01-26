package portForward

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
				LocalAddress:  "127.0.0.1",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getForwardedPort(tt.args.mapping, tt.args.s)
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
