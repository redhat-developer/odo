package labels

import (
	"errors"

	"github.com/redhat-developer/odo/pkg/version"
	"k8s.io/apimachinery/pkg/labels"
)

// kubernetesInstanceLabel identifies the component name
const kubernetesInstanceLabel = "app.kubernetes.io/instance"

// kubernetesManagedByLabel identifies the manager of the component
const kubernetesManagedByLabel = "app.kubernetes.io/managed-by"

// kubernetesManagedByVersionLabel identifies the version of manager used to deploy the resource
const kubernetesManagedByVersionLabel = "app.kubernetes.io/managed-by-version"

// kubernetesPartOfLabel groups all object that belong to one application
const kubernetesPartOfLabel = "app.kubernetes.io/part-of"

// kubernetesStorageNameLabel is the label key that is applied to all storage resources
// that are created
const kubernetesStorageNameLabel = "app.kubernetes.io/storage-name"

// ComponentDevMode indicates the resource is deployed using dev command
const ComponentDevMode = "Dev"

// ComponentDeployMode indicates the resource is deployed using deploy command
const ComponentDeployMode = "Deploy"

//  ComponentAnyMode is used to search resources deployed using either dev or deploy comamnd
const ComponentAnyMode = ""

// odoModeLabel ...
const odoModeLabel = "odo.dev/mode"

// odoProjectTypeAnnotation ...
const odoProjectTypeAnnotation = "odo.dev/project-type"

// app is the default name used when labeling
const app = "app"

const odoManager = "odo"

// devfileStorageLabel is the label key that is applied to all storage resources for devfile components
// that are created
const devfileStorageLabel = "storage-name"

const sourcePVCLabel = "odo-source-pvc"

// GetLabels return labels that should be applied to every object for given component in active application
// if you need labels to filter component then use GetSelector instead
func GetLabels(componentName string, applicationName string, mode string) map[string]string {
	labels := getLabels(componentName, applicationName, mode, true)
	return labels
}
func AddStorageInfo(labels map[string]string, storageName string, isSourceVolume bool) {
	labels[kubernetesStorageNameLabel] = storageName
	labels["component"] = labels[kubernetesInstanceLabel]
	labels[devfileStorageLabel] = storageName
	if isSourceVolume {
		labels[sourcePVCLabel] = storageName
	}
}

func GetStorageName(labels map[string]string) string {
	return labels[kubernetesStorageNameLabel]
}

func GetDevfileStorageName(labels map[string]string) string {
	return labels[devfileStorageLabel]
}

func GetComponentName(labels map[string]string) string {
	return labels[kubernetesInstanceLabel]
}

func GetAppName(labels map[string]string) string {
	return labels[kubernetesPartOfLabel]
}

func GetManagedBy(labels map[string]string) string {
	return labels[kubernetesManagedByLabel]
}

func IsManagedByOdo(labels map[string]string) bool {
	return labels[kubernetesManagedByLabel] == odoManager
}

func GetMode(labels map[string]string) string {
	return labels[odoModeLabel]
}

func GetProjectType(labels map[string]string, annotations map[string]string) (string, error) {
	// For backwards compatibility with previously deployed components that could be non-odo, check the annotation first
	// then check to see if there is a label with the project type
	if typ, ok := annotations[odoProjectTypeAnnotation]; ok {
		return typ, nil
	}
	if typ, ok := labels[odoProjectTypeAnnotation]; ok {
		return typ, nil
	}
	return "", errors.New("component type not found in labels or annotations")
}

func SetProjectType(annotations map[string]string, value string) {
	annotations[odoProjectTypeAnnotation] = value
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
	labels[kubernetesInstanceLabel] = componentName
	if mode != ComponentAnyMode {
		labels[odoModeLabel] = mode
	}
	return labels
}

// GetLabels return labels that identifies given application
// additional labels are used only when creating object
// if you are creating something use additional=true
// if you need labels to filter component then use additional=false
func getApplicationLabels(application string, additional bool) labels.Set {
	labels := labels.Set{
		kubernetesPartOfLabel:    application,
		kubernetesManagedByLabel: odoManager,
	}
	if additional {
		labels[app] = application
		labels[kubernetesManagedByVersionLabel] = version.VERSION
	}
	return labels
}
