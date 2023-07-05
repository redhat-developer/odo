package preference

import (
	"reflect"

	"github.com/redhat-developer/odo/pkg/api"
)

func (o *preferenceInfo) NewPreferenceList() api.PreferenceList {
	return api.PreferenceList{
		Items: toPreferenceItems(*o),
	}
}

func toPreferenceItems(prefInfo preferenceInfo) []api.PreferenceItem {
	settings := prefInfo.OdoSettings
	return []api.PreferenceItem{
		{
			Name:        UpdateNotificationSetting,
			Value:       settings.UpdateNotification,
			Default:     true,
			Type:        getType(prefInfo.GetUpdateNotification()), // use the Getter here to determine type
			Description: UpdateNotificationSettingDescription,
		},
		{
			Name:        TimeoutSetting,
			Value:       settings.Timeout,
			Default:     DefaultTimeout,
			Type:        getType(prefInfo.GetTimeout()),
			Description: TimeoutSettingDescription,
		},
		{
			Name:        PushTimeoutSetting,
			Value:       settings.PushTimeout,
			Default:     DefaultPushTimeout,
			Type:        getType(prefInfo.GetPushTimeout()),
			Description: PushTimeoutSettingDescription,
		},
		{
			Name:        RegistryCacheTimeSetting,
			Value:       settings.RegistryCacheTime,
			Default:     DefaultRegistryCacheTime,
			Type:        getType(prefInfo.GetRegistryCacheTime()),
			Description: RegistryCacheTimeSettingDescription,
		},
		{
			Name:        ConsentTelemetrySetting,
			Value:       settings.ConsentTelemetry,
			Default:     DefaultConsentTelemetrySetting,
			Type:        getType(prefInfo.GetConsentTelemetry()),
			Description: ConsentTelemetrySettingDescription,
		},
		{
			Name:        EphemeralSetting,
			Value:       settings.Ephemeral,
			Default:     DefaultEphemeralSetting,
			Type:        getType(prefInfo.GetEphemeral()),
			Description: EphemeralSettingDescription,
		},
		{
			Name:        ImageRegistrySetting,
			Value:       settings.ImageRegistry,
			Default:     DefaultDevfileRegistryURL,
			Type:        getType(prefInfo.GetImageRegistry()),
			Description: ImageRegistrySettingDescription,
		},
	}
}

func getType(v interface{}) string {

	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr {
		return rv.Elem().Kind().String()
	}

	return rv.Kind().String()
}
