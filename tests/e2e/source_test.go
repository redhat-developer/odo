package e2e

import (
	//"fmt"
	//"strings"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/e2e/helper"
)

const sourceExamples = "examples/source/"

var _ = Describe("odoSourceE2e", func() {
	//const t = "source"
	//var projName = fmt.Sprintf("odo-%s", t)

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.AfterFailed()
		os.RemoveAll(".odo")
	})

	// Create a separate project for source
	Context("create source project", func() {
		It("should create a new source project", func() {
			session := runCmdShouldPass("odo project create odo-source -w")
			Expect(session).To(ContainSubstring("odo-source"))
		})
	})

	Context("odo component creation", func() {

		It("Should be able to deploy a python source application", func() {
			// waitForCmdOut("odo project set "+projName, 4, false, func(output string) bool {
			// 	return strings.Contains(output, "Already on project : "+projName)
			// })
			runCmdShouldPass("odo create python python-app --project odo-source --context " + sourceExamples + "python/")
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("python-app"))

			// Push changes
			runCmdShouldPass("odo push --context " + sourceExamples + "python/")

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "WSGI", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo app delete python-app --project odo-source -f")
		})

		It("Should be able to deploy an openjdk source application", func() {
			importOpenJDKImage()

			runCmdShouldPass("odo create java openjdk-app --project odo-source --context " + sourceExamples + "openjdk/")
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("openjdk-app"))

			// Push changes
			runCmdShouldPass("odo push --context " + sourceExamples + "openjdk/")

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "Javalin", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo app delete openjdk-app --project odo-source -f")
		})

		It("Should be able to deploy a wildfly source application", func() {
			runCmdShouldPass("odo create wildfly wildfly-app --project odo-source --context " + sourceExamples + "wildfly/")
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("wildfly-app"))

			// Push changes
			runCmdShouldPass("odo push --context " + sourceExamples + "wildfly/")

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "Insult", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo app delete wildfly-app --project odo-source  -f")
		})

		It("Should be able to deploy a nodejs source application", func() {
			runCmdShouldPass("odo create nodejs nodejs-app --project odo-source --context " + sourceExamples + "nodejs/")
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("nodejs-app"))

			// Push changes
			runCmdShouldPass("odo push --context " + sourceExamples + "nodejs/")

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "node.js", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo app delete nodejs-app --project odo-source  -f")
		})

		It("Should be able to deploy a dotnet source application", func() {
			runCmdShouldPass("odo create dotnet:2.0 dotnet-app --project odo-source --context " + sourceExamples + "dotnet/")
			//cmpList := runCmdShouldPass("odo list")
			//Expect(cmpList).To(ContainSubstring("dotnet-app"))

			// Push changes
			runCmdShouldPass("odo push --context " + sourceExamples + "dotnet/")

			// Create a URL
			//runCmdShouldPass("odo url create --port 8080")
			//routeURL := determineRouteURL()

			// Ping said URL
			//responseStringMatchStatus := matchResponseSubString(routeURL, "dotnet", 30, 1)
			//Expect(responseStringMatchStatus).Should(BeTrue())

			// Delete the component
			runCmdShouldPass("odo app delete dotnet-app --project odo-source -f")
		})

	})

	// Delete the project
	Context("source project delete", func() {
		It("should delete source project", func() {
			ocDeleteProject("odo-source")
			runCmdShouldPass("rm -rf .odo")

		})
	})
})
