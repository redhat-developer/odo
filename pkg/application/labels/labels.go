package labels

// ApplicationLabel is label key that is used to group all object that belong to one application
// It should be save to use just this label to filter application
const ApplicationLabel = "app.kubernetes.io/name"

// AdditionalApplicationLabels additional labels that are applied to all objects belonging to one application
// Those labels are not used for filtering or grouping, they are used just when creating and they are mend to be used by other tools
var AdditionalApplicationLabels = []string{
	// OpenShift Web console uses this label for grouping
	"app",
}

// GetLabels return labels that identifies given application
// additional labels are used only when creating object
// if you are creating something use additional=true
// if you need labels to filter component than use additional=false
func GetLabels(application string, additional bool) map[string]string {
	labels := map[string]string{
		ApplicationLabel: application,
	}

	if additional {
		for _, additionalLabel := range AdditionalApplicationLabels {
			labels[additionalLabel] = application
		}
	}

	return labels
}
