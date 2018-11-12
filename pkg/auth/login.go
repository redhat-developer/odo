package auth

import (
	"os"

	"github.com/openshift/origin/pkg/oc/cli/login"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kclientapi "k8s.io/client-go/tools/clientcmd/api"
)

// Login takes of authentication part and returns error if there any
func Login(server, username, password, token, caAuth string, skipTLS bool) error {
	var config *restclient.Config
	if server == "" {

		// initialize client-go client
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, _ = kubeConfig.ClientConfig()
	}

	a := login.LoginOptions{
		Server:      server,
		CommandName: "odo",
		CAFile:      caAuth,
		InsecureTLS: skipTLS,
		Username:    username,
		Password:    password,
		Project:     "",
		Token:       token,
		StartingKubeConfig: &kclientapi.Config{
			Clusters:  map[string]*kclientapi.Cluster{},
			AuthInfos: map[string]*kclientapi.AuthInfo{},
			Contexts:  map[string]*kclientapi.Context{},
		},
		Config:         config,
		PathOptions:    &clientcmd.PathOptions{GlobalFile: clientcmd.RecommendedHomeFile, EnvVar: clientcmd.RecommendedConfigPathEnvVar, ExplicitFileFlag: "config", LoadingRules: &clientcmd.ClientConfigLoadingRules{ExplicitPath: ""}},
		RequestTimeout: 0,
		IOStreams:      genericclioptions.IOStreams{Out: os.Stdout, In: os.Stdin},
	}

	err := a.Run()
	if err != nil {
		return err
	}

	return nil
}
