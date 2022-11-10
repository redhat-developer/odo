package feature

// OdoFeature represents a free-form, but uniquely identifiable feature of odo.
// It can either be a CLI command or flag.
//
// To mark a given feature as experimental for example, it should be explicitly listed in _experimentalFeatures.
type OdoFeature struct {
	id          string
	description string
}

var (
	// GenericRunOnFlag is the feature supporting the `--run-on` generic CLI flag.
	GenericRunOnFlag = OdoFeature{
		id:          "generic-run-on",
		description: "flag: --run-on",
	}
)
