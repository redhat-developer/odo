package pipelines

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mitchellh/go-homedir"
	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/dryrun"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/namespaces"
	"github.com/openshift/odo/pkg/pipelines/pipelines"
	"github.com/openshift/odo/pkg/pipelines/resources"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/roles"
	"github.com/openshift/odo/pkg/pipelines/routes"
	"github.com/openshift/odo/pkg/pipelines/scm"
	"github.com/openshift/odo/pkg/pipelines/secrets"
	"github.com/openshift/odo/pkg/pipelines/tasks"
	"github.com/openshift/odo/pkg/pipelines/triggers"
	"github.com/openshift/odo/pkg/pipelines/yaml"
	"github.com/spf13/afero"

	v1rbac "k8s.io/api/rbac/v1"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
)

// InitOptions is a struct that provides flags for the Init command.
type InitOptions struct {
	GitOpsRepoURL            string // This is where the pipelines and configuration are.
	GitOpsWebhookSecret      string // This is the secret for authenticating hooks from your GitOps repo.
	Prefix                   string
	DockerConfigJSONFilename string
	InternalRegistryHostname string // This is the internal registry hostname used for pushing images.
	ImageRepo                string // This is where built images are pushed to.
	OutputPath               string // Where to write the bootstrapped files to?
	SealedSecretsNamespace   string // Where do we find the SealedSecrets service?
}

// PolicyRules to be bound to service account
var (
	Rules = []v1rbac.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"namespaces", "services"},
			Verbs:     []string{"patch", "get", "create"},
		},
		{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"clusterroles"},
			Verbs:     []string{"bind", "patch"},
		},
		{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"clusterrolebindings", "rolebindings"},
			Verbs:     []string{"get", "create", "patch"},
		},
		{
			APIGroups: []string{"bitnami.com"},
			Resources: []string{"sealedsecrets"},
			Verbs:     []string{"get", "patch", "create"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments"},
			Verbs:     []string{"get", "create", "patch"},
		},
		{
			APIGroups: []string{"argoproj.io"},
			Resources: []string{"applications"},
			Verbs:     []string{"get", "create", "patch"},
		},
	}
)

const (
	// Kustomize constants for kustomization.yaml
	Kustomize = "kustomization.yaml"

	namespacesPath           = "01-namespaces/cicd-environment.yaml"
	rolesPath                = "02-rolebindings/pipeline-service-role.yaml"
	rolebindingsPath         = "02-rolebindings/pipeline-service-rolebinding.yaml"
	serviceAccountPath       = "02-rolebindings/pipeline-service-account.yaml"
	secretsPath              = "03-secrets/gitops-webhook-secret.yaml"
	dockerConfigPath         = "03-secrets/docker-config.yaml"
	gitopsTasksPath          = "04-tasks/deploy-from-source-task.yaml"
	appTaskPath              = "04-tasks/deploy-using-kubectl-task.yaml"
	ciPipelinesPath          = "05-pipelines/ci-dryrun-from-pr-pipeline.yaml"
	appCiPipelinesPath       = "05-pipelines/app-ci-pipeline.yaml"
	cdPipelinesPath          = "05-pipelines/cd-deploy-from-push-pipeline.yaml"
	prTemplatePath           = "07-templates/ci-dryrun-from-pr-template.yaml"
	pushTemplatePath         = "07-templates/cd-deploy-from-push-template.yaml"
	appCIBuildPRTemplatePath = "07-templates/app-ci-build-pr-template.yaml"
	eventListenerPath        = "08-eventlisteners/cicd-event-listener.yaml"
	routePath                = "09-routes/gitops-webhook-event-listener.yaml"

	dockerSecretName = "regcred"

	saName              = "pipeline"
	roleBindingName     = "pipelines-service-role-binding"
	webhookSecretLength = 20
)

// Init bootstraps a GitOps pipelines and repository structure.
func Init(o *InitOptions, fs afero.Fs) error {
	exists, err := ioutils.IsExisting(fs, o.OutputPath)
	if exists {
		return err
	}
	if o.GitOpsWebhookSecret == "" {
		gitSecret, err := secrets.GenerateString(webhookSecretLength)
		if err != nil {
			return fmt.Errorf("failed to generate GitOps webhook secret: %v", err)
		}
		o.GitOpsWebhookSecret = gitSecret
	}
	gitOpsRepo, err := scm.NewRepository(o.GitOpsRepoURL)
	if err != nil {
		return err
	}

	outputs, err := createInitialFiles(fs, gitOpsRepo, o.Prefix, o.GitOpsWebhookSecret, o.DockerConfigJSONFilename, o.SealedSecretsNamespace)
	if err != nil {
		return err
	}
	_, err = yaml.WriteResources(fs, o.OutputPath, outputs)
	return err
}

func createInitialFiles(fs afero.Fs, repo scm.Repository, prefix, gitOpsWebhookSecret, dockerConfigPath, sealedSecretsNS string) (res.Resources, error) {
	cicd := &config.PipelinesConfig{Name: prefix + "cicd"}
	pipelineConfig := &config.Config{Pipelines: cicd}
	pipelines := createManifest(repo.URL(), pipelineConfig)
	initialFiles := res.Resources{
		pipelinesFile: pipelines,
	}
	resources, err := createCICDResources(fs, repo, cicd, gitOpsWebhookSecret, dockerConfigPath, sealedSecretsNS)
	if err != nil {
		return nil, err
	}

	files := getResourceFiles(resources)
	prefixedResources := addPrefixToResources(pipelinesPath(pipelines.Config), resources)
	initialFiles = res.Merge(prefixedResources, initialFiles)

	pipelinesConfigKustomizations := addPrefixToResources(
		config.PathForPipelines(pipelines.Config.Pipelines),
		getCICDKustomization(files))
	initialFiles = res.Merge(pipelinesConfigKustomizations, initialFiles)

	return initialFiles, nil
}

// createDockerSecret creates a secret that allows pushing images to upstream
// repositories.
func createDockerSecret(fs afero.Fs, dockerConfigJSONFilename, secretNS, sealedSecretsNS string) (*ssv1alpha1.SealedSecret, error) {
	if dockerConfigJSONFilename == "" {
		return nil, errors.New("failed to generate path to file: --dockerconfigjson flag is not provided")
	}
	authJSONPath, err := homedir.Expand(dockerConfigJSONFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to generate path to file: %v", err)
	}
	f, err := fs.Open(authJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Docker config %#v : %s", authJSONPath, err)
	}
	defer f.Close()

	dockerSecret, err := secrets.CreateSealedDockerConfigSecret(meta.NamespacedName(secretNS, dockerSecretName), f, sealedSecretsNS)
	if err != nil {
		return nil, err
	}

	return dockerSecret, nil
}

// createCICDResources creates resources assocated to pipelines.
func createCICDResources(fs afero.Fs, repo scm.Repository, pipelineConfig *config.PipelinesConfig, gitOpsWebhookSecret, dockerConfigJSONPath, sealedSecretsNS string) (res.Resources, error) {
	cicdNamespace := pipelineConfig.Name
	// key: path of the resource
	// value: YAML content of the resource
	outputs := map[string]interface{}{}
	githubSecret, err := secrets.CreateSealedSecret(meta.NamespacedName(cicdNamespace, eventlisteners.GitOpsWebhookSecret),
		gitOpsWebhookSecret, eventlisteners.WebhookSecretKey, sealedSecretsNS)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub Webhook Secret: %v", err)
	}

	outputs[secretsPath] = githubSecret
	outputs[namespacesPath] = namespaces.Create(cicdNamespace)
	outputs[rolesPath] = roles.CreateClusterRole(meta.NamespacedName("", roles.ClusterRoleName), Rules)

	sa := roles.CreateServiceAccount(meta.NamespacedName(cicdNamespace, saName))

	if dockerConfigJSONPath != "" {
		dockerSecret, err := createDockerSecret(fs, dockerConfigJSONPath, cicdNamespace, sealedSecretsNS)
		if err != nil {
			return nil, err
		}
		outputs[dockerConfigPath] = dockerSecret

		// add secret and sa to outputs
		outputs[serviceAccountPath] = roles.AddSecretToSA(sa, dockerSecretName)
	}

	outputs[rolebindingsPath] = roles.CreateClusterRoleBinding(meta.NamespacedName("", roleBindingName), sa, "ClusterRole", roles.ClusterRoleName)
	script, err := dryrun.MakeScript("kubectl", cicdNamespace)
	if err != nil {
		return nil, err
	}
	outputs[gitopsTasksPath] = tasks.CreateDeployFromSourceTask(cicdNamespace, script)
	outputs[appTaskPath] = tasks.CreateDeployUsingKubectlTask(cicdNamespace)
	outputs[ciPipelinesPath] = pipelines.CreateCIPipeline(meta.NamespacedName(cicdNamespace, "ci-dryrun-from-pr-pipeline"), cicdNamespace)
	outputs[cdPipelinesPath] = pipelines.CreateCDPipeline(meta.NamespacedName(cicdNamespace, "cd-deploy-from-push-pipeline"), cicdNamespace)
	outputs[appCiPipelinesPath] = pipelines.CreateAppCIPipeline(meta.NamespacedName(cicdNamespace, "app-ci-pipeline"))
	createTriggerBindings(repo, outputs, cicdNamespace)
	outputs[prTemplatePath] = triggers.CreateCIDryRunTemplate(cicdNamespace, saName)
	outputs[pushTemplatePath] = triggers.CreateCDPushTemplate(cicdNamespace, saName)
	outputs[appCIBuildPRTemplatePath] = triggers.CreateDevCIBuildPRTemplate(cicdNamespace, saName)
	outputs[eventListenerPath] = eventlisteners.Generate(repo, cicdNamespace, saName, eventlisteners.GitOpsWebhookSecret)
	route, err := routes.Generate(cicdNamespace)
	if err != nil {
		return nil, err
	}
	outputs[routePath] = route
	return outputs, nil
}

// Trigger bindings for repository types will be created during bootstrap
func createTriggerBindings(r scm.Repository, outputs res.Resources, ns string) {
	prBinding, prBindingName := r.CreatePRBinding(ns)
	outputs[filepath.Join("06-bindings", prBindingName+".yaml")] = prBinding
	pushBinding, pushBindingName := r.CreatePushBinding(ns)
	outputs[filepath.Join("06-bindings", pushBindingName+".yaml")] = pushBinding
}

func createManifest(gitOpsRepoURL string, configEnv *config.Config, envs ...*config.Environment) *config.Manifest {
	return &config.Manifest{
		GitOpsURL:    gitOpsRepoURL,
		Environments: envs,
		Config:       configEnv,
	}
}

func getCICDKustomization(files []string) res.Resources {
	return res.Resources{
		"base/kustomization.yaml": resources.Kustomization{
			Bases: []string{"./pipelines"},
		},
		"overlays/kustomization.yaml": resources.Kustomization{
			Bases: []string{"../base"},
		},
		"base/pipelines/kustomization.yaml": resources.Kustomization{
			Resources: files,
		},
	}
}

func pipelinesPath(m *config.Config) string {
	return filepath.Join(config.PathForPipelines(m.Pipelines), "base/pipelines")
}

func addPrefixToResources(prefix string, files res.Resources) map[string]interface{} {
	updated := map[string]interface{}{}
	for k, v := range files {
		updated[filepath.Join(prefix, k)] = v
	}
	return updated
}

func getResourceFiles(res res.Resources) []string {
	files := []string{}
	for k := range res {
		files = append(files, k)
	}
	sort.Strings(files)
	return files
}
