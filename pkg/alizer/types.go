package alizer

type DevfileLocation struct {
	// name of the Devfile in Devfile registry (required if DevfilePath is not defined)
	Devfile string `json:"devfile,omitempty"`

	// name of the devfile registry (as configured in odo registry). It can be used in combination with Devfile, but not with DevfilePath (optional)
	DevfileRegistry string `json:"devfile-registry,omitempty"`

	// path to a devfile. This is alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL (required if Devfile is not defined)
	DevfilePath string `json:"devfile-path,omitempty"`
}
