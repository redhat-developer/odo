package service

import (
	"github.com/go-openapi/spec"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OperatorExample struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              map[string]interface{} `json:"spec,omitempty"`
}

func NewOperatorExample(almExample map[string]interface{}) OperatorExample {
	return OperatorExample{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OperatorExample",
			APIVersion: "odo.dev/v1alpha1",
		},
		Spec: almExample,
	}
}

const OperatorBackedServiceKind = "Service"
const OperatorBackedServiceCRListKind = "ServiceList"

type OperatorBackedServiceSpec struct {
	Kind        string       `json:"kind"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	DisplayName string       `json:"displayName"`
	Schema      *spec.Schema `json:"schema"`
}

type OperatorBackedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              OperatorBackedServiceSpec `json:"spec"`
}

func NewOperatorBackedService(name string, kind string, version string, description string, displayName string, spec *spec.Schema) OperatorBackedService {
	return OperatorBackedService{
		TypeMeta: metav1.TypeMeta{
			Kind:       OperatorBackedServiceKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: OperatorBackedServiceSpec{
			Kind:        kind,
			Version:     version,
			Description: description,
			DisplayName: displayName,
			Schema:      spec,
		},
	}
}

type OperatorServiceCRItem struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

type OperatorServiceCRListSpec struct {
	DisplayName string                  `json:"displayName"`
	Description string                  `json:"description"`
	CRDS        []OperatorServiceCRItem `json:"crds"`
}

type OperatorBackedServiceCRList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              OperatorServiceCRListSpec `json:"spec"`
}

func NewOperatorBackedCRList(name string, displayName string, description string) *OperatorBackedServiceCRList {
	return &OperatorBackedServiceCRList{
		TypeMeta: metav1.TypeMeta{
			Kind:       OperatorBackedServiceCRListKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: OperatorServiceCRListSpec{
			DisplayName: displayName,
			Description: description,
		},
	}
}
