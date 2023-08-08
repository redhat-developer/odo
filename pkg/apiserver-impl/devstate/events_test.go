package devstate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
)

func TestDevfileState_UpdateEvents(t *testing.T) {
	type args struct {
		event    string
		commands []string
	}
	tests := []struct {
		name    string
		state   func(t *testing.T) DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "set preStart event",
			state: func(t *testing.T) DevfileState {
				return NewDevfileState()
			},
			args: args{
				event:    "preStart",
				commands: []string{"command1"},
			},
			want: DevfileContent{
				Content: `events:
  preStart:
  - command1
metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
				Events: Events{
					PreStart: []string{"command1"},
				},
			},
		}, {
			name: "set postStart event when preStart is already set",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.UpdateEvents("preStart", []string{"command1"})
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				event:    "postStart",
				commands: []string{"command2"},
			},
			want: DevfileContent{
				Content: `events:
  postStart:
  - command2
  preStart:
  - command1
metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
				Events: Events{
					PreStart:  []string{"command1"},
					PostStart: []string{"command2"},
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state(t)
			got, err := o.UpdateEvents(tt.args.event, tt.args.commands)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.UpdateEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.UpdateEvents() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
