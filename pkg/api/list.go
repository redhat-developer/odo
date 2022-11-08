package api

// ResourcesList is the result of the `odo list` command
type ResourcesList struct {
	// ComponentInDevfile is the component name present in the local Devfile when `odo list` is executed, or empty
	ComponentInDevfile string `json:"componentInDevfile,omitempty"`
	// Components is a list of components deployed in the cluster or present in the local Devfile
	Components []ComponentAbstract `json:"components,omitempty"`

	// BindingsInDevfile is the list of binding names present in the local devfile
	BindingsInDevfile []string `json:"bindingsInDevfile,omitempty"`
	// Bindings is a list of bindings in the local devfile and/or cluster
	Bindings []ServiceBinding `json:"bindings,omitempty"`

	// BindableServices is the list of bindable services that could be bound to the component
	BindableServices []BindableService `json:"bindableServices,omitempty"`

	// Namespaces is the list of namespces available for the user on the cluster
	Namespaces []Project `json:"namespaces,omitempty"`
}
