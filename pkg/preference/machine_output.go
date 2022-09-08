package preference

import (
	"reflect"
)

type PreferenceList struct {
	Items []PreferenceItem `json:"items,omitempty"`
}

type PreferenceItem struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`       // The value set by the user, this will be nil if the user hasn't set it
	Default     interface{} `json:"default"`     // default value of the preference if the user hasn't set the value
	Type        string      `json:"type"`        // the type of the preference, possible values int, string, bool
	Description string      `json:"description"` // The description of the preference
}

func (o *preferenceInfo) NewPreferenceList() PreferenceList {
	return PreferenceList{
		Items: toPreferenceItems(*o),
	}
}

func toPreferenceItems(prefInfo preferenceInfo) []PreferenceItem {
	settings := prefInfo.OdoSettings
	return []PreferenceItem{
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
	}
}

func getType(v interface{}) string {

	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr {
		return rv.Elem().Kind().String()
	}

	return rv.Kind().String()
}
