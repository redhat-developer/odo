package url

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// URL is
type Url struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UrlSpec `json:"spec,omitempty"`
}

// UrlSpec is
type UrlSpec struct {
	URL      string `json:"path,omitempty"`
	Protocol string
	Port     int
}

// AppList is a list of applications
type UrlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Url `json:"items"`
}
