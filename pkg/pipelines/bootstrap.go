package pipelines

import (
	"errors"
	"fmt"
	"os"
	"path"

	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"

	"github.com/mitchellh/go-homedir"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/routes"
	"github.com/openshift/odo/pkg/pipelines/tasks"
	"sigs.k8s.io/yaml"
)

var (
	dockerSecretName = "regcred"
	saName           = "demo-sa"
	roleName         = "tekton-triggers-openshift-demo"
	roleBindingName  = "tekton-triggers-openshift-binding"

	// PolicyRules to be bound to service account
	rules = []v1rbac.PolicyRule{
		v1rbac.PolicyRule{
			APIGroups: []string{"tekton.dev"},
			Resources: []string{"eventlisteners", "triggerbindings", "triggertemplates", "tasks", "taskruns"},
			Verbs:     []string{"get"},
		},
		v1rbac.PolicyRule{
			APIGroups: []string{"tekton.dev"},
			Resources: []string{"pipelineruns", "pipelineresources", "taskruns"},
			Verbs:     []string{"create"},
		},
	}
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
		return errors.New("failed due to Tekton Pipelines or Triggers are not installed")
	}

	outputs := make([]interface{}, 0)

	//  Create GitHub Secret
	githubAuth, err := createGithubSecret()
	if err != nil {
		return err
	}
	outputs = append(outputs, githubAuth)

	// Create Docker Secret
	dockerSecret, err := createDockerSecret(quayUsername)
	if err != nil {
		return err
	}
	outputs = append(outputs, dockerSecret)

	tasks := tasks.Generate(githubAuth.GetName())
	for _, task := range tasks {
		outputs = append(outputs, task)
	}

	eventListener := eventlisteners.Generate(baseRepo)
	outputs = append(outputs, eventListener)

	route := routes.Generate()
	outputs = append(outputs, route)

	//  Create Service Account, Role, Role Bindings, and ClusterRole Bindings
	outputs = append(outputs, createServiceAccount(saName, dockerSecretName))
	outputs = append(outputs, createRole(roleName, rules))
	outputs = append(outputs, createRoleBinding(roleBindingName, saName, "Role", roleName))
	outputs = append(outputs, createRoleBinding("edit-clusterrole-binding", saName, "ClusterRole", "edit"))

	// Marshall
	for _, r := range outputs {
		data, err := yaml.Marshal(r)
		if err != nil {
			return err
		}
		fmt.Printf("%s---\n", data)
	}

	return nil
}

// createGithubSecret creates Github secret
func createGithubSecret() (*corev1.Secret, error) {
	tokenPath, err := pathToDownloadedFile("token")
	if err != nil {
		return nil, fmt.Errorf("failed to generate path to file: %w", err)
	}
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file %s due to %w", tokenPath, err)
	}
	defer f.Close()

	githubAuth, err := createOpaqueSecret("github-auth", f)
	if err != nil {
		return nil, err
	}

	return githubAuth, nil
}

// createDockerSecret creates Docker secret
func createDockerSecret(quayUsername string) (*corev1.Secret, error) {
	authJSONPath, err := pathToDownloadedFile(quayUsername + "-auth.json")
	if err != nil {
		return nil, fmt.Errorf("failed to generate path to file: %w", err)
	}

	f, err := os.Open(authJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read docker file '%s' due to %w", authJSONPath, err)
	}
	defer f.Close()

	dockerSecret, err := createDockerConfigSecret(dockerSecretName, f)
	if err != nil {
		return nil, err
	}

	return dockerSecret, nil

}
func pathToDownloadedFile(fname string) (string, error) {
	return homedir.Expand(path.Join("~/Downloads/", fname))
}

// create and invoke a Tekton Checker
func checkTektonInstall() (bool, error) {
	tektonChecker, err := newTektonChecker()
	if err != nil {
		return false, err
	}
	return tektonChecker.checkInstall()
}
