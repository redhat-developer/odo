package labels

import (
	"errors"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/redhat-developer/odo/pkg/version"
)

// GetLabels return labels that should be applied to every object for given component in active application
// if you need labels to filter component then use GetSelector instead
func GetLabels(componentName string, applicationName string, mode string) map[string]string {
	labels := getLabels(componentName, applicationName, mode, true)
	return labels
}

// AddStorageInfo adds labels for storage resources
func AddStorageInfo(labels map[string]string, storageName string, isSourceVolume bool) {
	labels[kubernetesStorageNameLabel] = storageName
	labels[componentLabel] = labels[kubernetesInstanceLabel]
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

// GetSelector returns a selector string used for selection of resources which are part of the given component in given mode
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
		labels[appLabel] = application
		labels[kubernetesManagedByVersionLabel] = version.VERSION
	}
	return labels
}
