package v1alpha2

type PluginComponent struct {
	BaseComponent   `json:",inline"`
	ImportReference `json:",inline"`
	PluginOverrides `json:",inline"`
}
