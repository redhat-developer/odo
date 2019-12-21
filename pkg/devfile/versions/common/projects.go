package common

import "fmt"

// Errors
var (
	ErrorNoProjects         = "no projects present in devfile"
	ErrorInvalidProjectType = "invalid project type '%s' in devfile. " + fmt.Sprintf("Supported types are: '%s'", SupportedDevfileProjectTypes)
)

// isProjectTypeSupported checks if the given project type is supported in ODO
func isProjectTypeSupported(typ DevfileProjectType) bool {
	for _, supportedType := range SupportedDevfileProjectTypes {
		if supportedType == typ {
			return true
		}
	}
	return false
}

// validate() validates individual projects in the devfile
func (p *DevfileProject) validate() error {

	// Check if the project type is supported
	if !isProjectTypeSupported(p.Source.Type) {
		return fmt.Errorf(fmt.Sprintf(ErrorInvalidProjectType, p.Source.Type))
	}

	// Successful
	return nil
}

// ValidateProjects validates all the devfile projects
func ValidateProjects(projects []DevfileProject) error {

	// 1. Check if there are non-zero projects
	if len(projects) < 1 {
		return fmt.Errorf(ErrorNoProjects)
	}

	// 2. Validate each project
	for _, p := range projects {
		if err := p.validate(); err != nil {
			return err
		}
	}

	// Sucessful
	return nil
}
