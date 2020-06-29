package component

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"

	adaptercommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/envinfo"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const TestRecommendedCommandName = "test"

// TestOptions encapsulates the options for the odo command
type TestOptions struct {
	commandName      string
	componentContext string
	namespace        string
	devfilePath      string
	testCommand      common.DevfileCommand
	EnvSpecificInfo  *envinfo.EnvSpecificInfo
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
	if !experimental.IsExperimentalModeEnabled() {
		return fmt.Errorf("'odo test' is only supported under experimental mode. Run 'odo preference set experimental true' to enable experimental mode. ")
	}
	to.devfilePath = filepath.Join(to.componentContext, DevfilePath)
	envInfo, err := envinfo.NewEnvSpecificInfo(to.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve configuration information")
	}
	to.EnvSpecificInfo = envInfo
	to.Context = genericclioptions.NewDevfileContext(cmd)

	if !pushtarget.IsPushTargetDocker() {
		// The namespace was retrieved from the --project flag (or from the kube client if not set) and stored in kclient when initalizing the context
		to.namespace = to.KClient.Namespace
	}

	return
}

// Validate validates the TestOptions based on completed values
func (to *TestOptions) Validate() (err error) {
	devObj, err := devfileParser.ParseAndValidate(to.devfilePath)
	if err != nil {
		return errors.Wrap(err, "fail to parse devfile")
	}
	if reflect.DeepEqual(devObj.Ctx.GetApiVersion(), "1.0.0") {
		return fmt.Errorf("'odo test' is not supported in devfile 1.0.0")
	}
	to.testCommand, err = adaptercommon.GetTestCommand(devObj.Data, to.commandName)
	if err != nil {
		return errors.Wrap(err, "fail to get test command")
	}
	if reflect.DeepEqual(common.DevfileCommand{}, to.testCommand) {
		return fmt.Errorf("the test command is empty")
	}
	return
}

// Run contains the logic for the odo command
func (to *TestOptions) Run() (err error) {
	return to.RunTestCommand()
}

// NewCmdTest implements the odo tets command
func NewCmdTest(name, fullName string) *cobra.Command {
	to := NewTestOptions()
	testCmd := &cobra.Command{
		Use:     name,
		Short:   "Run test command defined in devfile",
		Long:    "Run test command defined in devfile",
		Example: fmt.Sprintf(testExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(to, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	testCmd.Annotations = map[string]string{"command": "main"}
	testCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	testCmd.Flags().StringVar(&to.commandName, "test-command", "", "command name to run")
	//Adding `--context` flag
	genericclioptions.AddContextFlag(testCmd, &to.componentContext)
	//Adding `--project` flag
	projectCmd.AddProjectFlag(testCmd)
	completion.RegisterCommandHandler(testCmd, completion.ComponentNameCompletionHandler)
	return testCmd
}
