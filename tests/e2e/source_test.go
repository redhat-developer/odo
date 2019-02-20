package e2e

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const sourceExamples = "examples/source/"

var _ = Describe("odoSourceE2e", func() {
	const t = "source"
	var projName = fmt.Sprintf("odo-%s", t)

	// Create a separate project for source
	Context("create source project", func() {
		It("should create a new source project", func() {
			session := runCmdShouldPass("odo project create " + projName)
			Expect(session).To(ContainSubstring(projName))
		})
	})

	Context("odo component creation", func() {

		It("Should be able to deploy a python source application", func() {
			waitForCmdOut("odo project set "+projName, 4, false, func(output string) bool {
				return strings.Contains(output, "Already on project : "+projName)
			})
			runCmdShouldPass("odo create python python-app --local " + sourceExamples + "/python/")
			cmpList := runCmdShouldPass("odo list")
			Expect(cmpList).To(ContainSubstring("python-app"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'WSGI' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmdShouldPass("odo delete python-app -f")
		})

		It("Should be able to deploy an openjdk source application", func() {
			importOpenJDKImage()

			runCmdShouldPass("odo create openjdk18 openjdk-app --local " + sourceExamples + "/openjdk/")
			cmpList := runCmdShouldPass("odo list")
			Expect(cmpList).To(ContainSubstring("openjdk-app"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'Javalin' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmdShouldPass("odo delete openjdk-app -f")
		})

		It("Should be able to deploy a wildfly source application", func() {
			runCmdShouldPass("odo create wildfly wildfly-app --local " + sourceExamples + "/wildfly/")
			cmpList := runCmdShouldPass("odo list")
			Expect(cmpList).To(ContainSubstring("wildfly-app"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'Insult' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmdShouldPass("odo delete wildfly-app -f")
		})

		It("Should be able to deploy a nodejs source application", func() {
			runCmdShouldPass("odo create nodejs nodejs-app --local " + sourceExamples + "/nodejs/")
			cmpList := runCmdShouldPass("odo list")
			Expect(cmpList).To(ContainSubstring("nodejs-app"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'node.js' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmdShouldPass("odo delete nodejs-app -f")
		})

		It("Should be able to deploy a dotnet source application", func() {
			runCmdShouldPass("odo create dotnet dotnet-app --local " + sourceExamples + "/dotnet/")
			cmpList := runCmdShouldPass("odo list")
			Expect(cmpList).To(ContainSubstring("dotnet-app"))

			// Push changes
			runCmdShouldPass("odo push")

			// Create a URL
			runCmdShouldPass("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'dotnet' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmdShouldPass("odo delete dotnet-app -f")
		})

	})

	// Delete the project
	Context("source project delete", func() {
		It("should delete source project", func() {
			deleteProject(projName)
		})
	})
})
