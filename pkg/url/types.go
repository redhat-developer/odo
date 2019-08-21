package url

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// URL is
type Url struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UrlSpec   `json:"spec,omitempty"`
	Status            UrlStatus `json:"status,omitempty"`
}

// UrlSpec is
type UrlSpec struct {
	Host     string `json:"host,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Port     int    `json:"port,omitempty"`
}

// AppList is a list of applications
type UrlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Url `json:"items"`
}

// UrlStatus is Status of url
type UrlStatus struct {
	// "Pushed" or "Not Pushed" or "Locally Delted"
	State StateType `json:"state"`
}

type StateType string

const (
	// StateTypePushed means that Url is present both locally and on cluster
	StateTypePushed = "Pushed"
	// StateTypeNotPushed means that Url is only in local config, but not on the cluster
	StateTypeNotPushed = "Not Pushed"
	// StateTypeLocallyDeleted means that Url was deleted from the local config, but it is still present on the cluster
	StateTypeLocallyDeleted = "Locally Deleted"
)
