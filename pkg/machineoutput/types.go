package machineoutput

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Kind is what kind we should use in the machine readable output
const Kind = "Error"

// APIVersion is the current API version we are using
const APIVersion = "odo.openshift.io/v1alpha1"

// GenericError for machine readable output error messages
type GenericError struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Message           string `json:"message"`
}

// GenericSuccess same as above, but copy-and-pasted just in case
// we change the output in the future
type GenericSuccess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Message           string `json:"message"`
}

// OutputSuccess outputs a "successful" machine-readable output format in json
func OutputSuccess(machineOutput interface{}) {
	printableOutput, err := json.Marshal(machineOutput)

	// If we error out... there's no way to output it (since we disable logging when using -o json)
	if err != nil {
		fmt.Fprintf(log.GetStderr(), "Unable to unmarshal JSON: %s\n", err.Error())
	} else {
		fmt.Fprintf(log.GetStdout(), "%s\n", string(printableOutput))
	}
}

// OutputError outputs a "successful" machine-readable output format in json
func OutputError(machineOutput interface{}) {
	printableOutput, err := json.Marshal(machineOutput)

	// If we error out... there's no way to output it (since we disable logging when using -o json)
	if err != nil {
		fmt.Fprintf(log.GetStderr(), "Unable to unmarshal JSON: %s\n", err.Error())
	} else {
		fmt.Fprintf(log.GetStderr(), "%s\n", string(printableOutput))
	}
}

// CatalogListServices `odo catalog list services` standard machine readable output
func CatalogListServices(services []occlient.Service) {

	data := struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`
		Items             []occlient.Service `json:"items"`
	}{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CatalogListServices",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		Items: services,
	}

	OutputSuccess(data)
}

// ProjectSuccess outputs a success output that includes
// project information and namespace
func ProjectSuccess(projectName string, message string) {
	machineOutput := GenericSuccess{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Project",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectName,
			Namespace: projectName,
		},
		Message: message,
	}

	OutputSuccess(machineOutput)
}
