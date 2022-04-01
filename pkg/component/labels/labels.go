package labels

import (
	"github.com/redhat-developer/odo/pkg/version"
	"k8s.io/apimachinery/pkg/labels"
)

// KubernetesInstanceLabel is a label key used to identify the component name
const KubernetesInstanceLabel = "app.kubernetes.io/instance"

// KubernetesManagedByLabel ...
const KubernetesManagedByLabel = "app.kubernetes.io/managed-by"

// ManagerVersion is a Kubernetes label that adds what version of odo is being ran.
const KubernetesManagedByVersionLabel = "app.kubernetes.io/managed-by-version"

// KubernetesPartOfLabel is label key that is used to group all object that belong to one application
// It should be save to use just this label to filter application
const KubernetesPartOfLabel = "app.kubernetes.io/part-of"

// KubernetesStorageNameLabel is the label key that is applied to all storage resources
// that are created
const KubernetesStorageNameLabel = "app.kubernetes.io/storage-name"

// ComponentDevMode ...
const ComponentDevMode = "Dev"

// ComponentDeployMode ...
const ComponentDeployMode = "Deploy"

// ComponentDeployName ...
const ComponentAnyMode = ""

// OdoModeLabel ...
const OdoModeLabel = "odo.dev/mode"

// OdoProjectTypeAnnotation ...
const OdoProjectTypeAnnotation = "odo.dev/project-type"

// App is the default name used when labeling
const App = "app"

// DevfileStorageLabel is the label key that is applied to all storage resources for devfile components
// that are created
const DevfileStorageLabel = "storage-name"

const SourcePVCLabel = "odo-source-pvc"

// GetLabels return labels that should be applied to every object for given component in active application
// if you need labels to filter component then use GetSelector instead
func GetLabels(componentName string, applicationName string, mode string) map[string]string {
	labels := getLabels(componentName, applicationName, mode, true)
	return labels
}

// GetSelector is used for selection of resources which are a part of the given component
func GetSelector(componentName string, applicationName string, mode string) string {
	labels := getLabels(componentName, applicationName, mode, false)
	return labels.String()
}

// GetLabels return labels that should be applied to every object for given component in active application
// additional labels are used only for creating object
// if you are creating something use additional=true
// if you need labels to filter component then use additional=false
func getLabels(componentName string, applicationName string, mode string, additional bool) labels.Set {
	labels := getApplicationLabels(applicationName, additional)
	labels[KubernetesInstanceLabel] = componentName
	if mode != ComponentAnyMode {
		labels[OdoModeLabel] = mode
	}
	return labels
}

// GetLabels return labels that identifies given application
// additional labels are used only when creating object
// if you are creating something use additional=true
// if you need labels to filter component then use additional=false
func getApplicationLabels(application string, additional bool) labels.Set {
	labels := labels.Set{
		KubernetesPartOfLabel:    application,
		KubernetesManagedByLabel: "odo",
	}

	if additional {
		labels[App] = application
		labels[KubernetesManagedByVersionLabel] = version.VERSION
	}

	return labels
}
