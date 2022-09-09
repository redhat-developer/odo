package preference

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/preference"
)

func TestView(t *testing.T) {
	ctrl := gomock.NewController(t)
	prefClient := preference.NewMockClient(ctrl)
	opts := NewViewOptions()
	opts.SetClientset(&clientset.Clientset{
		PreferenceClient: prefClient,
	})

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
	boolValue := true
	intValue := 5
	var intNilValue *int = nil
	var boolNilValue *bool = nil

	preferenceList := preference.PreferenceList{
		Items: []preference.PreferenceItem{
			{
				Name:    preference.UpdateNotificationSetting,
				Value:   boolNilValue,
				Default: false,
			},
			{
				Name:    preference.PushTimeoutSetting,
				Value:   &intValue,
				Default: preference.DefaultPushTimeout,
			},
			{
				Name:    preference.RegistryCacheTimeSetting,
				Value:   intNilValue,
				Default: preference.DefaultRegistryCacheTime,
			},
			{
				Name:    preference.ConsentTelemetrySetting,
				Value:   &boolValue,
				Default: preference.DefaultConsentTelemetrySetting,
			},
			{
				Name:    preference.TimeoutSetting,
				Value:   intNilValue,
				Default: preference.DefaultTimeout,
			},
			{
				Name:    preference.EphemeralSetting,
				Value:   &boolValue,
				Default: preference.DefaultEphemeralSetting,
			},
		},
	}
	registryList := []preference.Registry{
		{
			Name:   preference.DefaultDevfileRegistryName,
			URL:    preference.DefaultDevfileRegistryURL,
			Secure: false,
		},
		{
			Name:   "StagingRegistry",
			URL:    "https://registry.staging.devfile.io",
			Secure: true,
		},
	}
	prefClient.EXPECT().NewPreferenceList().Return(preferenceList)
	prefClient.EXPECT().RegistryList().Return(&registryList)

	err = opts.Run(context.Background())
	if err != nil {
		t.Errorf(`Expected nil error, got %s`, err)
	}
}
