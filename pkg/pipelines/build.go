package pipelines

import (
	"fmt"

	"github.com/openshift/odo/pkg/pipelines/argocd"
	"github.com/openshift/odo/pkg/pipelines/config"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/yaml"
	"github.com/spf13/afero"
)

// BuildParameters is a struct that provides flags for the BuildResources
// command.
type BuildParameters struct {
	ManifestFilename string
	OutputPath       string
	RepositoryURL    string
}

// BuildResources builds all resources from a pipelines.
func BuildResources(o *BuildParameters, appFs afero.Fs) error {
	m, err := config.ParseFile(appFs, o.ManifestFilename)
	if err != nil {
		return fmt.Errorf("failed to parse pipelines: %w", err)
	}
	if err := m.Validate(); err != nil {
		return err
	}
	resources, err := buildResources(appFs, o, m)
	if err != nil {
		return err
	}
	_, err = yaml.WriteResources(appFs, o.OutputPath, resources)
	return err
}

func buildResources(fs afero.Fs, o *BuildParameters, m *config.Manifest) (res.Resources, error) {
	resources := res.Resources{}
	envs, err := buildEnvironments(fs, m)
	if err != nil {
		return nil, err
	}
	resources = res.Merge(envs, resources)
	elFiles, err := buildEventListenerResources(o.RepositoryURL, m)
	if err != nil {
		return nil, err
	}
	resources = res.Merge(elFiles, resources)

	argoApps, err := argocd.Build(argocd.ArgoCDNamespace, o.RepositoryURL, m)
	if err != nil {
		return nil, err
	}
	resources = res.Merge(argoApps, resources)
	return resources, nil
}
