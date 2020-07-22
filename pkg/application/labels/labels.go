package labels

import (
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/pkg/version"
)

// ApplicationLabel is label key that is used to group all object that belong to one application
// It should be save to use just this label to filter application
const ApplicationLabel = "app.kubernetes.io/part-of"

//////////////////////////////
// ADDITIONALLY USED LABELS
//////////////////////////////

// App is the default name used when labeling
const App = "app"

// OdoManagedBy notes that this is managed by odo
const OdoManagedBy = "app.kubernetes.io/managed-by"

// OdoVersion is a Kubernetes label that adds what version of odo is being ran.
const OdoVersion = "app.kubernetes.io/managed-by-version"

// GetLabels return labels that identifies given application
// additional labels are used only when creating object
// if you are creating something use additional=true
// if you need labels to filter component then use additional=false
func GetLabels(application string, additional bool) map[string]string {
	labels := map[string]string{
		ApplicationLabel: application,
	}

	if additional {
		labels[App] = application
		labels[OdoVersion] = version.VERSION
		labels[OdoManagedBy] = "odo"
	}

	return labels
}

// GetSelector are supposed to be used for selection of any resource part of an application
func GetSelector(application string) string {
	labels := map[string]string{
		ApplicationLabel: application,
		App:              application,
		OdoManagedBy:     "odo",
	}

	return util.ConvertLabelsToSelector(labels)
}
