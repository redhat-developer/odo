package binding

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pkg/errors"
)

type UnstructuredResourceReader func(namespace string, name string) (*unstructured.Unstructured, error)

type secretConfigMapReader struct {
	configMapReader UnstructuredResourceReader
	secretReader    UnstructuredResourceReader
}

type annotationBackedDefinitionBuilder struct {
	*secretConfigMapReader
	name  string
	value string
}

var _ DefinitionBuilder = (*annotationBackedDefinitionBuilder)(nil)

type modelKey string

const (
	pathModelKey                    modelKey = "path"
	objectTypeModelKey              modelKey = "objectType"
	sourceKeyModelKey               modelKey = "sourceKey"
	sourceValueModelKey             modelKey = "sourceValue"
	elementTypeModelKey             modelKey = "elementType"
	AnnotationPrefix                         = "service.binding"
	ProvisionedServiceAnnotationKey          = "servicebinding.io/provisioned-service"
	TypeKey                                  = AnnotationPrefix + "/type"
)

func NewDefinitionBuilder(annotationName string, annotationValue string, configMapReader UnstructuredResourceReader, secretReader UnstructuredResourceReader) *annotationBackedDefinitionBuilder {
	return &annotationBackedDefinitionBuilder{
		name:  annotationName,
		value: annotationValue,
		secretConfigMapReader: &secretConfigMapReader{
			configMapReader: configMapReader,
			secretReader:    secretReader,
		},
	}
}

func IsServiceBindingAnnotation(annotationKey string) (bool, error) {
	if annotationKey == ProvisionedServiceAnnotationKey {
		return false, nil
	}

	if annotationKey == AnnotationPrefix {
		return true, nil
	} else if strings.HasPrefix(annotationKey, AnnotationPrefix) {
		if strings.HasPrefix(annotationKey, AnnotationPrefix+"/") {
			return true, nil
		}

		// it starts with AnnotationPrefix, but has extra text at the end not
		// separated by a /, so treat it as an error
		return false, fmt.Errorf("can't process annotation with name %q", annotationKey)
	}

	// bail out when the annotation name doesn't start with "service.binding"
	return false, nil
}

func (m *annotationBackedDefinitionBuilder) isServiceBindingAnnotation() (bool, error) {
	return IsServiceBindingAnnotation(m.name)
}

func (m *annotationBackedDefinitionBuilder) outputName() string {

	if p := strings.SplitN(m.name, "/", 2); len(p) == 2 && len(p[1]) > 0 {
		return p[1]
	}

	return ""
}

func (m *annotationBackedDefinitionBuilder) Build() (Definition, error) {

	if valid, err := m.isServiceBindingAnnotation(); !valid || err != nil {
		return nil, err
	}

	outputName := m.outputName()

	mod, err := newModel(m.value)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create binding model for annotation key %s and value %s", m.name, m.value)
	}

	switch {
	case (mod.isStringElementType() && mod.isStringObjectType()) || mod.value != "":
		return &stringDefinition{
			outputName: outputName,
			value:      mod.value,
			definition: definition{
				path: mod.path,
			},
		}, nil

	case mod.isStringElementType() && mod.hasDataField():
		return &stringFromDataFieldDefinition{
			secretConfigMapReader: m.secretConfigMapReader,
			objectType:            mod.objectType,
			outputName:            outputName,
			definition: definition{
				path: mod.path,
			},
			sourceKey: mod.sourceKey,
		}, nil

	case mod.isMapElementType() && mod.hasDataField():
		return &mapFromDataFieldDefinition{
			secretConfigMapReader: m.secretConfigMapReader,
			objectType:            mod.objectType,
			outputName:            outputName,
			definition: definition{
				path: mod.path,
			},
			sourceValue: mod.sourceValue,
		}, nil

	case mod.isMapElementType() && mod.isStringObjectType():
		return &stringOfMapDefinition{
			outputName: outputName,
			definition: definition{
				path: mod.path,
			},
		}, nil

	case mod.isSliceOfMapsElementType():
		return &sliceOfMapsFromPathDefinition{
			outputName: outputName,
			definition: definition{
				path: mod.path,
			},
			sourceKey:   mod.sourceKey,
			sourceValue: mod.sourceValue,
		}, nil

	case mod.isSliceOfStringsElementType():
		return &sliceOfStringsFromPathDefinition{
			outputName: outputName,
			definition: definition{
				path: mod.path,
			},
			sourceValue: mod.sourceValue,
		}, nil
	}
	return nil, fmt.Errorf("Annotation %s: %s not implemented!", m.name, m.value)
}
