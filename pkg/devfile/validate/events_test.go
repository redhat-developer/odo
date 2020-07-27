package validate

import (
	"reflect"
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

func TestValidateEvents(t *testing.T) {

	tests := []struct {
		name               string
		events             common.DevfileEvents
		commands           []common.DevfileCommand
		wantErr            bool
		errorShouldContain []string
	}{
		{
			name: "Case 1: Valid events",
			events: common.DevfileEvents{
				PostStart: []string{
					"event1",
				},
				PreStop: []string{
					"event2",
				},
			},
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						Id: "event1",
					},
				},
				{
					Composite: &common.Composite{
						Id: "event2",
					},
				},
			},
			wantErr:            false,
			errorShouldContain: nil,
		},
		{
			name: "Case 1: Valid events",
			events: common.DevfileEvents{
				PostStop: []string{
					"event1",
				},
				PreStart: []string{
					"event2",
				},
			},
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						Id: "event11",
					},
				},
				{
					Composite: &common.Composite{
						Id: "event22",
					},
				},
			},
			wantErr: true,
			errorShouldContain: []string{
				"preStart type event event2 invalid",
				"postStop type event event1 invalid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEvents(tt.events, tt.commands)
			if err != nil && !tt.wantErr {
				t.Errorf("TestValidateEvents error - %v", err)
			} else if err != nil && tt.wantErr {
				for _, errorString := range tt.errorShouldContain {
					if !strings.Contains(err.Error(), errorString) {
						t.Errorf("TestValidateEvents error mismatch, %v does not contain %s", err.Error(), errorString)
					}
				}
			}
		})
	}

}

func TestIsEventValid(t *testing.T) {

	tests := []struct {
		name       string
		eventNames []string
		eventType  string
		commands   []common.DevfileCommand
		wantErr    bool
	}{
		{
			name: "Case 1: Valid events",
			eventNames: []string{
				"event1",
				"event2",
			},
			eventType: "preStart",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						Id: "event1",
					},
				},
				{
					Composite: &common.Composite{
						Id: "event2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: Invalid events",
			eventNames: []string{
				"event1",
				"event2",
			},
			eventType: "postStop",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						Id: "event11",
					},
				},
				{
					Composite: &common.Composite{
						Id: "event22",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isEventValid(tt.eventNames, tt.eventType, tt.commands)
			if err != nil && !tt.wantErr {
				t.Errorf("TestIsEventValid error: %v", err)
			} else if err != nil && tt.wantErr {
				want := &InvalidEventError{eventType: tt.eventType, event: strings.Join(tt.eventNames, ",")}
				if !reflect.DeepEqual(err, want) {
					t.Errorf("TestIsEventValid error mismatch - got: %v want: %v", err, want)
				}
			}
		})
	}

}
