package helper

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
)

// DevfileUpdater is a helper type that can mutate a Devfile object.
// It is intended to be used in conjunction with the UpdateDevfileContent function.
type DevfileUpdater func(*parser.DevfileObj) error

// DevfileMetadataNameSetter sets the 'metadata.name' field into the given Devfile
var DevfileMetadataNameSetter = func(name string) DevfileUpdater {
	return func(d *parser.DevfileObj) error {
		return d.SetMetadataName(name)
	}
}

// DevfileMetadataNameRemover removes the 'metadata.name' field from the given Devfile
var DevfileMetadataNameRemover = DevfileMetadataNameSetter("")

// DevfileCommandGroupUpdater updates the group definition of the specified command.
// It returns an error if the command was not found in the Devfile, or if there are multiple commands with the same name and type.
var DevfileCommandGroupUpdater = func(cmdName string, cmdType v1alpha2.CommandType, group *v1alpha2.CommandGroup) DevfileUpdater {
	return func(d *parser.DevfileObj) error {
		cmds, err := d.Data.GetCommands(parsercommon.DevfileOptions{
			CommandOptions: parsercommon.CommandOptions{
				CommandType: cmdType,
			},
			FilterByName: cmdName,
		})
		if err != nil {
			return err
		}
		if len(cmds) != 1 {
			return fmt.Errorf("found %v command(s) with name %q", len(cmds), cmdName)
		}
		cmd := cmds[0]
		switch cmdType {
		case v1alpha2.ApplyCommandType:
			cmd.Apply.Group = group
		case v1alpha2.CompositeCommandType:
			cmd.Composite.Group = group
		case v1alpha2.CustomCommandType:
			cmd.Custom.Group = group
		case v1alpha2.ExecCommandType:
			cmd.Exec.Group = group
		default:
			return fmt.Errorf("command type not handled: %q", cmdType)
		}
		return nil
	}
}

// UpdateDevfileContent parses the Devfile at the given path, then updates its content using the given handlers, and writes the updated Devfile to the given path.
//
// The handlers are invoked in the order they are provided.
//
// No operation is performed if no handler function is specified.
//
// See DevfileMetadataNameRemover for an example of handler function that can operate on the Devfile content.
func UpdateDevfileContent(path string, handlers []DevfileUpdater) {
	if len(handlers) == 0 {
		//Nothing to do => skip
		return
	}

	d, err := parser.ParseDevfile(parser.ParserArgs{
		Path:               path,
		FlattenedDevfile:   pointer.Bool(false),
		SetBooleanDefaults: pointer.Bool(false),
	})
	Expect(err).NotTo(HaveOccurred())
	for _, h := range handlers {
		err = h(&d)
		Expect(err).NotTo(HaveOccurred())
	}
	err = d.WriteYamlDevfile()
	Expect(err).NotTo(HaveOccurred())
}
