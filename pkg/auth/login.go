package auth

import (
	"bytes"
	"fmt"
	"os"

	"github.com/openshift/origin/pkg/oc/cli/login"
	odolog "github.com/redhat-developer/odo/pkg/log"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

// Login takes of authentication part and returns error if there any
func Login(server, username, password, token, caAuth string, skipTLS bool) error {
	//loginOutBuffer is created to intercept out msgs of login code
	loginOutBuffer := &bytes.Buffer{}
	a := login.LoginOptions{
		Server:         server,
		CommandName:    "odo",
		CAFile:         caAuth,
		InsecureTLS:    skipTLS,
		Username:       username,
		Password:       password,
		Project:        "",
		Token:          token,
		PathOptions:    &clientcmd.PathOptions{GlobalFile: clientcmd.RecommendedHomeFile, EnvVar: clientcmd.RecommendedConfigPathEnvVar, ExplicitFileFlag: "config", LoadingRules: &clientcmd.ClientConfigLoadingRules{ExplicitPath: ""}},
		RequestTimeout: 0,
		IOStreams:      genericclioptions.IOStreams{Out: loginOutBuffer, In: os.Stdin},
	}

	// initialize client-go client and read starting kubeconfig file

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	kubeconfig, _ := kubeConfig.RawConfig()

	a.StartingKubeConfig = &kubeconfig

	// if server URL is not given as argument, we will look for current context from kubeconfig file
	if len(a.Server) == 0 {
		if defaultContext, defaultContextExists := a.StartingKubeConfig.Contexts[a.StartingKubeConfig.CurrentContext]; defaultContextExists {
			if cluster, exists := a.StartingKubeConfig.Clusters[defaultContext.Cluster]; exists {
				a.Server = cluster.Server
			}
		}
	}

	err := a.Run()
	if err != nil {
		return err
	}
	// Process the messages returned by openshift login code and print our message
	originalOutMsg := loginOutBuffer.Bytes()
	var extraOut string
	// If new-project is present in original message, then ask user to do the same in our message
	if bytes.Contains(originalOutMsg, []byte("new-project")) {
		extraOut = "\nYou don't have any projects. You can try to create a new project, by running\n\n\todo project create <projectname>\n"
	}
	odolog.Successf("Logged in to \"%s\" as \"%s\"", a.Server, a.Username)
	if len(extraOut) > 0 {
		fmt.Println(extraOut)
	}
	return nil
}
