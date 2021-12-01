package component

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/odo/cli/component/ui"
)

// getStarterProjectInteractiveMode gets starter project value by asking user in interactive mode.
func getStarterProjectInteractiveMode(projects []devfilev1.StarterProject) *devfilev1.StarterProject {
	projectName := ui.SelectStarterProject(projects)

	// if user do not wish to download starter project or there are no projects in devfile, project name would be empty
	if projectName == "" {
		return nil
	}

	var project devfilev1.StarterProject

	for _, value := range projects {
		if value.Name == projectName {
			project = value
			break
		}
	}

	return &project
}
