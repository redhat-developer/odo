package manifest

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/openshift/odo/pkg/manifest/config"
	"github.com/openshift/odo/pkg/manifest/eventlisteners"
	"github.com/openshift/odo/pkg/manifest/ioutils"
	"github.com/openshift/odo/pkg/manifest/meta"
	"github.com/openshift/odo/pkg/manifest/pipelines"
	"github.com/openshift/odo/pkg/manifest/roles"
	"github.com/openshift/odo/pkg/manifest/routes"
	"github.com/openshift/odo/pkg/manifest/secrets"
	"github.com/openshift/odo/pkg/manifest/tasks"
	"github.com/openshift/odo/pkg/manifest/triggers"
	"github.com/openshift/odo/pkg/manifest/yaml"

	v1rbac "k8s.io/api/rbac/v1"
)

type resources map[string]interface{}

// InitParameters is a struct that provides flags for the Init command.
type InitParameters struct {
	GitOpsRepo          string
	GitOpsWebhookSecret string
	Output              string
	Prefix              string
	SkipChecks          bool
}

// PolicyRules to be bound to service account
var (
	Rules = []v1rbac.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"namespaces"},
			Verbs:     []string{"patch"},
		},
		{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"clusterroles"},
			Verbs:     []string{"bind", "patch"},
		},
		{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"rolebindings"},
			Verbs:     []string{"get", "patch"},
		},
		{
			APIGroups: []string{"bitnami.com"},
			Resources: []string{"sealedsecrets"},
			Verbs:     []string{"get", "patch"},
		},
	}
)

const (
	pipelineDir = "pipelines"

	// CICDDir constants for CICD directory name
	CICDDir = "cicd"

	// EnvsDir constants for environment directory name
	EnvsDir = "environments"

	// BaseDir constant for base directory name
	BaseDir = "base"

	// Kustomize constants for kustomization.yaml
	Kustomize = "kustomization.yaml"

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

	//dockerSecretName     = "regcred"
	saName          = "pipeline"
	roleName        = "pipelines-service-role"
	roleBindingName = "pipelines-service-role-binding"
	//devRoleBindingName   = "pipeline-edit-dev"
	//stageRoleBindingName = "pipeline-edit-stage"
)

// Init bootstraps a GitOps manifest and repository structure.
func Init(o *InitParameters) error {

	if !o.SkipChecks {
		installed, err := pipelines.CheckTektonInstall()
		if err != nil {
			return fmt.Errorf("failed to run Tekton Pipelines installation check: %w", err)
		}
		if !installed {
			return errors.New("failed due to Tekton Pipelines or Triggers are not installed")
		}
	}

	exists, err := ioutils.IsExisting(o.Output)
	if exists {
		return err
	}

	outputs, err := createInitialFiles(o.Prefix, o.GitOpsRepo, o.GitOpsWebhookSecret)
	if err != nil {
		return err
	}

	_, err = yaml.WriteResources(o.Output, outputs)
	return err
}

// CreateResources creates resources assocated to pipelines
func CreateResources(prefix, gitOpsRepo, gitOpsWebhook string) (map[string]interface{}, error) {

	// key: path of the resource
	// value: YAML content of the resource
	outputs := map[string]interface{}{}
	cicdNamespace := AddPrefix(prefix, "cicd")

	githubSecret, err := secrets.CreateSealedSecret(meta.NamespacedName(cicdNamespace, eventlisteners.GitOpsWebhookSecret),
		gitOpsWebhook, eventlisteners.WebhookSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub Webhook Secret: %w", err)
	}

	outputs[secretsPath] = githubSecret
	outputs[namespacesPath] = CreateNamespace(cicdNamespace)
	outputs[rolesPath] = roles.CreateClusterRole(meta.NamespacedName("", roles.ClusterRoleName), Rules)

	sa := roles.CreateServiceAccount(meta.NamespacedName(cicdNamespace, saName))
	outputs[rolebindingsPath] = roles.CreateRoleBinding(meta.NamespacedName(cicdNamespace, roleBindingName), sa, "ClusterRole", roles.ClusterRoleName)

	outputs[tasksPath] = tasks.CreateDeployFromSourceTask(cicdNamespace, GetPipelinesDir("", prefix))
	outputs[ciPipelinesPath] = pipelines.CreateCIPipeline(meta.NamespacedName(cicdNamespace, "ci-dryrun-from-pr-pipeline"), cicdNamespace)
	outputs[cdPipelinesPath] = pipelines.CreateCDPipeline(meta.NamespacedName(cicdNamespace, "cd-deploy-from-push-pipeline"), cicdNamespace)
	outputs[prBindingPath] = triggers.CreatePRBinding(cicdNamespace)
	outputs[pushBindingPath] = triggers.CreatePushBinding(cicdNamespace)
	outputs[prTemplatePath] = triggers.CreateCIDryRunTemplate(cicdNamespace, saName)
	outputs[pushTemplatePath] = triggers.CreateCDPushTemplate(cicdNamespace, saName)
	outputs[eventListenerPath] = eventlisteners.Generate(gitOpsRepo, cicdNamespace, saName)

	outputs[routePath] = routes.Generate(cicdNamespace)
	return outputs, nil
}

func createInitialFiles(prefix, gitOpsRepo, gitOpsWebhook string) (resources, error) {
	manifest := createManifest(prefix)
	initialFiles := resources{
		"manifest.yaml": manifest,
	}

	cicdResources, err := CreateResources(prefix, gitOpsRepo, gitOpsWebhook)
	if err != nil {
		return nil, err
	}
	files := getResourceFiles(cicdResources)

	prefixedResources := addPrefixToResources(pipelinesPath(manifest), cicdResources)
	initialFiles = merge(prefixedResources, initialFiles)

	cicdKustomizations := addPrefixToResources(cicdEnvironmentPath(manifest), getCICDKustomization(files))
	initialFiles = merge(cicdKustomizations, initialFiles)

	return initialFiles, nil
}

func createManifest(prefix string) *config.Manifest {
	return &config.Manifest{
		Environments: []*config.Environment{
			{
				Name:   prefix + "cicd",
				IsCICD: true,
			},
		},
	}
}

func getCICDKustomization(files []string) resources {
	return resources{
		"base/kustomization.yaml": map[string]interface{}{
			"bases": []string{"./pipelines"},
		},
		"overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
		"base/pipelines/kustomization.yaml": map[string]interface{}{
			"resources": files,
		},
	}
}

func pathForEnvironment(env *config.Environment) string {
	return filepath.Join("environments", env.Name)
}

func pipelinesPath(m *config.Manifest) string {
	return filepath.Join(cicdEnvironmentPath(m), "base/pipelines")
}

func addPrefixToResources(prefix string, files resources) map[string]interface{} {
	updated := map[string]interface{}{}
	for k, v := range files {
		updated[filepath.Join(prefix, k)] = v
	}
	return updated
}

func merge(from, to resources) resources {
	merged := resources{}
	for k, v := range to {
		merged[k] = v
	}
	for k, v := range from {
		merged[k] = v
	}
	return merged
}

// TODO: this should probably use the .FindCICDEnvironment on the manifest.
func cicdEnvironmentPath(m *config.Manifest) string {
	return pathForEnvironment(m.Environments[0])
}

func getResourceFiles(res resources) []string {
	files := []string{}
	for k := range res {
		files = append(files, k)
	}
	sort.Strings(files)
	return files
}

// GetPipelinesDir gets pipelines directory
func GetPipelinesDir(rootPath, prefix string) string {
	return filepath.Join(rootPath, EnvsDir, AddPrefix(prefix, CICDDir), BaseDir, pipelineDir)
}
