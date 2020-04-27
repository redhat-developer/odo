package manifest

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/openshift/odo/pkg/manifest/config"
	"github.com/openshift/odo/pkg/manifest/deployment"
	"github.com/openshift/odo/pkg/manifest/eventlisteners"
	"github.com/openshift/odo/pkg/manifest/meta"
	res "github.com/openshift/odo/pkg/manifest/resources"
	"github.com/openshift/odo/pkg/manifest/secrets"
	"github.com/openshift/odo/pkg/manifest/yaml"
)

const manifestFile = "manifest.yaml"
const bootstrapImage = "nginxinc/nginx-unprivileged:latest"

// BootstrapOptions is a struct that provides the optional flags
type BootstrapOptions struct {
	GitOpsRepoURL            string // This is where the manifest and configuration are.
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

// Bootstrap bootstraps a GitOps manifest and repository structure.
func Bootstrap(o *BootstrapOptions) error {
	bootstrapped, err := bootstrapResources(o)
	if err != nil {
		return fmt.Errorf("failed to bootstrap resources: %w", err)
	}
	appFs := afero.NewOsFs()

	buildParams := &BuildParameters{
		ManifestFilename: manifestFile,
		OutputPath:       o.OutputPath,
		RepositoryURL:    o.GitOpsRepoURL,
	}

	m := bootstrapped[manifestFile].(*config.Manifest)
	built, err := buildResources(appFs, buildParams, m)
	if err != nil {
		return fmt.Errorf("failed to build resources: %w", err)
	}
	bootstrapped = res.Merge(built, bootstrapped)
	_, err = yaml.WriteResources(appFs, o.OutputPath, bootstrapped)
	return err
}

func bootstrapResources(o *BootstrapOptions) (res.Resources, error) {
	orgRepo, err := orgRepoFromURL(o.GitOpsRepoURL)
	if err != nil {
		return nil, err
	}
	repoName, err := repoFromURL(o.AppRepoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid app repo URL: %w", err)
	}

	bootstrapped, err := createInitialFiles(o.Prefix, orgRepo, o.GitOpsWebhookSecret, o.DockerConfigJSONFilename, "")
	if err != nil {
		return nil, err
	}
	ns := NamespaceNames(o.Prefix)
	secretName := "github-webhook-secret-" + repoName + "-svc"
	envs, err := bootstrapEnvironments(o.Prefix, o.AppRepoURL, secretName, ns)
	if err != nil {
		return nil, err
	}
	m := createManifest(envs...)
	bootstrapped[manifestFile] = m
	env, err := m.GetEnvironment(ns["dev"])
	if err != nil {
		return nil, err
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
	secretsPath := filepath.Join(config.PathForEnvironment(cicdEnv), "base", "pipelines", "03-secrets", secretName+".yaml")
	bootstrapped[secretsPath] = hookSecret
	bootstrapped = res.Merge(svcFiles, bootstrapped)
	return bootstrapped, nil
}

func bootstrapServiceDeployment(dev *config.Environment) (res.Resources, error) {
	svc := dev.Apps[0].Services[0]
	svcBase := filepath.Join(config.PathForService(dev, svc), "base", "config")
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
			app, err := applicationFromRepo(repoURL, secretName, ns["cicd"])
			if err != nil {
				return nil, err
			}
			env.Apps = []*config.Application{app}
			env.Pipelines = defaultPipelines
		}
		envs = append(envs, env)
	}
	envs = append(envs, &config.Environment{Name: prefix + "argocd", IsArgoCD: true})
	sort.Sort(config.ByName(envs))
	return envs, nil
}

func applicationFromRepo(repoURL, secretName, secretNS string) (*config.Application, error) {
	repo, err := repoFromURL(repoURL)
	if err != nil {
		return nil, err
	}
	return &config.Application{
		Name: repo,
		Services: []*config.Service{
			{
				Name:      repo + "-svc",
				SourceURL: repoURL,
				Webhook: &config.Webhook{
					Secret: &config.Secret{
						Name:      secretName,
						Namespace: secretNS,
					},
				},
			},
		},
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
