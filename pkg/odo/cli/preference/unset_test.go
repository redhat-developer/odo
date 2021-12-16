package preference

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/preference"
)

func TestUnsetForce(t *testing.T) {
	ctrl := gomock.NewController(t)
	prefClient := preference.NewMockClient(ctrl)
	opts := NewUnsetOptions(prefClient)
	opts.forceFlag = true

	cmdline := cmdline.NewMockCmdline(ctrl)

	args := []string{"Arg1"}
	err := opts.Complete("unset", cmdline, args)
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

	prefClient.EXPECT().DeleteConfiguration("arg1")
	err = opts.Run()
}

func TestUnset(t *testing.T) {
	ctrl := gomock.NewController(t)
	prefClient := preference.NewMockClient(ctrl)
	opts := NewUnsetOptions(prefClient)
	opts.forceFlag = false

	cmdline := cmdline.NewMockCmdline(ctrl)

	args := []string{"Arg1"}
	err := opts.Complete("unset", cmdline, args)
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

	prefClient.EXPECT().IsSet("arg1").Return(false)
	err = opts.Run()
	if err == nil || !strings.Contains(err.Error(), "preference already unset") {
		t.Errorf(`Expected error "preference already unset", got nil`)
	}
}
