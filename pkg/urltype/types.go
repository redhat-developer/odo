package urltype

import (
	"github.com/openshift/odo/pkg/localConfigProvider"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// URL is
type URL struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Spec          URLSpec   `json:"spec,omitempty"`
	Status        URLStatus `json:"status,omitempty"`
}

// URLSpec is
type URLSpec struct {
	Host         string                      `json:"host,omitempty"`
	Protocol     string                      `json:"protocol,omitempty"`
	Port         int                         `json:"port,omitempty"`
	Secure       bool                        `json:"secure"`
	Kind         localConfigProvider.URLKind `json:"kind,omitempty"`
	TLSSecret    string                      `json:"tlssecret,omitempty"`
	ExternalPort int                         `json:"externalport,omitempty"`
	Path         string                      `json:"path,omitempty"`
}

// URLList is a list of applications
type URLList struct {
	v1.TypeMeta `json:",inline"`
	v1.ListMeta `json:"metadata,omitempty"`
	Items       []URL `json:"items"`
}

// URLStatus is Status of url
type URLStatus struct {
	// "Pushed" or "Not Pushed" or "Locally Delted"
	State StateType `json:"state"`
}

type StateType string

const (
	// StateTypePushed means that URL is present both locally and on cluster/container
	StateTypePushed = "Pushed"
	// StateTypeNotPushed means that URL is only in local config, but not on the cluster/container
	StateTypeNotPushed = "Not Pushed"
	// StateTypeLocallyDeleted means that URL was deleted from the local config, but it is still present on the cluster/container
	StateTypeLocallyDeleted = "Locally Deleted"
)

// Get returns URL definition for given URL name
func (urls URLList) Get(urlName string) URL {
	for _, url := range urls.Items {
		if url.Name == urlName {
			return url
		}
	}
	return URL{}

}
