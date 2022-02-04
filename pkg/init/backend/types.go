package backend

type DevfileLocation struct {
	// Name of the component to create (required)
	//	Name string
	// name of the Devfile in Devfile registry (required if --Devfile-path is not defined)
	Devfile string
	// name of the devfile registry (as configured in odo registry). It can be used in combination with --devfile, but not with --devfile-path (optional)
	DevfileRegistry string
	// name of the Starter project (optional)
	//	Starter string
	// path to a devfile. This is alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL (required if --devfile is not defined)
	DevfilePath string
}
