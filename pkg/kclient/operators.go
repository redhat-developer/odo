package kclient

import (
	"github.com/golang/glog"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ErrNoSuchOperator = odoError("Could not find specified operator")
)

type odoError string

func (e odoError) Error() string {
	return string(e)
}

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

// GetClusterServiceVersion returns a particular CSV from a list of CSVs
func (c *Client) GetClusterServiceVersion(name string) (olm.ClusterServiceVersion, error) {
	csvs, err := c.GetClusterServiceVersionList()
	if err != nil {
		return olm.ClusterServiceVersion{}, err
	}
	for _, item := range csvs.Items {
		if item.Name == name {
			return item, nil
		}
	}
	return olm.ClusterServiceVersion{}, ErrNoSuchOperator
}
