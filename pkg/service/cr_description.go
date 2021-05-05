package service

import (
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JSONCRDDescriptionRepr struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	DescriptionSpec   *olm.CRDDescription `json:"spec" yaml:"spec"`
}

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

func ConvertCRDToJSONRepr(crd *olm.CRDDescription) JSONCRDDescriptionRepr {
	return JSONCRDDescriptionRepr{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CRDDescription",
			APIVersion: "odo.dev/v1alpha1",
		},
		DescriptionSpec: crd,
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
