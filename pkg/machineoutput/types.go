package machineoutput

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redhat-developer/odo/pkg/log"
)

// ListKind is the kind used for all lists in the machine readable output
const ListKind = "List"

// APIVersion is the current API version used in the machine readable output
const APIVersion = "odo.dev/v1alpha1"

// unindentedMutex prevents multiple JSON objects from being outputted simultaneously on the same line. This is only
// required for OutputSuccessUnindented's 'unindented' JSON objects, since objects printed by other methods are not written from
// multiple threads.
var unindentedMutex = &sync.Mutex{}

// OutputSuccessUnindented outputs a "successful" machine-readable output format in unindented json
func OutputSuccessUnindented(machineOutput interface{}) {
	printableOutput, err := json.Marshal(machineOutput)

	unindentedMutex.Lock()
	defer unindentedMutex.Unlock()

	// If we error out... there's no way to output it (since we disable logging when using -o json)
	if err != nil {
		fmt.Fprintf(log.GetStderr(), "Unable to unmarshal JSON: %s\n", err.Error())
	} else {
		fmt.Fprintf(log.GetStdout(), "%s\n", string(printableOutput))
	}
}

// OutputSuccess outputs a "successful" machine-readable output format in json
func OutputSuccess(machineOutput interface{}) {
	printableOutput, err := marshalJSONIndented(machineOutput)

	// If we error out... there's no way to output it (since we disable logging when using -o json)
	if err != nil {
		fmt.Fprintf(log.GetStderr(), "Unable to unmarshal JSON: %s\n", err.Error())
	} else {
		fmt.Fprintf(log.GetStdout(), "%s\n", string(printableOutput))
	}
}

// OutputError outputs a "successful" machine-readable output format in json
func OutputError(machineOutput interface{}) {
	printableOutput, err := marshalJSONIndented(machineOutput)

	// If we error out... there's no way to output it (since we disable logging when using -o json)
	if err != nil {
		fmt.Fprintf(log.GetStderr(), "Unable to unmarshal JSON: %s\n", err.Error())
	} else {
		fmt.Fprintf(log.GetStderr(), "%s\n", string(printableOutput))
	}
}

// marshalJSONIndented returns indented json representation of obj
func marshalJSONIndented(obj interface{}) ([]byte, error) {
	return json.MarshalIndent(obj, "", "	")
}
