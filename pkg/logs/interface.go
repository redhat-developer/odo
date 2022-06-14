package logs

import "io"

type Client interface {
	// DevModeLogs gets logs for the Dev mode containers of the provided component name and namespace. A component could
	// have multiple pods and containers running on the cluster. It returns a slice of maps where container name is the
	// key and its logs are the value.
	DevModeLogs(componentName string, namespace string) ([]map[string]io.ReadCloser, error)

	// DeployModeLogs gets logs for the Deploy mode containers of the provided component name and namespace. A component
	// could have multiple pods and containers running on the cluster. It returns a slice of maps where container name
	// is the key and its logs are the value.
	DeployModeLogs(componentName string, namespce string) ([]map[string]io.ReadCloser, error)

	// AllModeLogs gets logs for both Dev and Deploy mode containers of the provided component name and namespace
	AllModeLogs(componentName string, namespace string) ([]map[string]io.ReadCloser, error)
}
