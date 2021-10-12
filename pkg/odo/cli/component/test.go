package component

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	"github.com/openshift/odo/pkg/util"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// TestRecommendedCommandName is the recommended test command name
const TestRecommendedCommandName = "test"

// TestOptions encapsulates the options for the odo command
type TestOptions struct {
	commandName      string
	componentContext string
	devfilePath      string
	show             bool
	devObj           devfileParser.DevfileObj
	*genericclioptions.Context
}

var testExample = templates.Examples(`
  # Run default test command
  %[1]s

  # Run a specific test command
  %[1]s --test-command <command name>

`)

// NewTestOptions creates a new TestOptions instance
func NewTestOptions() *TestOptions {
	return &TestOptions{}
}

// Complete completes TestOptions after they've been created
func (to *TestOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	to.devfilePath = devfile.DevfileLocation(to.componentContext)
	to.Context, err = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the TestOptions based on completed values
func (to *TestOptions) Validate() (err error) {

	if !util.CheckPathExists(to.devfilePath) {
		return fmt.Errorf("unable to find devfile, odo test command is only supported by devfile components")
	}

	devObj, err := devfile.ParseFromFile(to.devfilePath)
	if err != nil {
		return err
	}
	to.devObj = devObj
	return
}

// Run contains the logic for the odo command
func (to *TestOptions) Run(cmd *cobra.Command) (err error) {
	return to.RunTestCommand()
}

// NewCmdTest implements the odo test command
func NewCmdTest(name, fullName string) *cobra.Command {
	to := NewTestOptions()
	testCmd := &cobra.Command{
		Use:     name,
		Short:   "Run the test command defined in the devfile",
		Long:    "Run the test command defined in the devfile",
		Example: fmt.Sprintf(testExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(to, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	testCmd.Annotations = map[string]string{"command": "main"}
	testCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	testCmd.Flags().StringVar(&to.commandName, "test-command", "", "Devfile Test Command to execute")
	testCmd.Flags().BoolVar(&to.show, "show-log", false, "If enabled, logs will be shown when running the test command")
	//Adding `--context` flag
	genericclioptions.AddContextFlag(testCmd, &to.componentContext)
	//Adding `--project` flag
	projectCmd.AddProjectFlag(testCmd)
	// Adding `--app` flag
	appCmd.AddApplicationFlag(testCmd)
	completion.RegisterCommandHandler(testCmd, completion.ComponentNameCompletionHandler)
	return testCmd
}
