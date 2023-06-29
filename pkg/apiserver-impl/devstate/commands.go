package devstate

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
)

func (o *DevfileState) AddExecCommand(name string, component string, commandLine string, workingDir string, hotReloadCapable bool) (DevfileContent, error) {
	command := v1alpha2.Command{
		Id: name,
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				Component:        component,
				CommandLine:      commandLine,
				WorkingDir:       workingDir,
				HotReloadCapable: &hotReloadCapable,
			},
		},
	}
	err := o.Devfile.Data.AddCommands([]v1alpha2.Command{command})
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) AddApplyCommand(name string, component string) (DevfileContent, error) {
	command := v1alpha2.Command{
		Id: name,
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				Component: component,
			},
		},
	}
	err := o.Devfile.Data.AddCommands([]v1alpha2.Command{command})
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) AddCompositeCommand(name string, parallel bool, commands []string) (DevfileContent, error) {
	command := v1alpha2.Command{
		Id: name,
		CommandUnion: v1alpha2.CommandUnion{
			Composite: &v1alpha2.CompositeCommand{
				Parallel: &parallel,
				Commands: commands,
			},
		},
	}
	err := o.Devfile.Data.AddCommands([]v1alpha2.Command{command})
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) DeleteCommand(name string) (DevfileContent, error) {
	err := o.checkCommandUsed(name)
	if err != nil {
		return DevfileContent{}, fmt.Errorf("error deleting command %q: %w", name, err)
	}
	err = o.Devfile.Data.DeleteCommand(name)
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) checkCommandUsed(name string) error {
	commands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.CompositeCommandType,
		},
	})
	if err != nil {
		return err
	}
	for _, command := range commands {
		for _, subcommand := range command.Composite.Commands {
			if subcommand == name {
				return fmt.Errorf("command %q is used by composite command %q", name, command.Id)
			}
		}
	}
	return nil
}
