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
			if runtime.GOARCH == "s390x" || runtime.GOARCH == "ppc64le" {
				Skip("Skipping test because there is no supported builder image.")
			}
		})

		// issue https://github.com/openshift/odo/issues/4623
		// It("Should be able to deploy a wildfly source application", func() {
		// 	helper.CopyExample(filepath.Join("source", "wildfly"), commonVar.Context)
		// 	helper.Cmd("odo", "create", "--s2i", "wildfly", "wildfly-app", "--project",
		// 		commonVar.Project, "--context", commonVar.Context).ShouldPass()

		// 	// Push changes
		// 	helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
		// 	cmpList := helper.Cmd("odo", "list", "--context", commonVar.Context).ShouldPass().Out()
		// 	Expect(cmpList).To(ContainSubstring("wildfly-app"))

		// 	// Create a URL
		// 	helper.Cmd("odo", "url", "create", "--port", "8080", "--context", commonVar.Context).ShouldPass()
		// 	helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
		// 	routeURL := helper.DetermineRouteURL(commonVar.Context)

		// 	// Ping said URL
		// 	helper.HttpWaitFor(routeURL, "Insult", 30, 1)

		// 	// Delete the component
		// 	helper.Cmd("odo", "app", "delete", "app", "--project", commonVar.Project, "-f").ShouldPass()
		// })

		// issue https://github.com/openshift/odo/issues/4623
		// It("Should be able to deploy a dotnet source application", func() {
		// 	oc.ImportDotnet20IS(commonVar.Project)
		// 	helper.CopyExample(filepath.Join("source", "dotnet"), commonVar.Context)
		// 	helper.Cmd("odo", "create", "--s2i", "dotnet:2.0", "dotnet-app", "--project",
		// 		commonVar.Project, "--context", commonVar.Context).ShouldPass()

		// 	// Push changes
		// 	helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
		// 	cmpList := helper.Cmd("odo", "list", "--context", commonVar.Context).ShouldPass().Out()
		// 	Expect(cmpList).To(ContainSubstring("dotnet-app"))

		// 	// Create a URL
		// 	helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
		// 	routeURL := helper.DetermineRouteURL(commonVar.Context)

		// 	// Ping said URL
		// 	helper.HttpWaitFor(routeURL, "dotnet", 30, 1)

		// 	// Delete the component
		// 	helper.Cmd("odo", "app", "delete", "app", "--project", commonVar.Project, "-f").ShouldPass()
		// })
	})

	Context("odo component creation", func() {

		It("Should be able to deploy an openjdk source application", func() {
			oc.ImportJavaIS(commonVar.Project)
			helper.CopyExample(filepath.Join("source", "openjdk"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "java:8", "openjdk-app", "--project",
				commonVar.Project, "--context", commonVar.Context).ShouldPass()

			// Push changes
			helper.Cmd("odo", "push", "--context", commonVar.Context, "-v", "4").ShouldPass()
			cmpList := helper.Cmd("odo", "list", "--context", commonVar.Context).ShouldPass().Out()
			Expect(cmpList).To(ContainSubstring("openjdk-app"))

			// Create a URL
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Javalin", 30, 1)

			// Delete the component
			helper.Cmd("odo", "app", "delete", "app", "--project", commonVar.Project, "-f").ShouldPass()
		})

		It("Should be able to deploy a nodejs source application", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs-app", "--project",
				commonVar.Project, "--context", commonVar.Context).ShouldPass()

			// Push changes
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			cmpList := helper.Cmd("odo", "list", "--context", commonVar.Context).ShouldPass().Out()
			Expect(cmpList).To(ContainSubstring("nodejs-app"))

			// Create a URL
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "node.js", 30, 1)

			// Delete the component
			helper.Cmd("odo", "app", "delete", "app", "--project", commonVar.Project, "-f").ShouldPass()
		})

	})

	Context("odo component creation, Skip tests for ppc64le arch", func() {
		JustBeforeEach(func() {
			if runtime.GOARCH == "ppc64le" {
				Skip("Skipping test on Power because python is not fully supported by odo, and it is not guaranteed to work.")
			}
		})

		It("Should be able to deploy a python source application", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "python", "python-app", "--project",
				commonVar.Project, "--context", commonVar.Context).ShouldPass()

			// Push changes
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			cmpList := helper.Cmd("odo", "list", "--context", commonVar.Context).ShouldPass().Out()
			Expect(cmpList).To(ContainSubstring("python-app"))

			// Create a URL
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "WSGI", 30, 1)

			// Delete the component
			helper.Cmd("odo", "app", "delete", "app", "--project", commonVar.Project, "-f").ShouldPass()
		})
	})
})
