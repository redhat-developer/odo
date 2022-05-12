package registry

import "github.com/redhat-developer/odo/pkg/api"

// DevfileStackList lists all the Devfile Stacks
type DevfileStackList struct {
	DevfileRegistries []api.Registry
	Items             []api.DevfileStack
}

// TypesWithDetails is the list of project types in devfile registries, and their associated devfiles
type TypesWithDetails map[string][]api.DevfileStack
