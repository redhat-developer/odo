package logs

import "io"

type Client interface {
	// GetLogsForMode gets logs of the containers for the specified mode (Dev or Deploy) of the provided
	// component name and namespace. It returns a slice of maps where container name is the key and its logs are
	// the value. Each map is a key-value pair of container name and its logs
	GetLogsForMode(mode string, componentName string, namespace string) ([]map[string]io.ReadCloser, error)
}
