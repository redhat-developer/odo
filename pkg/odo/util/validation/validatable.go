package validation

// Validatable represents a common ancestor for validatable parameters
type Validatable struct {
	Required             bool        `json:"required,omitempty"`
	Type                 string      `json:"type"`
	AdditionalValidators []Validator `json:"-"`
}

// AsValidatable allows avoiding type casts in client code
func (v Validatable) AsValidatable() Validatable {
	return v
}
