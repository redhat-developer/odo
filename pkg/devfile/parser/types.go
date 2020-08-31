package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	devfileCtx "github.com/openshift/odo/pkg/devfile/parser/context"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
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
func (d DevfileObj) OverrideComponents(overridePatch []common.DevfileComponent) error {
	for _, patchComponent := range overridePatch {
		found := false
		for _, originalComponent := range d.Data.GetComponents() {
			if strings.ToLower(patchComponent.Name) == originalComponent.Name {
				found = true

				var updatedContainer common.Container

				merged, err := handleMerge(originalComponent.Container, patchComponent.Container, common.Container{})
				if err != nil {
					return err
				}

				err = json.Unmarshal(merged, &updatedContainer)
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal override components")
				}

				d.Data.UpdateComponent(common.DevfileComponent{Name: patchComponent.Name, Container: &updatedContainer})
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
func (d DevfileObj) OverrideCommands(overridePatch []common.DevfileCommand) error {
	for _, patchCommand := range overridePatch {
		found := false
		for _, originalCommand := range d.Data.GetCommands() {
			if strings.ToLower(patchCommand.Id) == originalCommand.Id {
				found = true
				var updatedExec common.Exec

				merged, err := handleMerge(originalCommand.Exec, patchCommand.Exec, common.Exec{})
				if err != nil {
					return err
				}

				err = json.Unmarshal(merged, &updatedExec)
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal override commands")
				}

				d.Data.UpdateCommand(common.DevfileCommand{Id: patchCommand.Id, Exec: &updatedExec})
			}
		}
		if !found {
			return fmt.Errorf("the command to override is not found in the parent")
		}
	}
	return nil
}

// OverrideProjects overrides the projects of the parent devfile
// overridePatch contains the patches to be applied to the parent's projects
func (d DevfileObj) OverrideProjects(overridePatch []common.DevfileProject) error {
	for _, patchProject := range overridePatch {
		found := false
		for _, originalProject := range d.Data.GetProjects() {
			if strings.ToLower(patchProject.Name) == originalProject.Name {
				found = true
				var updatedProject common.DevfileProject

				merged, err := handleMerge(originalProject, patchProject, common.DevfileProject{})
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
			return fmt.Errorf("the project to override is not found in the parent")
		}
	}
	return nil
}

// OverrideStarterProjects overrides the starter projects of the parent devfile
// overridePatch contains the patches to be applied to the parent's starter projects
func (d DevfileObj) OverrideStarterProjects(overridePatch []common.DevfileStarterProject) error {
	for _, patchProject := range overridePatch {
		found := false
		for _, originalProject := range d.Data.GetStarterProjects() {
			if strings.ToLower(patchProject.Name) == originalProject.Name {
				found = true
				var updatedProject common.DevfileStarterProject

				merged, err := handleMerge(originalProject, patchProject, common.DevfileStarterProject{})
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
	if reflect.TypeOf(original) != reflect.TypeOf(patch) {
		return nil, fmt.Errorf("type of original and patch doesn't match")
	}

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
