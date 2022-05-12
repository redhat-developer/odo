package registry

import "github.com/redhat-developer/odo/pkg/api"

// Registry is the main struct of devfile registry
type Registry struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secure bool   `json:"secure"`
	// Priority of the registry for listing purposes. The higher the number, the higher the priority
	Priority int `json:"-"`
}

// DevfileStack is the main struct for devfile catalog components
type DevfileStack struct {
	Name            string   `json:"name"`
	DisplayName     string   `json:"displayName"`
	Description     string   `json:"description"`
	Registry        Registry `json:"registry"`
	Language        string   `json:"language"`
	Tags            []string `json:"tags"`
	ProjectType     string   `json:"projectType"`
	Version         string   `json:"version"`
	StarterProjects []string `json:"starterProjects"`

	DevfileData *api.DevfileData `json:"devfileData,omitempty"`
}

// DevfileStackList lists all the Devfile Stacks
type DevfileStackList struct {
	DevfileRegistries []Registry
	Items             []DevfileStack
}

// TypesWithDetails is the list of project types in devfile registries, and their associated devfiles
type TypesWithDetails map[string][]DevfileStack
