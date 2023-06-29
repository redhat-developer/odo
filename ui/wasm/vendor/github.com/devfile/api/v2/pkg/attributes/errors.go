package attributes

import "fmt"

// KeyNotFoundError returns an error if no key is found for the attribute
type KeyNotFoundError struct {
	Key string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("Attribute with key %q does not exist", e.Key)
}
