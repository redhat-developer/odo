package test

import (
	"errors"
	"fmt"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "test"

// TestOptions encapsulates the options for the odo command
type TestOptions struct {
	testCommand string
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
func (o *TestOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if !experimental.IsExperimentalModeEnabled() {
		return errors.New("'odo test' is only supported under experimental mode. Run 'odo preference set experimental true' to enable experimental mode. ")
	}
	return
}

// Validate validates the TestOptions based on completed values
func (o *TestOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo command
func (o *TestOptions) Run() (err error) {

	return
}

// NewCmdTest implements the odo tets command
func NewCmdTest(name, fullName string) *cobra.Command {
	o := NewTestOptions()
	testCmd := &cobra.Command{
		Use:     name,
		Short:   "Run test command defined in devfile",
		Long:    "Run test command defined in devfile",
		Example: fmt.Sprintf(testExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	testCmd.Annotations = map[string]string{"command": "main"}
	testCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	testCmd.Flags().StringVar(&o.testCommand, "test-command", "", "command name to run")
	return testCmd
}
