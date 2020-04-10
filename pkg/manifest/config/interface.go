package config

// EnvironmentVisitor is an interface for accessing environments from the manifest.
type EnvironmentVisitor interface {
	Environment(*Environment) error
}

// ApplicationVisitor is an interface for accessing applications from the manifest.
type ApplicationVisitor interface {
	Application(*Environment, *Application) error
}

// ServiceVisitor is an interface for accessing services from the manifest.
type ServiceVisitor interface {
	Service(*Environment, *Application, *Service) error
}

type ManifestVisitor interface {
	Environment(*Environment) error
	Application(*Environment, *Application) error
	Service(*Environment, *Application, *Service) error
}
