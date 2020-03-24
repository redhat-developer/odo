package pipelines

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/routes"
	"github.com/openshift/odo/pkg/pipelines/tasks"
	"github.com/openshift/odo/pkg/pipelines/triggers"
	v1rbac "k8s.io/api/rbac/v1"
)

// InitParameters is a struct that provides flags for initialise command
type InitParameters struct {
	GitOpsRepo          string
	GitOpsWebhookSecret string
	Output              string
	Prefix              string
	SkipChecks          bool
}

// PolicyRules to be bound to service account
var (
	rules = []v1rbac.PolicyRule{
		v1rbac.PolicyRule{
			APIGroups: []string{""},
			Resources: []string{"namespaces"},
			Verbs:     []string{"patch"},
		},
		v1rbac.PolicyRule{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"clusterroles"},
			Verbs:     []string{"bind", "patch"},
		},
		v1rbac.PolicyRule{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"rolebindings"},
			Verbs:     []string{"get", "patch"},
		},
	}
)

const (
	pipelineDir       = "pipelines"
	cicdDir           = "cicd-environment"
	envsDir           = "envs"
	baseDir           = "base"
	kustomize         = "kustomization.yaml"
	namespacesPath    = "01-namespaces/cicd-environment.yaml"
	rolesPath         = "02-rolebindings/pipeline-service-role.yaml"
	rolebindingsPath  = "02-rolebindings/pipeline-service-rolebinding.yaml"
	secretsPath       = "03-secrets/gitops-webhook-secret.yaml"
	tasksPath         = "04-tasks/deploy-from-source-task.yaml"
	ciPipelinesPath   = "05-pipelines/ci-dryrun-from-pr-pipeline.yaml"
	cdPipelinesPath   = "05-pipelines/cd-deploy-from-push-pipeline.yaml"
	prBindingPath     = "06-bindings/github-pr-binding.yaml"
	pushBindingPath   = "06-bindings/github-push-binding.yaml"
	prTemplatePath    = "07-templates/ci-dryrun-from-pr-template.yaml"
	pushTemplatePath  = "07-templates/cd-deploy-from-push-template.yaml"
	eventListenerPath = "08-eventlisteners/cicd-event-listener.yaml"
	routePath         = "09-routes/gitops-webhook-event-listener.yaml"
)

// Init function will initialise the gitops directory
func Init(o *InitParameters) error {

	if !o.SkipChecks {
		installed, err := checkTektonInstall()
		if err != nil {
			return fmt.Errorf("failed to run Tekton Pipelines installation check: %w", err)
		}
		if !installed {
			return errors.New("failed due to Tekton Pipelines or Triggers are not installed")
		}
	}

	namespaces := namespaceNames(o.Prefix)

	gitopsName := getGitopsRepoName(o.GitOpsRepo)
	gitopsPath := filepath.Join(o.Output, gitopsName)

	// check if the gitops dir already exists
	exists, _ := isExisting(gitopsPath)
	if exists {
		return fmt.Errorf("%s already exists at %s", gitopsName, gitopsPath)
	}

	// key: path of the resource
	// value: YAML content of the resource
	outputs := map[string]interface{}{}

	if o.GitOpsWebhookSecret != "" {
		githubSecret, err := createOpaqueSecret(meta.NamespacedName(namespaces["cicd"], eventlisteners.GitOpsWebhookSecret), o.GitOpsWebhookSecret, eventlisteners.WebhookSecretKey)
		if err != nil {
			return fmt.Errorf("failed to generate GitHub Webhook Secret: %w", err)
		}

		outputs[secretsPath] = githubSecret
	}

	// create gitops pipeline
	files := createPipelineResources(outputs, namespaces, o.GitOpsRepo, o.Prefix)

	pipelinesPath := getPipelinesDir(gitopsPath, o.Prefix)

	fileNames, err := writeResources(pipelinesPath, files)
	if err != nil {
		return err
	}

	sort.Strings(fileNames)
	// kustomize file should refer all the pipeline resources
	if err := addKustomize("resources", fileNames, filepath.Join(pipelinesPath, kustomize)); err != nil {
		return err
	}

	if err := addKustomize("bases", []string{"./pipelines"}, filepath.Join(getCICDDir(gitopsPath, o.Prefix), kustomize)); err != nil {
		return err
	}

	if err := addKustomize("bases", []string{}, filepath.Join(gitopsPath, envsDir, baseDir, kustomize)); err != nil {
		return err
	}

	return nil
}

func getCICDDir(path, prefix string) string {
	return filepath.Join(path, envsDir, addPrefix(prefix, cicdDir))
}

func createPipelineResources(outputs map[string]interface{}, namespaces map[string]string, gitopsRepo, prefix string) map[string]interface{} {

	outputs[namespacesPath] = createNamespace(namespaces["cicd"])

	outputs[rolesPath] = createClusterRole(meta.NamespacedName("", clusterRoleName), rules)

	sa := createServiceAccount(meta.NamespacedName(namespaces["cicd"], saName))

	outputs[rolebindingsPath] = createRoleBinding(meta.NamespacedName(namespaces["cicd"], roleBindingName), sa, "ClusterRole", clusterRoleName)

	outputs[tasksPath] = tasks.CreateDeployFromSourceTask(namespaces["cicd"], getPipelinesDir("", prefix))

	outputs[ciPipelinesPath] = createCIPipeline(meta.NamespacedName(namespaces["cicd"], "ci-dryrun-from-pr-pipeline"), namespaces["cicd"])

	outputs[cdPipelinesPath] = createCDPipeline(meta.NamespacedName(namespaces["cicd"], "cd-deploy-from-push-pipeline"), namespaces["cicd"])

	outputs[prBindingPath] = triggers.CreatePRBinding(namespaces["cicd"])

	outputs[pushBindingPath] = triggers.CreatePushBinding(namespaces["cicd"])

	outputs[prTemplatePath] = triggers.CreateCIDryRunTemplate(namespaces["cicd"], saName)

	outputs[pushTemplatePath] = triggers.CreateCDPushTemplate(namespaces["cicd"], saName)

	outputs[eventListenerPath] = eventlisteners.Generate(gitopsRepo, namespaces["cicd"], saName)

	outputs[routePath] = routes.Generate(namespaces["cicd"])

	return outputs
}

func writeResources(path string, files map[string]interface{}) ([]string, error) {
	filenames := make([]string, 0)
	for filename, item := range files {
		err := marshalItemsToFile(filepath.Join(path, filename), list(item))
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, filename)
	}
	return filenames, nil
}

func marshalItemsToFile(filename string, items []interface{}) error {
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return fmt.Errorf("failed to MkDirAll for %s: %v", filename, err)
	}
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to Create file %s: %v", filename, err)
	}
	defer f.Close()
	return marshalOutputs(f, items)
}

func list(i interface{}) []interface{} {
	return []interface{}{i}
}

func getPipelinesDir(rootPath, prefix string) string {
	return filepath.Join(rootPath, envsDir, addPrefix(prefix, cicdDir), pipelineDir)
}

func addKustomize(name string, items []string, path string) error {
	content := make([]interface{}, 0)
	content = append(content, map[string]interface{}{name: items})
	return marshalItemsToFile(path, content)
}

func checkTektonInstall() (bool, error) {
	tektonChecker, err := newTektonChecker()
	if err != nil {
		return false, err
	}
	return tektonChecker.checkInstall()
}

func getGitopsRepoName(repo string) string {
	return strings.Split(repo, "/")[1]
}

func addPrefix(prefix, name string) string {
	if prefix != "" {
		return prefix + name
	}
	return name
}

func isExisting(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}
