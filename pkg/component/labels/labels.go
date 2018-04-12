package labels

import (
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
)

// ComponentLabel is a label key used to identify component
const ComponentLabel = "app.kubernetes.io/component-name"

// ComponentTypeLabel is kubernetes that identifies type of a component
const ComponentTypeLabel = "app.kubernetes.io/component-type"

// GetLabels return labels that should be applied to every object for given component in active application
// additional labels are used only for creating object
// if you are creating something use additional=true
// if you need labels to filter component that use additional=false
func GetLabels(componentName string, applicationName string, additional bool) map[string]string {
	labels := applabels.GetLabels(applicationName, additional)
	labels[ComponentLabel] = componentName
	return labels
}
