package auth

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/openshift/oc/pkg/cli/login"
	odolog "github.com/redhat-developer/odo/pkg/log"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct{}

func NewKubernetesClient() *KubernetesClient {
	return &KubernetesClient{}
}

// Login takes care of authentication part and returns error, if any
func (o KubernetesClient) Login(server, username, password, token, caAuth string, skipTLS bool) error {
	// Here we are grabbing the stdout output and then
	// throwing it through "copyAndFilter" in order to get
	// a correctly filtered result from `odo login`
	filteredReader, filteredWriter := io.Pipe()
	go func() {
		defer filteredWriter.Close()
		_, _ = copyAndFilter(odolog.GetStdout(), filteredReader)
	}()

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
		IOStreams:      genericclioptions.IOStreams{Out: filteredWriter, In: os.Stdin, ErrOut: odolog.GetStderr()},
	}

	// initialize client-go client and read starting kubeconfig file

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawKubeConfig, _ := kubeConfig.RawConfig()

	a.StartingKubeConfig = &rawKubeConfig

	// if server URL is not given as argument, we will look for current context from kubeconfig file
	if len(a.Server) == 0 {
		if defaultContext, defaultContextExists := a.StartingKubeConfig.Contexts[a.StartingKubeConfig.CurrentContext]; defaultContextExists {
			if cluster, exists := a.StartingKubeConfig.Clusters[defaultContext.Cluster]; exists {
				a.Server = cluster.Server
			}
		}
	}

	// if defaultNamespace is not defined, we will look for current namespace from kubeconfig file if defined
	if len(a.DefaultNamespace) == 0 {
		if defaultContext, defaultContextExists := a.StartingKubeConfig.Contexts[a.StartingKubeConfig.CurrentContext]; defaultContextExists {
			if len(defaultContext.Namespace) > 0 {
				a.DefaultNamespace = defaultContext.Namespace
			}
		}
	}

	// 1. Say we're connecting
	odolog.Info("Connecting to the OpenShift cluster\n")

	// 2. Handle the error messages here. This is copied over from:
	// https://github.com/openshift/origin/blob/master/pkg/oc/cli/login/login.go#L60
	// as unauthorized errors are handled MANUALLY by oc.
	if err := a.GatherInfo(); err != nil {
		if kapierrors.IsUnauthorized(err) {
			fmt.Println("Login failed (401 Unauthorized)")
			fmt.Println("Verify you have provided correct credentials.")

			if err, isStatusErr := err.(*kapierrors.StatusError); isStatusErr {
				if details := err.Status().Details; details != nil {
					for _, cause := range details.Causes {
						fmt.Println(cause.Message)
					}
				}
			}
		}
		return err
	}

	// 3. Correctly save the configuration
	newFileCreated, err := a.SaveConfig()
	if err != nil {
		return err
	}

	// If a new file has been created, we output what to do next (obviously odo help). This is taken from:
	// https://github.com/openshift/origin/blob/4c293b86b111d9aaeba7bb1e72ee57410652ae9d/pkg/oc/cli/login/login.go#L184
	if newFileCreated {
		odolog.Infof("\nWelcome! See '%s help' to get started.", a.CommandName)
	}

	return nil
}

// copyAndFilter captures the output, filters it and then spits it back out to stdout.
// Kindly taken from https://stackoverflow.com/questions/54570268/filtering-the-output-of-a-terminal-output-using-golang
func copyAndFilter(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			if _, e := w.Write(filteredInformation(d)); e != nil {
				return out, e
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

// filteredInformation takes a list of strings ([]byte), replaces them and spits it back out
// This is used since we utilize `oc login` with odo and require certain strings to be filtered / changed
// to their odo equivalent
func filteredInformation(s []byte) []byte {

	// List of strings to correctly filter
	s = bytes.Replace(s, []byte("oc new-project"), []byte("odo project create"), -1)
	s = bytes.Replace(s, []byte("<projectname>"), []byte("<project-name>"), -1)
	s = bytes.Replace(s, []byte("project <project-name>"), []byte("project set <project-name>"), -1)
	s = bytes.Replace(s, []byte("odo projects"), []byte("odo project list"), -1)

	return s
}
