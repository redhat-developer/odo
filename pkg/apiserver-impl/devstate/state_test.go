package devstate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDevfileState_AddContainer(t *testing.T) {
	type args struct {
		name       string
		image      string
		command    []string
		args       []string
		memRequest string
		memLimit   string
		cpuRequest string
		cpuLimit   string
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			state: func() DevfileState {
				return NewDevfileState()
			},
			args: args{
				name:       "a-name",
				image:      "an-image",
				command:    []string{"run", "command"},
				args:       []string{"arg1", "arg2"},
				memRequest: "1Gi",
				memLimit:   "2Gi",
				cpuRequest: "100m",
				cpuLimit:   "200m",
			},
			want: DevfileContent{
				Content: `components:
- container:
    args:
    - arg1
    - arg2
    command:
    - run
    - command
    cpuLimit: 200m
    cpuRequest: 100m
    image: an-image
    memoryLimit: 2Gi
    memoryRequest: 1Gi
  name: a-name
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{},
				Containers: []Container{
					{
						Name:          "a-name",
						Image:         "an-image",
						Command:       []string{"run", "command"},
						Args:          []string{"arg1", "arg2"},
						MemoryRequest: "1Gi",
						MemoryLimit:   "2Gi",
						CpuRequest:    "100m",
						CpuLimit:      "200m",
					},
				},
				Images:    []Image{},
				Resources: []Resource{},
				Events:    Events{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.AddContainer(tt.args.name, tt.args.image, tt.args.command, tt.args.args, tt.args.memRequest, tt.args.memLimit, tt.args.cpuRequest, tt.args.cpuLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.AddContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.AddContainer() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.AddContainer() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
