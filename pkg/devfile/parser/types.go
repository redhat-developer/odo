package parser

import (
	"encoding/json"
	"fmt"
	devfileCtx "github.com/openshift/odo/pkg/devfile/parser/context"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"reflect"
	"strings"
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

func (d DevfileObj) OverrideComponents(overridePatch []common.DevfileComponent) error {
	for _, patchComponent := range overridePatch {
		found := false
		for _, originalComponent := range d.Data.GetComponents() {
			if strings.ToLower(patchComponent.Container.Name) == originalComponent.Container.Name {
				found = true

				var updatedComponent common.Container

				merged, err := handleMerge(originalComponent.Container, patchComponent.Container, common.Container{})
				if err != nil {
					return err
				}

				err = json.Unmarshal(merged, &updatedComponent)
				if err != nil {
					return err
				}

				d.Data.UpdateComponent(patchComponent.Container.Name, common.DevfileComponent{Container: &updatedComponent})
			}
		}
		if !found {
			return fmt.Errorf("the component to override is not found in the parent")
		}
	}
	return nil
}

func (d DevfileObj) OverrideCommands(overridePatch []common.DevfileCommand) error {
	for _, patchCommand := range overridePatch {
		found := false
		for _, originalCommand := range d.Data.GetCommands() {
			if strings.ToLower(patchCommand.Exec.Id) == originalCommand.Exec.Id {
				found = true
				var updatedCommand common.Exec

				merged, err := handleMerge(originalCommand.Exec, patchCommand.Exec, common.Exec{})
				if err != nil {
					return err
				}

				err = json.Unmarshal(merged, &updatedCommand)
				if err != nil {
					return err
				}

				d.Data.UpdateCommand(patchCommand.Exec.Id, common.DevfileCommand{Exec: &updatedCommand})
			}
		}
		if !found {
			return fmt.Errorf("the command to override is not found in the parent")
		}
	}
	return nil
}

func (d DevfileObj) OverrideEvents(overridePatch common.DevfileEvents) error {
	var updatedEvents common.DevfileEvents

	merged, err := handleMerge(d.Data.GetEvents(), overridePatch, common.DevfileEvents{})
	if err != nil {
		return err
	}

	err = json.Unmarshal(merged, &updatedEvents)
	if err != nil {
		return err
	}

	d.Data.UpdateEvents(updatedEvents.PostStart,
		updatedEvents.PostStop,
		updatedEvents.PreStart,
		updatedEvents.PreStop)
	return nil
}

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
					return err
				}

				d.Data.UpdateProject(patchProject.Name, updatedProject)
			}
		}
		if !found {
			return fmt.Errorf("the command to override is not found in the parent")
		}
	}
	return nil
}

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
