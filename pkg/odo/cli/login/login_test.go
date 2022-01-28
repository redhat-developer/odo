package login

import (
	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/auth"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"testing"
)

func TestLoginOptions_Complete(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "Case 1: No arguments",
			args: []string{},
			want: "",
		},
		{
			name: "Case 2: One argument",
			args: []string{"one-server"},
			want: "one-server",
		},
		{
			name: "Case 3: Multiple arguments",
			args: []string{"blah", "blah"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fake Cobra
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cmdline := cmdline.NewMockCmdline(ctrl)

			loginOptions := NewLoginOptions(nil)
			_ = loginOptions.Complete(cmdline, tt.args)

			if loginOptions.server != tt.want {
				t.Errorf("got %s; wanted %s", loginOptions.server, tt.want)
			}
		})
	}
}

func TestLoginOptions_Validate(t *testing.T) {
	tests := []struct {
		name            string
		serverParam     string
		serverFlag      string
		wantErr         bool
		wantEqualValues bool
	}{
		{
			name:            "Case 1: Both --server flag and server link as parameter, but different values",
			serverParam:     "value1",
			serverFlag:      "value2",
			wantErr:         true,
			wantEqualValues: false,
		},
		{
			name:            "Case 2: Only server flag provided",
			serverParam:     "",
			serverFlag:      "value",
			wantErr:         false,
			wantEqualValues: false,
		},
		{
			name:            "Case 3: Only server link provided as parameter",
			serverParam:     "value1",
			serverFlag:      "",
			wantErr:         false,
			wantEqualValues: true,
		},
		{
			name:            "Case 4: Same value provided for both flag and parameter",
			serverParam:     "value",
			serverFlag:      "value",
			wantErr:         false,
			wantEqualValues: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fake Cobra
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := auth.NewMockClient(ctrl)

			loginOptions := NewLoginOptions(client)
			loginOptions.server = tt.serverParam
			loginOptions.serverFlag = tt.serverFlag

			err := loginOptions.Validate()
			if err == nil && tt.wantErr {
				t.Errorf("got no error when one expected")
			}

			if tt.wantEqualValues && loginOptions.serverFlag != loginOptions.server {
				t.Errorf("wanted equal values for server flag and parameter, but values differ\nserver flag=%s server parame=%s", loginOptions.serverFlag, loginOptions.server)
			}
		})
	}
}
