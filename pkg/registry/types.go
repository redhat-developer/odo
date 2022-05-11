package registry

import "github.com/redhat-developer/odo/pkg/api"

// Registry is the main struct of devfile registry
type Registry struct {
	Name     string
	URL      string
	Secure   bool
	Priority int // The "priority" of the registry for listing purposes. The higher the number, the higher the priority
}

// DevfileStack is the main struct for devfile catalog components
type DevfileStack struct {
	Name                 string
	DisplayName          string
	Description          string
	Link                 string
	Registry             Registry
	Language             string
	Tags                 []string
	ProjectType          string
	Version              string
	StarterProjects      []string
	SupportedOdoFeatures api.SupportedOdoFeatures
}

// DevfileStackList lists all the Devfile Stacks
type DevfileStackList struct {
	DevfileRegistries []Registry
	Items             []DevfileStack
}

// TypesWithDetails is the list of project types in devfile registries, and their associated devfiles
type TypesWithDetails map[string][]DevfileStack
