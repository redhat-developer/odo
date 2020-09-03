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

var _ = Describe("odo java e2e tests", func() {
	var project string
	var context string
	var oc helper.OcRunner
	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	// contains a minimal javaee app
	const warGitRepo = "https://github.com/lordofthejars/book-insultapp"

	// contains a minimal javalin app
	const jarGitRepo = "https://github.com/geoand/javalin-helloworld"

	// Test wildfly
	Context("odo wildfly component creation ", func() {
		JustBeforeEach(func() {
			if runtime.GOARCH == "s390x" {
				Skip("Skipping test because there is no supported wildfly builder image.")
			}
		})

		It("Should be able to deploy a git repo that contains a wildfly application without wait flag", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "wildfly", "wo-wait-javaee-git-test", "--project",
				project, "--ref", "master", "--git", warGitRepo, "--context", context)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "gitrepo", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "-v", "4", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Insult", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "wo-wait-javaee-git-test", "-f", "--context", context)
		})
	})
	// Test Java
	Context("odo component creation", func() {
		It("Should be able to deploy a .war file using wildfly", func() {
			helper.CopyExample(filepath.Join("binary", "java", "wildfly"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "wildfly", "javaee-war-test", "--project",
				project, "--binary", filepath.Join(context, "ROOT.war"), "--context", context)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "warfile", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Sample", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "javaee-war-test", "-f", "--context", context)
		})

		It("Should be able to deploy a git repo that contains a java uberjar application using openjdk", func() {
			oc.ImportJavaIS(project)

			// Deploy the git repo / wildfly example
			helper.CmdShouldPass("odo", "create", "--s2i", "java:8", "uberjar-git-test", "--project",
				project, "--ref", "master", "--git", jarGitRepo, "--context", context)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "uberjar", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Hello World", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "uberjar-git-test", "-f", "--context", context)
		})

		It("Should be able to deploy a spring boot uberjar file using openjdk", func() {
			oc.ImportJavaIS(project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), context)

			helper.CmdShouldPass("odo", "create", "--s2i", "java:8", "sb-jar-test", "--project",
				project, "--binary", filepath.Join(context, "sb.jar"), "--context", context)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "uberjaropenjdk", "--port", "8080", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			routeURL := helper.DetermineRouteURL(context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "HTTP Booster", 300, 1)

			// Delete the component
			helper.CmdShouldPass("odo", "delete", "sb-jar-test", "-f", "--context", context)
		})

	})
})
