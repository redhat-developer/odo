package validate

import (
	"github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"testing"
)

func Test_validateEvents(t *testing.T) {

	var tests = []struct {
		name    string
		events  v1alpha2.Events
		wantErr bool
	}{
		{
			name: "just postStart event present",
			events: v1alpha2.Events{
				WorkspaceEvents: v1alpha2.WorkspaceEvents{
					PostStart: []string{"asdf"},
				},
			},
			wantErr: false,
		},
		{
			name: "preStart event present",
			events: v1alpha2.Events{
				WorkspaceEvents: v1alpha2.WorkspaceEvents{
					PostStart: []string{"asdf"},
					PreStart:  []string{"asdf"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateEvents(tt.events); (err != nil) != tt.wantErr {
				t.Errorf("validateEvents() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
