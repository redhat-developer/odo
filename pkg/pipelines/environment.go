package pipelines

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/pkg/pipelines/config"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/scm"
	"github.com/openshift/odo/pkg/pipelines/yaml"
	"github.com/spf13/afero"
)

// EnvParameters encapsulates parameters for add env command
type EnvParameters struct {
	ManifestFilename string
	EnvName          string
	Cluster          string
}

// AddEnv adds a new environment to the manifest.
func AddEnv(o *EnvParameters, appFs afero.Fs) error {
	m, err := config.ParseFile(appFs, o.ManifestFilename)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}
	env := m.GetEnvironment(o.EnvName)
	if env != nil {
		return fmt.Errorf("environment %s already exists", o.EnvName)
	}
	files := res.Resources{}
	newEnv, err := newEnvironment(m, o.EnvName)
	if err != nil {
		return err
	}
	if o.Cluster != "" {
		newEnv.Cluster = o.Cluster
	}
	m.Environments = append(m.Environments, newEnv)
	files[pipelinesFile] = m
	outputPath := filepath.Dir(o.ManifestFilename)
	buildParams := &BuildParameters{
		ManifestFilename: o.ManifestFilename,
		OutputPath:       outputPath,
	}
	built, err := buildResources(appFs, buildParams, m)
	if err != nil {
		return fmt.Errorf("failed to build resources: %w", err)
	}
	files = res.Merge(built, files)
	_, err = yaml.WriteResources(appFs, outputPath, files)
	return err
}

func newEnvironment(m *config.Manifest, name string) (*config.Environment, error) {
	pipelinesConfig := m.GetPipelinesConfig()
	if pipelinesConfig != nil && m.GitOpsURL != "" {
		r, err := scm.NewRepository(m.GitOpsURL)
		if err != nil {
			return nil, err
		}
		return &config.Environment{
			Name:      name,
			Pipelines: defaultPipelines(r),
		}, nil
	}

	return &config.Environment{
		Name: name,
	}, nil
}
