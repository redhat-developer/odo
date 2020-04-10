package manifest

import (
	"path/filepath"

	"github.com/openshift/odo/pkg/manifest/config"
	"github.com/openshift/odo/pkg/manifest/yaml"
)

type resources map[string]interface{}

// InitParameters is a struct that provides flags for the Init command.
type InitParameters struct {
	GitOpsRepo          string
	GitOpsWebhookSecret string
	Output              string
	Prefix              string
}

// Init bootstraps a GitOps manifest and repository structure.
func Init(o *InitParameters) error {
	outputs, err := createInitialFiles(o.Prefix)
	if err != nil {
		return err
	}
	_, err = yaml.WriteResources(o.Output, outputs)
	return err
}

func createInitialFiles(prefix string) (map[string]interface{}, error) {
	manifest := createManifest(prefix)
	initialFiles := map[string]interface{}{
		"manifest.yaml": manifest,
	}
	cicdKustomizations := addPrefixToResources(cicdEnvironmentPath(manifest), getCICDKustomization())
	initialFiles = merge(cicdKustomizations, initialFiles)
	return initialFiles, nil
}

func createManifest(prefix string) *config.Manifest {
	return &config.Manifest{
		Environments: []*config.Environment{
			{
				Name:   prefix + "cicd",
				IsCICD: true,
			},
		},
	}
}

func getCICDKustomization() resources {
	return resources{
		"base/kustomization.yaml": map[string]interface{}{
			"resources": []string{},
		},
		"overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
	}
}

func pathForEnvironment(env *config.Environment) string {
	return filepath.Join("environments", env.Name)
}

func addPrefixToResources(prefix string, files resources) map[string]interface{} {
	updated := map[string]interface{}{}
	for k, v := range files {
		updated[filepath.Join(prefix, k)] = v
	}
	return updated
}

func merge(from, to resources) resources {
	merged := resources{}
	for k, v := range to {
		merged[k] = v
	}
	for k, v := range from {
		merged[k] = v
	}
	return merged
}

// TODO: this should probably use the .FindCICDEnvironment on the manifest.
func cicdEnvironmentPath(m *config.Manifest) string {
	return pathForEnvironment(m.Environments[0])
}
