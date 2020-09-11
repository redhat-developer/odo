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
	var context, frontendContext, backendContext string
	var oc helper.OcRunner

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
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
			// Check for -vmodule moduleSpec output which is in additional flags
			Expect(appHelp).To(ContainSubstring("--vmodule moduleSpec"))
		})
	})

	Context("when running help for unlink command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "unlink", "-h")
			// Check for -vmodule moduleSpec output which is in additional flags
			Expect(appHelp).To(ContainSubstring("--vmodule moduleSpec"))
		})
	})

	Context("When link between components using wrong port", func() {
		JustBeforeEach(func() {
			frontendContext = helper.CreateNewContext()
			backendContext = helper.CreateNewContext()
		})
		JustAfterEach(func() {
			helper.DeleteDir(frontendContext)
			helper.DeleteDir(backendContext)
		})
		It("should fail", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), frontendContext)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "frontend", "--context", frontendContext, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", frontendContext)
			helper.CopyExample(filepath.Join("source", "python"), backendContext)
			helper.CmdShouldPass("odo", "create", "--s2i", "python", "backend", "--context", backendContext, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", backendContext)
			stdErr := helper.CmdShouldFail("odo", "link", "backend", "--context", frontendContext, "--port", "1234")
			Expect(stdErr).To(ContainSubstring("Unable to properly link to component backend using port 1234"))
		})
	})

	Context("When handling link/unlink between components", func() {
		JustBeforeEach(func() {
			frontendContext = helper.CreateNewContext()
			backendContext = helper.CreateNewContext()
		})
		JustAfterEach(func() {
			helper.DeleteDir(frontendContext)
			helper.DeleteDir(backendContext)
		})
		It("should link the frontend application to the backend and then unlink successfully", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), frontendContext)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "frontend", "--context", frontendContext, "--project", project)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", frontendContext)
			helper.CmdShouldPass("odo", "push", "--context", frontendContext)
			frontendURL := helper.DetermineRouteURL(frontendContext)
			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("source", "openjdk"), backendContext)
			helper.CmdShouldPass("odo", "create", "--s2i", "java:8", "backend", "--project", project, "--context", backendContext)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", backendContext)
			helper.CmdShouldPass("odo", "push", "--context", backendContext)

			// we link
			helper.CmdShouldPass("odo", "link", "backend", "--context", frontendContext, "--port", "8778")
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", "app", project)
			Expect(envFromOutput).To(ContainSubstring("backend"))

			dcName := oc.GetDcName("frontend", project)
			// wait for DeploymentConfig rollout to finish, so we can check if application is successfully running
			oc.WaitForDCRollout(dcName, project, 20*time.Second)
			helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)

			outputErr := helper.CmdShouldFail("odo", "link", "backend", "--port", "8778", "--context", frontendContext)
			Expect(outputErr).To(ContainSubstring("been linked"))
			helper.CmdShouldPass("odo", "unlink", "backend", "--port", "8778", "--context", frontendContext)
		})

		It("Wait till frontend dc rollout properly after linking the frontend application to the backend", func() {
			appName := helper.RandString(7)
			helper.CopyExample(filepath.Join("source", "nodejs"), frontendContext)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "frontend", "--app", appName, "--context", frontendContext, "--project", project)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", frontendContext)
			helper.CmdShouldPass("odo", "push", "--context", frontendContext)
			frontendURL := helper.DetermineRouteURL(frontendContext)

			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("source", "openjdk"), backendContext)
			helper.CmdShouldPass("odo", "create", "--s2i", "java:8", "backend", "--app", appName, "--project", project, "--context", backendContext)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", backendContext)
			helper.CmdShouldPass("odo", "push", "--context", backendContext)

			// link both component and wait till frontend dc rollout properly
			helper.CmdShouldPass("odo", "link", "backend", "--port", "8080", "--wait", "--context", frontendContext)
			helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)

			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", appName, project)
			Expect(envFromOutput).To(ContainSubstring("backend"))

			helper.CmdShouldPass("odo", "unlink", "backend", "--port", "8080", "--context", frontendContext)
		})

		It("should successfully delete component after linked component is deleted", func() {
			// first create the two components
			helper.CopyExample(filepath.Join("source", "nodejs"), frontendContext)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "frontend", "--context", frontendContext, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", frontendContext)
			helper.CopyExample(filepath.Join("source", "nodejs"), backendContext)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "backend", "--context", backendContext, "--project", project)
			helper.CmdShouldPass("odo", "push", "--context", backendContext)

			// now link frontend to the backend component
			helper.CmdShouldPass("odo", "link", "backend", "--port", "8080", "--context", frontendContext)

			// now delete the backend component and then the frontend component
			// this didn't work earlier: https://github.com/openshift/odo/issues/2355
			helper.CmdShouldPass("odo", "delete", "-f", "--context", backendContext)
			helper.CmdShouldPass("odo", "delete", "-f", "--context", frontendContext)
		})
	})
})
