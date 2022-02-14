package libdevfile

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/redhat-developer/odo/pkg/libdevfile/generator"
)

func Test_applyCommand_Execute(t *testing.T) {

	command1 := generator.GetApplyCommand(generator.ApplyCommandParams{
		Id:        "command1",
		Component: "component",
	})
	component := generator.GetContainerComponent(generator.ContainerComponentParams{
		Name: "component",
	})
	component1 := generator.GetContainerComponent(generator.ContainerComponentParams{
		Name: "component1",
	})
	component2 := generator.GetContainerComponent(generator.ContainerComponentParams{
		Name: "component2",
	})

	type fields struct {
		command    v1alpha2.Command
		devfileObj func() parser.DevfileObj
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "execute an apply command",
			fields: fields{
				command: command1,
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command1})
					_ = data.AddComponents([]v1alpha2.Component{component, component1, component2})
					return parser.DevfileObj{
						Data: data,
					}
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &applyCommand{
				command:    tt.fields.command,
				devfileObj: tt.fields.devfileObj(),
			}
			// TODO handler
			if err := o.Execute(nil); (err != nil) != tt.wantErr {
				t.Errorf("applyCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
