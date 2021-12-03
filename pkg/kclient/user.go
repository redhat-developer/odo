package kclient

import (
	"context"
	"fmt"
	"io"

	oauthv1client "github.com/openshift/client-go/oauth/clientset/versioned/typed/oauth/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

// RunLogout logs out the current user from cluster
func (c *Client) RunLogout(stdout io.Writer) error {
	output, err := c.userClient.Users().Get(context.TODO(), "~", metav1.GetOptions{})
	if err != nil {
		klog.V(1).Infof("%v : unable to get userinfo", err)
	}

	// read the current config form ~/.kube/config
	conf, err := c.KubeConfig.ClientConfig()
	if err != nil {
		klog.V(1).Infof("%v : unable to get client config", err)
	}
	// initialising oauthv1client
	client, err := oauthv1client.NewForConfig(conf)
	if err != nil {
		klog.V(1).Infof("%v : unable to create a new OauthV1Client", err)
	}

	// deleting token form the server
	if e := client.OAuthAccessTokens().Delete(context.TODO(), conf.BearerToken, metav1.DeleteOptions{}); e != nil {
		klog.V(1).Infof("%v", e)
	}

	rawConfig, err := c.KubeConfig.RawConfig()
	if err != nil {
		klog.V(1).Infof("%v : unable to switch to  project", err)
	}

	// deleting token for the current server from local config
	for key, value := range rawConfig.AuthInfos {
		if key == rawConfig.Contexts[rawConfig.CurrentContext].AuthInfo {
			value.Token = ""
		}
	}
	err = clientcmd.ModifyConfig(clientcmd.NewDefaultClientConfigLoadingRules(), rawConfig, true)
	if err != nil {
		klog.V(1).Infof("%v : unable to write config to config file", err)
	}

	_, err = io.WriteString(stdout, fmt.Sprintf("Logged \"%v\" out on \"%v\"\n", output.Name, conf.Host))
	return err
}
