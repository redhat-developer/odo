package v1alpha2

// +k8s:deepcopy-gen=false

// Union is an interface that allows managing structs defined as
// Kubernetes unions with discriminators, according to the following KEP:
// https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/20190325-unions.md
type Union interface {
	discriminator() *string

	// Normalize allows normalizing the union, according to the following rules:
	// - When only one field of the union is set and no discriminator is set, set the discriminator according to the union value.
	// - When several fields are set and a discrimnator is set, remove (== reset to zero value) all the values that do not match the discriminator.
	// - When only one union value is set and it matches discriminator, just do nothing.
	// - In other case, something is inconsistent or ambiguous: an error is thrown.
	Normalize() error

	// Simplify allows removing the union discriminator,
	// but only after normalizing it if necessary.
	Simplify()
}
