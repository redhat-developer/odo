package kclient

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/odo/cli/catalog/util"

	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

var (
	ErrNoSuchOperator = errors.New("Could not find specified operator")
)

const (
	apiVersion = "odo.dev/v1alpha1"
)

// GetClusterServiceVersionList returns a list of CSVs in the cluster
// It is equivalent to doing `oc get csvs` using oc cli
func (c *Client) GetClusterServiceVersionList() (*olm.ClusterServiceVersionList, error) {
	klog.V(3).Infof("Fetching list of operators installed in cluster")
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

// GetCustomResourcesFromCSV returns a list of CRs provided by an operator/CSV.
func (c *Client) GetCustomResourcesFromCSV(csv *olm.ClusterServiceVersion) *[]olm.CRDDescription {
	// we will return a list of CRs owned by the csv
	return &csv.Spec.CustomResourceDefinitions.Owned
}

// SearchClusterServiceVersionList searches for whether the operator/CSV contains
// given keyword then return it
func (c *Client) SearchClusterServiceVersionList(name string) (*olm.ClusterServiceVersionList, error) {
	var result []olm.ClusterServiceVersion
	csvs, err := c.GetClusterServiceVersionList()
	if err != nil {
		return &olm.ClusterServiceVersionList{}, errors.Wrap(err, "unable to list services")
	}

	// do a partial search in all the services
	for _, service := range csvs.Items {
		if strings.Contains(service.ObjectMeta.Name, name) {
			result = append(result, service)
		} else if strings.Contains(util.CsvOperators(service.Spec.CustomResourceDefinitions), name) {
			result = append(result, service)
		}
	}

	return &olm.ClusterServiceVersionList{
		TypeMeta: v1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		Items: result,
	}, nil
}

// GetCustomResource returns the CR matching the name
func (c *Client) GetCustomResource(customResource string) (*olm.CRDDescription, error) {
	// Get all csvs in the namespace
	csvs, err := c.GetClusterServiceVersionList()
	if err != nil {
		return &olm.CRDDescription{}, err
	}

	// iterate of csvs to find if CR of our interest is provided by any of those
	for _, csv := range csvs.Items {
		clusSerVer := csv
		crs := c.GetCustomResourcesFromCSV(&clusSerVer)

		for _, cr := range *crs {
			if cr.Kind == customResource {
				return &cr, nil
			}
		}
	}

	return &olm.CRDDescription{}, fmt.Errorf("Couldn't find a Custom Resource named %q in the namespace", customResource)
}

// GetCSVWithCR returns the CSV (Operator) that contains the CR (service)
func (c *Client) GetCSVWithCR(name string) (*olm.ClusterServiceVersion, error) {
	csvs, err := c.GetClusterServiceVersionList()
	if err != nil {
		return &olm.ClusterServiceVersion{}, errors.Wrap(err, "unable to list services")
	}

	for _, csv := range csvs.Items {
		clusterServiceVersion := csv
		for _, cr := range *c.GetCustomResourcesFromCSV(&clusterServiceVersion) {
			if cr.Kind == name {
				return &csv, nil
			}
		}
	}
	return &olm.ClusterServiceVersion{}, fmt.Errorf("Could not find any Operator containing requested CR: %s", name)
}
