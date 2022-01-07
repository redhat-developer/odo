package component

import (
	"encoding/json"
	"github.com/redhat-developer/odo/pkg/storage"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComponentFullDescriptionSpec represents the complete description of the component
type ComponentFullDescriptionSpec struct {
	App     string              `json:"app,omitempty"`
	Type    string              `json:"type,omitempty"`
	URL     urlpkg.URLList      `json:"urls,omitempty"`
	Storage storage.StorageList `json:"storages,omitempty"`
	Env     []corev1.EnvVar     `json:"env,omitempty"`
	Ports   []string            `json:"ports,omitempty"`
}

// ComponentFullDescription describes a component fully
type ComponentFullDescription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentFullDescriptionSpec `json:"spec,omitempty"`
	Status            ComponentStatus              `json:"status,omitempty"`
}

// copyFromComponentDescription copies over all fields from Component that can be copied
func (cfd *ComponentFullDescription) copyFromComponentDesc(component *Component) error {
	d, err := json.Marshal(component)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, cfd)
}

// fillEmptyFields fills any fields that are empty in the ComponentFullDescription
func (cfd *ComponentFullDescription) fillEmptyFields(componentDesc Component, componentName string, applicationName string, projectName string) {
	// fix missing names in case it is not in description
	if len(cfd.Name) <= 0 {
		cfd.Name = componentName
	}

	if len(cfd.Namespace) <= 0 {
		cfd.Namespace = projectName
	}

	if len(cfd.Kind) <= 0 {
		cfd.Kind = "Component"
	}

	if len(cfd.APIVersion) <= 0 {
		cfd.APIVersion = apiVersion
	}

	if len(cfd.Spec.App) <= 0 {
		cfd.Spec.App = applicationName
	}
	cfd.Spec.Ports = componentDesc.Spec.Ports
}

// GetComponent returns a component representation
func (cfd *ComponentFullDescription) GetComponent() Component {
	cmp := NewComponent(cfd.Name)
	cmp.Spec.App = cfd.Spec.App
	cmp.Spec.Ports = cfd.Spec.Ports
	cmp.Spec.Type = cfd.Spec.Type
	cmp.Spec.StorageSpec = cfd.Spec.Storage.Items
	cmp.Spec.URLSpec = cfd.Spec.URL.Items
	for _, url := range cfd.Spec.URL.Items {
		cmp.Spec.URL = append(cmp.Spec.URL, url.Name)
	}
	for _, storage := range cfd.Spec.Storage.Items {
		cmp.Spec.Storage = append(cmp.Spec.URL, storage.Name)
	}
	cmp.ObjectMeta.Namespace = cfd.ObjectMeta.Namespace
	cmp.Status = cfd.Status
	cmp.Spec.Env = cfd.Spec.Env
	return cmp
}
