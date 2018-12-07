package application

import (
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Application
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AppSeco `json:"spec,omitempty"`
}

type AppSeco struct {
	Components []string
}

// AppSpec holds all information about application
type AppSpec struct {
	Name       string                    `json:"applicationName,omitempty"`
	Components []component.ComponentSpec `json:"components,omitempty"`
}

// AppList is a list of applications
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []config.ApplicationInfo `json:"items"`
}
