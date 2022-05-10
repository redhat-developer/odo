package api

// ResourcesList is the result of the `odo list` command
type ResourcesList struct {
	// ComponentsInDevfile is the list of components names present in the local Devfile when `odo list` is executed
	ComponentsInDevfile []string `json:"componentsInDevfile"`
	// Components is a list of components deployed in the cluster or present in the local Devfile
	Components []ComponentAbstract `json:"components"`
}
