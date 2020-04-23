package pipelines

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openshift/odo/pkg/manifest"
	"github.com/openshift/odo/pkg/manifest/ioutils"
	"github.com/spf13/afero"

	"github.com/openshift/odo/pkg/manifest/yaml"
)

// InitParameters is a struct that provides flags for initialise command
type InitParameters struct {
	GitOpsRepo               string
	GitOpsWebhookSecret      string
	Output                   string
	Prefix                   string
	DeploymentPath           string
	ImageRepo                string
	InternalRegistryHostname string
	Dockercfgjson            string
}

// Init function will initialise the gitops directory
func Init(o *InitParameters) error {

	// check if the gitops dir already exists
	exists, err := ioutils.IsExisting(o.Output)
	if exists {
		return err
	}

	files, err := manifest.CreateResources(o.Prefix, o.GitOpsRepo, o.GitOpsWebhookSecret, "", o.ImageRepo)
	if err != nil {
		return err
	}

	pipelinesPath := manifest.GetPipelinesDir(o.Output, o.Prefix)
	appFs := afero.NewOsFs()
	fileNames, err := yaml.WriteResources(appFs, pipelinesPath, files)
	if err != nil {
		return err
	}

	sort.Strings(fileNames)
	// kustomize file should refer all the pipeline resources
	if err := yaml.AddKustomize(appFs, "resources", fileNames, filepath.Join(pipelinesPath, manifest.Kustomize)); err != nil {
		return err
	}

	if err := yaml.AddKustomize(appFs, "bases", []string{"./pipelines"}, filepath.Join(getCICDDir(o.Output, o.Prefix), manifest.BaseDir, manifest.Kustomize)); err != nil {
		return err
	}

	// Add overlays
	if err := yaml.AddKustomize(appFs, "bases", []string{"../base"}, filepath.Join(getCICDDir(o.Output, o.Prefix), "overlays", manifest.Kustomize)); err != nil {
		return err
	}

	return nil
}

func getCICDDir(path, prefix string) string {
	return filepath.Join(path, manifest.EnvsDir, manifest.AddPrefix(prefix, manifest.CICDDir))
}

// validatingImageRepo validates the input image repo.  It determines if it is
// for internal registry and prepend internal registry hostname if neccessary.
func validatingImageRepo(o *InitParameters) (bool, string, error) {
	components := strings.Split(o.ImageRepo, "/")

	// repo url has minimum of 2 components
	if len(components) < 2 {
		return false, "", imageRepoValidationErrors(o.ImageRepo)
	}

	for _, v := range components {
		// check for empty components
		if strings.TrimSpace(v) == "" {
			return false, "", imageRepoValidationErrors(o.ImageRepo)
		}
		// check for white spaces
		if len(v) > len(strings.TrimSpace(v)) {
			return false, "", imageRepoValidationErrors(o.ImageRepo)
		}
	}

	if len(components) == 2 {
		if components[0] == "docker.io" || components[0] == "quay.io" {
			// we recognize docker.io and quay.io.  It is missing one component
			return false, "", imageRepoValidationErrors(o.ImageRepo)
		}
		// We have format like <project>/<app> which is an internal registry.
		// We prepend the internal registry hostname.
		return true, o.InternalRegistryHostname + "/" + o.ImageRepo, nil
	}

	// Check the first component to see if it is an internal registry
	if len(components) == 3 {
		return components[0] == o.InternalRegistryHostname, o.ImageRepo, nil
	}

	// > 3 components.  invalid repo
	return false, "", imageRepoValidationErrors(o.ImageRepo)
}

func imageRepoValidationErrors(imageRepo string) error {
	return fmt.Errorf("failed to parse image repo:%s, expected image repository in the form <registry>/<username>/<repository> or <project>/<app> for internal registry", imageRepo)
}
