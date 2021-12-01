package labels

import (
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
)

// StorageLabel is the label key that is applied to all storage resources
// that are created
const StorageLabel = "app.kubernetes.io/storage-name"

// DevfileStorageLabel is the label key that is applied to all storage resources for devfile components
// that are created
const DevfileStorageLabel = "storage-name"

// SourceStorageLabel
const SourcePVCLabel = "odo-source-pvc"

// GetLabels gets the labels to be applied to the given storage besides the
// component labels and application labels.
func GetLabels(storageName string, componentName string, applicationName string, additional bool) map[string]string {
	labels := componentlabels.GetLabels(componentName, applicationName, additional)
	labels[StorageLabel] = storageName
	return labels
}
