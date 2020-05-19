package e2escenarios

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo java e2e tests", func() {
	var globals helper.Globals
	var oc helper.OcRunner

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		globals = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeach(globals)
	})

	// contains a minimal javaee app
	const warGitRepo = "https://github.com/lordofthejars/book-insultapp"

	// contains a minimal javalin app
	const jarGitRepo = "https://github.com/geoand/javalin-helloworld"

	// Test Java
	Context("odo component creation", func() {
		It("Should be able to deploy a git repo that contains a wildfly application without wait flag", func() {
			helper.CmdShouldPass("odo", "create", "wildfly", "wo-wait-javaee-git-test", "--project",
				globals.Project, "--ref", "master", "--git", warGitRepo, "--context", globals.Context)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "gitrepo", "--port", "8080", "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "-v", "4", "--context", globals.Context)
			routeURL := helper.DetermineRouteURL(globals.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Insult", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "wo-wait-javaee-git-test", "-f", "--context", globals.Context)
		})

		It("Should be able to deploy a .war file using wildfly", func() {
			helper.CopyExample(filepath.Join("binary", "java", "wildfly"), globals.Context)
			helper.CmdShouldPass("odo", "create", "wildfly", "javaee-war-test", "--project",
				globals.Project, "--binary", filepath.Join(globals.Context, "ROOT.war"), "--context", globals.Context)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "warfile", "--port", "8080", "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			routeURL := helper.DetermineRouteURL(globals.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Sample", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "javaee-war-test", "-f", "--context", globals.Context)
		})

		It("Should be able to deploy a git repo that contains a java uberjar application using openjdk", func() {
			oc.ImportJavaIS(globals.Project)

			// Deploy the git repo / wildfly example
			helper.CmdShouldPass("odo", "create", "java:8", "uberjar-git-test", "--project",
				globals.Project, "--ref", "master", "--git", jarGitRepo, "--context", globals.Context)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "uberjar", "--port", "8080", "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			routeURL := helper.DetermineRouteURL(globals.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Hello World", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "uberjar-git-test", "-f", "--context", globals.Context)
		})

		It("Should be able to deploy a spring boot uberjar file using openjdk", func() {
			oc.ImportJavaIS(globals.Project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), globals.Context)

			helper.CmdShouldPass("odo", "create", "java:8", "sb-jar-test", "--project",
				globals.Project, "--binary", filepath.Join(globals.Context, "sb.jar"), "--context", globals.Context)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "uberjaropenjdk", "--port", "8080", "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			routeURL := helper.DetermineRouteURL(globals.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "HTTP Booster", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "sb-jar-test", "-f", "--context", globals.Context)
		})

	})
})
