package config

import (
	"errors"
	"sort"
)

// Manifest describes a set of environments, apps and services for deployment.
type Manifest struct {
	Environments []*Environment `yaml:"environments,omitempty"`
}

func (m *Manifest) GetCICDEnvironment() (*Environment, error) {
	envs := []*Environment{}
	for _, env := range m.Environments {
		if env.IsCICD {
			envs = append(envs, env)
		}
	}
	if len(envs) > 1 {
		return nil, errors.New("found multiple CI/CD environments")
	}
	if len(envs) == 0 {
		return nil, errors.New("could not find CI/CD environment")
	}
	return envs[0], nil
}

// Environment is a slice of Apps, these are the named apps in the namespace.
//
// The CICD environment will be used to automatically generate CI/CD resources
// into.
// The CICD environment should not have any applications defined.
type Environment struct {
	Name      string         `yaml:"name,omitempty"`
	Pipelines *Pipelines     `yaml:"pipelines,omitempty"`
	Apps      []*Application `yaml:"apps,omitempty"`
	// TODO: this should check that there is 0 or 1 CICD environment in the
	// manfifest.
	IsCICD bool `yaml:"cicd,omitempty"`
}

func (e Environment) GoString() string {
	return e.Name
}

// Application has many services.
//
// The ConfigRepo indicates that the configuration for this application lives in
// another repository.
// TODO: validate that an app with a ConfigRepo has no services.
type Application struct {
	Name       string      `yaml:"name,omitempty"`
	Services   []*Service  `yaml:"services,omitempty"`
	ConfigRepo *Repository `yaml:"config_repo,omitempty"`
}

// Service has an upstream source.
type Service struct {
	Name      string `yaml:"name,omitempty"`
	SourceURL string `yaml:"source_url,omitempty"`
}

// Repository refers to an upstream source for reading additional config from.
type Repository struct {
	URL string `yaml:"url,omitempty"`
	// TargetRevision defines the commit, tag, or branch in which to sync the application to.
	// If omitted, will sync to HEAD
	TargetRevision string `yaml:"target_revision,omitempty"`
	// Path is a directory path within the Git repository.
	Path string `yaml:"path,omitempty"`
}

// Pipelines describes the names for pipelines to be executed for CI and CD.
//
// These pipelines will be executed with a Git clone URL and commit SHA.
type Pipelines struct {
	Integration *TemplateBinding `yaml:"integration,omitempty"`
	Deployment  *TemplateBinding `yaml:"deployment,omitempty"`
}

// TemplateBinding is a combination of the template and binding to be used for a
// pipeline execution.
type TemplateBinding struct {
	Template string `yaml:"template,omitempty"`
	Binding  string `yaml:"binding,omitempty"`
}

// Walk implements post-node visiting of each element in the manifest.
//
// Every App, Service and Environment is called once, and any error from the
// handling function terminates the Walk.
//
// The environments are sorted using a custom sorting mechanism, that orders by
// name, but, moves CICD environments to the bottom of the list.
func (m Manifest) Walk(visitor interface{}) error {
	sort.Sort(byName(m.Environments))
	for _, env := range m.Environments {
		for _, app := range env.Apps {
			for _, svc := range app.Services {
				if v, ok := visitor.(ServiceVisitor); ok {
					err := v.Service(env, app, svc)
					if err != nil {
						return err
					}
				}
			}
			if v, ok := visitor.(ApplicationVisitor); ok {
				err := v.Application(env, app)
				if err != nil {
					return err
				}
			}
		}
		if v, ok := visitor.(EnvironmentVisitor); ok {
			err := v.Environment(env)
			if err != nil {
				return nil
			}
		}
	}
	return nil
}

type byName []*Environment

func (a byName) Len() int      { return len(a) }
func (a byName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool {
	if a[j].IsCICD {
		return true
	}
	return a[i].Name < a[j].Name
}
