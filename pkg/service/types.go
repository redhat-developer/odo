package service

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/odo/pkg/odo/util/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []Service `json:"items"`
}

// ServicePlan holds the information about service catalog plans associated to service classes
type ServicePlan struct {
	Name        string
	DisplayName string
	Description string
	Parameters  servicePlanParameters
}

func (sp *ServicePlanParameter) UnmarshalJSON(data []byte) error {
	tempServicePlanParameter := struct {
		Name                   string      `json:"name"`
		Title                  string      `json:"title,omitempty"`
		Description            string      `json:"description,omitempty"`
		Default                interface{} `json:"default,omitempty"`
		validation.Validatable `json:",inline,omitempty"`
	}{}

	err := json.Unmarshal(data, &tempServicePlanParameter)
	if err != nil {
		panic(err)
	}
	sp.Default = fmt.Sprint(tempServicePlanParameter.Default)

	sp.Name = tempServicePlanParameter.Name
	sp.Title = tempServicePlanParameter.Title
	sp.Description = tempServicePlanParameter.Description
	sp.Validatable = tempServicePlanParameter.Validatable

	return nil
}
