package service

import olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

type CRDDescriptionRepr struct {
	Kind        string      `yaml:"Kind"`
	Version     string      `yaml:"Version"`
	Description string      `yaml:"Description"`
	Parameters  []Parameter `yaml:"Parameters"`
}

type Parameter struct {
	Path        string `yaml:"Path"`
	DisplayName string `yaml:"DisplayName,omitempty"`
	Description string `yaml:"Description,omitempty"`
}

func ConvertCRDToRepr(crd *olm.CRDDescription) CRDDescriptionRepr {
	return CRDDescriptionRepr{
		Kind:        crd.Kind,
		Description: crd.Description,
		Version:     crd.Version,
		Parameters:  convertSpecDescriptorsToParameters(crd.SpecDescriptors),
	}
}

func convertSpecDescriptorsToParameters(specDescriptors []olm.SpecDescriptor) []Parameter {
	params := []Parameter{}
	for _, desc := range specDescriptors {
		params = append(params, Parameter{
			Path:        desc.Path,
			DisplayName: desc.DisplayName,
			Description: desc.Description,
		})
	}
	return params
}
