package pipelines

import (
	"errors"
	"fmt"
	"os"
	"path"

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

	//
	//  Create GitHub Secret
	//
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

	//
	//  Create Docker Secret
	//
	dockerSecret, err := createDockerConfigSecret(dockerSecretName, f)
	if err != nil {
		return err
	}
	outputs = append(outputs, dockerSecret)

	tasks := tasks.GenerateTasks()
	for _, task := range tasks {
		outputs = append(outputs, task)
	}

	eventListener := eventlisteners.GenerateEventListener(baseRepo)
	outputs = append(outputs, eventListener)

	route := routes.GenerateRoute()
	outputs = append(outputs, route)

	//
	//  Create Service Account, Role, Role Bindings, and ClusterRole Bindings
	//
	outputs = append(outputs, createServiceAccount(saName, dockerSecretName))
	outputs = append(outputs, createRole(roleName, rules))
	outputs = append(outputs, createRoleBinding(roleBindingName, saName, "Role", roleName))
	outputs = append(outputs, createRoleBinding("edit-clusterrole-binding", saName, "ClusterRole", "edit"))

	//
	// Marshall outputs to yamls
	//
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

// create and invoke a Tekton Checker
func checkTektonInstall() (bool, error) {
	tektonChecker, err := newTektonChecker()
	if err != nil {
		return false, err
	}
	return tektonChecker.checkInstall()
}
