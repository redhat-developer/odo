package preference

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/preference"
)

func TestView(t *testing.T) {
	ctrl := gomock.NewController(t)
	prefClient := preference.NewMockClient(ctrl)
	opts := NewViewOptions(prefClient)

	cmdline := cmdline.NewMockCmdline(ctrl)

	args := []string{}
	err := opts.Complete("view", cmdline, args)
	if err != nil {
		t.Errorf("Expected nil error, got %s", err)
		return
	}

	err = opts.Validate()
	if err != nil {
		t.Errorf("Expected nil error, got %s", err)
		return
	}

	prefClient.EXPECT().UpdateNotification()
	prefClient.EXPECT().NamePrefix()
	prefClient.EXPECT().Timeout()
	prefClient.EXPECT().BuildTimeout()
	prefClient.EXPECT().PushTimeout()
	prefClient.EXPECT().EphemeralSourceVolume()
	prefClient.EXPECT().ConsentTelemetry()

	err = opts.Run()
	if err != nil {
		t.Errorf(`Expected nil error, got %s`, err)
	}
}
