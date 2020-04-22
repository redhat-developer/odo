package manifest

import (
	"path/filepath"

	"github.com/openshift/odo/pkg/manifest/ioutils"
	"github.com/openshift/odo/pkg/manifest/meta"
	"github.com/openshift/odo/pkg/manifest/roles"
	"github.com/openshift/odo/pkg/manifest/yaml"
	"github.com/spf13/afero"
)

const (
	envNamespace   = "namespace.yaml"
	envRoleBinding = "rolebinding.yaml"
)

// EnvParameters encapsulates parameters for add env command
type EnvParameters struct {
	EnvName string
	Output  string
	Prefix  string
}

// Env will bootstrap a new environment directory
func Env(o *EnvParameters) error {
	envName := AddPrefix(o.Prefix, o.EnvName)
	envPath := getEnvPath(o.Output, o.EnvName, o.Prefix)
	// check if the gitops dir exists
	exists, err := ioutils.IsExisting(o.Output)
	if !exists {
		return err
	}

	// check if the environment dir already exists
	exists, err = ioutils.IsExisting(envPath)
	if exists {
		return err
	}

	appFs := afero.NewOsFs()
	err = yaml.AddKustomize(appFs, "resources", []string{envNamespace, envRoleBinding}, filepath.Join(envPath, "base", Kustomize))
	if err != nil {
		return err
	}

	err = yaml.AddKustomize(appFs, "bases", []string{"../base"}, filepath.Join(envPath, "overlays", Kustomize))
	if err != nil {
		return err
	}

	if err = addEnvResources(appFs, o.Prefix, envPath, envName); err != nil {
		return err
	}
	return nil
}

func addEnvResources(fs afero.Fs, prefix, envPath, envName string) error {
	namespaces := NamespaceNames(prefix)
	outputs := map[string]interface{}{}
	basePath := filepath.Join(envPath, "base")

	outputs[envNamespace] = CreateNamespace(envName)

	sa := roles.CreateServiceAccount(meta.NamespacedName(namespaces["cicd"], saName))

	outputs[envRoleBinding] = roles.CreateRoleBinding(meta.NamespacedName(envName, roleBindingName), sa, "ClusterRole", "edit")
	_, err := yaml.WriteResources(fs, basePath, outputs)
	if err != nil {
		return err
	}
	return nil
}

func getEnvPath(gitopsPath, envName, prefix string) string {
	return filepath.Join(gitopsPath, EnvsDir, AddPrefix(prefix, envName))
}
