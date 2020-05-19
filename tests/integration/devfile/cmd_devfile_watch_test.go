package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"
)

var _ = Describe("odo devfile watch command tests", func() {
	var globals helper.Globals

	var cliRunner helper.CliRunner

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is run after every Spec (It)
	var _ = BeforeEach(func() {

		globals = helper.CommonBeforeEach()

		helper.Chdir(globals.Context)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeach(globals)

	})

	Context("when running help for watch command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "watch", "-h")
			Expect(appHelp).To(ContainSubstring("Watch for changes"))
		})
	})

	Context("when executing watch without pushing a devfile component", func() {
		It("should fail", func() {
			cmpName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, "--context", globals.Context, cmpName)
			output := helper.CmdShouldFail("odo", "watch", "--context", globals.Context)
			Expect(output).To(ContainSubstring("component does not exist. Please use `odo push` to create your component"))
		})
	})

	Context("when executing watch without a valid devfile", func() {
		It("should fail", func() {
			output := helper.CmdShouldFail("odo", "watch", "--devfile", "fake-devfile.yaml")
			Expect(output).To(ContainSubstring("The current directory does not represent an odo component"))
		})
	})

	Context("when executing odo watch with devfile flag without experimental mode", func() {
		It("should fail", func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), globals.Context)
			output := helper.CmdShouldFail("odo", "watch", "--devfile", filepath.Join(globals.Context, "devfile.yaml"))
			Expect(output).To(ContainSubstring("Error: unknown flag: --devfile"))
		})
	})

	Context("when executing odo watch after odo push", func() {
		It("should listen for file changes", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			watchFlag := ""
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing devrun command"},
			}
			// odo watch and validate
			utils.OdoWatch(utils.OdoV1Watch{}, odoV2Watch, namespace, context, watchFlag, cliRunner, "kube")
		})
	})

	Context("when executing odo watch after odo push with custom commands", func() {
		It("should listen for file changes", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--build-command", "build", "--run-command", "run", "--devfile", "devfile.yaml", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			watchFlag := "--build-command build --run-command run"
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing build command", "Executing run command"},
			}
			// odo watch and validate
			utils.OdoWatch(utils.OdoV1Watch{}, odoV2Watch, namespace, context, watchFlag, cliRunner, "kube")
		})
	})
})
