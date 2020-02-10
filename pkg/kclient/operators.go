package kclient

import (
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetClusterServiceVersionList returns a list of CSVs in the cluster
// It is equivalent to doing `oc get csvs` using oc cli
func (c *Client) GetClusterServiceVersionList() (*olmv1alpha1.ClusterServiceVersionList, error) {
	csvs, err := c.OperatorClient.ClusterServiceVersions(c.Namespace).List(v1.ListOptions{})
	if err != nil {
		return &olmv1alpha1.ClusterServiceVersionList{}, err
	}
	return csvs, nil
}
