package labels

import (
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
)

// UrlLabel is the label key that is applied to all url resources
// that are created
const UrlLabel = "app.kubernetes.io/url-name"

// GetLabels gets the labels to be applied to the given url besides the
// component labels and application labels.
func GetLabels(urlName string, componentName string, applicationName string, additional bool) map[string]string {
	labels := componentlabels.GetLabels(componentName, applicationName, additional)
	labels[UrlLabel] = urlName
	return labels
}
