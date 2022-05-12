package kclient

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/olekukonko/tablewriter"
	corev1 "k8s.io/api/core/v1"
)

const FieldManager = "odo"

// GetInputEnvVarsFromStrings generates corev1.EnvVar values from the array of string key=value pairs
// envVars is the array containing the key=value pairs
func GetInputEnvVarsFromStrings(envVars []string) ([]corev1.EnvVar, error) {
	var inputEnvVars []corev1.EnvVar
	var keys = make(map[string]int)
	for _, env := range envVars {
		splits := strings.SplitN(env, "=", 2)
		if len(splits) < 2 {
			return nil, errors.New("invalid syntax for env, please specify a VariableName=Value pair")
		}
		_, ok := keys[splits[0]]
		if ok {
			return nil, fmt.Errorf("multiple values found for VariableName: %s", splits[0])
		}

		keys[splits[0]] = 1

		inputEnvVars = append(inputEnvVars, corev1.EnvVar{
			Name:  splits[0],
			Value: splits[1],
		})
	}
	return inputEnvVars, nil
}

// getErrorMessageFromEvents generates a error message from the given events
func getErrorMessageFromEvents(failedEvents map[string]corev1.Event) strings.Builder {
	// Create an output table
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")

	// Header
	table.SetHeader([]string{"Name", "Count", "Reason", "Message"})

	// List of events
	for name, event := range failedEvents {
		table.Append([]string{name, strconv.Itoa(int(event.Count)), event.Reason, event.Message})
	}

	// Here we render the table as well as a helpful error message
	table.Render()

	return *tableString
}

// GetGVRFromCR parses and returns the values for group, version and resource
// for a given Custom Resource (CR).
func GetGVRFromCR(cr *olm.CRDDescription) schema.GroupVersionResource {
	var group, version, resource string
	version = cr.Version

	gr := strings.SplitN(cr.Name, ".", 2)
	resource = gr[0]
	group = gr[1]

	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
}

// ConvertK8sResourceToUnstructured converts any K8s resource to unstructured.Unstructured format
// TODO: Remove this method and use https://github.com/redhat-developer/service-binding-operator/blob/master/pkg/converter/unstructured.go#L11
func ConvertK8sResourceToUnstructured(resource interface{}) (unstructuredResource unstructured.Unstructured, err error) {
	var data []byte
	data, err = json.Marshal(&resource)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &unstructuredResource.Object)
	if err != nil {
		return
	}
	return unstructuredResource, nil
}
