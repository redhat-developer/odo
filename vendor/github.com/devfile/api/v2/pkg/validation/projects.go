package validation

import (
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// ValidateStarterProjects checks if starter project has only one remote configured
// and if the checkout remote matches the renote configured
func ValidateStarterProjects(starterProjects []v1alpha2.StarterProject) error {

	var projectErrorsList []string
	for _, starterProject := range starterProjects {
		var gitSource v1alpha2.GitLikeProjectSource
		if starterProject.Git != nil {
			gitSource = starterProject.Git.GitLikeProjectSource
		} else {
			continue
		}

		switch len(gitSource.Remotes) {
		case 0:
			starterProjectErr := fmt.Errorf("starterProject %s should have at least one remote", starterProject.Name)
			newErr := resolveErrorMessageWithImportAttributes(starterProjectErr, starterProject.Attributes)
			projectErrorsList = append(projectErrorsList, newErr.Error())
		case 1:
			if gitSource.CheckoutFrom != nil && gitSource.CheckoutFrom.Remote != "" {
				err := validateRemoteMap(gitSource.Remotes, gitSource.CheckoutFrom.Remote, starterProject.Name)
				if err != nil {
					newErr := resolveErrorMessageWithImportAttributes(err, starterProject.Attributes)
					projectErrorsList = append(projectErrorsList, newErr.Error())
				}
			}
		default: // len(gitSource.Remotes) >= 2
			starterProjectErr := fmt.Errorf("starterProject %s should have one remote only", starterProject.Name)
			newErr := resolveErrorMessageWithImportAttributes(starterProjectErr, starterProject.Attributes)
			projectErrorsList = append(projectErrorsList, newErr.Error())
		}
	}

	var err error
	if len(projectErrorsList) > 0 {
		projectErrors := fmt.Sprintf("\n%s", strings.Join(projectErrorsList, "\n"))
		err = fmt.Errorf("error validating starter projects:%s", projectErrors)
	}

	return err
}

// ValidateProjects checks if the project has more than one remote configured then a checkout
// remote is mandatory and if the checkout remote matches the renote configured
func ValidateProjects(projects []v1alpha2.Project) error {

	var projectErrorsList []string
	for _, project := range projects {
		var gitSource v1alpha2.GitLikeProjectSource
		if project.Git != nil {
			gitSource = project.Git.GitLikeProjectSource
		} else {
			continue
		}
		switch len(gitSource.Remotes) {
		case 0:
			projectErr := fmt.Errorf("projects %s should have at least one remote", project.Name)
			newErr := resolveErrorMessageWithImportAttributes(projectErr, project.Attributes)
			projectErrorsList = append(projectErrorsList, newErr.Error())
		case 1:
			if gitSource.CheckoutFrom != nil && gitSource.CheckoutFrom.Remote != "" {
				if err := validateRemoteMap(gitSource.Remotes, gitSource.CheckoutFrom.Remote, project.Name); err != nil {
					newErr := resolveErrorMessageWithImportAttributes(err, project.Attributes)
					projectErrorsList = append(projectErrorsList, newErr.Error())
				}
			}
		default: // len(gitSource.Remotes) >= 2
			if gitSource.CheckoutFrom == nil || gitSource.CheckoutFrom.Remote == "" {
				projectErr := fmt.Errorf("project %s has more than one remote defined, but has no checkoutfrom remote defined", project.Name)
				newErr := resolveErrorMessageWithImportAttributes(projectErr, project.Attributes)
				projectErrorsList = append(projectErrorsList, newErr.Error())
				continue
			}
			if err := validateRemoteMap(gitSource.Remotes, gitSource.CheckoutFrom.Remote, project.Name); err != nil {
				newErr := resolveErrorMessageWithImportAttributes(err, project.Attributes)
				projectErrorsList = append(projectErrorsList, newErr.Error())
			}
		}
	}

	var err error
	if len(projectErrorsList) > 0 {
		projectErrors := fmt.Sprintf("\n%s", strings.Join(projectErrorsList, "\n"))
		err = fmt.Errorf("error validating projects:%s", projectErrors)
	}

	return err
}

// validateRemoteMap checks if the checkout remote is present in the project remote map
func validateRemoteMap(remotes map[string]string, checkoutRemote, projectName string) error {

	if _, ok := remotes[checkoutRemote]; !ok {
		return fmt.Errorf("unable to find the checkout remote %s in the remotes for project %s", checkoutRemote, projectName)
	}

	return nil
}
