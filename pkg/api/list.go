package api

// ResourcesList is the result of the `odo list` command
type ResourcesList struct {
	// ComponentInDevfile is the component name present in the local Devfile when `odo list` is executed, or empty
	ComponentInDevfile string `json:"componentInDevfile"`
	// Components is a list of components deployed in the cluster or present in the local Devfile
	Components []ComponentAbstract `json:"components"`
}
