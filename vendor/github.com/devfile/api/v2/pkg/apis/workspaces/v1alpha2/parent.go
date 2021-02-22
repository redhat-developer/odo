package v1alpha2

type Parent struct {
	ImportReference `json:",inline"`
	ParentOverrides `json:",inline"`
}
