package labels

import (
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/util"
)

// KubernetesInstanceLabel is a label key used to identify the component name
const KubernetesInstanceLabel = "app.kubernetes.io/instance"

// KubernetesNameLabel is Kubernetes label that identifies the type of a component being used
const KubernetesNameLabel = "app.kubernetes.io/name"

// KubernetesManagedByLabel ...
const KubernetesManagedByLabel = "app.kubernetes.io/managed-by"

// ComponentDevName ...
const ComponentDevName = "Dev"

// ComponentDeployName ...
const ComponentDeployName = "Deploy"

// ComponentNoneName ...
const ComponentNoneName = "None"

// OdoModeLabel ...
const OdoModeLabel = "odo.dev/mode"

// OdoProjectTypeAnnotation ...
const OdoProjectTypeAnnotation = "odo.dev/project-type"

// GetLabels return labels that should be applied to every object for given component in active application
// additional labels are used only for creating object
// if you are creating something use additional=true
// if you need labels to filter component then use additional=false
func GetLabels(componentName string, applicationName string, additional bool) map[string]string {
	labels := applabels.GetLabels(applicationName, additional)
	labels[KubernetesInstanceLabel] = componentName
	return labels
}

// GetSelector are supposed to be used for selection of resources which are a part of the given component
func GetSelector(componentName string, applicationName string) string {
	return util.ConvertLabelsToSelector(GetLabels(componentName, applicationName, false))
}
