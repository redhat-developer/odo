package validation

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/hashicorp/go-multierror"
)

// ValidateStarterProjects checks if starter project has only one remote configured
// and if the checkout remote matches the remote configured
func ValidateStarterProjects(starterProjects []v1alpha2.StarterProject) (returnedErr error) {

	for _, starterProject := range starterProjects {
		var gitSource v1alpha2.GitLikeProjectSource
		if starterProject.Git != nil {
			gitSource = starterProject.Git.GitLikeProjectSource
		} else {
			continue
		}

		if starterProjectErr := validateSingleRemoteGitSrc("starterProject", starterProject.Name, gitSource); starterProjectErr != nil {
			newErr := resolveErrorMessageWithImportAttributes(starterProjectErr, starterProject.Attributes)
			returnedErr = multierror.Append(returnedErr, newErr)
		}
	}

	return returnedErr
}

// ValidateProjects checks if the project has more than one remote configured then a checkout
// remote is mandatory and if the checkout remote matches the renote configured
func ValidateProjects(projects []v1alpha2.Project) (returnedErr error) {

	for _, project := range projects {
		var gitSource v1alpha2.GitLikeProjectSource
		if project.Git != nil {
			gitSource = project.Git.GitLikeProjectSource
		} else {
			continue
		}
		switch len(gitSource.Remotes) {
		case 0:

			newErr := resolveErrorMessageWithImportAttributes(&MissingProjectRemoteError{projectName: project.Name}, project.Attributes)
			returnedErr = multierror.Append(returnedErr, newErr)
		case 1:
			if gitSource.CheckoutFrom != nil && gitSource.CheckoutFrom.Remote != "" {
				if err := validateRemoteMap(gitSource.Remotes, gitSource.CheckoutFrom.Remote, "project", project.Name); err != nil {
					newErr := resolveErrorMessageWithImportAttributes(err, project.Attributes)
					returnedErr = multierror.Append(returnedErr, newErr)
				}
			}
		default: // len(gitSource.Remotes) >= 2
			if gitSource.CheckoutFrom == nil || gitSource.CheckoutFrom.Remote == "" {

				newErr := resolveErrorMessageWithImportAttributes(&MissingProjectCheckoutFromRemoteError{projectName: project.Name}, project.Attributes)
				returnedErr = multierror.Append(returnedErr, newErr)
				continue
			}
			if err := validateRemoteMap(gitSource.Remotes, gitSource.CheckoutFrom.Remote, "project", project.Name); err != nil {
				newErr := resolveErrorMessageWithImportAttributes(err, project.Attributes)
				returnedErr = multierror.Append(returnedErr, newErr)
			}
		}
	}

	return returnedErr
}

// validateRemoteMap checks if the checkout remote is present in the project remote map
func validateRemoteMap(remotes map[string]string, checkoutRemote, objectType, objectName string) error {

	if _, ok := remotes[checkoutRemote]; !ok {

		return &InvalidProjectCheckoutRemoteError{objectName: objectName, objectType: objectType, checkoutRemote: checkoutRemote}
	}

	return nil
}

// validateSingleRemoteGitSrc validates a git src for a single remote only
func validateSingleRemoteGitSrc(objectType, objectName string, gitSource v1alpha2.GitLikeProjectSource) (err error) {
	switch len(gitSource.Remotes) {
	case 0:
		err = &MissingRemoteError{objectType: objectType, objectName: objectName}
	case 1:
		if gitSource.CheckoutFrom != nil && gitSource.CheckoutFrom.Remote != "" {
			err = validateRemoteMap(gitSource.Remotes, gitSource.CheckoutFrom.Remote, objectType, objectName)
		}
	default: // len(gitSource.Remotes) >= 2
		err = &MultipleRemoteError{objectType: objectType, objectName: objectName}
	}

	return err
}
