package pipelines

import (
	"errors"
	"fmt"
	"io"
	"os"

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

// BootstrapOptions is a struct that provides the optional flags
type BootstrapOptions struct {
	GithubToken      string
	GitRepo          string
	Prefix           string
	QuayAuthFileName string
	QuayUserName     string
}

// Bootstrap is the main driver for getting OpenShift pipelines for GitOps
// configured with a basic configuration.
func Bootstrap(o *BootstrapOptions) error {
	// First, check for Tekton.  We proceed only if Tekton is installed
	installed, err := checkTektonInstall()
	if err != nil {
		return fmt.Errorf("failed to run Tekton Pipelines installation check: %w", err)
	}
	if !installed {
		return errors.New("failed due to Tekton Pipelines or Triggers are not installed")
	}
	outputs := make([]interface{}, 0)
	names := namespaceNames(o.Prefix)
	for _, n := range createNamespaces(values(names)) {
		outputs = append(outputs, n)
	}

	githubAuth, err := createOpaqueSecret(namespacedName(names["cicd"], "github-auth"), o.GithubToken)
	if err != nil {
		return fmt.Errorf("failed to generate path to file: %w", err)
	}
	outputs = append(outputs, githubAuth)

	// Create Docker Secret
	dockerSecret, err := createDockerSecret(o.QuayAuthFileName, names["cicd"])
	if err != nil {
		return err
	}
	outputs = append(outputs, dockerSecret)

	// Create Tasks
	tasks := tasks.Generate(githubAuth.GetName(), names["cicd"])
	for _, task := range tasks {
		outputs = append(outputs, task)
	}

	// Create Event Listener
	eventListener := eventlisteners.Generate(o.GitRepo, names["cicd"])
	outputs = append(outputs, eventListener)

	// Create route
	route := routes.Generate()
	outputs = append(outputs, route)

	//  Create Service Account, Role, Role Bindings, and ClusterRole Bindings
	sa := createServiceAccount(namespacedName(names["cicd"], saName), dockerSecretName)
	outputs = append(outputs, sa)
	role := createRole(namespacedName(names["cicd"], roleName), rules)
	outputs = append(outputs, role)
	outputs = append(outputs, createRoleBinding(namespacedName(roleBindingName, names["ci-cd"]), sa, role.Kind, role.Name))
	outputs = append(outputs, createRoleBinding(namespacedName("edit-clusterrole-binding", ""), sa, "ClusterRole", "edit"))

	return marshalOutputs(os.Stdout, outputs)
}

// createDockerSecret creates Docker secret
func createDockerSecret(quayIOAuthFilename, ns string) (*corev1.Secret, error) {

	authJSONPath, err := homedir.Expand(quayIOAuthFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to generate path to file: %w", err)
	}

	f, err := os.Open(authJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read docker file '%s' : %w", authJSONPath, err)
	}
	defer f.Close()

	dockerSecret, err := createDockerConfigSecret(namespacedName(dockerSecretName, ns), f)
	if err != nil {
		return nil, err
	}

	return dockerSecret, nil

}

// create and invoke a Tekton Checker
func checkTektonInstall() (bool, error) {
	tektonChecker, err := newTektonChecker()
	if err != nil {
		return false, err
	}
	return tektonChecker.checkInstall()
}

func values(m map[string]string) []string {
	values := []string{}
	for _, v := range m {
		values = append(values, v)

	}
	return values
}

// marshalOutputs marshal outputs to given writer
func marshalOutputs(out io.Writer, outputs []interface{}) error {
	for _, r := range outputs {
		data, err := yaml.Marshal(r)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		_, err = fmt.Fprintf(out, "%s---\n", data)
		if err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}
	return nil
}
