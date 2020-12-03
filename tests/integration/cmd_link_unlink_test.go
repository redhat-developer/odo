package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link and unlink command tests", func() {

	/*
		Uncomment when we uncomment the test specs
		var frontendContext, backendContext string
		var oc helper.OcRunner
	*/

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		// oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("when running help for link and unlink command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "link", "-h")
			helper.MatchAllInOutput(appHelp, []string{"Link component to a service ", "backed by an Operator or Service Catalog", "or component", "works only with s2i components"})
			appHelp = helper.CmdShouldPass("odo", "unlink", "-h")
			Expect(appHelp).To(ContainSubstring("Unlink component or service from a component"))
		})
	})

	/*
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
				helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "frontend", "--context", frontendContext, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", frontendContext)
				helper.CopyExample(filepath.Join("source", "python"), backendContext)
				helper.CmdShouldPass("odo", "create", "--s2i", "python", "backend", "--context", backendContext, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", backendContext)
				stdErr := helper.CmdShouldFail("odo", "link", "backend", "--context", frontendContext, "--port", "1234")
				Expect(stdErr).To(ContainSubstring("Unable to properly link to component backend using port 1234"))
			})
		})
	*/

	/*
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
				helper.CmdShouldPass("odo", "create", "nodejs", "--s2i", "frontend", "--context", frontendContext, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", frontendContext)
				helper.CmdShouldPass("odo", "push", "--context", frontendContext)
				frontendURL := helper.DetermineRouteURL(frontendContext)
				oc.ImportJavaIS(commonVar.Project)
				helper.CopyExample(filepath.Join("source", "openjdk"), backendContext)
				helper.CmdShouldPass("odo", "create", "java:8", "--s2i", "backend", "--project", commonVar.Project, "--context", backendContext)
				helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", backendContext)
				helper.CmdShouldPass("odo", "push", "--context", backendContext)

				// we link
				helper.CmdShouldPass("odo", "link", "backend", "--context", frontendContext, "--port", "8778")
				// ensure that the proper envFrom entry was created
				envFromOutput := oc.GetEnvFromEntry("frontend", "app", commonVar.Project)
				Expect(envFromOutput).To(ContainSubstring("backend"))

				dcName := oc.GetDcName("frontend", commonVar.Project)
				// wait for DeploymentConfig rollout to finish, so we can check if application is successfully running
				oc.WaitForDCRollout(dcName, commonVar.Project, 20*time.Second)
				helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)

				outputErr := helper.CmdShouldFail("odo", "link", "backend", "--port", "8778", "--context", frontendContext)
				Expect(outputErr).To(ContainSubstring("been linked"))
				helper.CmdShouldPass("odo", "unlink", "backend", "--port", "8778", "--context", frontendContext)
			})

			It("Wait till frontend dc rollout properly after linking the frontend application to the backend", func() {
				appName := helper.RandString(7)
				helper.CopyExample(filepath.Join("source", "nodejs"), frontendContext)
				helper.CmdShouldPass("odo", "create", "nodejs", "--s2i", "frontend", "--app", appName, "--context", frontendContext, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", frontendContext)
				helper.CmdShouldPass("odo", "push", "--context", frontendContext)
				frontendURL := helper.DetermineRouteURL(frontendContext)

				oc.ImportJavaIS(commonVar.Project)
				helper.CopyExample(filepath.Join("source", "openjdk"), backendContext)
				helper.CmdShouldPass("odo", "create", "java:8", "--s2i", "backend", "--app", appName, "--project", commonVar.Project, "--context", backendContext)
				helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", backendContext)
				helper.CmdShouldPass("odo", "push", "--context", backendContext)

				// link both component and wait till frontend dc rollout properly
				helper.CmdShouldPass("odo", "link", "backend", "--port", "8080", "--wait", "--context", frontendContext)
				helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)

				// ensure that the proper envFrom entry was created
				envFromOutput := oc.GetEnvFromEntry("frontend", appName, commonVar.Project)
				Expect(envFromOutput).To(ContainSubstring("backend"))

				helper.CmdShouldPass("odo", "unlink", "backend", "--port", "8080", "--context", frontendContext)
			})

			It("should successfully delete component after linked component is deleted", func() {
				// first create the two components
				helper.CopyExample(filepath.Join("source", "nodejs"), frontendContext)
				helper.CmdShouldPass("odo", "create", "nodejs", "--s2i", "frontend", "--context", frontendContext, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", frontendContext)
				helper.CopyExample(filepath.Join("source", "nodejs"), backendContext)
				helper.CmdShouldPass("odo", "create", "nodejs", "--s2i", "backend", "--context", backendContext, "--project", commonVar.Project)
				helper.CmdShouldPass("odo", "push", "--context", backendContext)

				// now link frontend to the backend component
				helper.CmdShouldPass("odo", "link", "backend", "--port", "8080", "--context", frontendContext)

				// now delete the backend component and then the frontend component
				// this didn't work earlier: https://github.com/openshift/odo/issues/2355
				helper.CmdShouldPass("odo", "delete", "-f", "--context", backendContext)
				helper.CmdShouldPass("odo", "delete", "-f", "--context", frontendContext)
			})
		})
	*/
})
