package component

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
)

func (co *CreateOptions) SetComponentSettings(args []string) error {
	err := co.setComponentSourceAttributes()
	if err != nil {
		return err
	}
	err = co.setComponentName(args)
	if err != nil {
		return err
	}

	var portList []string
	if len(co.componentPorts) > 0 {
		portList = co.componentPorts
	} else {
		portList, err = co.Client.GetPortsFromBuilderImage(*co.componentSettings.Type)
		if err != nil {
			return err
		}
	}

	co.componentSettings.Ports = &(portList)
	co.componentSettings.Project = &(co.Context.Project)
	envs, err := config.NewEnvVarListFromSlice(co.componentEnvVars)
	if err != nil {
		return err
	}
	co.componentSettings.Envs = envs
	co.ignores = []string{}
	return nil
}

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
