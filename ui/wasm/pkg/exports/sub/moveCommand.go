package sub

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/feloy/devfile-builder/wasm/pkg/utils"
)

func MoveCommand(commands []v1alpha2.Command, previousGroup, newGroup string, previousIndex, newIndex int) (map[string][]v1alpha2.Command, error) {
	commandsByGroup := map[string][]v1alpha2.Command{}

	for _, command := range commands {
		group := utils.GetGroup(command)
		commandsByGroup[group] = append(commandsByGroup[group], command)
	}

	if len(commandsByGroup[previousGroup]) < previousIndex {
		return nil, fmt.Errorf("unable to find command at index #%d in group %q", previousIndex, previousGroup)
	}

	commandToMove := commandsByGroup[previousGroup][previousIndex]
	utils.SetGroup(&commandToMove, newGroup)

	commandsByGroup[previousGroup] = append(
		commandsByGroup[previousGroup][:previousIndex],
		commandsByGroup[previousGroup][previousIndex+1:]...,
	)

	end := append([]v1alpha2.Command{}, commandsByGroup[newGroup][newIndex:]...)
	commandsByGroup[newGroup] = append(commandsByGroup[newGroup][:newIndex], commandToMove)
	commandsByGroup[newGroup] = append(commandsByGroup[newGroup], end...)

	return commandsByGroup, nil
}
