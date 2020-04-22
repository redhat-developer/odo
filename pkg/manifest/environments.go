package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openshift/odo/pkg/manifest/config"
	res "github.com/openshift/odo/pkg/manifest/resources"
	"github.com/spf13/afero"
)

const kustomization = "kustomization.yaml"

func buildEnvironments(fs afero.Fs, m *config.Manifest) (res.Resources, error) {
	files := make(res.Resources)
	eb := &envBuilder{fs: fs, files: files, appServices: make(map[string][]string)}
	err := m.Walk(eb)
	return eb.files, err
}

type envBuilder struct {
	// this is a mapping of app.Name to environment relative service paths.
	appServices map[string][]string
	files       res.Resources
	fs          afero.Fs
}

func (b *envBuilder) Application(env *config.Environment, app *config.Application) error {
	appPath := filepath.Join(config.PathForApplication(env, app))
	appFiles, err := filesForApplication(appPath, app, b.appServices[app.Name])
	if err != nil {
		return err
	}
	b.files = res.Merge(appFiles, b.files)
	return nil
}

func (b *envBuilder) Service(env *config.Environment, app *config.Application, svc *config.Service) error {
	if b.appServices[app.Name] == nil {
		b.appServices[app.Name] = []string{}
	}
	svcPath := config.PathForService(env, svc)
	b.appServices[app.Name] = append(b.appServices[app.Name], svcPath)
	svcFiles, err := filesForService(svcPath, svc)
	if err != nil {
		return err
	}
	b.files = res.Merge(svcFiles, b.files)

	return nil
}

func (b *envBuilder) Environment(env *config.Environment) error {
	if env.IsSpecial() {
		return nil
	}
	envPath := filepath.Join(config.PathForEnvironment(env), "env")
	basePath := filepath.Join(envPath, "base")
	envFiles := filesForEnvironment(basePath, env)
	kustomizedFilenames, err := listFiles(b.fs, basePath)
	if err != nil {
		return fmt.Errorf("failed to list initial files for %s: %s", basePath, err)
	}
	for k, _ := range envFiles {
		kustomizedFilenames[filepath.Base(k)] = true
	}
	envFiles[filepath.Join(basePath, kustomization)] = &res.Kustomization{Resources: extractFilenames(kustomizedFilenames)}

	overlaysPath := filepath.Join(envPath, "overlays")
	relPath, err := filepath.Rel(overlaysPath, basePath)
	if err != nil {
		return err
	}
	envFiles[filepath.Join(overlaysPath, kustomization)] = &res.Kustomization{Bases: []string{relPath}}
	b.files = res.Merge(envFiles, b.files)
	return nil
}

func filesForEnvironment(basePath string, env *config.Environment) res.Resources {
	envFiles := res.Resources{}
	filename := filepath.Join(basePath, fmt.Sprintf("%s-environment.yaml", env.Name))
	envFiles[filename] = CreateNamespace(env.Name)
	return envFiles
}

func filesForApplication(appPath string, app *config.Application, services []string) (res.Resources, error) {
	envFiles := res.Resources{}
	basePath := filepath.Join(appPath, "base")
	overlaysPath := filepath.Join(appPath, "overlays")
	overlaysFile := filepath.Join(overlaysPath, kustomization)
	overlayRel, err := filepath.Rel(overlaysPath, basePath)
	if err != nil {
		return nil, err
	}
	baseKustomization := filepath.Join(appPath, "base", kustomization)
	relServices := []string{}
	for _, v := range services {
		relService, err := filepath.Rel(filepath.Dir(baseKustomization), v)
		if err != nil {
			return nil, err
		}
		relServices = append(relServices, relService)
	}
	envFiles[filepath.Join(appPath, kustomization)] = &res.Kustomization{Bases: []string{"overlays"}}
	envFiles[filepath.Join(appPath, "base", kustomization)] = &res.Kustomization{Bases: relServices}
	envFiles[overlaysFile] = &res.Kustomization{Bases: []string{overlayRel}}

	return envFiles, nil
}

func filesForService(svcPath string, app *config.Service) (res.Resources, error) {
	envFiles := res.Resources{}
	basePath := filepath.Join(svcPath, "base")
	overlaysPath := filepath.Join(svcPath, "overlays")
	overlaysFile := filepath.Join(overlaysPath, kustomization)
	overlayRel, err := filepath.Rel(overlaysPath, basePath)
	if err != nil {
		return nil, err
	}
	envFiles[filepath.Join(svcPath, kustomization)] = &res.Kustomization{Bases: []string{"overlays"}}
	envFiles[filepath.Join(svcPath, "base", kustomization)] = &res.Kustomization{Bases: []string{"./config"}}
	envFiles[overlaysFile] = &res.Kustomization{Bases: []string{overlayRel}}

	return envFiles, nil
}

func extractFilenames(f map[string]bool) []string {
	names := []string{}
	for k, _ := range f {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func listFiles(fs afero.Fs, base string) (map[string]bool, error) {
	files := map[string]bool{}
	err := afero.Walk(fs, base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := strings.TrimPrefix(path, base+"/")
		if filename == kustomization {
			return nil
		}
		files[filename] = true
		return nil
	})
	return files, err
}
