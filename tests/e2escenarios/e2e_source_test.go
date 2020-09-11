package e2escenarios

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo source e2e tests", func() {
	var project string
	var context string
	var oc helper.OcRunner

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("odo component creation from source", func() {
		JustBeforeEach(func() {
			if runtime.GOARCH == "s390x" {
				Skip("Skipping test because there is no supported builder image.")
			}
		})

		It("Should be able to deploy a wildfly source application", func() {
			helper.CopyExample(filepath.Join("source", "wildfly"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "wildfly", "wildfly-app", "--project",
				project, "--context", context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context)
			cmpList := helper.CmdShouldPass("odo", "list", "--context", context)
			Expect(cmpList).To(ContainSubstring("wildfly-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Insult", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})

		It("Should be able to deploy a dotnet source application", func() {
			oc.ImportDotnet20IS(project)
			helper.CopyExample(filepath.Join("source", "dotnet"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "dotnet:2.0", "dotnet-app", "--project",
				project, "--context", context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context)
			cmpList := helper.CmdShouldPass("odo", "list", "--context", context)
			Expect(cmpList).To(ContainSubstring("dotnet-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "dotnet", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})
	})

	Context("odo component creation", func() {

		It("Should be able to deploy a python source application", func() {
			helper.CopyExample(filepath.Join("source", "python"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "python", "python-app", "--project",
				project, "--context", context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context)
			cmpList := helper.CmdShouldPass("odo", "list", "--context", context)
			Expect(cmpList).To(ContainSubstring("python-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "WSGI", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})

		It("Should be able to deploy an openjdk source application", func() {
			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("source", "openjdk"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "java:8", "openjdk-app", "--project",
				project, "--context", context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context, "-v", "4")
			cmpList := helper.CmdShouldPass("odo", "list", "--context", context)
			Expect(cmpList).To(ContainSubstring("openjdk-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Javalin", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})

		It("Should be able to deploy a nodejs source application", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "nodejs-app", "--project",
				project, "--context", context)

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context)
			cmpList := helper.CmdShouldPass("odo", "list", "--context", context)
			Expect(cmpList).To(ContainSubstring("nodejs-app"))

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "node.js", 30, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})

	})
})
