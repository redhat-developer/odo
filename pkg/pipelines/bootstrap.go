package pipelines

import (
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"sigs.k8s.io/yaml"
)

// Bootstrap is the main driver for getting OpenShift pipelines for GitOps
// configured with a basic configuration.
func Bootstrap(quayUsername, baseRepo, prefix string) error {

	// First, check for Tekton.  We proceed only if Tekton is installed
	installed, err := checkTektonInstall()
	if err != nil {
		return fmt.Errorf("failed to run Tekton Pipelines installation check: %w", err)
	}
	if !installed {
		return fmt.Errorf("failed due to Tekton Pipelines or Triggers are not installed")
	}

	outputs := make([]interface{}, 0)

	tokenPath, err := pathToDownloadedFile("token")
	if err != nil {
		return fmt.Errorf("failed to generate path to file: %w", err)
	}
	f, err := os.Open(tokenPath)
	if err != nil {
		return err
	}
	defer f.Close()

	githubAuth, err := createOpaqueSecret("github-auth", f)
	if err != nil {
		return err
	}
	outputs = append(outputs, githubAuth)

	authJSONPath, err := pathToDownloadedFile(quayUsername + "-auth.json")
	if err != nil {
		return fmt.Errorf("failed to generate path to file: %w", err)
	}

	f, err = os.Open(authJSONPath)
	if err != nil {
		return err
	}
	defer f.Close()

	dockerSecret, err := createDockerConfigSecret("regcred", f)
	if err != nil {
		return err
	}
	outputs = append(outputs, dockerSecret)

	for _, r := range outputs {
		data, err := yaml.Marshal(r)
		if err != nil {
			return err
		}
		fmt.Printf("%s---\n", data)
	}

	return nil
}

func pathToDownloadedFile(fname string) (string, error) {
	return homedir.Expand(path.Join("~/Downloads/", fname))
}

// create and invoke a Tetton Checker
func checkTektonInstall() (bool, error) {
	tektonChecker, err := newTektonChecker()
	if err != nil {
		return false, err
	}
	return tektonChecker.checkInstall()
}
