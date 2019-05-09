package integration

import (
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odoSourceE2e", func() {
	var project string
	var context string

	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
	})

	var _ = AfterEach(func() {
		helper.DeleteProject(project)
	})

	Context("odo component creation", func() {

		It("Should be able to deploy a python source application", func() {
			// waitForCmdOut("odo project set "+projName, 4, false, func(output string) bool {
			// 	return strings.Contains(output, "Already on project : "+projName)
			// })
			helper.CopyExample(filepath.Join("source", "python"), context)
			helper.CmdShouldPass("odo", "create", "python", "python-app", "--project",
				project, "--context", context)
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("python-app"))

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context)

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "WSGI", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})

		//https://github.com/openshift/odo/issues/1698
		/*It("Should be able to deploy an openjdk source application", func() {
			oc.ImportJavaIsToNspace(project)
			helper.CopyExample(filepath.Join("source", "openjdk"), context)
			helper.CmdShouldPass("odo", "create", "java", "openjdk-app", "--project",
				project, "--context", context)
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("openjdk-app"))

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context, "-v", "4")

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "Javalin", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})*/

		It("Should be able to deploy a wildfly source application", func() {
			helper.CopyExample(filepath.Join("source", "wildfly"), context)
			helper.CmdShouldPass("odo", "create", "wildfly", "wildfly-app", "--project",
				project, "--context", context)
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("wildfly-app"))

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context)

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "Insult", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})

		It("Should be able to deploy a nodejs source application", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs-app", "--project",
				project, "--context", context)
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("nodejs-app"))

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context)

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "node.js", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})

		It("Should be able to deploy a dotnet source application", func() {
			helper.CopyExample(filepath.Join("source", "dotnet"), context)
			helper.CmdShouldPass("odo", "create", "dotnet:2.0", "dotnet-app", "--project",
				project, "--context", context)
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("dotnet-app"))

			// Push changes
			helper.CmdShouldPass("odo", "push", "--context", context)

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "dotnet", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			helper.CmdShouldPass("odo", "app", "delete", "app", "--project", project, "-f")
		})
	})
})
