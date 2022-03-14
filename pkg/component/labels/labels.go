package labels

import (
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/util"
)

// ComponentKubernetesInstanceLabel is a label key used to identify the component name
const ComponentKubernetesInstanceLabel = "app.kubernetes.io/instance"

// ComponentKubernetesNameLabel is Kubernetes label that identifies the type of a component being used
const ComponentKubernetesNameLabel = "app.kubernetes.io/name"

// ComponentUnknownLabel is the label that is used to display something we do not know
const ComponentUnknownLabel = "<unknown>"

// KubernetesManagedByLabel ...
const KubernetesManagedByLabel = "app.kubernetes.io/managed-by"

// ComponentModeLabel ...
const ComponentModeLabel = "odo.dev/mode"

// ComponentDevName ...
const ComponentDevName = "Dev"

// ComponentDeployName ...
const ComponentDeployName = "Deploy"

// ComponentNoneName ...
const ComponentNoneName = "None"

// ComponentPushedName ...
const ComponentPushedName = "Pushed"

// ComponentProjectTypeAnnotation ...
const ComponentProjectTypeAnnotation = "odo.dev/project-type"

// GetLabels return labels that should be applied to every object for given component in active application
// additional labels are used only for creating object
// if you are creating something use additional=true
// if you need labels to filter component then use additional=false
func GetLabels(componentName string, applicationName string, additional bool) map[string]string {
	labels := applabels.GetLabels(applicationName, additional)
	labels[ComponentKubernetesInstanceLabel] = componentName
	return labels
}

// GetSelector are supposed to be used for selection of resources which are a part of the given component
func GetSelector(componentName string, applicationName string) string {
	return util.ConvertLabelsToSelector(GetLabels(componentName, applicationName, false))
}
