package v1alpha1

// ParamSpec defines an arbitrary named  input whose value can be supplied by a
// `Param`.
type ParamSpec struct {
	// Name declares the name by which a parameter is referenced.
	Name string `json:"name"`
	// Description is a user-facing description of the parameter that may be
	// used to populate a UI.
	// +optional
	Description string `json:"description,omitempty"`
	// Default is the value a parameter takes if no input value via a Param is supplied.
	// +optional
	Default *string `json:"default,omitempty"`
}

// Param defines a string value to be used for a ParamSpec with the same name.
type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
