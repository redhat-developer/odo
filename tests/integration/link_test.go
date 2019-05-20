package integration

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odoLinkE2e", func() {

	//new clean project and context for each test
	var project string
	var context1, context2, context3, context4 string
	var originalDir string

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		os.RemoveAll(".odo")
	})

	Context("odo link/unlink handling between components", func() {
		JustBeforeEach(func() {
			context1 = helper.CreateNewContext()
			context2 = helper.CreateNewContext()
		})
		JustAfterEach(func() {
			helper.DeleteDir(context1)
			helper.DeleteDir(context2)
		})
		It("reports error when using wrong port", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context1)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2)
			helper.CmdShouldPass("odo", "push", "--context", context2)
			stdErr := helper.CmdShouldFail("odo", "link", "backend", "--component", "frontend", "--context", context2, "--port", "1234")
			Expect(stdErr).To(ContainSubstring("Unable to properly link to component backend using port 1234"))
		})
		It("link the frontend application to the backend", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context1)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2)
			helper.CmdShouldPass("odo", "push", "--context", context2)
			helper.CmdShouldPass("odo", "link", "backend", "--component", "frontend", "--context", context2)
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", "app")
			Expect(envFromOutput).To(ContainSubstring("backend"))
			outputErr := helper.CmdShouldFail("odo", "link", "backend", "--component", "frontend", "--context", context2)
			Expect(outputErr).To(ContainSubstring("been linked"))
			helper.CmdShouldPass("odo", "unlink", "backend", "--component", "frontend", "--context", context2)
		})
	})

	Context("odo link/unlink handling between components and service", func() {
		JustBeforeEach(func() {
			context3 = helper.CreateNewContext()
			context4 = helper.CreateNewContext()
			originalDir = helper.Getwd()
		})
		JustAfterEach(func() {
			helper.Chdir(originalDir)
			helper.DeleteDir(context3)
			helper.DeleteDir(context4)
		})
		It("should link backend to service", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context3)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context3)
			helper.CmdShouldPass("odo", "push", "--context", context3)
			helper.CopyExample(filepath.Join("source", "python"), context4)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context4)
			helper.CmdShouldPass("odo", "push", "--context", context4)
			helper.CmdShouldPass("odo", "link", "backend", "--component", "frontend", "--context", context4)
			helper.Chdir(context4)
			helper.CmdShouldPass("odo", "service", "create", "mysql-persistent")

			ocArgs := []string{"get", "serviceinstance", "-o", "name"}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "mysql-persistent")
			})
			helper.CmdShouldPass("odo", "link", "mysql-persistent", "--wait-for-target", "--component", "backend")
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("backend", "app")
			Expect(envFromOutput).To(ContainSubstring("mysql-persistent"))

			outputErr := helper.CmdShouldFail("odo", "link", "mysql-persistent", "--component", "backend", "--context", context4)
			Expect(outputErr).To(ContainSubstring("been linked"))
		})

		It("Delete service and unlink the backend from the frontend", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context3)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context3)
			helper.CmdShouldPass("odo", "push", "--context", context3)
			helper.CopyExample(filepath.Join("source", "python"), context4)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context4)
			helper.CmdShouldPass("odo", "push", "--context", context4)
			helper.CmdShouldPass("odo", "link", "backend", "--component", "frontend", "--context", context4)
			helper.Chdir(context4)
			helper.CmdShouldPass("odo", "service", "create", "mysql-persistent")

			ocArgs := []string{"get", "serviceinstance", "-o", "name"}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "mysql-persistent")
			})
			helper.CmdShouldPass("odo", "service", "delete", "mysql-persistent", "-f")
			// ensure that the backend no longer has an envFrom value
			backendEnvFromOutput := oc.GetEnvFromEntry("backend", "app")
			Expect(backendEnvFromOutput).To(Equal("''"))
			// ensure that the frontend envFrom was not changed
			frontEndEnvFromOutput := oc.GetEnvFromEntry("frontend", "app")
			Expect(frontEndEnvFromOutput).To(ContainSubstring("backend"))
			helper.CmdShouldPass("odo", "unlink", "backend", "--component", "frontend")
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", "app")
			Expect(envFromOutput).To(Equal("''"))
		})
	})
})
