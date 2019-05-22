package integration

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

const javaFiles = "examples/binary/java/"

var _ = Describe("odoJavaE2e", func() {
	var project string
	var oc helper.OcRunner
	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		os.RemoveAll(".odo")
	})

	// contains a minimal javaee app
	const warGitRepo = "https://github.com/lordofthejars/book-insultapp"

	// contains a minimal javalin app
	const jarGitRepo = "https://github.com/geoand/javalin-helloworld"

	// Test Java
	Context("odo component creation", func() {
		It("Should be able to deploy a git repo that contains a wildfly application without wait flag", func() {
			helper.CmdShouldPass("odo", "create", "wildfly", "wo-wait-javaee-git-test", "--project",
				project, "--ref", "master", "--git", warGitRepo)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "gitrepo", "--port", "8080")
			helper.CmdShouldPass("odo", "push", "-v", "4")
			routeURL := helper.DetermineRouteURL("")

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Insult", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "wo-wait-javaee-git-test", "-f")
		})

		It("Should be able to deploy a .war file using wildfly", func() {
			helper.CmdShouldPass("odo", "create", "wildfly", "javaee-war-test", "--project",
				project, "--binary", "../examples/binary/java/wildfly/ROOT.war")

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "warfile", "--port", "8080")
			helper.CmdShouldPass("odo", "push")
			routeURL := helper.DetermineRouteURL("")

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Sample", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "javaee-war-test", "-f")
		})

		// https://github.com/openshift/odo/pull/1634
		/*It("Should be able to deploy a git repo that contains a java uberjar application using openjdk", func() {
			oc.ImportJavaIsToNspace(project)

			// Deploy the git repo / wildfly example
			helper.CmdShouldPass("odo", "create", "java", "uberjar-git-test", "--project",
				project, "--ref", "master", "--git", jarGitRepo)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "uberjar", "--port", "8080")
			helper.CmdShouldPass("odo", "push")
			routeURL := helper.DetermineRouteURL("")

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Hello World", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "uberjar-git-test", "-f")
		})*/

		It("Should be able to deploy a spring boot uberjar file using openjdk", func() {
			oc.ImportJavaIsToNspace(project)

			helper.CmdShouldPass("odo", "create", "java", "sb-jar-test", "--project",
				project, "--binary", "../examples/binary/java/openjdk/sb.jar")

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "uberjaropenjdk", "--port", "8080")
			helper.CmdShouldPass("odo", "push")
			routeURL := helper.DetermineRouteURL("")

			// Ping said URL
			helper.HttpWaitFor(routeURL, "HTTP Booster", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "sb-jar-test", "-f")
		})

	})
})
