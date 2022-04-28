package dev

import (
	"reflect"
	"testing"

	"github.com/redhat-developer/odo/pkg/api"
)

func Test_getForwardedPort(t *testing.T) {
	type args struct {
		mapping map[string][]int
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
				mapping: map[string][]int{
					"container1": {3000, 4200},
					"container2": {80, 8080},
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
				mapping: map[string][]int{
					"container1": {3000, 4200},
					"container2": {80, 8080},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getForwardedPort() = %v, want %v", got, tt.want)
			}
		})
	}
}
