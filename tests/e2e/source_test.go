package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const sourceExamples = "examples/source/"

var _ = Describe("odo-source-e2e", func() {
	const t = "source"
	var projName = fmt.Sprintf("odo-%s", t)

	// Create a separate project for source
	Context("create source project", func() {
		It("should create a new source project", func() {
			session := runCmd("odo project create " + projName)
			Expect(session).To(ContainSubstring(projName))
		})
	})

	Context("odo component creation", func() {

		It("Should be able to deploy a python source application", func() {
			runCmd("odo create python python-app --local " + sourceExamples + "/python/")
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("python-app"))

			// Push changes
			runCmd("odo push")

			// Create a URL
			runCmd("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'WSGI' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmd("odo delete python-app -f")
		})

		It("Should be able to deploy an openjdk source application", func() {
			importOpenJDKImage()

			runCmd("odo create openjdk18 openjdk-app --local " + sourceExamples + "/openjdk/")
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("openjdk-app"))

			// Push changes
			runCmd("odo push")

			// Create a URL
			runCmd("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'Javalin' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmd("odo delete openjdk-app -f")
		})

		It("Should be able to deploy a wildfly source application", func() {
			runCmd("odo create wildfly wildfly-app --local " + sourceExamples + "/wildfly/")
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("wildfly-app"))

			// Push changes
			runCmd("odo push")

			// Create a URL
			runCmd("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'Insult' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmd("odo delete wildfly-app -f")
		})

		It("Should be able to deploy a nodejs source application", func() {
			runCmd("odo create nodejs nodejs-app --local " + sourceExamples + "/nodejs/")
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("nodejs-app"))

			// Push changes
			runCmd("odo push")

			// Create a URL
			runCmd("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'node.js' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmd("odo delete nodejs-app -f")
		})

		It("Should be able to deploy a dotnet source application", func() {
			runCmd("odo create dotnet dotnet-app --local " + sourceExamples + "/dotnet/")
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("dotnet-app"))

			// Push changes
			runCmd("odo push")

			// Create a URL
			runCmd("odo url create --port 8080")
			routeURL := determineRouteURL()

			// Ping said URL
			waitForEqualCmd("curl -s "+routeURL+" | grep 'dotnet' | wc -l | tr -d '\n'", "1", 10)

			// Delete the component
			runCmd("odo delete dotnet-app -f")
		})

	})

	// Delete the project
	Context("source project delete", func() {
		It("should delete source project", func() {
			session := runCmd("odo project delete " + projName + " -f")
			Expect(session).To(ContainSubstring(projName))
		})
	})
})
