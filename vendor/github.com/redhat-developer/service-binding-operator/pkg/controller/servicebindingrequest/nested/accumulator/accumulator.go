package accumulator

import (
	"errors"

	"github.com/imdario/mergo"
)

const valuesKey = "values"

// Accumulator is a value accumulator.
type Accumulator map[string]interface{}

// UnsupportedTypeErr is returned when an unsupported type is encountered.
var UnsupportedTypeErr = errors.New("unsupported type")

// Accumulate accumulates the `val` value. An error is returned in the case
// `val` contains an unsupported type.
func (a Accumulator) Accumulate(val interface{}) error {
	b := NewAccumulator()
	switch v := val.(type) {
	case map[string]interface{}:
		b[valuesKey] = []map[string]interface{}{v}
	case string:
		b[valuesKey] = []string{v}
	case int:
		b[valuesKey] = []int{v}
	case []map[string]interface{}, []string, []int:
		b[valuesKey] = v
	default:
		return UnsupportedTypeErr
	}
	return mergo.Merge(&a, b, mergo.WithAppendSlice, mergo.WithOverride, mergo.WithTypeCheck)
}

// Value returns the accumulated values.
func (a Accumulator) Value() interface{} {
	return a[valuesKey]
}

// NewAccumulator returns a new value accumulator
func NewAccumulator() Accumulator {
	return Accumulator{}
}
