package validation

// Validatable represents a common ancestor for validatable parameters
type Validatable struct {
	Required             bool
	Type                 string
	AdditionalValidators []Validator
}

// AsValidatable allows avoiding type casts in client code
func (v Validatable) AsValidatable() Validatable {
	return v
}
