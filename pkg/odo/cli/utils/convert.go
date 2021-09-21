package utils

import (
	"fmt"
	"github.com/fatih/color"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile/convert"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	convertCommandName = "convert-to-devfile"
)

var convertLongDesc = ktemplates.LongDesc(`Converts odo specific configuration from s2i to devfile. 
It generates devfile.yaml and .odo/env/env.yaml for s2i components`)

//var convertExample = ktemplates.Examples(`odo utils convert-to-devfile`)

var convertExample = ktemplates.Examples(`  # Convert s2i component to devfile component

Note: Run all commands from  s2i component context directory

1. Generate devfile.yaml and env.yaml for s2i component.
%[1]s  

2. Push the devfile component to the cluster.
odo push

3. Verify if devfile component is deployed sucessfully.
odo list

4. Jump to 'rolling back conversion', if devfile component deployment failed.

5. Delete the s2i component.
odo delete --s2i -a

Congratulations, you have successfully converted s2i component to devfile component.

# Rolling back the conversion

1. If devfile component deployment failed, delete the devfile component with 'odo delete -a'. 
   It would delete only devfile component, your s2i component should still be running.
 
   To complete the migration seek help from odo dev community.

`)

// ConvertOptions encapsulates the options for the command
type ConvertOptions struct {
	context          *genericclioptions.Context
	componentContext string
	componentName    string
}

// NewConvertOptions creates a new ConvertOptions instance
func NewConvertOptions() *ConvertOptions {
	return &ConvertOptions{}
}

// Complete completes ConvertOptions after they've been created
func (co *ConvertOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	co.context, err = genericclioptions.NewContext(cmd)
	if err != nil {
		return err
	}
	co.componentName = co.context.LocalConfigInfo.GetName()
	return nil

}

// Validate validates the ConvertOptions based on completed values
func (co *ConvertOptions) Validate() (err error) {
	if co.context.LocalConfigInfo.GetSourceType() == config.GIT {
		return errors.New("migration of git type s2i components to devfile is not supported by odo")
	}

	return nil
}

// Run contains the logic for the command
func (co *ConvertOptions) Run(cmd *cobra.Command) (err error) {

	/* NOTE: This data is not used in devfile currently so cannot be converted
	   minMemory := context.LocalConfigInfo.GetMinMemory()
	   minCPU := context.LocalConfigInfo.GetMinCPU()
	   maxCPU := context.LocalConfigInfo.GetMaxCPU()
	*/

	err = convert.GenerateDevfileYaml(co.context.Client, co.context.LocalConfigInfo, co.componentContext)
	if err != nil {
		return errors.Wrap(err, "Error in generating devfile.yaml")
	}

	co.context.EnvSpecificInfo, err = convert.GenerateEnvYaml(co.context.Client, co.context.LocalConfigInfo, co.componentContext)

	if err != nil {
		return errors.Wrap(err, "Error in generating env.yaml")
	}

	printOutput()

	return nil
}

// NewCmdConvert implements the odo utils convert-to-devfile command
func NewCmdConvert(name, fullName string) *cobra.Command {
	o := NewConvertOptions()
	convertCmd := &cobra.Command{
		Use:     name,
		Short:   "converts s2i based components to devfile based components",
		Long:    convertLongDesc,
		Example: fmt.Sprintf(convertExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	genericclioptions.AddContextFlag(convertCmd, &o.componentContext)

	return convertCmd
}

func printOutput() {

	infoMessage := "devfile.yaml is available in the current directory."

	nextSteps := `
To complete the conversion, run the following steps:

NOTE: At all steps your s2i component is running, It would not be deleted until you do 'odo delete --s2i -a'

1. Deploy devfile component.
$ odo push

2. Verify if the component gets deployed successfully. 
$ odo list

3. If the devfile component was deployed successfully, your application is up, you can safely delete the s2i component. 
$ odo delete --s2i -a

congratulations you have successfully converted s2i component to devfile component :).
`

	rollBackMessage := ` If you see an error or your application not coming up, delete the devfile component with 'odo delete -a' and report this to odo dev community.`

	log.Infof(infoMessage)
	log.Italicf(nextSteps)
	yellow := color.New(color.FgYellow).SprintFunc()
	log.Warning(yellow(rollBackMessage))
}
