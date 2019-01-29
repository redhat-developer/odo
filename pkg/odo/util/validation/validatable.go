package validation

// Validatable represents a common ancestor for validatable parameters
type Validatable struct {
	// Required indicates whether this Validatable is a required value in the context it's supposed to be used
	Required bool `json:"required,omitempty"`
	// Type specifies the type of values this Validatable accepts so that some validation can be performed based on it
	Type string `json:"type"`
	// AdditionalValidators allows users to specify validators (in addition to default ones) to validate this Validatable's value
	AdditionalValidators []Validator `json:"-"`
}

// AsValidatable allows avoiding type casts in client code
func (v Validatable) AsValidatable() Validatable {
	return v
}
