package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link and unlink command tests", func() {

	var frontendContext, backendContext string
	var oc helper.OcRunner

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("when running help for link and unlink command", func() {
		It("should display the help", func() {
			appHelp := helper.Cmd("odo", "link", "-h").ShouldPass().Out()
			helper.MatchAllInOutput(appHelp, []string{"Link component to a service ", "backed by an Operator or Service Catalog", "or component"})
			appHelp = helper.Cmd("odo", "unlink", "-h").ShouldPass().Out()
			Expect(appHelp).To(ContainSubstring("Unlink component or service from a component"))
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
			helper.Cmd("odo", "create", "nodejs", "frontend", "--context", frontendContext, "--project", commonVar.Project).ShouldPass()
			helper.Cmd("odo", "url", "create", "--port", "8080", "--context", frontendContext).ShouldPass()
			helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()
			frontendURL := helper.DetermineRouteURL(frontendContext)
			helper.CopyExample(filepath.Join("source", "python"), backendContext)
			helper.Cmd("odo", "create", "python", "backend", "--project", commonVar.Project, "--context", backendContext).ShouldPass()
			helper.Cmd("odo", "url", "create", "--port", "8080", "--context", backendContext).ShouldPass()
			helper.Cmd("odo", "push", "--context", backendContext).ShouldPass()

			// we link
			helper.Cmd("odo", "link", "backend", "--context", frontendContext).ShouldPass()
			helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()
			// ensure that the proper envFrom entry was created
			envFromOutput := oc.GetEnvFromEntry("frontend", "app", commonVar.Project, "deployment")
			Expect(envFromOutput).To(ContainSubstring("backend"))

			helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)
			outputErr := helper.Cmd("odo", "link", "backend", "--context", frontendContext).ShouldFail().Err()
			Expect(outputErr).To(ContainSubstring("already linked"))
			helper.Cmd("odo", "unlink", "backend", "--context", frontendContext).ShouldPass()
			helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()
		})

		It("should successfully delete component after linked component is deleted", func() {
			// first create the two components
			helper.CopyExample(filepath.Join("source", "nodejs"), frontendContext)
			helper.Cmd("odo", "create", "nodejs", "frontend", "--context", frontendContext, "--project", commonVar.Project).ShouldPass()
			helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()
			helper.CopyExample(filepath.Join("source", "python"), backendContext)
			helper.Cmd("odo", "create", "python", "backend", "--context", backendContext, "--project", commonVar.Project).ShouldPass()
			helper.Cmd("odo", "push", "--context", backendContext).ShouldPass()

			// now link frontend to the backend component
			helper.Cmd("odo", "link", "backend", "--context", frontendContext).ShouldPass()
			helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()

			// now delete the backend component and then the frontend component
			// this didn't work earlier: https://github.com/openshift/odo/issues/2355
			helper.Cmd("odo", "delete", "-f", "--context", backendContext).ShouldPass()
			helper.Cmd("odo", "delete", "-f", "--context", frontendContext).ShouldPass()
		})
	})

})
