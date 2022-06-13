package preference

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/preference"
)

func TestUnsetForce(t *testing.T) {
	tests := []struct {
		name           string
		forceFlag      bool
		exists         bool
		expectedRunErr string
	}{
		{
			name:      "force && parameter exists",
			forceFlag: true,
			exists:    true,
		},
		{
			name:           "no force and parameter not exists",
			forceFlag:      false,
			exists:         false,
			expectedRunErr: "is already unset",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			opts := NewUnsetOptions()
			opts.SetClientset(&clientset.Clientset{
				PreferenceClient: prefClient,
			})
			opts.forceFlag = tt.forceFlag

			cmdline := cmdline.NewMockCmdline(ctrl)

			args := []string{"Arg1"}
			err := opts.Complete(cmdline, args)
			if err != nil {
				t.Errorf("Expected nil error, got %s", err)
				return
			}

			if opts.paramName != "arg1" {
				t.Errorf("Expected paramName %q, got %q", "arg1", opts.paramName)
			}

			err = opts.Validate()
			if err != nil {
				t.Errorf("Expected nil error, got %s", err)
				return
			}

			if tt.exists || tt.forceFlag {
				prefClient.EXPECT().DeleteConfiguration("arg1")
			}
			if !tt.forceFlag {
				prefClient.EXPECT().IsSet("arg1").Return(tt.exists)
			}
			err = opts.Run(context.Background())

			if err == nil && tt.expectedRunErr != "" {
				t.Errorf("Expected %v, got no error", tt.expectedRunErr)
				return
			}
			if err != nil && tt.expectedRunErr == "" {
				t.Errorf("Expected no error, got %v", err.Error())
				return
			}
			if err != nil && tt.expectedRunErr != "" && !strings.Contains(err.Error(), tt.expectedRunErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedRunErr, err.Error())
				return
			}
		})
	}

}
