package helper

import (
	"github.com/onsi/ginkgo/v2"
)

const (
	// LabelCluster is the default label, indicating tests that want to run with a cluster.
	// If no label is set, tests are assumed to be running in cluster mode.
	LabelCluster   = "cluster"
	LabelNoCluster = "nocluster"
	LabelUnauth    = "unauth"
	LabelPodman    = "podman"
)

func NeedsCluster(labels []string) bool {
	for _, label := range labels {
		if label == LabelCluster {
			return true
		}
		if label == LabelNoCluster {
			return false
		}
		if label == LabelPodman {
			// Check if there is any "cluster" label
			for _, l := range labels {
				if l == LabelCluster {
					return true
				}
			}
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
