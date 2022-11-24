package helper

import (
	"github.com/onsi/ginkgo/v2"
)

const (
	LabelNoCluster = "nocluster"
	LabelPodman    = "podman"
)

func NeedsCluster(labels []string) bool {
	for _, label := range labels {
		if label == LabelNoCluster {
			return false
		}
		if label == LabelPodman {
			return false
		}
	}
	return true
}

func LabelPodmanIf(value bool, args ...interface{}) []interface{} {
	res := []interface{}{}
	if value {
		res = append(res, ginkgo.Label(LabelPodman))
	}
	res = append(res, args...)
	return res
}
