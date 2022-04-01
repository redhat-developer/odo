package labels

import (
	"errors"

	"github.com/redhat-developer/odo/pkg/version"
	"k8s.io/apimachinery/pkg/labels"
)

// KubernetesInstanceLabel identifies the component name
const KubernetesInstanceLabel = "app.kubernetes.io/instance"

// KubernetesManagedByLabel identifies the manager of the component
const KubernetesManagedByLabel = "app.kubernetes.io/managed-by"

// KubernetesManagedByVersionLabel identifies the version of manager used to deploy the resource
const KubernetesManagedByVersionLabel = "app.kubernetes.io/managed-by-version"

// KubernetesPartOfLabel groups all object that belong to one application
const KubernetesPartOfLabel = "app.kubernetes.io/part-of"

// KubernetesStorageNameLabel is the label key that is applied to all storage resources
// that are created
const KubernetesStorageNameLabel = "app.kubernetes.io/storage-name"

// ComponentDevMode indicates the resource is deployed using dev command
const ComponentDevMode = "Dev"

// ComponentDeployMode indicates the resource is deployed using deploy command
const ComponentDeployMode = "Deploy"

//  ComponentAnyMode is used to search resources deployed using either dev or deploy comamnd
const ComponentAnyMode = ""

// OdoModeLabel ...
const OdoModeLabel = "odo.dev/mode"

// OdoProjectTypeAnnotation ...
const OdoProjectTypeAnnotation = "odo.dev/project-type"

// App is the default name used when labeling
const App = "app"

const odoManager = "odo"

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
func AddStorageInfo(labels map[string]string, storageName string, isSourceVolume bool) {
	labels[KubernetesStorageNameLabel] = storageName
	labels["component"] = labels[KubernetesInstanceLabel]
	labels[DevfileStorageLabel] = storageName
	if isSourceVolume {
		labels[SourcePVCLabel] = storageName
	}
}

func GetStorageName(labels map[string]string) string {
	return labels[KubernetesStorageNameLabel]
}

func GetDevfileStorageName(labels map[string]string) string {
	return labels[DevfileStorageLabel]
}

func GetComponentName(labels map[string]string) string {
	return labels[KubernetesInstanceLabel]
}

func GetAppName(labels map[string]string) string {
	return labels[KubernetesPartOfLabel]
}

func GetManagedBy(labels map[string]string) string {
	return labels[KubernetesManagedByLabel]
}

func IsManagedByOdo(labels map[string]string) bool {
	return labels[KubernetesManagedByLabel] == odoManager
}

func GetMode(labels map[string]string) string {
	return labels[OdoModeLabel]
}

func GetProjectType(labels map[string]string, annotations map[string]string) (string, error) {
	// For backwards compatibility with previously deployed components that could be non-odo, check the annotation first
	// then check to see if there is a label with the project type
	if typ, ok := annotations[OdoProjectTypeAnnotation]; ok {
		return typ, nil
	}
	if typ, ok := labels[OdoProjectTypeAnnotation]; ok {
		return typ, nil
	}
	return "", errors.New("component type not found in labels or annotations")
}

func SetProjectType(annotations map[string]string, value string) {
	annotations[OdoProjectTypeAnnotation] = value
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
		KubernetesManagedByLabel: odoManager,
	}
	if additional {
		labels[App] = application
		labels[KubernetesManagedByVersionLabel] = version.VERSION
	}
	return labels
}
