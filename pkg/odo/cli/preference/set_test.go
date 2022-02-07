package preference

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/preference"
)

func TestSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	prefClient := preference.NewMockClient(ctrl)
	opts := NewSetOptions()
	opts.SetClientset(&clientset.Clientset{
		PreferenceClient: prefClient,
	})
	opts.forceFlag = true

	cmdline := cmdline.NewMockCmdline(ctrl)

	args := []string{"Arg1", "Arg2"}
	err := opts.Complete(cmdline, args)
	if err != nil {
		t.Errorf("Expected nil error, got %s", err)
		return
	}

	if opts.paramName != "arg1" {
		t.Errorf("Expected paramName %q, got %q", "arg1", opts.paramName)
	}
	if opts.paramValue != "Arg2" {
		t.Errorf("Expected paramValue %q, got %q", "Arg2", opts.paramName)
	}

	err = opts.Validate()
	if err != nil {
		t.Errorf("Expected nil error, got %s", err)
		return
	}

	prefClient.EXPECT().SetConfiguration("arg1", "Arg2")
	err = opts.Run()
	if err != nil {
		t.Errorf("Expected nil error, got %s", err)
	}
}
