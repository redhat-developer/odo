package schema

import (
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

/*
Sample index file:
[
  {
    "name": "java-maven",
    "version": "1.1.0",
    "displayName": "Maven Java",
    "description": "Upstream Maven and OpenJDK 11",
    "type": "stack",
    "tags": [
      "Java",
      "Maven"
    ],
    "projectType": "maven",
    "language": "java",
    "links": {
      "self": "devfile-catalog/java-maven:latest"
    },
    "resources": [
      "devfile.yaml"
    ],
    "starterProjects": [
      "springbootproject"
    ]
  },
  {
    "name": "java-openliberty",
    "version": "0.5.0",
    "displayName": "Open Liberty",
    "description": "Java application stack using Open Liberty runtime",
    "type": "stack",
    "projectType": "docker",
    "language": "java",
    "links": {
      "self": "devfile-catalog/java-openliberty:latest"
    },
    "resources": [
      "devfile.yaml"
    ],
    "starterProjects": [
      "user-app"
    ]
  }
]
*/

/*
Index file schema definition
name: string - The stack name
version: string - The stack version
attributes: map[string]apiext.JSON - Map of implementation-dependant free-form YAML attributes
displayName: string - The display name of devfile
description: string - The description of devfile
type: DevfileType - The type of the devfile, currently supports stack and sample
tags: string[] - The tags associated to devfile
icon: string - The devfile icon
globalMemoryLimit: string - The devfile global memory limit
projectType: string - The project framework that is used in the devfile
language: string - The project language that is used in the devfile
links: map[string]string - Links related to the devfile
resources: []string - The file resources that compose a devfile stack.
starterProjects: string[] - The project templates that can be used in the devfile
git: *git - The information of remote repositories
provider: string - The devfile provider information
*/

// Schema is the index file schema
type Schema struct {
	Name              string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Version           string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Attributes        map[string]apiext.JSON `yaml:"attributes,omitempty" json:"attributes,omitempty"`
	DisplayName       string                 `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description       string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Type              DevfileType            `yaml:"type,omitempty" json:"type,omitempty"`
	Tags              []string               `yaml:"tags,omitempty" json:"tags,omitempty"`
	Architectures     []string               `yaml:"architectures,omitempty" json:"architectures,omitempty"`
	Icon              string                 `yaml:"icon,omitempty" json:"icon,omitempty"`
	GlobalMemoryLimit string                 `yaml:"globalMemoryLimit,omitempty" json:"globalMemoryLimit,omitempty"`
	ProjectType       string                 `yaml:"projectType,omitempty" json:"projectType,omitempty"`
	Language          string                 `yaml:"language,omitempty" json:"language,omitempty"`
	Links             map[string]string      `yaml:"links,omitempty" json:"links,omitempty"`
	Resources         []string               `yaml:"resources,omitempty" json:"resources,omitempty"`
	StarterProjects   []string               `yaml:"starterProjects,omitempty" json:"starterProjects,omitempty"`
	Git               *Git                   `yaml:"git,omitempty" json:"git,omitempty"`
	Provider          string                 `yaml:"provider,omitempty" json:"provider,omitempty"`
	SupportUrl        string                 `yaml:"supportUrl,omitempty" json:"supportUrl,omitempty"`
}

// DevfileType describes the type of devfile
type DevfileType string

const (
	// SampleDevfileType represents a sample devfile
	SampleDevfileType DevfileType = "sample"

	// StackDevfileType represents a stack devfile
	StackDevfileType DevfileType = "stack"
)

// StarterProject is the devfile starter project
type StarterProject struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}

// Devfile is the devfile structure that is used by index component
type Devfile struct {
	Meta            Schema           `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	StarterProjects []StarterProject `yaml:"starterProjects,omitempty" json:"starterProjects,omitempty"`
}

// Git stores the information of remote repositories
type Git struct {
	Remotes map[string]string `yaml:"remotes,omitempty" json:"remotes,omitempty"`
}

// ExtraDevfileEntries is the extraDevfileEntries structure that is used by index component
type ExtraDevfileEntries struct {
	Samples []Schema `yaml:"samples,omitempty" json:"samples,omitempty"`
	Stacks  []Schema `yaml:"stacks,omitempty" json:"stacks,omitempty"`
}
