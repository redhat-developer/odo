package version200

import (
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// GetComponents returns the slice of DevfileComponent objects parsed from the Devfile
func (d *Devfile200) GetComponents() []common.DevfileComponent {
	var comps []common.DevfileComponent
	for _, v := range d.Components {
		comps = append(comps, convertV2ComponentToCommon(v))
	}
	return comps
}

// GetAliasedComponents returns the slice of DevfileComponent objects that each have an alias
func (d *Devfile200) GetAliasedComponents() []common.DevfileComponent {
	var comps []common.DevfileComponent
	for _, v := range d.Components {
		comps = append(comps, convertV2ComponentToCommon(v))
	}

	var aliasedComponents = []common.DevfileComponent{}
	for _, comp := range comps {
		if comp.Alias != nil {
			aliasedComponents = append(aliasedComponents, comp)
		}
	}
	return aliasedComponents
}

// GetProjects returns the slice of DevfileProject objects parsed from the Devfile
func (d *Devfile200) GetProjects() []common.DevfileProject {
	var projects []common.DevfileProject
	for _, v := range d.Projects {
		if v.Git != nil {
			projects = append(projects, convertV2ProjectToCommon(v))
		}

	}

	return projects
}

// GetCommands returns the slice of DevfileCommand objects parsed from the Devfile
func (d *Devfile200) GetCommands() []common.DevfileCommand {
	var commands []common.DevfileCommand
	for _, v := range d.Commands {
		// currently we are supporting only exec command.
		if v.Exec != nil {
			cmd := convertV2CommandToCommon(v)
			cmd.Name = strings.ToLower(cmd.Name)
			commands = append(commands, cmd)
		}
	}

	return commands
}

func convertV2CommandToCommon(c Command) (d common.DevfileCommand) {
	var actions []common.DevfileCommandAction

	// TODO: Need to implement for composite, custom and other types of command.
	var ex common.DevfileCommandType = common.DevfileCommandTypeExec

	action := common.DevfileCommandAction{
		Command:   &c.Exec.CommandLine,
		Component: &c.Exec.Component,
		Type:      &ex,
		Workdir:   c.Exec.WorkingDir,
	}

	actions = append(actions, action)

	// TODO: Previewurl no matching type found
	return common.DevfileCommand{
		Actions:    actions,
		Attributes: c.Exec.Attributes,
		Name:       c.Exec.Id,
	}

}

func convertV2ComponentToCommon(c Component) (d common.DevfileComponent) {
	// TODO: for other component types, custom cheplugin, cheeditor etc.
	// TODO: Support for Volume, SourceMapping

	if c.Container != nil {
		d = common.DevfileComponent{
			Alias:                       &c.Container.Name,
			MountSources:                c.Container.MountSources,
			Type:                        common.DevfileComponentTypeDockerimage,
			DevfileComponentDockerimage: convertV2ContainerToCommon(*c.Container),
		}
	}

	return d
}

func convertV2ContainerToCommon(c Container) (d common.DevfileComponentDockerimage) {
	var endpoints []common.DockerimageEndpoint
	for _, v := range c.Endpoints {
		endpoints = append(endpoints, convertV2EndpointsToCommon(*v))
	}

	var envs []common.DockerimageEnv
	for _, v := range c.Env {
		envs = append(envs, convertV2EnvToCommon(*v))
	}

	var volumes []common.DockerimageVolume
	for _, v := range c.VolumeMounts {
		volumes = append(volumes, convertV2VolumeToCommon(*v))
	}

	// TODO: Args, Command not converted (as no matching types found)
	return common.DevfileComponentDockerimage{
		Image:       &c.Image,
		MemoryLimit: &c.MemoryLimit,
		Endpoints:   endpoints,
		Env:         envs,
		Volumes:     volumes,
	}
}

func convertV2EndpointsToCommon(e Endpoint) common.DockerimageEndpoint {
	return common.DockerimageEndpoint{
		Name: &e.Name,
		Port: &e.TargetPort,
	}
}

func convertV2EnvToCommon(e Env) common.DockerimageEnv {
	return common.DockerimageEnv{
		Name:  &e.Name,
		Value: &e.Value,
	}
}

func convertV2VolumeToCommon(v VolumeMount) common.DockerimageVolume {
	return common.DockerimageVolume{
		Name:          &v.Name,
		ContainerPath: &v.Path,
	}
}

func convertV2ProjectToCommon(p Project) common.DevfileProject {
	var projectType common.DevfileProjectType = common.DevfileProjectTypeGit

	// TODO: Need to clarify on Tag, CommitId and StartPoint
	// https://github.com/devfile/kubernetes-api/blob/master/pkg/apis/workspaces/v1alpha1/projects.go#L80
	// TODO: Need to add support for other type of projects github, custom, zip.
	src := common.DevfileProjectSource{
		Type:              projectType,
		Location:          p.Git.Location,
		Branch:            &p.Git.Branch,
		CommitId:          &p.Git.StartPoint,
		Tag:               &p.Git.StartPoint,
		StartPoint:        &p.Git.StartPoint,
		SparseCheckoutDir: &p.Git.SparseCheckoutDir,
	}

	return common.DevfileProject{
		ClonePath: p.ClonePath,
		Name:      p.Name,
		Source:    src,
	}
}
