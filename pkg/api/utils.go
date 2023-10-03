package api

import (
	v1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/libdevfile"
)

func GetDevfileData(devfileObj parser.DevfileObj) (*DevfileData, error) {
	commands, err := getDevfileCommands(devfileObj.Data)
	if err != nil {
		return nil, err
	}
	return &DevfileData{
		Devfile:              devfileObj.Data,
		Commands:             commands,
		SupportedOdoFeatures: getSupportedOdoFeatures(devfileObj.Data),
	}, nil
}

func getSupportedOdoFeatures(devfileData data.DevfileData) *SupportedOdoFeatures {
	return &SupportedOdoFeatures{
		Dev:    libdevfile.HasRunCommand(devfileData),
		Deploy: libdevfile.HasDeployCommand(devfileData),
		Debug:  libdevfile.HasDebugCommand(devfileData),
	}
}

func getDevfileCommands(devfileData data.DevfileData) ([]DevfileCommand, error) {
	commands, err := devfileData.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}

	toGroupFn := func(g *v1alpha2.CommandGroup) (group DevfileCommandGroup, isDefault *bool) {
		if g == nil {
			return "", nil
		}
		switch g.Kind {
		case v1alpha2.BuildCommandGroupKind:
			group = BuildCommandGroup
		case v1alpha2.RunCommandGroupKind:
			group = RunCommandGroup
		case v1alpha2.DebugCommandGroupKind:
			group = DebugCommandGroup
		case v1alpha2.TestCommandGroupKind:
			group = TestCommandGroup
		case v1alpha2.DeployCommandGroupKind:
			group = DeployCommandGroup
		}

		return group, g.IsDefault
	}

	var result []DevfileCommand
	for _, cmd := range commands {
		var (
			cmdType      DevfileCommandType
			cmdComponent string
			cmdCompType  DevfileComponentType
			cmdLine      string
		)
		var cmdGroup *v1alpha2.CommandGroup
		switch {
		case cmd.Apply != nil:
			cmdType = ApplyCommandType
			cmdComponent = cmd.Apply.Component
			cmdGroup = cmd.Apply.Group
		case cmd.Exec != nil:
			cmdType = ExecCommandType
			cmdComponent = cmd.Exec.Component
			cmdGroup = cmd.Exec.Group
			cmdLine = cmd.Exec.CommandLine
		case cmd.Composite != nil:
			cmdType = CompositeCommandType
			cmdGroup = cmd.Composite.Group
		}

		var imageName string
		var comp v1alpha2.Component
		if cmdComponent != "" {
			var ok bool
			comp, ok, err = libdevfile.FindComponentByName(devfileData, cmdComponent)
			if err != nil {
				return nil, err
			}
			if ok {
				switch {
				case comp.Kubernetes != nil:
					cmdCompType = KubernetesComponentType
				case comp.Openshift != nil:
					cmdCompType = OpenshiftComponentType
				case comp.Container != nil:
					cmdCompType = ContainerComponentType
				case comp.Image != nil:
					cmdCompType = ImageComponentType
					imageName = comp.Image.ImageName
				}
			}
		}
		g, isDefault := toGroupFn(cmdGroup)
		c := DevfileCommand{
			Name:          cmd.Id,
			Type:          cmdType,
			Group:         g,
			IsDefault:     isDefault,
			CommandLine:   cmdLine,
			Component:     cmdComponent,
			ComponentType: cmdCompType,
			ImageName:     imageName,
		}
		result = append(result, c)
	}

	return result, nil
}
