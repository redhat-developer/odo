package kclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/go-openapi/spec"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

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

// GetCustomResourcesFromCSV returns a list of CRs provided by an operator/CSV.
func (c *Client) GetCustomResourcesFromCSV(csv *olm.ClusterServiceVersion) *[]olm.CRDDescription {
	// we will return a list of CRs owned by the csv
	return &csv.Spec.CustomResourceDefinitions.Owned
}

// GetCSVWithCR returns the CSV (Operator) that contains the CR (service)
func (c *Client) GetCSVWithCR(name string) (*olm.ClusterServiceVersion, error) {
	csvs, err := c.ListClusterServiceVersions()
	if err != nil {
		return &olm.ClusterServiceVersion{}, fmt.Errorf("unable to list services: %w", err)
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
	err := json.Unmarshal(data, schema)
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

// toOpenAPISpec transforms Spec descriptors from a CRD description to an OpenAPI schema
func toOpenAPISpec(repr *olm.CRDDescription) *spec.Schema {
	if len(repr.SpecDescriptors) == 0 {
		return nil
	}
	schema := new(spec.Schema).Typed("object", "")
	schema.AdditionalProperties = &spec.SchemaOrBool{
		Allows: false,
	}
	for _, param := range repr.SpecDescriptors {
		addParam(schema, param)
	}
	return schema
}

// addParam adds a Spec Descriptor parameter to an OpenAPI schema
func addParam(schema *spec.Schema, param olm.SpecDescriptor) {
	parts := strings.SplitN(param.Path, ".", 2)
	if len(parts) == 1 {
		child := spec.StringProperty()
		if len(param.XDescriptors) == 1 {
			switch param.XDescriptors[0] {
			case "urn:alm:descriptor:com.tectonic.ui:podCount":
				child = spec.Int32Property()
				// TODO(feloy) more cases, based on
				// - https://github.com/openshift/console/blob/master/frontend/packages/operator-lifecycle-manager/src/components/descriptors/reference/reference.md
				// - https://docs.google.com/document/d/17Tdmpu4R6pA5UC4LumyJ2EP6AcotMWM127Jy728hYCk
			}
		}
		child = child.WithTitle(param.DisplayName).WithDescription(param.Description)
		schema.SetProperty(parts[0], *child)
	} else {
		var child *spec.Schema
		if _, ok := schema.Properties[parts[0]]; ok {
			c := schema.Properties[parts[0]]
			child = &c
		} else {
			child = new(spec.Schema).Typed("object", "")
		}
		param.Path = parts[1]
		addParam(child, param)
		schema.SetProperty(parts[0], *child)
	}
}

// GetRestMappingFromUnstructured returns rest mappings from unstructured data
func (c *Client) GetRestMappingFromUnstructured(u unstructured.Unstructured) (*meta.RESTMapping, error) {
	gvk := u.GroupVersionKind()
	return c.restmapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}

func (c *Client) GetRestMappingFromGVK(gvk schema.GroupVersionKind) (*meta.RESTMapping, error) {
	return c.restmapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}

func (c *Client) GetGVKFromGVR(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	return c.restmapper.KindFor(gvr)
}

func (c *Client) GetGVRFromGVK(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	mapping, err := c.restmapper.RESTMapping(gvk.GroupKind())
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return mapping.Resource, nil
}

// GetOperatorGVRList creates a slice of rest mappings that are provided by Operators (CSV)
func (c *Client) GetOperatorGVRList() ([]meta.RESTMapping, error) {
	var operatorGVRList []meta.RESTMapping

	// ignoring the error because
	csvs, err := c.ListClusterServiceVersions()
	if err != nil {
		return operatorGVRList, err
	}
	for _, c := range csvs.Items {
		owned := c.Spec.CustomResourceDefinitions.Owned
		for i := range owned {
			operatorGVRList = append(operatorGVRList, meta.RESTMapping{
				Resource: GetGVRFromCR(&owned[i]),
			})
		}
	}
	return operatorGVRList, nil
}

func ConvertUnstructuredToResource(u unstructured.Unstructured, obj interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), obj)
}

func ConvertUnstructuredListToResource(u unstructured.UnstructuredList, obj interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), obj)
}
