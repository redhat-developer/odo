package v1alpha2

// +k8s:deepcopy-gen=false
type Overrides interface {
	TopLevelListContainer
	isOverride()
}

// OverridesBase is used in the Overrides generator in order to provide a common base for the generated Overrides
// So please be careful when renaming
type OverridesBase struct{}
