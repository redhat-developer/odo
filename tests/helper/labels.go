package helper

import (
	"github.com/onsi/ginkgo/v2"
)

const (
	LabelNoCluster = "nocluster"
	LabelUnauth    = "unauth"
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

func IsAuth(labels []string) bool {
	for _, label := range labels {
		if label == LabelUnauth {
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
