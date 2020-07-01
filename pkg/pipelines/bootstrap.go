package pipelines

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/deployment"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/imagerepo"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/namespaces"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/roles"
	"github.com/openshift/odo/pkg/pipelines/scm"
	"github.com/openshift/odo/pkg/pipelines/secrets"
	"github.com/openshift/odo/pkg/pipelines/yaml"
)

const (
	pipelinesFile     = "pipelines.yaml"
	bootstrapImage    = "nginxinc/nginx-unprivileged:latest"
	appCITemplateName = "app-ci-template"
)

// BootstrapOptions is a struct that provides the optional flags
type BootstrapOptions struct {
	GitOpsRepoURL            string // This is where the pipelines and configuration are.
	GitOpsWebhookSecret      string // This is the secret for authenticating hooks from your GitOps repo.
	ServiceRepoURL           string // This is the full URL to your GitHub repository for your service source.
	ServiceWebhookSecret     string // This is the secret for authenticating hooks from your service source.
	InternalRegistryHostname string // This is the internal registry hostname used for pushing images.
	ImageRepo                string // This is where built images are pushed to.
	Prefix                   string // Used to prefix generated environment names in a shared cluster.
	OutputPath               string // Where to write the bootstrapped files to?
	DockerConfigJSONFilename string
}

// Bootstrap bootstraps a GitOps pipelines and repository structure.
func Bootstrap(o *BootstrapOptions, appFs afero.Fs) error {
	if o.GitOpsWebhookSecret == "" {
		gitopsSecret, err := secrets.GenerateString(webhookSecretLength)
		if err != nil {
			return fmt.Errorf("failed to generate GitOps webhook secret: %v", err)
		}
		o.GitOpsWebhookSecret = gitopsSecret
	}
	if o.ServiceWebhookSecret == "" {
		appSecret, err := secrets.GenerateString(webhookSecretLength)
		if err != nil {
			return fmt.Errorf("failed to generate application webhook secret: %v", err)
		}
		o.ServiceWebhookSecret = appSecret
	}
	bootstrapped, err := bootstrapResources(o, appFs)
	if err != nil {
		return fmt.Errorf("failed to bootstrap resources: %v", err)
	}

	buildParams := &BuildParameters{
		PipelinesFilePath: pipelinesFile,
		OutputPath:        o.OutputPath,
	}

	m := bootstrapped[pipelinesFile].(*config.Manifest)
	built, err := buildResources(appFs, buildParams, m)
	if err != nil {
		return fmt.Errorf("failed to build resources: %v", err)
	}
	bootstrapped = res.Merge(built, bootstrapped)
	_, err = yaml.WriteResources(appFs, o.OutputPath, bootstrapped)
	return err
}

func bootstrapResources(o *BootstrapOptions, appFs afero.Fs) (res.Resources, error) {
	isInternalRegistry, imageRepo, err := imagerepo.ValidateImageRepo(o.ImageRepo, o.InternalRegistryHostname)
	if err != nil {
		return nil, err
	}
	gitOpsRepo, err := scm.NewRepository(o.GitOpsRepoURL)
	if err != nil {
		return nil, err
	}
	appRepo, err := scm.NewRepository(o.ServiceRepoURL)
	if err != nil {
		return nil, err
	}
	repoName, err := repoFromURL(appRepo.URL())
	if err != nil {
		return nil, fmt.Errorf("invalid app repo URL: %v", err)
	}

	bootstrapped, err := createInitialFiles(appFs, gitOpsRepo, o.Prefix, o.GitOpsWebhookSecret, o.DockerConfigJSONFilename)
	if err != nil {
		return nil, err
	}

	ns := namespaces.NamesWithPrefix(o.Prefix)
	appName := repoToAppName(repoName)
	serviceName := repoName
	secretName := secrets.MakeServiceWebhookSecretName(ns["dev"], serviceName)
	envs, configEnv, err := bootstrapEnvironments(appRepo, o.Prefix, secretName, ns)
	if err != nil {
		return nil, err
	}
	m := createManifest(gitOpsRepo.URL(), configEnv, envs...)

	devEnv := m.GetEnvironment(ns["dev"])
	if devEnv == nil {
		return nil, errors.New("unable to bootstrap without dev environment")
	}
	svcFiles, err := bootstrapServiceDeployment(devEnv, appName)
	if err != nil {
		return nil, err
	}
	hookSecret, err := secrets.CreateSealedSecret(
		meta.NamespacedName(ns["cicd"], secretName),
		o.ServiceWebhookSecret,
		eventlisteners.WebhookSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub Webhook Secret: %v", err)
	}
	cfg := m.GetPipelinesConfig()
	if cfg == nil {
		return nil, errors.New("failed to find a pipeline configuration - unable to continue bootstrap")
	}
	secretFilename := filepath.Join("03-secrets", secretName+".yaml")
	secretsPath := filepath.Join(config.PathForPipelines(cfg), "base", "pipelines", secretFilename)
	bootstrapped[secretsPath] = hookSecret

	bindingName, imageRepoBindingFilename, svcImageBinding := createSvcImageBinding(cfg, devEnv, serviceName, imageRepo, !isInternalRegistry)
	bootstrapped = res.Merge(svcImageBinding, bootstrapped)

	kustomizePath := filepath.Join(config.PathForPipelines(cfg), "base", "pipelines", "kustomization.yaml")
	k, ok := bootstrapped[kustomizePath].(res.Kustomization)
	if !ok {
		return nil, fmt.Errorf("no kustomization for the %s environment found", kustomizePath)
	}
	if isInternalRegistry {
		filenames, resources, err := imagerepo.CreateInternalRegistryResources(cfg, roles.CreateServiceAccount(meta.NamespacedName(cfg.Name, saName)), imageRepo)
		if err != nil {
			return nil, fmt.Errorf("failed to get resources for internal image repository: %v", err)
		}
		bootstrapped = res.Merge(resources, bootstrapped)
		k.Resources = append(k.Resources, filenames...)
	}

	// This is specific to bootstrap, because there's only one service.
	devEnv.Services[0].Pipelines = &config.Pipelines{
		Integration: &config.TemplateBinding{
			Bindings: append([]string{bindingName}, devEnv.Pipelines.Integration.Bindings[:]...),
		},
	}
	bootstrapped[pipelinesFile] = m

	k.Resources = append(k.Resources, secretFilename, imageRepoBindingFilename)
	sort.Strings(k.Resources)
	bootstrapped[kustomizePath] = k

	bootstrapped = res.Merge(svcFiles, bootstrapped)
	return bootstrapped, nil
}

func bootstrapServiceDeployment(dev *config.Environment, appName string) (res.Resources, error) {
	svc := dev.Services[0]
	svcBase := filepath.Join(config.PathForService(dev, svc.Name), "base", "config")
	resources := res.Resources{}
	// TODO: This should change if we add Namespace to Environment.
	// We'd need to create the resources in the namespace _of_ the Environment.
	resources[filepath.Join(svcBase, "100-deployment.yaml")] = deployment.Create(appName, dev.Name, svc.Name, bootstrapImage, deployment.ContainerPort(8080))
	resources[filepath.Join(svcBase, "200-service.yaml")] = createBootstrapService(appName, dev.Name, svc.Name)
	resources[filepath.Join(svcBase, "kustomization.yaml")] = &res.Kustomization{Resources: []string{"100-deployment.yaml", "200-service.yaml"}}
	return resources, nil
}

func bootstrapEnvironments(repo scm.Repository, prefix, secretName string, ns map[string]string) ([]*config.Environment, *config.Config, error) {
	envs := []*config.Environment{}
	var pipelinesConfig *config.PipelinesConfig
	for k, v := range ns {
		if k == "cicd" {
			pipelinesConfig = &config.PipelinesConfig{Name: prefix + "cicd"}
		} else {
			env := &config.Environment{Name: v}
			if k == "dev" {
				svc, err := serviceFromRepo(repo.URL(), secretName, ns["cicd"])
				if err != nil {
					return nil, nil, err
				}
				app, err := applicationFromRepo(repo.URL(), svc.Name)
				if err != nil {
					return nil, nil, err
				}
				env.Apps = []*config.Application{app}
				env.Services = []*config.Service{svc}
				env.Pipelines = defaultPipelines(repo)
			}
			envs = append(envs, env)
		}
	}
	cfg := &config.Config{Pipelines: pipelinesConfig, ArgoCD: &config.ArgoCDConfig{Namespace: "argocd"}}
	return envs, cfg, nil
}

func serviceFromRepo(repoURL, secretName, secretNS string) (*config.Service, error) {
	repo, err := repoFromURL(repoURL)
	if err != nil {
		return nil, err
	}
	return &config.Service{
		Name:      repo,
		SourceURL: repoURL,
		Webhook: &config.Webhook{
			Secret: &config.Secret{
				Name:      secretName,
				Namespace: secretNS,
			},
		},
	}, nil
}

func applicationFromRepo(repoURL, serviceName string) (*config.Application, error) {
	repo, err := repoFromURL(repoURL)
	if err != nil {
		return nil, err
	}
	return &config.Application{
		Name:        repoToAppName(repo),
		ServiceRefs: []string{serviceName},
	}, nil
}

func repoFromURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Path, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".git"), nil
}

func orgRepoFromURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Path, "/")
	orgRepo := strings.Join(parts[len(parts)-2:], "/")
	return strings.TrimSuffix(orgRepo, ".git"), nil
}

func createBootstrapService(appName, ns, name string) *corev1.Service {
	svc := &corev1.Service{
		TypeMeta:   meta.TypeMeta("Service", "v1"),
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, name)),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080)},
			},
		},
	}
	labels := map[string]string{
		deployment.KubernetesAppNameLabel: name,
		deployment.KubernetesPartOfLabel:  appName,
	}
	svc.ObjectMeta.Labels = labels
	svc.Spec.Selector = labels
	return svc
}

func repoToAppName(repoName string) string {
	return "app-" + repoName
}

func defaultPipelines(r scm.Repository) *config.Pipelines {
	return &config.Pipelines{
		Integration: &config.TemplateBinding{
			Template: appCITemplateName,
			Bindings: []string{r.PRBindingName()},
		},
	}
}
