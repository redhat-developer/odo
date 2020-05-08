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
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/namespaces"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/secrets"
	"github.com/openshift/odo/pkg/pipelines/yaml"
)

const pipelinesFile = "pipelines.yaml"
const bootstrapImage = "nginxinc/nginx-unprivileged:latest"

// BootstrapOptions is a struct that provides the optional flags
type BootstrapOptions struct {
	GitOpsRepoURL            string // This is where the pipelines and configuration are.
	GitOpsWebhookSecret      string // This is the secret for authenticating hooks from your GitOps repo.
	AppRepoURL               string // This is the full URL to your GitHub repository for your app source.
	AppWebhookSecret         string // This is the secret for authenticating hooks from your app source.
	InternalRegistryHostname string // This is the internal registry hostname used for pushing images.
	ImageRepo                string // This is where built images are pushed to.
	Prefix                   string // Used to prefix generated environment names in a shared cluster.
	OutputPath               string // Where to write the bootstrapped files to?
	DockerConfigJSONFilename string
}

var defaultPipelines = &config.Pipelines{
	Integration: &config.TemplateBinding{
		Template: "app-ci-template",
		Binding:  "github-pr-binding",
	},
}

// Bootstrap bootstraps a GitOps pipelines and repository structure.
func Bootstrap(o *BootstrapOptions, appFs afero.Fs) error {
	bootstrapped, err := bootstrapResources(o, appFs)
	if err != nil {
		return fmt.Errorf("failed to bootstrap resources: %w", err)
	}

	buildParams := &BuildParameters{
		ManifestFilename: pipelinesFile,
		OutputPath:       o.OutputPath,
		RepositoryURL:    o.GitOpsRepoURL,
	}

	m := bootstrapped[pipelinesFile].(*config.Manifest)
	built, err := buildResources(appFs, buildParams, m)
	if err != nil {
		return fmt.Errorf("failed to build resources: %w", err)
	}
	bootstrapped = res.Merge(built, bootstrapped)
	_, err = yaml.WriteResources(appFs, o.OutputPath, bootstrapped)
	return err
}

func bootstrapResources(o *BootstrapOptions, appFs afero.Fs) (res.Resources, error) {
	repoName, err := repoFromURL(o.AppRepoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid app repo URL: %w", err)
	}

	bootstrapped, err := createInitialFiles(appFs, o.Prefix, o.GitOpsRepoURL, o.GitOpsWebhookSecret, o.DockerConfigJSONFilename, o.ImageRepo)
	if err != nil {
		return nil, err
	}
	ns := namespaces.NamesWithPrefix(o.Prefix)
	secretName := secrets.MakeServiceWebhookSecretName(repoName)
	envs, err := bootstrapEnvironments(o.Prefix, o.AppRepoURL, secretName, ns)
	if err != nil {
		return nil, err
	}
	m := createManifest(o.GitOpsRepoURL, envs...)
	bootstrapped[pipelinesFile] = m
	env := m.GetEnvironment(ns["dev"])
	if env == nil {
		return nil, errors.New("unable to bootstrap without dev environment")
	}
	svcFiles, err := bootstrapServiceDeployment(env)
	if err != nil {
		return nil, err
	}
	hookSecret, err := secrets.CreateSealedSecret(
		meta.NamespacedName(ns["cicd"], secretName),
		o.AppWebhookSecret,
		eventlisteners.WebhookSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub Webhook Secret: %w", err)
	}
	cicdEnv, err := m.GetCICDEnvironment()
	if err != nil {
		return nil, fmt.Errorf("bootstrap environments: %w", err)
	}
	secretFilename := filepath.Join("03-secrets", secretName+".yaml")
	secretsPath := filepath.Join(config.PathForEnvironment(cicdEnv), "base", "pipelines", secretFilename)
	bootstrapped[secretsPath] = hookSecret
	kustomizePath := filepath.Join(config.PathForEnvironment(cicdEnv), "base", "pipelines", "kustomization.yaml")
	k, ok := bootstrapped[kustomizePath].(res.Kustomization)
	if !ok {
		return nil, fmt.Errorf("no kustomization for the %s environment found", kustomizePath)
	}
	k.Resources = append(k.Resources, secretFilename)
	bootstrapped[kustomizePath] = k
	bootstrapped = res.Merge(svcFiles, bootstrapped)
	return bootstrapped, nil
}

func bootstrapServiceDeployment(dev *config.Environment) (res.Resources, error) {
	svc := dev.Services[0]
	svcBase := filepath.Join(config.PathForService(dev, svc.Name), "base", "config")
	resources := res.Resources{}
	// TODO: This should change if we add Namespace to Environment.
	resources[filepath.Join(svcBase, "100-deployment.yaml")] = deployment.Create(dev.Name, svc.Name, bootstrapImage, deployment.ContainerPort(8080))
	resources[filepath.Join(svcBase, "200-service.yaml")] = createBootstrapService(dev.Name, svc.Name)
	resources[filepath.Join(svcBase, "kustomization.yaml")] = &res.Kustomization{Resources: []string{"100-deployment.yaml", "200-service.yaml"}}
	return resources, nil
}

func bootstrapEnvironments(prefix, repoURL, secretName string, ns map[string]string) ([]*config.Environment, error) {
	envs := []*config.Environment{}
	for k, v := range ns {
		env := &config.Environment{Name: v}
		if k == "cicd" {
			env.IsCICD = true
		}
		if k == "dev" {
			svc, err := serviceFromRepo(repoURL, secretName, ns["cicd"])
			if err != nil {
				return nil, err
			}
			app, err := applicationFromRepo(repoURL, svc.Name)
			if err != nil {
				return nil, err
			}
			env.Apps = []*config.Application{app}
			env.Services = []*config.Service{svc}
			env.Pipelines = defaultPipelines
		}
		envs = append(envs, env)
	}
	envs = append(envs, &config.Environment{Name: prefix + "argocd", IsArgoCD: true})
	sort.Sort(config.ByName(envs))
	return envs, nil
}

func serviceFromRepo(repoURL, secretName, secretNS string) (*config.Service, error) {
	repo, err := repoFromURL(repoURL)
	if err != nil {
		return nil, err
	}
	return &config.Service{
		Name:      repo + "-svc",
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
		Name:        repo,
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

func createBootstrapService(ns, name string) *corev1.Service {
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
	}
	svc.ObjectMeta.Labels = labels
	svc.Spec.Selector = labels
	return svc
}
