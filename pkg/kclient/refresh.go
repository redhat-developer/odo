package kclient

import (
	"fmt"
	"reflect"

	"k8s.io/client-go/tools/clientcmd"
)

// Refresh re-creates a new Kubernetes client and checks if the Config changes
// If config changed, updates the Kubernetes client with the new configuration and returns true
// If the namespace or cluster of the current context has changed since the last time
// the config has been loaded, the function will not update the configuration
func (c *Client) Refresh() (bool, error) {
	newClient, err := New()
	if err != nil {
		return false, err
	}

	oldCluster, oldNs, err := getContext(c)
	if err != nil {
		return false, err
	}
	newCluster, newNs, err := getContext(newClient)
	if err != nil {
		return false, err
	}

	if oldCluster != newCluster {
		return false, fmt.Errorf("cluster changed (%q -> %q), won't refresh the configuration", oldCluster, newCluster)
	}
	if oldNs != newNs {
		return false, fmt.Errorf("namespace changed (%q -> %q), won't refresh the configuration", oldNs, newNs)
	}

	updated, err := isConfigUpdated(c.GetConfig(), newClient.GetConfig())
	if err != nil {
		return false, err
	}

	if updated {
		*c = *newClient
	}
	return updated, nil
}

func getContext(c *Client) (cluster string, namespace string, err error) {
	raw, err := c.GetConfig().RawConfig()
	if err != nil {
		return "", "", err
	}

	currentCtx := raw.Contexts[raw.CurrentContext]
	return currentCtx.Cluster, currentCtx.Namespace, nil
}

func isConfigUpdated(oldC, newC clientcmd.ClientConfig) (bool, error) {
	oldRaw, err := oldC.RawConfig()
	if err != nil {
		return false, err
	}
	newRaw, err := newC.RawConfig()
	if err != nil {
		return false, err
	}
	return !reflect.DeepEqual(oldRaw, newRaw), nil
}
