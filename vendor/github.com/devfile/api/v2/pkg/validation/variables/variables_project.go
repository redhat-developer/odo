package variables

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// ValidateAndReplaceForProjects validates the projects data for global variable references and replaces them with the variable value.
// Returns a map of project names and invalid variable references if present.
func ValidateAndReplaceForProjects(variables map[string]string, projects []v1alpha2.Project) map[string][]string {

	projectsWarningMap := make(map[string][]string)

	for i := range projects {
		var err error

		invalidKeys := make(map[string]bool)

		// Validate project clonepath
		if projects[i].ClonePath, err = validateAndReplaceDataWithVariable(projects[i].ClonePath, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate project source
		if err = validateandReplaceForProjectSource(variables, &projects[i].ProjectSource); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		err = newInvalidKeysError(invalidKeys)
		if verr, ok := err.(*InvalidKeysError); ok {
			projectsWarningMap[projects[i].Name] = verr.Keys
		}
	}

	return projectsWarningMap
}

// ValidateAndReplaceForStarterProjects validates the starter projects data for global variable references and replaces them with the variable value.
// Returns a map of starter project names and invalid variable references if present.
func ValidateAndReplaceForStarterProjects(variables map[string]string, starterProjects []v1alpha2.StarterProject) map[string][]string {

	starterProjectsWarningMap := make(map[string][]string)

	for i := range starterProjects {
		var err error

		invalidKeys := make(map[string]bool)

		// Validate starter project description
		if starterProjects[i].Description, err = validateAndReplaceDataWithVariable(starterProjects[i].Description, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate starter project sub dir
		if starterProjects[i].SubDir, err = validateAndReplaceDataWithVariable(starterProjects[i].SubDir, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		// Validate starter project source
		if err = validateandReplaceForProjectSource(variables, &starterProjects[i].ProjectSource); err != nil {
			checkForInvalidError(invalidKeys, err)
		}

		err = newInvalidKeysError(invalidKeys)
		if verr, ok := err.(*InvalidKeysError); ok {
			starterProjectsWarningMap[starterProjects[i].Name] = verr.Keys
		}
	}

	return starterProjectsWarningMap
}

// validateandReplaceForProjectSource validates a project source location for global variable references and replaces them with the variable value
func validateandReplaceForProjectSource(variables map[string]string, projectSource *v1alpha2.ProjectSource) error {

	var err error

	invalidKeys := make(map[string]bool)

	if projectSource != nil {
		switch {
		case projectSource.Zip != nil:
			if projectSource.Zip.Location, err = validateAndReplaceDataWithVariable(projectSource.Zip.Location, variables); err != nil {
				checkForInvalidError(invalidKeys, err)
			}
		case projectSource.Git != nil:
			gitProject := &projectSource.Git.GitLikeProjectSource

			if gitProject.CheckoutFrom != nil {
				// validate git checkout revision
				if gitProject.CheckoutFrom.Revision, err = validateAndReplaceDataWithVariable(gitProject.CheckoutFrom.Revision, variables); err != nil {
					checkForInvalidError(invalidKeys, err)
				}

				// // validate git checkout remote
				if gitProject.CheckoutFrom.Remote, err = validateAndReplaceDataWithVariable(gitProject.CheckoutFrom.Remote, variables); err != nil {
					checkForInvalidError(invalidKeys, err)
				}
			}

			// validate git remotes
			for k := range gitProject.Remotes {
				// validate remote map value
				if gitProject.Remotes[k], err = validateAndReplaceDataWithVariable(gitProject.Remotes[k], variables); err != nil {
					checkForInvalidError(invalidKeys, err)
				}

				// validate remote map key
				var updatedKey string
				if updatedKey, err = validateAndReplaceDataWithVariable(k, variables); err != nil {
					checkForInvalidError(invalidKeys, err)
				} else if updatedKey != k {
					gitProject.Remotes[updatedKey] = gitProject.Remotes[k]
					delete(gitProject.Remotes, k)
				}
			}
		}
	}

	return newInvalidKeysError(invalidKeys)
}
