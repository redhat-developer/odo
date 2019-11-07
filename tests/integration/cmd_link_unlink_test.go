package integration

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link and unlink command tests", func() {

	//new clean project and context for each test
	var project string
	var context, context1, context2 string
	var oc helper.OcRunner

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		oc = helper.NewOcRunner("oc")
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when running help for link command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "link", "-h")
			Expect(appHelp).To(ContainSubstring("Link component to a service or component"))
		})
	})

	Context("when running help for unlink command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "unlink", "-h")
			Expect(appHelp).To(ContainSubstring("Unlink component or service from a component"))
		})
	})

	Context("When link between components using wrong port", func() {
		JustBeforeEach(func() {
			context1 = helper.CreateNewContext()
			context2 = helper.CreateNewContext()
		})
		JustAfterEach(func() {
			helper.DeleteDir(context1)
			helper.DeleteDir(context2)
		})
		It("should fail", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context1)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", context2)
			stdErr := helper.CmdShouldFail("odo", "link", "backend", "--component", "frontend", "--project", project, "--context", context2, "--port", "1234")
			Expect(stdErr).To(ContainSubstring("Unable to properly link to component backend using port 1234"))
		})
	})

	Context("When handling link/unlink between components", func() {
		JustBeforeEach(func() {
			context1 = helper.CreateNewContext()
			context2 = helper.CreateNewContext()
		})
		JustAfterEach(func() {
			helper.DeleteDir(context1)
			helper.DeleteDir(context2)
		})
		It("should link the frontend application to the backend and then unlink successfully", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context1)
			helper.CmdShouldPass("odo", "create", "nodejs", "frontend", "--context", context1, "--project", project)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context1)
			helper.CmdShouldPass("odo", "push", "--context", context1)
			frontendURL := helper.DetermineRouteURL(context1)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.CmdShouldPass("odo", "create", "python", "backend", "--context", context2, "--project", project)
			helper.CmdShouldPass("odo", "url", "create", "--context", context2)
			helper.CmdShouldPass("odo", "push", "--context", context2)

			helper.CmdShouldPass("odo", "link", "backend", "--component", "frontend", "--project", project, "--context", context2)
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", "app", project)
			Expect(envFromOutput).To(ContainSubstring("backend"))

			dcName := oc.GetDcName("frontend", project)
			// wait for DeploymentConfig rollout to finish, so we can check if application is successfully running
			oc.WaitForDCRollout(dcName, project, 20*time.Second)
			helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)

			outputErr := helper.CmdShouldFail("odo", "link", "backend", "--component", "frontend", "--project", project, "--context", context2)
			Expect(outputErr).To(ContainSubstring("been linked"))
			helper.CmdShouldPass("odo", "unlink", "backend", "--component", "frontend", "--project", project, "--context", context2)
		})
	})
})
