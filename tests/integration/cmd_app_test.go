package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"path/filepath"
	"time"
)

var _ = Describe("odoCmdAppE2e", func() {
	var project string
	var context string
	var originalDir string

	appName := "app"
	cmpName := "nodejs"

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
		originalDir = helper.Getwd()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
	})

	Context("App test", func() {
		It("should pass inside a odo directory without app parameters", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			// changing directory to the context directory
			helper.Chdir(context)
			appListOutput := helper.CmdShouldPass("odo", "app", "list")
			Expect(appListOutput).To(ContainSubstring(appName))
			helper.CmdShouldPass("odo", "app", "describe")
			helper.CmdShouldPass("odo", "app", "delete", "-f")
			helper.Chdir(originalDir)
		})

		It("should fail outside a odo directory without app parameters(except the list command)", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			// list should pass as the project exists
			appListOutput := helper.CmdShouldPass("odo", "app", "list")
			Expect(appListOutput).To(ContainSubstring(appName))
			helper.CmdShouldFail("odo", "app", "describe")
			helper.CmdShouldFail("odo", "app", "delete")
		})

		It("should pass outside a odo directory with app parameters", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			appListOutput := helper.CmdShouldPass("odo", "app", "list")
			Expect(appListOutput).To(ContainSubstring(appName))
			helper.CmdShouldPass("odo", "app", "describe", appName)
			helper.CmdShouldPass("odo", "app", "delete", appName, "-f")
		})
	})

})
