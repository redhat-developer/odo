package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"
)

var _ = Describe("odo devfile watch command tests", func() {
	var namespace string
	var context string
	var cmpName string
	var currentWorkingDirectory string

	var cliRunner helper.CliRunner

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		if os.Getenv("KUBERNETES") == "true" {
			homeDir := helper.GetUserHomeDir()
			_ = helper.CopyKubeConfigFile(filepath.Join(homeDir, ".kube", "config"), filepath.Join(context, "config"))
			namespace = helper.CreateRandNamespace()
			cliRunner = helper.NewKubectlRunner("kubectl")
		} else {
			namespace = helper.CreateRandProject()
			cliRunner = helper.NewOcRunner("oc")
		}
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)

		// Set experimental mode to true
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		if os.Getenv("KUBERNETES") == "true" {
			helper.DeleteNamespace(namespace)
			os.Unsetenv("KUBECONFIG")
		} else {
			helper.DeleteProject(namespace)
		}
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
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
			helper.Chdir(currentWorkingDirectory)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--context", context, cmpName)
			output := helper.CmdShouldFail("odo", "watch", "--context", context)
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
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "false", "-f")
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			output := helper.CmdShouldFail("odo", "watch", "--devfile", filepath.Join(context, "devfile.yaml"))
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
