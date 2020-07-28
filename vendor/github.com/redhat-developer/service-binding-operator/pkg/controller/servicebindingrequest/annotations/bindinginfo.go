package annotations

import (
	"errors"
	"strings"
)

const (
	// ServiceBindingOperatorAnnotationPrefix is the prefix of Service Binding Operator related annotations.
	ServiceBindingOperatorAnnotationPrefix = "servicebindingoperator.redhat.io/"
)

// BindingInfo represents the pieces of a binding as parsed from an annotation.
type BindingInfo struct {
	// ResourceReferencePath is the field in the Backing Service CR referring to a bindable property, either
	// embedded or a reference to a related object..
	ResourceReferencePath string
	// SourcePath is the field that will be collected from the Backing Service CR or a related object.
	SourcePath string
	// Descriptor is the field reference to another manifest.
	Descriptor string
	// Value is the original annotation value.
	Value string
}

var ErrInvalidAnnotationPrefix = errors.New("invalid annotation prefix")
var ErrInvalidAnnotationName = errors.New("invalid annotation name")
var ErrEmptyAnnotationName = errors.New("empty annotation name")

// NewBindingInfo parses the encoded in the annotation name, returning its parts.
func NewBindingInfo(name string, value string) (*BindingInfo, error) {
	// do not process unknown annotations
	if !strings.HasPrefix(name, ServiceBindingOperatorAnnotationPrefix) {
		return nil, ErrInvalidAnnotationPrefix
	}

	cleanName := strings.TrimPrefix(name, ServiceBindingOperatorAnnotationPrefix)
	if len(cleanName) == 0 {
		return nil, ErrEmptyAnnotationName
	}

	parts := strings.SplitN(cleanName, "-", 2)

	resourceReferencePath := parts[0]
	sourcePath := parts[0]

	// the annotation is a reference to another manifest
	if len(parts) == 2 {
		sourcePath = parts[1]
	}

	return &BindingInfo{
		ResourceReferencePath: resourceReferencePath,
		SourcePath:            sourcePath,
		Descriptor:            strings.Join([]string{value, sourcePath}, ":"),
		Value:                 value,
	}, nil
}
