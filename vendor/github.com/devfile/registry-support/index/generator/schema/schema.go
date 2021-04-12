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
    ]
  },
  {
    "name": "java-openliberty",
    "version": "0.3.0",
    "displayName": "Open Liberty",
    "description": "Java application stack using Open Liberty runtime",
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
tags: string[] - The tags associated to devfile
icon: string - The devfile icon
globalMemoryLimit: string - The devfile global memory limit
projectType: string - The project framework that is used in the devfile
language: string - The project language that is used in the devfile
links: map[string]string - Links related to the devfile
resources: []string - The file resources that compose a devfile stack.
starterProjects: string[] - The project templates that can be used in the devfile
*/

// Schema is the index file schema
type Schema struct {
	Name              string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Version           string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Attributes        map[string]apiext.JSON `yaml:"attributes,omitempty" json:"attributes,omitempty"`
	DisplayName       string                 `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description       string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Tags              []string               `yaml:"tags,omitempty" json:"tags,omitempty"`
	Icon              string                 `yaml:"icon,omitempty" json:"icon,omitempty"`
	GlobalMemoryLimit string                 `yaml:"globalMemoryLimit,omitempty" json:"globalMemoryLimit,omitempty"`
	ProjectType       string                 `yaml:"projectType,omitempty" json:"projectType,omitempty"`
	Language          string                 `yaml:"language,omitempty" json:"language,omitempty"`
	Links             map[string]string      `yaml:"links,omitempty" json:"links,omitempty"`
	Resources         []string               `yaml:"resources,omitempty" json:"resources,omitempty"`
	StarterProjects   []string               `yaml:"starterProjects,omitempty" json:"starterProjects,omitempty"`
}

// StarterProject is the devfile starter project
type StarterProject struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
}

// Devfile is the devfile structure that is used by index component
type Devfile struct {
	Meta            Schema           `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	StarterProjects []StarterProject `yaml:"starterProjects,omitempty" json:"starterProjects,omitempty"`
}
