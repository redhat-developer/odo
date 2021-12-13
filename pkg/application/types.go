package application

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appAPIVersion = "odo.dev/v1alpha1"
	appKind       = "Application"
	appList       = "List"
)

// Application
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AppSpec `json:"spec,omitempty"`
}

// AppSpec is list of components present in given application
type AppSpec struct {
	Components []string `json:"components,omitempty"`
}

// AppList is a list of applications
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}
