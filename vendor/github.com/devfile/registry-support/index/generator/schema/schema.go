package schema

/*
Sample index file:
[
  {
    "name": "java-maven",
    "displayName": "Maven Java",
    "description": "Upstream Maven and OpenJDK 11",
    "tags": [
      "Java",
      "Maven"
    ],
    "projectType": "maven",
    "language": "java",
    "links": {
      "Link": "/devfile/java-maven/devfile.yaml"
    },
    "resources": [
      "devfile.yaml",
    ]
  },
  {
    "name": "nodejs",
    "displayName": "NodeJS Runtime",
    "description": "Stack with NodeJS 12",
    "tags": [
      "NodeJS",
      "Express",
      "ubi8"
    ],
    "projectType": "nodejs",
    "language": "nodejs",
    "links": {
      "Link": "/devfile/nodejs/devfile.yaml"
    },
    "resources": [
      "devfile.yaml",
      "node.vsx"
    ],
    "starterProjects": [
      "nodejs-starter"
    ]
  }
]
*/

/*
Index file schema definition
name: string - The stack name
displayName: string - The display name of devfile
description: string - The description of devfile
tags: string[] - The tags associated to devfile
projectType: string - The project framework that is used in the devfile
language: string - The project language that is used in the devfile
links: map[string]string - Links related to the devfile
resources: []string - The file resources that compose a devfile stack.
starterProjects: string[] - The project templates that can be used in the devfile
*/

// Schema is the index file schema
type Schema struct {
	Name            string            `yaml:"name,omitempty" json:"name,omitempty"`
	DisplayName     string            `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description     string            `yaml:"description,omitempty" json:"description,omitempty"`
	Tags            []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	ProjectType     string            `yaml:"projectType,omitempty" json:"projectType,omitempty"`
	Language        string            `yaml:"language,omitempty" json:"language,omitempty"`
	Links           map[string]string `yaml:"links,omitempty" json:"links,omitempty"`
	Resources       []string          `yaml:"resources,omitempty" json:"resources,omitempty"`
	StarterProjects []string          `yaml:"starterProjects,omitempty" json:"starterProjects,omitempty"`
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
