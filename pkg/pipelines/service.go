package pipelines

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/environments"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/meta"
	res "github.com/openshift/odo/pkg/pipelines/resources"

	"github.com/openshift/odo/pkg/pipelines/secrets"
	"github.com/openshift/odo/pkg/pipelines/yaml"
	"github.com/spf13/afero"
)

func AddService(gitRepoURL, envName, appName, serviceName, webhookSecret, manifest, imageRepo, internalRegistryHostname string, fs afero.Fs) error {

	m, err := config.ParseFile(fs, manifest)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	cicdEnv, err := m.GetCICDEnvironment()
	if err != nil {
		return err
	}
	outputPath := filepath.Dir(manifest)

	files, err := serviceResources(m, fs, gitRepoURL, envName, appName, serviceName, webhookSecret, manifest)
	if err != nil {
		return err
	}

	_, err = yaml.WriteResources(fs, outputPath, files)
	if err != nil {
		return err
	}
	if cicdEnv != nil {
		base := filepath.Join(outputPath, config.PathForEnvironment(cicdEnv), "base", "pipelines")
		err = updateKustomization(fs, base)
		if err != nil {
			return err
		}
	}
	return nil
}

func serviceResources(m *config.Manifest, fs afero.Fs, gitRepoURL, envName, appName, serviceName, webhookSecret, manifest string) (res.Resources, error) {
	files := res.Resources{}

	svc, err := createService(serviceName, gitRepoURL)
	if err != nil {
		return nil, err
	}

	cicdEnv, err := m.GetCICDEnvironment()
	if err != nil {
		return nil, err
	}
	if cicdEnv != nil && webhookSecret == "" && gitRepoURL != "" {
		return nil, fmt.Errorf("The webhook secret is required")
	}
	// add the secret only if CI/CD env is present
	if cicdEnv != nil {
		secretName := secrets.MakeServiceWebhookSecretName(svc.Name)
		hookSecret, err := secrets.CreateSealedSecret(meta.NamespacedName(cicdEnv.Name, secretName), webhookSecret, eventlisteners.WebhookSecretKey)
		if err != nil {
			return nil, err
		}
		svc.Webhook = &config.Webhook{
			Secret: &config.Secret{
				Name:      secretName,
				Namespace: cicdEnv.Name,
			},
		}
		secretPath := filepath.Join(config.PathForEnvironment(cicdEnv), "base", "pipelines")
		files[filepath.Join(secretPath, "03-secrets", secretName+".yaml")] = hookSecret
	}

	err = m.AddService(envName, appName, svc)
	if err != nil {
		return nil, err
	}
	err = m.Validate()
	if err != nil {
		return nil, err
	}

	files[filepath.Base(manifest)] = m
	outputPath := filepath.Dir(manifest)
	buildParams := &BuildParameters{
		ManifestFilename: manifest,
		OutputPath:       outputPath,
		RepositoryURL:    m.GitOpsURL,
	}
	built, err := buildResources(fs, buildParams, m)
	if err != nil {
		return nil, err
	}
	files = res.Merge(built, files)

	return files, nil

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
	list, err := environments.ListFiles(fs, base)
	if err != nil {
		return err
	}
	files[Kustomize] = &res.Kustomization{Resources: environments.ExtractFilenames(list)}
	_, err = yaml.WriteResources(fs, base, files)
	return err
}
