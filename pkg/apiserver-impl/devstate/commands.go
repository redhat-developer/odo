package devstate

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
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

func (o *DevfileState) MoveCommand(previousGroup, newGroup string, previousIndex, newIndex int) (DevfileContent, error) {
	commands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return DevfileContent{}, err
	}

	commandsByGroup, err := subMoveCommand(commands, previousGroup, newGroup, previousIndex, newIndex)
	if err != nil {
		return DevfileContent{}, err
	}

	// Deleting from the end as deleting from the beginning seems buggy
	for i := len(commands) - 1; i >= 0; i-- {
		err = o.Devfile.Data.DeleteCommand(commands[i].Id)
		if err != nil {
			return DevfileContent{}, err
		}
	}

	for _, group := range []string{"build", "run", "test", "debug", "deploy", ""} {
		err := o.Devfile.Data.AddCommands(commandsByGroup[group])
		if err != nil {
			return DevfileContent{}, err
		}
	}
	return o.GetContent()
}

func subMoveCommand(commands []v1alpha2.Command, previousGroup, newGroup string, previousIndex, newIndex int) (map[string][]v1alpha2.Command, error) {
	commandsByGroup := map[string][]v1alpha2.Command{}

	for _, command := range commands {
		group := GetGroup(command)
		commandsByGroup[group] = append(commandsByGroup[group], command)
	}

	if len(commandsByGroup[previousGroup]) <= previousIndex {
		return nil, fmt.Errorf("unable to find command at index #%d in group %q", previousIndex, previousGroup)
	}

	commandToMove := commandsByGroup[previousGroup][previousIndex]
	SetGroup(&commandToMove, newGroup)

	commandsByGroup[previousGroup] = append(
		commandsByGroup[previousGroup][:previousIndex],
		commandsByGroup[previousGroup][previousIndex+1:]...,
	)

	end := append([]v1alpha2.Command{}, commandsByGroup[newGroup][newIndex:]...)
	commandsByGroup[newGroup] = append(commandsByGroup[newGroup][:newIndex], commandToMove)
	commandsByGroup[newGroup] = append(commandsByGroup[newGroup], end...)

	return commandsByGroup, nil
}

func (o *DevfileState) SetDefaultCommand(commandName string, group string) (DevfileContent, error) {
	commands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return DevfileContent{}, err
	}

	for i, command := range commands {
		if GetGroup(command) == group {
			isDefault := command.Id == commandName
			SetDefault(&commands[i], isDefault)
			err = o.Devfile.Data.UpdateCommand(command)
			if err != nil {
				return DevfileContent{}, err
			}
		}
	}
	return o.GetContent()
}

func (o *DevfileState) UnsetDefaultCommand(commandName string) (DevfileContent, error) {
	commands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return DevfileContent{}, err
	}

	for i, command := range commands {
		if command.Id == commandName {
			SetDefault(&commands[i], false)
			err = o.Devfile.Data.UpdateCommand(command)
			if err != nil {
				return DevfileContent{}, err
			}
			break
		}
	}
	return o.GetContent()
}
