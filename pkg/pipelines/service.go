package pipelines

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/environments"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/imagerepo"
	"github.com/openshift/odo/pkg/pipelines/meta"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/roles"
	"github.com/openshift/odo/pkg/pipelines/triggers"

	"github.com/openshift/odo/pkg/pipelines/secrets"
	"github.com/openshift/odo/pkg/pipelines/yaml"
	"github.com/spf13/afero"
)

// AddServiceParameters are parameters passed to AddSerice function
type AddServiceParameters struct {
	AppName                  string
	EnvName                  string
	GitRepoURL               string
	ImageRepo                string
	InternalRegistryHostname string
	PipelinesFilePath        string
	ServiceName              string
	WebhookSecret            string
}

func AddService(p *AddServiceParameters, fs afero.Fs) error {
	m, err := config.ParseFile(fs, p.PipelinesFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse pipelines-file: %v", err)
	}

	outputPath := filepath.Dir(p.PipelinesFilePath)

	files, err := serviceResources(m, fs, p)
	if err != nil {
		return err
	}

	_, err = yaml.WriteResources(fs, outputPath, files)
	if err != nil {
		return err
	}
	cfg := m.GetPipelinesConfig()
	if cfg != nil {
		base := filepath.Join(outputPath, config.PathForPipelines(cfg), "base", "pipelines")
		err = updateKustomization(fs, base)
		if err != nil {
			return err
		}
	}
	return nil
}

func serviceResources(m *config.Manifest, fs afero.Fs, p *AddServiceParameters) (res.Resources, error) {
	files := res.Resources{}

	svc, err := createService(p.ServiceName, p.GitRepoURL)
	if err != nil {
		return nil, err
	}

	cfg := m.GetPipelinesConfig()
	if cfg != nil && p.WebhookSecret == "" && p.GitRepoURL != "" {
		return nil, fmt.Errorf("The webhook secret is required")
	}

	env := m.GetEnvironment(p.EnvName)
	if env == nil {
		return nil, fmt.Errorf("environment %s does not exist", p.EnvName)
	}

	// add the secret only if CI/CD env is present
	if cfg != nil {
		secretName := secrets.MakeServiceWebhookSecretName(p.EnvName, svc.Name)
		hookSecret, err := secrets.CreateSealedSecret(meta.NamespacedName(cfg.Name, secretName), p.WebhookSecret, eventlisteners.WebhookSecretKey)
		if err != nil {
			return nil, err
		}

		svc.Webhook = &config.Webhook{
			Secret: &config.Secret{
				Name:      secretName,
				Namespace: cfg.Name,
			},
		}
		secretFilename := filepath.Join("03-secrets", secretName+".yaml")
		secretsPath := filepath.Join(config.PathForPipelines(cfg), "base", "pipelines", secretFilename)
		files[secretsPath] = hookSecret

		if p.ImageRepo != "" {
			_, resources, bindingName, err := createImageRepoResources(m, cfg, env, p)
			if err != nil {
				return nil, err
			}

			files = res.Merge(resources, files)
			svc.Pipelines = &config.Pipelines{
				Integration: &config.TemplateBinding{
					Bindings: append([]string{bindingName}, env.Pipelines.Integration.Bindings[:]...),
				},
			}
		}
	}

	err = m.AddService(p.EnvName, p.AppName, svc)
	if err != nil {
		return nil, err
	}
	err = m.Validate()
	if err != nil {
		return nil, err
	}

	files[filepath.Base(p.PipelinesFilePath)] = m
	outputPath := filepath.Dir(p.PipelinesFilePath)
	buildParams := &BuildParameters{
		PipelinesFilePath: p.PipelinesFilePath,
		OutputPath:        outputPath,
	}
	built, err := buildResources(fs, buildParams, m)
	if err != nil {
		return nil, err
	}
	return res.Merge(built, files), nil
}

func createImageRepoResources(m *config.Manifest, cfg *config.PipelinesConfig, env *config.Environment, p *AddServiceParameters) ([]string, res.Resources, string, error) {
	isInternalRegistry, imageRepo, err := imagerepo.ValidateImageRepo(p.ImageRepo, p.InternalRegistryHostname)
	if err != nil {
		return nil, nil, "", err
	}

	resources := res.Resources{}
	filenames := []string{}

	bindingName, bindingFilename, svcImageBinding := createSvcImageBinding(cfg, env, p.ServiceName, imageRepo, !isInternalRegistry)
	resources = res.Merge(svcImageBinding, resources)
	filenames = append(filenames, bindingFilename)

	if isInternalRegistry {
		files, regRes, err := imagerepo.CreateInternalRegistryResources(cfg, roles.CreateServiceAccount(meta.NamespacedName(cfg.Name, saName)), imageRepo)
		if err != nil {
			return nil, nil, "", fmt.Errorf("failed to get resources for internal image repository: %v", err)
		}
		resources = res.Merge(regRes, resources)
		filenames = append(filenames, files...)
	}

	return filenames, resources, bindingName, nil
}

func createService(serviceName, url string) (*config.Service, error) {
	if url == "" {
		return &config.Service{
			Name: serviceName,
		}, nil
	}
	return &config.Service{
		Name:      serviceName,
		SourceURL: url,
	}, nil
}

func updateKustomization(fs afero.Fs, base string) error {
	files := res.Resources{}
	filenames, err := environments.ListFiles(fs, base)
	if err != nil {
		return err
	}
	files[Kustomize] = &res.Kustomization{Resources: filenames.Items()}
	_, err = yaml.WriteResources(fs, base, files)
	return err
}

func makeSvcImageBindingName(envName, svcName string) string {
	return fmt.Sprintf("%s-%s-binding", envName, svcName)
}

func makeSvcImageBindingFilename(bindingName string) string {
	return filepath.Join("06-bindings", bindingName+".yaml")
}

func makeImageBindingPath(cfg *config.PipelinesConfig, imageRepoBindingFilename string) string {
	return filepath.Join(config.PathForPipelines(cfg), "base", "pipelines", imageRepoBindingFilename)
}

func createSvcImageBinding(cfg *config.PipelinesConfig, env *config.Environment, svcName, imageRepo string, isTLSVerify bool) (string, string, res.Resources) {
	name := makeSvcImageBindingName(env.Name, svcName)
	filename := makeSvcImageBindingFilename(name)
	resourceFilePath := makeImageBindingPath(cfg, filename)
	return name, filename, res.Resources{resourceFilePath: triggers.CreateImageRepoBinding(cfg.Name, name, imageRepo, strconv.FormatBool(isTLSVerify))}
}
