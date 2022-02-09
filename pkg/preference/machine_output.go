package preference

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	prefAPIVersion = "odo.dev/v1alpha1"
	prefKind       = "PreferenceList"
)

type PreferenceList struct {
	metav1.TypeMeta `json:",inline"`
	Items           []PreferenceItem `json:"items,omitempty"`
}

type PreferenceItem struct {
	Name        string
	Value       interface{} // The value set by the user, this will be nil if the user hasn't set it
	Default     interface{} // default value of the preference if the user hasn't set the value
	Type        string      // the type of the preference, possible values int, string, bool
	Description string      // The description of the preference
}

func (o *preferenceInfo) NewPreferenceList() PreferenceList {
	return PreferenceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: prefAPIVersion,
			Kind:       prefKind,
		},
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
