package asker

import (
	"reflect"
	"testing"
)

func Test_buildPersonalizedConfigurationOptions(t *testing.T) {
	type args struct {
		configuration ContainerConfiguration
	}
	tests := []struct {
		name        string
		args        args
		wantOptions []string
		wantTracker []OperationOnContainer
	}{
		{
			name: "default options",
			args: args{configuration: ContainerConfiguration{
				Ports: []string{},
				Envs:  map[string]string{},
			}},
			wantOptions: []string{
				"NOTHING - configuration is correct",
				"Add new port",
				"Add new environment variable",
			},
			wantTracker: []OperationOnContainer{
				{
					Ops:  "Nothing",
					Kind: "",
					Key:  "",
				}, {
					Ops:  "Add",
					Kind: "Port",
					Key:  "",
				}, {
					Ops:  "Add",
					Kind: "EnvVar",
					Key:  "",
				}},
		},
		{
			name: "all options",
			args: args{configuration: ContainerConfiguration{
				Ports: []string{"7000", "8000"},
				Envs:  map[string]string{"foo": "bar"},
			}},
			wantOptions: []string{
				"NOTHING - configuration is correct",
				"Delete port \"7000\"",
				"Delete port \"8000\"",
				"Add new port",
				"Delete environment variable \"foo\"",
				"Add new environment variable",
			},
			wantTracker: []OperationOnContainer{
				{
					Ops:  "Nothing",
					Kind: "",
					Key:  "",
				}, {
					Ops:  "Delete",
					Kind: "Port",
					Key:  "7000",
				}, {
					Ops:  "Delete",
					Kind: "Port",
					Key:  "8000",
				}, {
					Ops:  "Add",
					Kind: "Port",
					Key:  "",
				}, {
					Ops:  "Delete",
					Kind: "EnvVar",
					Key:  "foo",
				}, {
					Ops:  "Add",
					Kind: "EnvVar",
					Key:  "",
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOptions, gotTracker := buildPersonalizedConfigurationOptions(tt.args.configuration)
			if !reflect.DeepEqual(gotOptions, tt.wantOptions) {
				t.Errorf("buildPersonalizedConfigurationOptions() gotOptions = %v, want %v", gotOptions, tt.wantOptions)
			}
			if !reflect.DeepEqual(gotTracker, tt.wantTracker) {
				t.Errorf("buildPersonalizedConfigurationOptions() gotTracker = %v, want %v", gotTracker, tt.wantTracker)
			}
		})
	}
}
