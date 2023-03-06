package api

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

type PreferenceView struct {
	Preferences []PreferenceItem `json:"preferences,omitempty"`
	Registries  []Registry       `json:"registries,omitempty"`
}
