package kclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	apiVersion = "odo.dev/v1alpha1"
)

// IsServiceBindingSupported checks if resource of type service binding request present on the cluster
func (c *Client) IsServiceBindingSupported() (bool, error) {
	return c.IsResourceSupported("binding.operators.coreos.com", "v1alpha1", "servicebindings")
}

// IsCSVSupported checks if resource of type service binding request present on the cluster
func (c *Client) IsCSVSupported() (bool, error) {
	return c.IsResourceSupported("operators.coreos.com", "v1alpha1", "clusterserviceversions")
}

// ListClusterServiceVersions returns a list of CSVs in the cluster
// It is equivalent to doing `oc get csvs` using oc cli
func (c *Client) ListClusterServiceVersions() (*olm.ClusterServiceVersionList, error) {
	klog.V(3).Infof("Fetching list of operators installed in cluster")
	csvs, err := c.OperatorClient.ClusterServiceVersions(c.Namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &olm.ClusterServiceVersionList{}, nil
		}
		return &olm.ClusterServiceVersionList{}, err
	}
	return csvs, nil
}

// GetClusterServiceVersion returns a particular CSV from a list of CSVs
func (c *Client) GetClusterServiceVersion(name string) (olm.ClusterServiceVersion, error) {
	csv, err := c.OperatorClient.ClusterServiceVersions(c.Namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return olm.ClusterServiceVersion{}, err
	}
	return *csv, nil
}

// GetCustomResourcesFromCSV returns a list of CRs provided by an operator/CSV.
func (c *Client) GetCustomResourcesFromCSV(csv *olm.ClusterServiceVersion) *[]olm.CRDDescription {
	// we will return a list of CRs owned by the csv
	return &csv.Spec.CustomResourceDefinitions.Owned
}

// CheckCustomResourceInCSV checks if the custom resource is present in the CSV.
func (c *Client) CheckCustomResourceInCSV(customResource string, csv *olm.ClusterServiceVersion) (bool, *olm.CRDDescription) {
	var cr *olm.CRDDescription
	hasCR := false
	CRs := c.GetCustomResourcesFromCSV(csv)
	for _, custRes := range *CRs {
		c := custRes
		if c.Kind == customResource {
			cr = &c
			hasCR = true
			break
		}
	}
	return hasCR, cr
}

// SearchClusterServiceVersionList searches for whether the operator/CSV contains
// given keyword then return it
func (c *Client) SearchClusterServiceVersionList(name string) (*olm.ClusterServiceVersionList, error) {
	var result []olm.ClusterServiceVersion
	csvs, err := c.ListClusterServiceVersions()
	if err != nil {
		return &olm.ClusterServiceVersionList{}, errors.Wrap(err, "unable to list services")
	}

	// do a partial search in all the services
	for _, service := range csvs.Items {
		if strings.Contains(service.ObjectMeta.Name, name) {
			result = append(result, service)
		} else {
			for _, crd := range service.Spec.CustomResourceDefinitions.Owned {
				if name == crd.Kind {
					result = append(result, service)
				}
			}
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
	csvs, err := c.ListClusterServiceVersions()
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

	return &olm.CRDDescription{}, fmt.Errorf("could not find a Custom Resource named %q in the namespace", customResource)
}

// GetCSVWithCR returns the CSV (Operator) that contains the CR (service)
func (c *Client) GetCSVWithCR(name string) (*olm.ClusterServiceVersion, error) {
	csvs, err := c.ListClusterServiceVersions()
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
	return &olm.ClusterServiceVersion{}, fmt.Errorf("could not find any Operator containing requested CR: %s", name)
}

// GetResourceSpecDefinition returns the OpenAPI v2 definition of the Kubernetes resource of a given group/version/kind
func (c *Client) GetResourceSpecDefinition(group, version, kind string) (*spec.Schema, error) {
	data, err := c.KubeClient.Discovery().RESTClient().Get().AbsPath("/openapi/v2").SetHeader("Accept", "application/json").Do(context.TODO()).Raw()
	if err != nil {
		return nil, err
	}
	return getResourceSpecDefinitionFromSwagger(data, group, version, kind)
}

// getResourceSpecDefinitionFromSwagger returns the OpenAPI v2 definition of the Kubernetes resource of a given group/version/kind, for a given swagger data
func getResourceSpecDefinitionFromSwagger(data []byte, group, version, kind string) (*spec.Schema, error) {
	schema := new(spec.Schema)
	err := json.Unmarshal([]byte(data), schema)
	if err != nil {
		return nil, err
	}

	var crd spec.Schema
	found := false
loopDefinitions:
	for _, definition := range schema.Definitions {
		extensions := definition.Extensions
		gvkI, ok := extensions["x-kubernetes-group-version-kind"]
		if !ok {
			continue
		}
		// The concrete type of this extension is expected to be an array of interface{}
		// If not, we ignore it
		gvkA, ok := gvkI.([]interface{})
		if !ok {
			continue
		}

		for i := range gvkA {
			// The concrete type of each element is expected to be a map[string]interface{}
			// If not, we ignore it
			gvk, ok := gvkA[i].(map[string]interface{})
			if !ok {
				continue
			}
			gvkGroup := gvk["group"].(string)
			gvkVersion := gvk["version"].(string)
			gvkKind := gvk["kind"].(string)
			if strings.HasSuffix(group, gvkGroup) && version == gvkVersion && kind == gvkKind {
				crd = definition
				found = true
				break loopDefinitions
			}
		}

	}
	if !found {
		return nil, errors.New("no definition found")
	}

	spec, ok := crd.Properties["spec"]
	if ok {
		return &spec, nil
	}
	return nil, nil
}
