package service

import (
	"github.com/go-openapi/spec"
	"github.com/openshift/odo/pkg/machineoutput"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OperatorExample struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              map[string]interface{} `json:"spec,omitempty"`
}

// ServiceInfo holds all important information about one service
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceSpec   `json:"spec,omitempty"`
	Status            ServiceStatus `json:"status,omitempty"`
}

// ServiceSpec ...
type ServiceSpec struct {
	Type string `json:"type,omitempty"`
	Plan string `json:"plan,omitempty"`
}

// ServiceStatus ...
type ServiceStatus struct {
	Status string `json:"status,omitempty"`
}

<<<<<<< HEAD
=======
// ServiceClass holds the information regarding a service catalog service class
type ServiceClass struct {
	Name              string
	Bindable          bool
	ShortDescription  string
	LongDescription   string
	Tags              []string
	VersionsAvailable []string
	ServiceBrokerName string
}

>>>>>>> 20cd8e28c (Removing remaining refs of service catalog code)
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func NewOperatorExample(almExample map[string]interface{}) OperatorExample {
	return OperatorExample{
<<<<<<< HEAD
		TypeMeta: metav1.TypeMeta{
			Kind:       "OperatorExample",
			APIVersion: "odo.dev/v1alpha1",
		},
		Spec: almExample,
	}
}

const OperatorBackedServiceKind = "Service"

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
=======
>>>>>>> 20cd8e28c (Removing remaining refs of service catalog code)
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
