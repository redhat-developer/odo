package component

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/util"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// TestRecommendedCommandName is the recommended test command name
const TestRecommendedCommandName = "test"

// TestOptions encapsulates the options for the odo command
type TestOptions struct {
	// Context
	*genericclioptions.Context

	// Flags
	testCommandFlag string
	contextFlag     string
	showLogFlag     bool

	// devfile content
	devObj devfileParser.DevfileObj
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
func (to *TestOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	to.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(to.contextFlag))
	return
}

// Validate validates the TestOptions based on completed values
func (to *TestOptions) Validate() (err error) {

	if !util.CheckPathExists(to.Context.GetDevfilePath()) {
		return fmt.Errorf("unable to find devfile, odo test command is only supported by devfile components")
	}

	devObj, err := devfile.ParseAndValidateFromFile(to.Context.GetDevfilePath())
	if err != nil {
		return err
	}
	to.devObj = devObj
	return
}

// Run contains the logic for the odo command
func (to *TestOptions) Run() (err error) {
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
	testCmd.Flags().StringVar(&to.testCommandFlag, "test-command", "", "Devfile Test Command to execute")
	testCmd.Flags().BoolVar(&to.showLogFlag, "show-log", false, "If enabled, logs will be shown when running the test command")
	//Adding `--context` flag
	odoutil.AddContextFlag(testCmd, &to.contextFlag)
	//Adding `--project` flag
	projectCmd.AddProjectFlag(testCmd)
	completion.RegisterCommandHandler(testCmd, completion.ComponentNameCompletionHandler)
	return testCmd
}
