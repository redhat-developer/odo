package feature

// OdoFeature represents a free-form, but uniquely identifiable feature of odo.
// It can either be a CLI command or flag.
//
// To mark a given feature as experimental for example, it should be explicitly listed in _experimentalFeatures.
type OdoFeature struct {
	id          string
	description string
}
