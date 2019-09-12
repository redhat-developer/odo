package catalog

import (
	imagev1 "github.com/openshift/api/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Catalog ...
type Catalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CatalogSpec `json:"spec,omitempty"`
}

// CatalogSpec ...
type CatalogSpec struct {
	Namespace      string              `json:"namespace"`
	AllTags        []string            `json:"allTags"`
	NonHiddenTags  []string            `json:"nonHiddenTags"`
	ImageStreamRef imagev1.ImageStream `json:"imageStreamRef"`
}

// CatalogImageList ...
type CatalogImageList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []Catalog `json:"items"`
}
