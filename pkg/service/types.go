package service

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/odo/pkg/odo/util/validation"
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

// ServicePlanParameter holds the information regarding a service catalog plan parameter
type ServicePlanParameter struct {
	Name                   string `json:"name"`
	Title                  string `json:"title,omitempty"`
	Description            string `json:"description,omitempty"`
	Default                string `json:"default,omitempty"`
	validation.Validatable `json:",inline,omitempty"`
}

type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

// ServicePlan holds the information about service catalog plans associated to service classes
type ServicePlan struct {
	Name        string
	DisplayName string
	Description string
	Parameters  servicePlanParameters
}

// UnmarshalJSON unmarshals the JSON for ServicePlanParameter instead of using
// the built in json.Unmarshal
func (sp *ServicePlanParameter) UnmarshalJSON(data []byte) error {
	// create a temporary struct similar to ServicePlanParameter but with
	// Default's type set to interface{} so that we can store any value in it
	tempServicePlanParameter := struct {
		Name                   string      `json:"name"`
		Title                  string      `json:"title,omitempty"`
		Description            string      `json:"description,omitempty"`
		Default                interface{} `json:"default,omitempty"`
		validation.Validatable `json:",inline,omitempty"`
	}{}

	// unmarshal the json obtained from server into the temporary struct
	err := json.Unmarshal(data, &tempServicePlanParameter)
	if err != nil {
		return err
	}

	// convert the value into a string so that it can be stored in ServicePlanParameter
	if tempServicePlanParameter.Default != nil {
		sp.Default = fmt.Sprint(tempServicePlanParameter.Default)
	}

	sp.Name = tempServicePlanParameter.Name
	sp.Title = tempServicePlanParameter.Title
	sp.Description = tempServicePlanParameter.Description
	sp.Validatable = tempServicePlanParameter.Validatable

	return nil
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
