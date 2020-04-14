package pipelines

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/openshift/odo/pkg/manifest"
	"github.com/openshift/odo/pkg/manifest/ioutils"
	pl "github.com/openshift/odo/pkg/manifest/pipelines"

	"github.com/openshift/odo/pkg/manifest/yaml"
)

// InitParameters is a struct that provides flags for initialise command
type InitParameters struct {
	GitOpsRepo          string
	GitOpsWebhookSecret string
	Output              string
	Prefix              string
	SkipChecks          bool
}

// Init function will initialise the gitops directory
func Init(o *InitParameters) error {

	if !o.SkipChecks {
		installed, err := pl.CheckTektonInstall()
		if err != nil {
			return fmt.Errorf("failed to run Tekton Pipelines installation check: %w", err)
		}
		if !installed {
			return errors.New("failed due to Tekton Pipelines or Triggers are not installed")
		}
	}

	// check if the gitops dir already exists
	exists, err := ioutils.IsExisting(o.Output)
	if exists {
		return err
	}

	files, err := manifest.CreateResources(o.Prefix, o.GitOpsRepo, o.GitOpsWebhookSecret)
	if err != nil {
		return err
	}

	pipelinesPath := manifest.GetPipelinesDir(o.Output, o.Prefix)

	fileNames, err := yaml.WriteResources(pipelinesPath, files)
	if err != nil {
		return err
	}

	sort.Strings(fileNames)
	// kustomize file should refer all the pipeline resources
	if err := yaml.AddKustomize("resources", fileNames, filepath.Join(pipelinesPath, manifest.Kustomize)); err != nil {
		return err
	}

	if err := yaml.AddKustomize("bases", []string{"./pipelines"}, filepath.Join(getCICDDir(o.Output, o.Prefix), manifest.BaseDir, manifest.Kustomize)); err != nil {
		return err
	}

	// Add overlays
	if err := yaml.AddKustomize("bases", []string{"../base"}, filepath.Join(getCICDDir(o.Output, o.Prefix), "overlays", manifest.Kustomize)); err != nil {
		return err
	}

	return nil
}

func getCICDDir(path, prefix string) string {
	return filepath.Join(path, manifest.EnvsDir, manifest.AddPrefix(prefix, manifest.CICDDir))
}
