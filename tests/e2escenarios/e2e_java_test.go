package e2escenarios

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo java e2e tests", func() {
	var oc helper.OcRunner
	var commonVar helper.CommonVar

	// contains a minimal javaee app
	const warGitRepo = "https://github.com/lordofthejars/book-insultapp"

	// contains a minimal javalin app
	const jarGitRepo = "https://github.com/geoand/javalin-helloworld"

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		// initialize oc runner
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
		oc.AddSecret(commonVar)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	// Test wildfly
	Context("odo wildfly component creation ", func() {
		JustBeforeEach(func() {
			if runtime.GOARCH == "s390x" || runtime.GOARCH == "ppc64le" {
				Skip("Skipping test because there is no supported wildfly builder image.")
			}
		})

		It("Should be able to deploy a git repo that contains a wildfly application without wait flag", func() {
			helper.Cmd("odo", "create", "--s2i", "wildfly", "wo-wait-javaee-git-test", "--project",
				commonVar.Project, "--ref", "master", "--git", warGitRepo, "--context", commonVar.Context).ShouldPass()

			// Create a URL
			helper.Cmd("odo", "url", "create", "gitrepo", "--port", "8080", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "-v", "4", "--context", commonVar.Context).ShouldPass()
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Insult", 90, 1)

			// Delete the component
			helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()
		})
	})
	// Test Java
	Context("odo component creation", func() {
		It("Should be able to deploy a .war file using wildfly", func() {
			helper.CopyExample(filepath.Join("binary", "java", "wildfly"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "wildfly", "javaee-war-test", "--project",
				commonVar.Project, "--binary", filepath.Join(commonVar.Context, "ROOT.war"), "--context", commonVar.Context).ShouldPass()

			// Create a URL
			helper.Cmd("odo", "url", "create", "warfile", "--port", "8080", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Sample", 90, 1)

			// Delete the component
			helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()
		})

		It("Should be able to deploy a git repo that contains a java uberjar application using openjdk", func() {
			oc.ImportJavaIS(commonVar.Project)

			// Deploy the git repo / wildfly example
			helper.Cmd("odo", "create", "--s2i", "java:8", "uberjar-git-test", "--project",
				commonVar.Project, "--ref", "master", "--git", jarGitRepo, "--context", commonVar.Context).ShouldPass()

			// Create a URL
			helper.Cmd("odo", "url", "create", "uberjar", "--port", "8080", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Hello World", 90, 1)

			// Delete the component
			helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()
		})

		It("Should be able to deploy a spring boot uberjar file using openjdk", func() {
			oc.ImportJavaIS(commonVar.Project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)

			helper.Cmd("odo", "create", "--s2i", "java:8", "sb-jar-test", "--project",
				commonVar.Project, "--binary", filepath.Join(commonVar.Context, "sb.jar"), "--context", commonVar.Context).ShouldPass()

			// Create a URL
			helper.Cmd("odo", "url", "create", "uberjaropenjdk", "--port", "8080", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "HTTP Booster", 300, 1)

			// Delete the component
			helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()
		})

	})
})
