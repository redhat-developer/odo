package catalog

// Registry is the main struct of devfile registry
type Registry struct {
	Name   string
	URL    string
	Secure bool
}

// DevfileComponentType is the main struct for devfile catalog components
type DevfileComponentType struct {
	Name        string
	DisplayName string
	Description string
	Link        string
	Registry    Registry
	Language    string
	Tags        []string
}

// DevfileComponentTypeList lists all the DevfileComponentType's
type DevfileComponentTypeList struct {
	DevfileRegistries []Registry
	Items             []DevfileComponentType
}
