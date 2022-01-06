package component

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/kclient"
	"os"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	"github.com/redhat-developer/odo/pkg/component"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/spf13/cobra"
)

// DescribeRecommendedCommandName is the recommended describe command name
const DescribeRecommendedCommandName = "describe"

var describeExample = ktemplates.Examples(`  # Describe nodejs component
%[1]s nodejs
`)

// DescribeOptions is a dummy container to attach complete, validate and run pattern
type DescribeOptions struct {
	// Component context
	*ComponentOptions
	// Clients
	componentClient component.Client
	// Flags
	contextFlag string
}

// NewDescribeOptions returns new instance of ListOptions
func NewDescribeOptions(client component.Client) *DescribeOptions {
	return &DescribeOptions{
		ComponentOptions: &ComponentOptions{},
		componentClient:  client,
	}
}

// Complete completes describe args
func (do *DescribeOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	if do.contextFlag == "" {
		do.contextFlag, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	err = do.ComponentOptions.Complete(cmdline, args)
	if err != nil {
		return err
	}
	return nil
}

// Validate validates the describe parameters
func (do *DescribeOptions) Validate() (err error) {

	if !((do.GetApplication() != "" && do.GetProject() != "") || do.EnvSpecificInfo.Exists()) {
		return fmt.Errorf("component %v does not exist", do.componentName)
	}

	return nil
}

// Run has the logic to perform the required actions as part of command
func (do *DescribeOptions) Run() (err error) {

	cfd, err := do.componentClient.NewComponentFullDescriptionFromClientAndLocalConfigProvider(do.EnvSpecificInfo, do.componentName, do.Context.GetApplication(), do.Context.GetProject(), do.contextFlag)
	if err != nil {
		return err
	}

	if log.IsJSON() {
		machineoutput.OutputSuccess(cfd)
	} else {
		err = cfd.Print(do.Context.KClient)
		if err != nil {
			return err
		}
	}
	return
}

// NewCmdDescribe implements the describe odo command
func NewCmdDescribe(name, fullName string) *cobra.Command {
	client, _ := kclient.New()
	do := NewDescribeOptions(component.NewClient(client))

	var describeCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s [component_name]", name),
		Short:       "Describe component",
		Long:        `Describe component.`,
		Example:     fmt.Sprintf(describeExample, fullName),
		Args:        cobra.RangeArgs(0, 1),
		Annotations: map[string]string{"machineoutput": "json", "command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(do, cmd, args)
		},
	}

	describeCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(describeCmd, completion.ComponentNameCompletionHandler)
	// Adding --context flag
	odoutil.AddContextFlag(describeCmd, &do.contextFlag)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(describeCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(describeCmd)

	return describeCmd
}
