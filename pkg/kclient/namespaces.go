package kclient

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
)

// SetCurrentNamespace sets the given namespace to current namespace
func (c *Client) SetCurrentNamespace(namespace string) error {
	fmt.Println(namespace)
	rawConfig, err := c.KubeConfig.RawConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to switch to %s namespace", namespace)
	}

	rawConfig.Contexts[rawConfig.CurrentContext].Namespace = namespace

	err = clientcmd.ModifyConfig(clientcmd.NewDefaultClientConfigLoadingRules(), rawConfig, true)
	if err != nil {
		return errors.Wrapf(err, "unable to switch to %s namespace", namespace)
	}

	c.Namespace = namespace
	return nil
}
