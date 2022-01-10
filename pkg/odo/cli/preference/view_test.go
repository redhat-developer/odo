package preference

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/preference"

	"k8s.io/utils/pointer"
)

func TestView(t *testing.T) {
	ctrl := gomock.NewController(t)
	prefClient := preference.NewMockClient(ctrl)
	opts := NewViewOptions(prefClient)

	cmdline := cmdline.NewMockCmdline(ctrl)

	args := []string{}
	err := opts.Complete(cmdline, args)
	if err != nil {
		t.Errorf("Expected nil error, got %s", err)
		return
	}

	err = opts.Validate()
	if err != nil {
		t.Errorf("Expected nil error, got %s", err)
		return
	}

	prefClient.EXPECT().UpdateNotification().Return(pointer.Bool(false))
	prefClient.EXPECT().NamePrefix().Return(pointer.String("aprefix"))
	prefClient.EXPECT().Timeout().Return(pointer.Int(10))
	prefClient.EXPECT().BuildTimeout().Return(pointer.Int(10))
	prefClient.EXPECT().PushTimeout().Return(pointer.Int(10))
	prefClient.EXPECT().EphemeralSourceVolume().Return(pointer.Bool(false))
	prefClient.EXPECT().ConsentTelemetry().Return(pointer.Bool(false))

	err = opts.Run()
	if err != nil {
		t.Errorf(`Expected nil error, got %s`, err)
	}
}
