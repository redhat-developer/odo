package auth

import (
	"bytes"
	"os"

	"github.com/fatih/color"
	odolog "github.com/openshift/odo/pkg/log"
	"github.com/openshift/origin/pkg/oc/cli/login"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

// Login takes of authentication part and returns error if there any
func Login(server, username, password, token, caAuth string, skipTLS bool) error {
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
		IOStreams:      genericclioptions.IOStreams{Out: odolog.GetStdout(), In: os.Stdin, ErrOut: odolog.GetStderr()},
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

	// 1. Say we're connecting
	odolog.Info("Connecting to the OpenShift cluster\n")

	// 2. Gather information and connect
	// We set the color to "yellow" to distinguish between odo and oc output
	color.Set(color.FgYellow)
	if err := a.GatherInfo(); err != nil {
		// Make sure we newline between the "Connecting" and error-out messages
		odolog.Info("")
		color.Unset()
		return err
	}
	color.Unset()

	// 3. Output the information in a correct format
	// In order to interpret the error message and manipulate the output of `oc`
	// we pass in a buffer in order to modify the output.
	loginOutBuffer := &bytes.Buffer{}
	a.IOStreams = genericclioptions.IOStreams{Out: loginOutBuffer, In: os.Stdin, ErrOut: os.Stderr}
	newFileCreated, err := a.SaveConfig()
	if err != nil {
		return err
	}

	// If a new file has been created, we output what to do next (obviously odo help). This is taken from:
	// https://github.com/openshift/origin/blob/4c293b86b111d9aaeba7bb1e72ee57410652ae9d/pkg/oc/cli/login/login.go#L184
	if newFileCreated {
		odolog.Infof("Welcome! See '%s help' to get started.", a.CommandName)
	}

	// Process the messages returned by openshift login code and print our message
	originalOutMsg := loginOutBuffer.Bytes()
	loginSuccessMsg := bytes.Replace(originalOutMsg, []byte("new-project"), []byte("project create"), -1)
	loginSuccessMsg = bytes.Replace(loginSuccessMsg, []byte("<projectname>"), []byte("<project-name>"), -1)

	// Add newline + output succeded
	odolog.Info("")
	if len(originalOutMsg) == 0 {
		odolog.Success("Login succeeded")
	} else {
		odolog.Successf("%s", loginSuccessMsg)
	}

	return nil
}
