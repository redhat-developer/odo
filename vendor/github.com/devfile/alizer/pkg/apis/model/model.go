/*******************************************************************************
 * Copyright (c) 2021 Red Hat, Inc.
 * Distributed under license by Red Hat, Inc. All rights reserved.
 * This program is made available under the terms of the
 * Eclipse Public License v2.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 * Red Hat, Inc.
 ******************************************************************************/

package model

import (
	"context"
	"regexp"
)

const (
	DockerFile PortDetectionAlgorithm = 0
	Compose    PortDetectionAlgorithm = 1
	Source     PortDetectionAlgorithm = 2
)

// All models inside model.go are sorted by name A-Z

// AngularCliJson represents the angular-cli.json file
type AngularCliJson struct {
	Defaults struct {
		Serve AngularHostPort `json:"serve"`
	} `json:"defaults"`
}

// AngularHostPort represents the value of AngularCliJson.Defaults.Serve
type AngularHostPort struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// AngularJson represents angular.json
type AngularJson struct {
	Projects map[string]AngularProjectBody `json:"projects"`
}

// AngularProjectBody represents the value of each key of the map for AngularJson.Projects
type AngularProjectBody struct {
	Architect struct {
		Serve struct {
			Options AngularHostPort `json:"options"`
		} `json:"serve"`
	} `json:"architect"`
}

// ApplicationFileInfo is the main struct used to select potential application files
// for detectors
type ApplicationFileInfo struct {
	// Context is the given context
	Context *context.Context

	// Root is the root path of the component
	Root string

	// Dir is the directory of the application file
	Dir string

	// File is the filename of the application file
	File string
}

// Component represents every component detected from analysis process
type Component struct {
	// Name is the name of the component
	Name string

	// Path is the root path of the component
	Path string

	// Languages is the slice of languages detected inside the component
	Languages []Language

	// Ports is the slice of integers (port values) detected
	Ports []int
}

// DetectionSettings represents the required settings for component detection
type DetectionSettings struct {
	// BasePath is the root path we need to apply detection process
	BasePath string

	// PortDetectionStrategy is the list of areas that we will apply port detection
	// Accepted values can be found at PortDetectionAlgorithm
	PortDetectionStrategy []PortDetectionAlgorithm
}

// DevfileFilter represents all filters passed to registry api upon requests
type DevfileFilter struct {
	// MinSchemaVersion is the minimum schemaVersion of the fetched devfiles
	MinSchemaVersion string

	// MaxSchemaVersion is the maximum schemaVersion of the fetched devfiles
	MaxSchemaVersion string
}

// DevfileScore represents the score that each devfile gets upon devfile matching process
type DevfileScore struct {
	// DevfileIndex is the index of the fetched registry stacks slice
	DevfileIndex int

	// Score is the score that a devfile has. The biggest score gets matched with a given source code
	Score int
}

// DevfileType represents a devfile.y(a)ml file
type DevfileType struct {
	// Name is the name of a devfile
	Name string

	// Language is the language of a devfile
	Language string

	// ProjectType is the projectType of a devfile
	ProjectType string

	// Tags is a slice of tags of a devfile
	Tags []string

	// Versions is a slice of versions of a devfile
	Versions []Version
}

// EnvVar represents an environment variable with a name and a corresponding value.
type EnvVar struct {
	// Name is the name of the environment variable.
	Name string

	// Value is the value associated with the environment variable.
	Value string
}

// Language represents every language detected from language analysis process
type Language struct {
	// Name is the name of the language
	Name string

	// Aliases is the slice of aliases for this language
	Aliases []string

	// Weight is the float value which shows the importance of this language inside a given source code
	Weight float64

	// Frameworks is the slice of frameworks detected for this language
	Frameworks []string

	// Tools is the slice of tools detected for this language
	Tools []string

	// CanBeComponent is the bool value shows if this language can be detected as component
	CanBeComponent bool

	// CanBeContainerComponent is the bool value shows if this language can be detected as container component
	CanBeContainerComponent bool
}

// MicronautApplicationProps represents the application.properties file of micronaut applications
type MicronautApplicationProps struct {
	Micronaut struct {
		Server struct {
			Port int `yaml:"port,omitempty"`
			SSL  struct {
				Enabled bool `yaml:"enabled,omitempty"`
				Port    int  `yaml:"port,omitempty"`
			} `yaml:"ssl,omitempty"`
		} `yaml:"server,omitempty"`
	} `yaml:"micronaut,omitempty"`
}

// OpenLibertyServerXml represents the server.xml file inside an open liberty application
type OpenLibertyServerXml struct {
	HttpEndpoint struct {
		HttpPort  string `xml:"httpPort,attr"`
		HttpsPort string `xml:"httpsPort,attr"`
	} `xml:"httpEndpoint"`
}

// PortDetectionAlgorithm represents one of port detection algorithm values
type PortDetectionAlgorithm int

// PortMatchRule represents a rule for port matching with a given regex and a string to replace
type PortMatchRule struct {
	// Regex is the regexp.Regexp value which will be used to match ports
	Regex *regexp.Regexp

	// ToReplace is the string value which will be replaced once the Regex is matched
	ToReplace string
}

// PortMatchRules represents a struct of rules and subrules for port matching
type PortMatchRules struct {
	// MatchIndexRegexes is a slice of PortMatchRule
	MatchIndexRegexes []PortMatchRule

	// MatchRegexes is a slice of PortMatchSubRule
	MatchRegexes []PortMatchSubRule
}

// PortMatchSubRule represents a sub rule for port matching
type PortMatchSubRule struct {
	// Regex is the primary regexp.Regexp value for the sub rule
	Regex *regexp.Regexp

	// Regex is the secondary regexp.Regexp value for the sub rule
	SubRegex *regexp.Regexp
}

// QuarkusApplicationYaml represents the application.yaml used for quarkus applications
type QuarkusApplicationYaml struct {
	Quarkus QuarkusHttp `yaml:"quarkus,omitempty"`
}

// QuarkusHttp represents the port field from application.yaml of quarkus applications
type QuarkusHttp struct {
	Http QuarkusHttpPort `yaml:"http,omitempty"`
}

// QuarkusHttpPort represents the port value from application.yaml of quarkus applications
type QuarkusHttpPort struct {
	Port             int    `yaml:"port,omitempty"`
	InsecureRequests string `yaml:"insecure-requests,omitempty"`
	SSLPort          int    `yaml:"ssl-port,omitempty"`
}

// SpringApplicationProsServer represents the application.properties file used for spring applications
type SpringApplicationProsServer struct {
	Server struct {
		Port int `yaml:"port,omitempty"`
		Http struct {
			Port int `yaml:"port,omitempty"`
		} `yaml:"http,omitempty"`
	} `yaml:"server,omitempty"`
}

// Version represents a version of a devfile
type Version struct {
	// SchemaVersion is the schemaVersion value of a devfile version
	SchemaVersion string

	// Default is the default value of a devfile version
	Default bool

	// Version is the version tag of a devfile version
	Version string
}

// VertxConf represents the config file for vertx applications
type VertxConf struct {
	Port         int                `json:"http.port,omitempty"`
	ServerConfig VertexServerConfig `json:"http.server,omitempty"`
}

// VertexServerConfig represents the server config file for vertx applications
type VertexServerConfig struct {
	Port int `json:"http.server.port,omitempty"`
}
