package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile watch command tests", func() {
	var namespace string
	var context string
	var currentWorkingDirectory string

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		if os.Getenv("KUBERNETES") == "true" {
			namespace = helper.CreateRandNamespace(context)
		} else {
			namespace = helper.CreateRandProject()
		}
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		if os.Getenv("KUBERNETES") == "true" {
			helper.DeleteNamespace(namespace)
		} else {
			helper.DeleteProject(namespace)
		}
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when running help for watch command", func() {
		It("should display the help", func() {
			// Devfile push requires experimental mode to be set
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

			appHelp := helper.CmdShouldPass("odo", "watch", "-h")
			Expect(appHelp).To(ContainSubstring("Watch for changes"))
		})
	})

	Context("when executing watch without pushing a devfile component", func() {
		It("should fail", func() {
			// Devfile push requires experimental mode to be set
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			cmpName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--context", context, cmpName)

			output := helper.CmdShouldFail("odo", "watch", "--devfile", filepath.Join(context, "devfile.yaml"), "--context", context)
			Expect(output).To(ContainSubstring("component does not exist. Please use `odo push` to create your component"))
		})
	})

	Context("when executing watch without a valid devfile", func() {
		It("should fail", func() {
			// Devfile push requires experimental mode to be set
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			output := helper.CmdShouldFail("odo", "watch", "--devfile", "fake-devfile.yaml")
			Expect(output).To(ContainSubstring("The current directory does not represent an odo component"))
		})
	})

	Context("when executing odo watch with devfile flag without experimental mode", func() {
		It("should fail", func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)
			output := helper.CmdShouldFail("odo", "watch", "--devfile", filepath.Join(context, "devfile.yaml"))
			Expect(output).To(ContainSubstring("Error: unknown flag: --devfile"))
		})
	})
})
