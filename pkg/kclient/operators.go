package kclient

import (
	"github.com/golang/glog"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetClusterServiceVersionList returns a list of CSVs in the cluster
// It is equivalent to doing `oc get csvs` using oc cli
func (c *Client) GetClusterServiceVersionList() (*olm.ClusterServiceVersionList, error) {
	glog.V(4).Infof("Fetching list of operators installed in cluster")
	csvs, err := c.OperatorClient.ClusterServiceVersions(c.Namespace).List(v1.ListOptions{})
	if err != nil {
		return &olm.ClusterServiceVersionList{}, err
	}
	return csvs, nil
}
