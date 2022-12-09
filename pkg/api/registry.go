package api

// Registry is the main struct of devfile registry
type Registry struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secure bool   `json:"secure"`
	// Priority of the registry for listing purposes. The higher the number, the higher the priority
	Priority int `json:"-"`
}

// DevfileStack is the main struct for devfile stack
type DevfileStack struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description"`
	Registry    Registry `json:"registry"`
	Language    string   `json:"language"`
	Tags        []string `json:"tags"`
	ProjectType string   `json:"projectType"`

	// DefaultVersion is the default version. Marshalled as "version" for backward compatibility.
	// Deprecated. Use Versions instead.
	DefaultVersion string                `json:"version"`
	Versions       []DevfileStackVersion `json:"versions,omitempty"`

	// DefaultStarterProjects is the list of starter projects for the default stack.
	// Marshalled as "starterProjects" for backward compatibility.
	// Deprecated. Use Versions.StarterProjects instead.
	DefaultStarterProjects []string     `json:"starterProjects"`
	DevfileData            *DevfileData `json:"devfileData,omitempty"`
}

type DevfileStackVersion struct {
	Version         string   `json:"version,omitempty"`
	IsDefault       bool     `json:"isDefault"`
	SchemaVersion   string   `json:"schemaVersion,omitempty"`
	StarterProjects []string `json:"starterProjects"`
}
