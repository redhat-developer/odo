package component

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
)

// decideAndDownloadStarterProject decides the starter project from the value passed by the user and
// downloads it
func decideAndDownloadStarterProject(devObj parser.DevfileObj, projectPassed string, token string, interactive bool, contextDir string) error {
	if projectPassed == "" && !interactive {
		return nil
	}

	// Retrieve starter projects
	starterProjects, err := devObj.Data.GetStarterProjects(parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}

	var starterProject *devfilev1.StarterProject
	if interactive {
		starterProject = getStarterProjectInteractiveMode(starterProjects)
	} else {
		starterProject, err = component.GetStarterProject(starterProjects, projectPassed)
		if err != nil {
			return err
		}
	}

	if starterProject == nil {
		return nil
	}

	return component.DownloadStarterProject(starterProject, token, contextDir)
}

// DevfileJSON creates the full json description of a devfile component is prints it
func (co *CreateOptions) DevfileJSON() error {
	client, err := genericclioptions.Client()
	if err == nil {
		co.Client = client
	}

	envInfo, err := envinfo.NewEnvSpecificInfo(co.componentContext)
	if err != nil {
		return err
	}

	cfd, err := component.NewComponentFullDescriptionFromClientAndLocalConfigProvider(co.Client, envInfo, envInfo.GetName(), envInfo.GetApplication(), co.GetProject(), co.GetComponentContext())
	if err != nil {
		return err
	}
	machineoutput.OutputSuccess(cfd.GetComponent())
	return nil
}
