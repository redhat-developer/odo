package labels

import (
	componentlabels "github.com/openshift/odo/pkg/component/labels"
)

// URLLabel is the label key that is applied to all url resources
// that are created
const URLLabel = "app.kubernetes.io/url-name"

// GetLabels gets the labels to be applied to the given url besides the
// component labels and application labels.
func GetLabels(urlName string, componentName string, applicationName string, additional bool) map[string]string {
	labels := componentlabels.GetLabels(componentName, applicationName, additional)
	labels[URLLabel] = urlName
	return labels
}
