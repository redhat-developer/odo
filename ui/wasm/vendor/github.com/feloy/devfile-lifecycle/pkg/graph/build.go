package graph

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/feloy/devfile-lifecycle/pkg/dftools"
)

func Build(devfileData data.DevfileData) (*Graph, error) {

	g := NewGraph()

	/* Get "container" node */
	containers, err := devfileData.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.ContainerComponentType,
		},
	})
	if err != nil {
		return nil, err
	}

	singleContainer := len(containers) == 1

	start := g.AddNode("start")

	var containerNode *Node
	var containerName string
	if len(containers) == 1 {
		container := containers[0]
		containerName = container.Name
		containerNode = g.AddNode(
			containerName,
			"container: "+container.Name,
			"image: "+container.Container.Image,
		)
	} else {
		texts := make([]string, 0, 2*len(containers))
		for _, container := range containers {
			texts = append(
				texts,
				"container: "+container.Name,
				"image: "+container.Container.Image,
			)
		}
		containerName = "containers"
		containerNode = g.AddNode(
			containerName,
			texts...,
		)
	}
	g.EntryNodeID = containerNode.ID

	_ = g.AddEdge(
		start,
		containerNode,
		"dev",
	)

	syncNodeStart := g.AddNode(
		"sync-all-"+containerName,
		"Sync All Sources",
	)

	_ = g.AddEdge(
		containerNode,
		syncNodeStart,
		"container running",
	)

	/* Get PostStart event */
	postStartEvents := devfileData.GetEvents().PostStart

	previousNode := syncNodeStart
	nextText := "sources synced"
	for _, postStartEvent := range postStartEvents {
		node := g.AddNode(postStartEvent, "Post Start", "command: "+postStartEvent)
		_ = g.AddEdge(
			previousNode,
			node,
			nextText,
		)
		previousNode = node
		nextText = postStartEvent + " done"
	}

	/* Get "build command" node */
	buildCommands, err := devfileData.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandGroupKind: v1alpha2.BuildCommandGroupKind,
		},
	})
	if err != nil {
		return nil, err
	}

	var defaultBuildCommand v1alpha2.Command
	for _, buildCommand := range buildCommands {
		if dftools.GetCommandGroup(buildCommand).IsDefault != nil && *dftools.GetCommandGroup(buildCommand).IsDefault {
			defaultBuildCommand = buildCommand
			break
		}
	}

	if defaultBuildCommand.Id == "" {
		return g, nil
	}

	buildNodeStart, buildNodeEnd, err := addCommand(g, devfileData, defaultBuildCommand, previousNode, singleContainer, nextText)
	if err != nil {
		return nil, err
	}

	for _, debug := range []bool{false, true} {
		/* Get "run command" node */

		kind := v1alpha2.RunCommandGroupKind
		if debug {
			kind = v1alpha2.DebugCommandGroupKind
		}
		runCommands, err := devfileData.GetCommands(common.DevfileOptions{
			CommandOptions: common.CommandOptions{
				CommandGroupKind: kind,
			},
		})
		if err != nil {
			return nil, err
		}

		var defaultRunCommand v1alpha2.Command
		for _, runCommand := range runCommands {
			if dftools.GetCommandGroup(runCommand).IsDefault != nil && *dftools.GetCommandGroup(runCommand).IsDefault {
				defaultRunCommand = runCommand
				break
			}
		}

		if defaultRunCommand.Id == "" {
			continue
		}

		edgeText := "build done, "
		if debug {
			edgeText += "with debug"
		} else {
			edgeText += "with run"
		}

		runNodeStart, runNodeEnd, err := addCommand(g, devfileData, defaultRunCommand, buildNodeEnd, singleContainer, edgeText)
		if err != nil {
			return nil, err
		}

		lines := []string{
			"Expose ports",
		}
		for _, container := range containers {
			for _, endpoint := range container.Container.Endpoints {
				if !debug && dftools.IsDebugPort(endpoint) {
					continue
				}
				lines = append(lines, fmt.Sprintf("%s: %d", endpoint.Name, endpoint.TargetPort))
			}
		}

		exposeNode := g.AddNode(
			containerName+"-"+runNodeStart.ID+"-expose",
			lines...,
		)

		_ = g.AddEdge(
			runNodeEnd,
			exposeNode,
			"command running",
		)

		/* Get PreStop event */
		preStopEvents := devfileData.GetEvents().PreStop

		previousNode := exposeNode
		nextText := "User quits"
		for _, preStopEvent := range preStopEvents {
			node := g.AddNode(preStopEvent, "Pre Stop", "command: "+preStopEvent)
			_ = g.AddEdge(
				previousNode,
				node,
				nextText,
			)
			previousNode = node
			nextText = preStopEvent + " done"
		}

		/* Add "stop container" node */

		stopNode := g.AddNode(containerName+"-stop", "Stop containers")

		/* Add "user quits" edge */

		_ = g.AddEdge(
			previousNode,
			stopNode,
			nextText,
		)

		_, syncNodeChangedExists := g.nodes["sync-modified-"+containerName]

		// Add "Sync" node
		syncNodeChanged := g.AddNode(
			"sync-modified-"+containerName,
			"Sync Modified Sources",
		)

		_ = g.AddEdge(
			exposeNode,
			syncNodeChanged,
			"source changed",
		)

		/* Add "source synced" edge */

		if !syncNodeChangedExists {
			if defaultRunCommand.Exec != nil && defaultRunCommand.Exec.HotReloadCapable != nil && *defaultRunCommand.Exec.HotReloadCapable {
				_ = g.AddEdge(
					syncNodeChanged,
					exposeNode,
					"source synced",
				)
			} else {
				_ = g.AddEdge(
					syncNodeChanged,
					buildNodeStart,
					"source synced",
				)
			}
		}

		/* Add "devfile changed" edge */

		_ = g.AddEdge(
			exposeNode,
			containerNode,
			"devfile changed",
		)
	}

	deployCommands, err := devfileData.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandGroupKind: v1alpha2.DeployCommandGroupKind,
		},
	})
	if err != nil {
		return nil, err
	}

	var defaultDeployCommand v1alpha2.Command
	for _, deployCommand := range deployCommands {
		if dftools.GetCommandGroup(deployCommand).IsDefault != nil && *dftools.GetCommandGroup(deployCommand).IsDefault {
			defaultDeployCommand = deployCommand
			break
		}
	}

	if defaultDeployCommand.Id != "" {
		_, lastNode, err := addCommand(g, devfileData, defaultDeployCommand, start, singleContainer, "deploy")
		if err != nil {
			return nil, err
		}

		finishNode := g.AddNode(
			"finish-deploy",
			"end",
		)

		_ = g.AddEdge(
			lastNode,
			finishNode,
			"command done",
		)
	}

	return g, nil
}

func addCommand(g *Graph, devfileData data.DevfileData, command v1alpha2.Command, nodeBefore *Node, singleContainer bool, text ...string) (start *Node, end *Node, err error) {
	if command.Exec != nil {
		return addExecCommand(g, command, nodeBefore, singleContainer, text...)
	}
	if command.Composite != nil {
		return addCompositeCommand(g, devfileData, command, nodeBefore, singleContainer, text...)
	}
	if command.Apply != nil {
		return addApplyCommand(g, command, nodeBefore, text...)
	}
	return nil, nil, fmt.Errorf("command type not implemented for %s", command.Id)
}

func addExecCommand(g *Graph, command v1alpha2.Command, nodeBefore *Node, singleContainer bool, text ...string) (*Node, *Node, error) {
	texts := []string{"command: " + command.Id}
	if !singleContainer {
		texts = append(texts, "container: "+command.Exec.Component)
	}
	node := g.AddNode(
		command.Id,
		texts...,
	)

	_ = g.AddEdge(
		nodeBefore,
		node,
		text...,
	)

	return node, node, nil

}

func addApplyCommand(g *Graph, command v1alpha2.Command, nodeBefore *Node, text ...string) (*Node, *Node, error) {
	node := g.AddNode(
		command.Id,
		"command: "+command.Id,
	)

	_ = g.AddEdge(
		nodeBefore,
		node,
		text...,
	)

	return node, node, nil

}

func addCompositeCommand(g *Graph, devfileData data.DevfileData, command v1alpha2.Command, nodeBefore *Node, singleContainer bool, text ...string) (*Node, *Node, error) {
	previousNode := nodeBefore
	var firstNode *Node
	for _, subcommandName := range command.Composite.Commands {
		subcommands, err := devfileData.GetCommands(common.DevfileOptions{
			FilterByName: subcommandName,
		})
		if err != nil {
			return nil, nil, err
		}
		if len(subcommands) != 1 {
			return nil, nil, fmt.Errorf("command not found: %s", subcommandName)
		}
		var first *Node
		first, previousNode, err = addCommand(g, devfileData, subcommands[0], previousNode, singleContainer, text...)
		if err != nil {
			return nil, nil, err
		}
		if firstNode == nil {
			firstNode = first
		}
		text = []string{
			subcommandName + " done",
		}
	}

	return firstNode, previousNode, nil
}
