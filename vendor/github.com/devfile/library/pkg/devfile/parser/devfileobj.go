package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// Default filenames for create devfile
const (
	OutputDevfileJsonPath = "devfile.json"
	OutputDevfileYamlPath = "devfile.yaml"
)

// DevfileObj is the runtime devfile object
type DevfileObj struct {

	// Ctx has devfile context info
	Ctx devfileCtx.DevfileCtx

	// Data has the devfile data
	Data data.DevfileData
}

// OverrideComponents overrides the components of the parent devfile
// overridePatch contains the patches to be applied to the parent's components
func (d DevfileObj) OverrideComponents(overridePatch []v1.ComponentParentOverride) error {
	// func (d DevfileObj) OverrideComponents(overridePatch interface{}) error {
	for _, patchComponent := range overridePatch {
		found := false
		for _, originalComponent := range d.Data.GetComponents() {
			if strings.ToLower(patchComponent.Name) == originalComponent.Name {
				found = true

				var updatedComponent v1.ContainerComponent

				merged, err := handleMerge(originalComponent.Container, patchComponent.Container, v1.ContainerComponent{})
				if err != nil {
					return err
				}
				err = json.Unmarshal(merged, &updatedComponent)
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal override components")
				}

				d.Data.UpdateComponent(v1.Component{
					Name: patchComponent.Name,
					ComponentUnion: v1.ComponentUnion{
						Container: &updatedComponent,
					},
				})
			}
		}
		if !found {
			return fmt.Errorf("the component to override is not found in the parent")
		}
	}
	return nil
}

// OverrideCommands overrides the commands of the parent devfile
// overridePatch contains the patches to be applied to the parent's commands
func (d DevfileObj) OverrideCommands(overridePatch []v1.CommandParentOverride) (err error) {
	for _, patchCommand := range overridePatch {
		found := false
		for _, originalCommand := range d.Data.GetCommands() {

			if strings.ToLower(patchCommand.Id) == originalCommand.Id {
				found = true

				var devfileCommand v1.Command

				if patchCommand.Exec != nil && originalCommand.Exec != nil {
					devfileCommand, err = overrideExecCommand(patchCommand, originalCommand)
					if err != nil {
						return err
					}

				} else if patchCommand.Composite != nil && originalCommand.Composite != nil {
					devfileCommand, err = overrideCompositeCommand(patchCommand, originalCommand)
					if err != nil {
						return err
					}
					// TODO: add other command types
				} else {
					// If the original command and patch command are different types, then we can't patch, so throw an error
					return fmt.Errorf("cannot overide command %q with a different type of command", originalCommand.Id)
				}

				d.Data.UpdateCommand(devfileCommand)
			}
		}
		if !found {
			return fmt.Errorf("the command to override is not found in the parent")
		}
	}
	return nil
}

// overrideCompositeCommand overrides the given parent composite commmand
// patchCommand contains the patches to be applied to the parent's command
func overrideCompositeCommand(patchCommand v1.CommandParentOverride, originalCommand v1.Command) (v1.Command, error) {
	var updatedComposite v1.CompositeCommand

	merged, err := handleMerge(originalCommand.Composite, patchCommand.Composite, v1.CompositeCommand{})
	if err != nil {
		return v1.Command{}, err
	}

	err = json.Unmarshal(merged, &updatedComposite)
	if err != nil {
		return v1.Command{}, errors.Wrap(err, "failed to unmarshal override commands")
	}
	return v1.Command{
		Id: patchCommand.Id,
		CommandUnion: v1.CommandUnion{
			Composite: &updatedComposite,
		},
	}, nil
}

// overrideExecCommand overrides the given parent Exec commmand
// patchCommand contains the patches to be applied to the parent's command
func overrideExecCommand(patchCommand v1.CommandParentOverride, originalCommand v1.Command) (v1.Command, error) {
	var updatedExec v1.ExecCommand
	merged, err := handleMerge(originalCommand.Exec, patchCommand.Exec, v1.ExecCommand{})
	if err != nil {
		return v1.Command{}, err
	}

	err = json.Unmarshal(merged, &updatedExec)
	if err != nil {
		return v1.Command{}, errors.Wrap(err, "failed to unmarshal override commands")
	}
	return v1.Command{
		Id: patchCommand.Id,
		CommandUnion: v1.CommandUnion{
			Exec: &updatedExec,
		},
	}, nil
}

// OverrideProjects overrides the projects of the parent devfile
// overridePatch contains the patches to be applied to the parent's projects
func (d DevfileObj) OverrideProjects(overridePatch []v1.ProjectParentOverride) error {
	for _, patchProject := range overridePatch {
		found := false
		for _, originalProject := range d.Data.GetProjects() {
			if strings.ToLower(patchProject.Name) == originalProject.Name {
				found = true
				var updatedProject v1.Project

				merged, err := handleMerge(originalProject, patchProject, v1.Project{})
				if err != nil {
					return err
				}

				err = json.Unmarshal(merged, &updatedProject)
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal override projects")
				}

				d.Data.UpdateProject(updatedProject)
			}
		}
		if !found {
			return fmt.Errorf("the command to override is not found in the parent")
		}
	}
	return nil
}

// OverrideStarterProjects overrides the starter projects of the parent devfile
// overridePatch contains the patches to be applied to the parent's starter projects
func (d DevfileObj) OverrideStarterProjects(overridePatch []v1.StarterProjectParentOverride) error {
	for _, patchProject := range overridePatch {
		found := false
		for _, originalProject := range d.Data.GetStarterProjects() {
			if strings.ToLower(patchProject.Name) == originalProject.Name {
				found = true
				var updatedProject v1.StarterProject

				merged, err := handleMerge(originalProject, patchProject, v1.StarterProject{})
				if err != nil {
					return err
				}

				err = json.Unmarshal(merged, &updatedProject)
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal override starter projects")
				}
				d.Data.UpdateStarterProject(updatedProject)
			}
		}
		if !found {
			return fmt.Errorf("the starterProject to override is not found in the parent")
		}
	}
	return nil
}

// handleMerge merges the patch to the original data
// dataStruct is the type of the original and the patch data
func handleMerge(original, patch, dataStruct interface{}) ([]byte, error) {
	// if reflect.TypeOf(original) != reflect.TypeOf(patch) {
	// 	return nil, fmt.Errorf("type of original and patch doesn't match")
	// }

	originalJson, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}

	patchJson, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	merged, err := strategicpatch.StrategicMergePatch(originalJson, patchJson, dataStruct)
	if err != nil {
		return nil, err
	}
	return merged, nil
}
