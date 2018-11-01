package e2e

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const files = "examples/java/"

var _ = Describe("odoJavaE2e", func() {
	var t = "java"
	var projName = fmt.Sprintf("odo-%s", t)
	var gitRepo = "https://github.com/openshift-evangelists/Wild-West-Backend"

	// Create a separate project for Java
	Context("create java project", func() {
		It("should create a new java project", func() {
			session := runCmd("odo project create " + projName)
			Expect(session).To(ContainSubstring(projName))
		})
	})

	// Test Java
	Context("odo component creation", func() {

		It("Should be able to deploy the git repo", func() {

			// Deploy the git repo / wildfly example
			runCmd("odo create wildfly git-test --git " + gitRepo)
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("git-test"))

			// Push changes
			runCmd("odo push")

			// Create a URL
			runCmd("odo url create")
			getRoute := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
			getRoute = strings.TrimSpace(getRoute)

			// Ping said URL
			waitForEqualCmd("curl -s "+getRoute+" | grep 'Welcome to WildFly' | wc -l | tr -d '\n'", "2", 10)

			// Delete the component
			runCmd("odo delete git-test -f")
		})

		It("Should be able to deploy a .jar file", func() {
			runCmd("odo create wildfly jar-test --binary " + files + "wildwest-1.0.jar")
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("jar-test"))

			// Push changes
			runCmd("odo push")

			// Create a URL
			runCmd("odo url create")
			getRoute := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
			getRoute = strings.TrimSpace(getRoute)

			// Ping said URL
			waitForEqualCmd("curl -s "+getRoute+" | grep 'Welcome to WildFly' | wc -l | tr -d '\n'", "2", 10)

			// Delete the component
			runCmd("odo delete jar-test -f")
		})

		It("Should be able to deploy a .war file", func() {
			runCmd("odo create wildfly war-test --binary " + files + "wildwest-1.0.war")
			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("war-test"))

			// Push changes
			runCmd("odo push")

			// Create a URL
			runCmd("odo url create")
			getRoute := runCmd("odo url list  | sed -n '1!p' | awk 'FNR==2 { print $2 }'")
			getRoute = strings.TrimSpace(getRoute)

			// Ping said URL
			waitForEqualCmd("curl -s "+getRoute+" | grep 'Welcome to WildFly' | wc -l | tr -d '\n'", "2", 10)

			// Delete the component
			runCmd("odo delete war-test -f")
		})

	})

	// Delete the project
	Context("java project delete", func() {
		It("should delete java project", func() {
			session := runCmd("odo project delete " + projName)
			Expect(session).To(ContainSubstring(projName))
		})
	})
})
