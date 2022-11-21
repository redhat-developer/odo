package helper

const (
	LabelNoCluster = "nocluster"
)

func NeedsCluster(labels []string) bool {
	for _, label := range labels {
		if label == LabelNoCluster {
			return false
		}
	}
	return true
}
