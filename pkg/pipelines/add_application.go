package pipelines

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/manifest"
	"github.com/openshift/odo/pkg/manifest/eventlisteners"
	"github.com/openshift/odo/pkg/manifest/ioutils"
	"github.com/openshift/odo/pkg/manifest/pipelines"
	"github.com/openshift/odo/pkg/manifest/yaml"
	"github.com/spf13/afero"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/types"

	"github.com/openshift/odo/pkg/manifest/meta"
	"github.com/openshift/odo/pkg/manifest/secrets"
)

// AddParameters is a struct that provides flags for add application command
type AddParameters struct {
	AppName              string
	EnvName              string
	Output               string
	Prefix               string
	ServiceWebhookSecret string
	ServiceGitRepo       string
	SkipChecks           bool
}

const (
	appDir           = "applications"
	appWebhookSecret = "app-webhook-secret"
	configDir        = "config"
	configSApath     = "base/config/serviceaccount.yaml"
	overlaysDir      = "overlays"

	// PatchPath path to eventlistener patch yaml
	PatchPath         = "overlays/eventlistener_patch.yaml"
	pipelinePatchPath = "overlays/pipeline_patch.yaml"

	servicesDir      = "services"
	secretPath       = "base/config/secret.yaml"
	webhookPath      = "base/config/app-webhook-secret.yaml"
	rolebindingPath  = "base/config/edit-rolebinding.yaml"
	kustomizeModPath = "base/config/kustomization.yaml"
	secretName       = "secret"
)

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.

type patchStringValue struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// CreateApplication creates an application
func CreateApplication(o *AddParameters) error {

	if !o.SkipChecks {
		installed, err := pipelines.CheckTektonInstall()
		if err != nil {
			return fmt.Errorf("failed to run Tekton Pipelines installation check: %w", err)
		}
		if !installed {
			return errors.New("failed due to Tekton Pipelines or Triggers are not installed")
		}
	}

	ServiceRepo := getGitopsRepoName(o.ServiceGitRepo)

	// we simpily output to the output dir, no gitops repo in the output path
	gitopsPath := o.Output

	outputs := map[string]interface{}{}

	exists, _ := ioutils.IsExisting(gitopsPath)

	if !exists {
		return fmt.Errorf("Output does not exist at %s", gitopsPath)
	}

	// check if the environment exists
	exists, _ = ioutils.IsExisting(filepath.Join(gitopsPath, "environments", o.EnvName))
	if !exists {
		return fmt.Errorf("Environment %s doesn't exist at %s", o.EnvName, gitopsPath)
	}

	configPath := filepath.Join(gitopsPath, servicesDir, ServiceRepo)

	CreatePatchKustomiseFile(outputs, filepath.Join(overlaysDir, manifest.Kustomize))

	environmentName := manifest.NamespaceNames(o.Prefix)

	createKustomizeMod(outputs, []string{fmt.Sprintf("../../../../environments/%s/overlays", environmentName["cicd"])}, []string{"app-webhook-secret.yaml"}, kustomizeModPath)
	kustomizeModPath := filepath.Join(gitopsPath, fmt.Sprintf("/environments/%s/base/kustomization.yaml", o.EnvName))

	fs := afero.Afero{Fs: afero.OsFs{}}
	createKustomizeEnv(fs, []string{fmt.Sprintf("../../../%s/%s/overlays", appDir, o.AppName)}, []string{"namespace.yaml", "rolebinding.yaml"}, kustomizeModPath)

	secretName := fmt.Sprintf("svc-%s-secret", ServiceRepo)
	files := createResourcesConfig(outputs, o.ServiceWebhookSecret, o.EnvName, secretName)

	createPatchFiles(outputs, o.ServiceGitRepo, o.EnvName, secretName)
	createPipelinePatch(outputs, o.EnvName)

	_, err := yaml.WriteResources(fs, configPath, files)

	if err != nil {
		return err
	}
	if err := yaml.AddKustomize(fs, "bases", []string{"overlays"}, filepath.Join(gitopsPath, appDir, o.AppName, manifest.Kustomize)); err != nil {
		return err
	}

	if err := yaml.AddKustomize(fs, "bases", []string{"../base"}, filepath.Join(gitopsPath, appDir, o.AppName, overlaysDir, manifest.Kustomize)); err != nil {
		return err
	}

	if err := yaml.AddKustomize(fs, "bases", []string{fmt.Sprintf("../../../services/%s/overlays", ServiceRepo)}, filepath.Join(gitopsPath, appDir, o.AppName, manifest.BaseDir, manifest.Kustomize)); err != nil {
		return err
	}
	if err := yaml.AddKustomize(fs, "bases", []string{"../config"}, filepath.Join(gitopsPath, servicesDir, ServiceRepo, manifest.BaseDir, manifest.Kustomize)); err != nil {
		return err
	}
	if err := yaml.AddKustomize(fs, "bases", []string{"./config"}, filepath.Join(gitopsPath, servicesDir, ServiceRepo, manifest.BaseDir, manifest.Kustomize)); err != nil {
		return err
	}

	return nil
}

func createResourcesConfig(outputs map[string]interface{}, serviceWebhookSecret, environmentName, secretName string) map[string]interface{} {

	githubSecret, _ := secrets.CreateSealedSecret(meta.NamespacedName(environmentName, secretName),
		serviceWebhookSecret, eventlisteners.WebhookSecretKey)
	outputs[webhookPath] = githubSecret

	return outputs
}

func createPatchFiles(outputs map[string]interface{}, serviceRepo, ns, secretName string) {
	t := []patchStringValue{
		{
			Op:    "add",
			Path:  "/spec/triggers/-",
			Value: eventlisteners.CreateListenerTrigger("app-ci-build-from-pr", eventlisteners.StageCIDryRunFilters, serviceRepo, "github-pr-binding", "app-ci-template", secretName, ns),
		},
		{
			Op:    "add",
			Path:  "/spec/triggers/-",
			Value: eventlisteners.CreateListenerTrigger("app-cd-deploy-from-master", eventlisteners.StageCDDeployFilters, serviceRepo, "github-push-binding", "app-cd-template", secretName, ns),
		},
	}
	outputs[PatchPath] = t

}

func createPipelinePatch(outputs map[string]interface{}, ns string) {
	outputs[pipelinePatchPath] = []patchStringValue{
		{
			Op:    "replace",
			Path:  "/spec/tasks/1/params/2/value",
			Value: ns,
		},
	}
}

// CreatePatchKustomiseFile creates patch kustomization file
func CreatePatchKustomiseFile(outputs map[string]interface{}, path string) {

	bases := []string{"../base"}

	GVK := gvk.Gvk{
		Group:   "tekton.dev",
		Version: "v1alpha1",
		Kind:    "EventListener",
	}
	target := &types.PatchTarget{
		Gvk:  GVK,
		Name: "cicd-event-listener",
	}
	Patches := []types.PatchJson6902{
		{
			Target: target,
			Path:   "eventlistener_patch.yaml",
		},
		{
			Target: pipelineTarget(),
			Path:   "pipeline_patch.yaml",
		},
	}
	file := types.Kustomization{
		Bases:           bases,
		PatchesJson6902: Patches,
	}
	outputs[path] = file

}

func pipelineTarget() *types.PatchTarget {
	return &types.PatchTarget{
		Gvk: gvk.Gvk{
			Group:   "tekton.dev",
			Version: "v1alpha1",
			Kind:    "Pipeline",
		},
		Name: "app-cd-pipeline",
	}
}

func createKustomizeMod(outputs map[string]interface{}, basesParams, resourcesParams []string, path string) {
	bases := basesParams
	resources := resourcesParams

	file := types.Kustomization{
		Resources: resources,
		Bases:     bases,
	}
	outputs[path] = file
}

func createKustomizeEnv(fs afero.Fs, basesParams, resourcesParams []string, path string) {
	bases := basesParams
	resources := resourcesParams
	file := types.Kustomization{
		Resources: resources,
		Bases:     bases,
	}

	yaml.MarshalItemToFile(fs, path, file)
}

func getGitopsRepoName(repo string) string {
	return strings.Split(repo, "/")[1]
}
