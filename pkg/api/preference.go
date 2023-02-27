package api

import "github.com/redhat-developer/odo/pkg/preference"

type PreferenceView struct {
	Preferences []preference.PreferenceItem `json:"preferences,omitempty"`
	Registries  []Registry                  `json:"registries,omitempty"`
}
