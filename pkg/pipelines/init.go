package pipelines

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/namespaces"
	"github.com/openshift/odo/pkg/pipelines/pipelines"
	"github.com/openshift/odo/pkg/pipelines/resources"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/roles"
	"github.com/openshift/odo/pkg/pipelines/routes"
	"github.com/openshift/odo/pkg/pipelines/secrets"
	"github.com/openshift/odo/pkg/pipelines/tasks"
	"github.com/openshift/odo/pkg/pipelines/triggers"
	"github.com/openshift/odo/pkg/pipelines/yaml"
	"github.com/spf13/afero"

	v1rbac "k8s.io/api/rbac/v1"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
)

// InitParameters is a struct that provides flags for the Init command.
type InitParameters struct {
	DockerConfigJSONFilename string
	GitOpsRepoURL            string
	GitOpsWebhookSecret      string
	ImageRepo                string
	InternalRegistryHostname string
	OutputPath               string
	Prefix                   string
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
			Resources: []string{"clusterrolebindings"},
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

	baseDir = "base"

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
	appCdPipelinesPath       = "05-pipelines/app-cd-pipeline.yaml"
	cdPipelinesPath          = "05-pipelines/cd-deploy-from-push-pipeline.yaml"
	prBindingPath            = "06-bindings/github-pr-binding.yaml"
	pushBindingPath          = "06-bindings/github-push-binding.yaml"
	prTemplatePath           = "07-templates/ci-dryrun-from-pr-template.yaml"
	pushTemplatePath         = "07-templates/cd-deploy-from-push-template.yaml"
	appCIBuildPRTemplatePath = "07-templates/app-ci-build-pr-template.yaml"
	appCDBuildPRTemplatePath = "07-templates/app-cd-build-pr-template.yaml"
	eventListenerPath        = "08-eventlisteners/cicd-event-listener.yaml"
	routePath                = "09-routes/gitops-webhook-event-listener.yaml"

	dockerSecretName = "regcred"

	saName          = "pipeline"
	roleName        = "pipelines-service-role"
	roleBindingName = "pipelines-service-role-binding"
)

// Init bootstraps a GitOps pipelines and repository structure.
func Init(o *InitParameters, fs afero.Fs) error {
	_, imageRepo, err := validateImageRepo(o.ImageRepo, o.InternalRegistryHostname)
	if err != nil {
		return err
	}

	exists, err := ioutils.IsExisting(fs, o.OutputPath)
	if exists {
		return err
	}

	outputs, err := createInitialFiles(fs, o.GitOpsRepoURL, o.Prefix, o.GitOpsWebhookSecret, o.DockerConfigJSONFilename, imageRepo)
	if err != nil {
		return err
	}
	_, err = yaml.WriteResources(fs, o.OutputPath, outputs)
	return err
}

// createCICDResources creates resources assocated to pipelines.
func createCICDResources(fs afero.Fs, cicdEnv *config.Environment, gitOpsRepo, gitOpsWebhookSecret, dockerConfigJSONPath, imageRepo string) (res.Resources, error) {
	cicdNamespace := cicdEnv.Name
	// key: path of the resource
	// value: YAML content of the resource
	outputs := map[string]interface{}{}
	githubSecret, err := secrets.CreateSealedSecret(meta.NamespacedName(cicdNamespace, eventlisteners.GitOpsWebhookSecret),
		gitOpsWebhookSecret, eventlisteners.WebhookSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub Webhook Secret: %w", err)
	}

	outputs[secretsPath] = githubSecret
	outputs[namespacesPath] = namespaces.Create(cicdNamespace)
	outputs[rolesPath] = roles.CreateClusterRole(meta.NamespacedName("", roles.ClusterRoleName), Rules)

	sa := roles.CreateServiceAccount(meta.NamespacedName(cicdNamespace, saName))

	if dockerConfigJSONPath != "" {
		dockerSecret, err := CreateDockerSecret(fs, dockerConfigJSONPath, cicdNamespace)
		if err != nil {
			return nil, err
		}
		outputs[dockerConfigPath] = dockerSecret

		// add secret and sa to outputs
		outputs[serviceAccountPath] = roles.AddSecretToSA(sa, dockerSecretName)
	}

	outputs[rolebindingsPath] = roles.CreateClusterRoleBinding(meta.NamespacedName("", roleBindingName), sa, "ClusterRole", roles.ClusterRoleName)
	outputs[gitopsTasksPath] = tasks.CreateDeployFromSourceTask(cicdNamespace, filepath.Join(config.PathForEnvironment(cicdEnv), "base"))
	outputs[appTaskPath] = tasks.CreateDeployUsingKubectlTask(cicdNamespace)
	outputs[ciPipelinesPath] = pipelines.CreateCIPipeline(meta.NamespacedName(cicdNamespace, "ci-dryrun-from-pr-pipeline"), cicdNamespace)
	outputs[cdPipelinesPath] = pipelines.CreateCDPipeline(meta.NamespacedName(cicdNamespace, "cd-deploy-from-push-pipeline"), cicdNamespace)
	outputs[appCiPipelinesPath] = pipelines.CreateAppCIPipeline(meta.NamespacedName(cicdNamespace, "app-ci-pipeline"), false)
	outputs[prBindingPath] = triggers.CreatePRBinding(cicdNamespace)
	outputs[pushBindingPath] = triggers.CreatePushBinding(cicdNamespace)
	outputs[prTemplatePath] = triggers.CreateCIDryRunTemplate(cicdNamespace, saName)
	outputs[pushTemplatePath] = triggers.CreateCDPushTemplate(cicdNamespace, saName)
	outputs[appCIBuildPRTemplatePath] = triggers.CreateDevCIBuildPRTemplate(cicdNamespace, saName, imageRepo)
	outputs[eventListenerPath] = eventlisteners.Generate(gitOpsRepo, cicdNamespace, saName, eventlisteners.GitOpsWebhookSecret)

	outputs[routePath] = routes.Generate(cicdNamespace)
	return outputs, nil
}

// CreateDockerSecret creates Docker secret
func CreateDockerSecret(fs afero.Fs, dockerConfigJSONFilename, ns string) (*ssv1alpha1.SealedSecret, error) {
	if dockerConfigJSONFilename == "" {
		return nil, errors.New("failed to generate path to file: --dockerconfigjson flag is not provided")
	}

	authJSONPath, err := homedir.Expand(dockerConfigJSONFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to generate path to file: %w", err)
	}
	f, err := fs.Open(authJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read docker file '%s' : %w", authJSONPath, err)
	}
	defer f.Close()

	dockerSecret, err := secrets.CreateSealedDockerConfigSecret(meta.NamespacedName(ns, dockerSecretName), f)
	if err != nil {
		return nil, err
	}

	return dockerSecret, nil

}

func createInitialFiles(fs afero.Fs, prefix, gitOpsURL, gitOpsWebhookSecret, dockerConfigPath, imageRepo string) (res.Resources, error) {
	cicdEnv := &config.Environment{Name: prefix + "cicd", IsCICD: true}
	pipelines := createManifest(gitOpsURL, cicdEnv)
	initialFiles := res.Resources{
		pipelinesFile: pipelines,
	}
	orgRepo, err := orgRepoFromURL(gitOpsURL)
	if err != nil {
		return nil, err
	}

	resources, err := createCICDResources(fs, cicdEnv, orgRepo, gitOpsWebhookSecret, dockerConfigPath, imageRepo)
	if err != nil {
		return nil, err
	}
	files := getResourceFiles(resources)

	prefixedResources := addPrefixToResources(pipelinesPath(pipelines), resources)
	initialFiles = res.Merge(prefixedResources, initialFiles)

	cicdKustomizations := addPrefixToResources(cicdEnvironmentPath(pipelines), getCICDKustomization(files))
	initialFiles = res.Merge(cicdKustomizations, initialFiles)

	return initialFiles, nil
}

func createManifest(gitOpsURL string, envs ...*config.Environment) *config.Manifest {
	return &config.Manifest{
		GitOpsURL:    gitOpsURL,
		Environments: envs,
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

func pathForEnvironment(env *config.Environment) string {
	return filepath.Join("environments", env.Name)
}

func pipelinesPath(m *config.Manifest) string {
	return filepath.Join(cicdEnvironmentPath(m), "base/pipelines")
}

func addPrefixToResources(prefix string, files res.Resources) map[string]interface{} {
	updated := map[string]interface{}{}
	for k, v := range files {
		updated[filepath.Join(prefix, k)] = v
	}
	return updated
}

// TODO: this should probably use the .FindCICDEnvironment on the pipelines.
func cicdEnvironmentPath(m *config.Manifest) string {
	return pathForEnvironment(m.Environments[0])
}

func getResourceFiles(res res.Resources) []string {
	files := []string{}
	for k := range res {
		files = append(files, k)
	}
	sort.Strings(files)
	return files
}

// validateImageRepo validates the input image repo.  It determines if it is
// for internal registry and prepend internal registry hostname if neccessary.
func validateImageRepo(imageRepo, registryURL string) (bool, string, error) {
	components := strings.Split(imageRepo, "/")

	// repo url has minimum of 2 components
	if len(components) < 2 {
		return false, "", imageRepoValidationErrors(imageRepo)
	}

	for _, v := range components {
		// check for empty components
		if strings.TrimSpace(v) == "" {
			return false, "", imageRepoValidationErrors(imageRepo)
		}
		// check for white spaces
		if len(v) > len(strings.TrimSpace(v)) {
			return false, "", imageRepoValidationErrors(imageRepo)
		}
	}

	if len(components) == 2 {
		if components[0] == "docker.io" || components[0] == "quay.io" {
			// we recognize docker.io and quay.io.  It is missing one component
			return false, "", imageRepoValidationErrors(imageRepo)
		}
		// We have format like <project>/<app> which is an internal registry.
		// We prepend the internal registry hostname.
		return true, registryURL + "/" + imageRepo, nil
	}

	// Check the first component to see if it is an internal registry
	if len(components) == 3 {
		return components[0] == registryURL, imageRepo, nil
	}

	// > 3 components.  invalid repo
	return false, "", imageRepoValidationErrors(imageRepo)
}

func imageRepoValidationErrors(imageRepo string) error {
	return fmt.Errorf("failed to parse image repo:%s, expected image repository in the form <registry>/<username>/<repository> or <project>/<app> for internal registry", imageRepo)
}
