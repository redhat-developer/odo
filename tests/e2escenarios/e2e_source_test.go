package e2escenarios

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo source e2e tests", func() {
	var oc helper.OcRunner
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		// initialize oc runner
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("odo component creation from source", func() {
		JustBeforeEach(func() {
			if runtime.GOARCH == "s390x" {
				Skip("Skipping test because there is no supported builder image.")
			}
		})

		It("Should be able to deploy a wildfly source application", func() {
			helper.CopyExample(filepath.Join("source", "wildfly"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "wildfly", "wildfly-app", "--project",
				commonVar.Project, "--context", commonVar.Context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			cmpList := helper.CmdShouldPass("odo", "list", "--context", commonVar.Context)
			Expect(cmpList).To(ContainSubstring("wildfly-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Insult", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", commonVar.Project, "-f")
		})

		It("Should be able to deploy a dotnet source application", func() {
			oc.ImportDotnet20IS(commonVar.Project)
			helper.CopyExample(filepath.Join("source", "dotnet"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "dotnet:2.0", "dotnet-app", "--project",
				commonVar.Project, "--context", commonVar.Context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			cmpList := helper.CmdShouldPass("odo", "list", "--context", commonVar.Context)
			Expect(cmpList).To(ContainSubstring("dotnet-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "dotnet", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", commonVar.Project, "-f")
		})
	})

	Context("odo component creation", func() {

		It("Should be able to deploy a python source application", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "python", "python-app", "--project",
				commonVar.Project, "--context", commonVar.Context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			cmpList := helper.CmdShouldPass("odo", "list", "--context", commonVar.Context)
			Expect(cmpList).To(ContainSubstring("python-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "WSGI", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", commonVar.Project, "-f")
		})

		It("Should be able to deploy an openjdk source application", func() {
			oc.ImportJavaIS(commonVar.Project)
			helper.CopyExample(filepath.Join("source", "openjdk"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "java:8", "openjdk-app", "--project",
				commonVar.Project, "--context", commonVar.Context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context, "-v", "4")
			cmpList := helper.CmdShouldPass("odo", "list", "--context", commonVar.Context)
			Expect(cmpList).To(ContainSubstring("openjdk-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Javalin", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", commonVar.Project, "-f")
		})

		It("Should be able to deploy a nodejs source application", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs-app", "--project",
				commonVar.Project, "--context", commonVar.Context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			cmpList := helper.CmdShouldPass("odo", "list", "--context", commonVar.Context)
			Expect(cmpList).To(ContainSubstring("nodejs-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "node.js", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", commonVar.Project, "-f")
		})

	})
})
