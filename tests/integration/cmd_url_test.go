package integration

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo url command tests", func() {
	var globals helper.Globals

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEach()

	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeach(globals)

	})

	Context("Listing urls", func() {
		It("should list appropriate URLs and push message", func() {
			var stdout string
			url1 := helper.RandString(5)
			url2 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", globals.Context, "--project", globals.Project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex", "--port", "8080,8000")
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			stdout = helper.CmdShouldFail("odo", "url", "list", "--context", globals.Context)
			Expect(stdout).To(ContainSubstring("no URLs found"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--context", globals.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed", url1, "odo push"})

			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})
			helper.DontMatchAllInOutput(stdout, []string{"Not Pushed", "odo push"})

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f", "--context", globals.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url1, "odo push"})

			helper.CmdShouldPass("odo", "url", "create", url2, "--port", "8000", "--context", globals.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url2, "Not Pushed", "odo push"})
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{url2, "Pushed"})
			helper.DontMatchAllInOutput(stdout, []string{url1, "Not Pushed", "odo push"})
		})

		It("should create a secure URL", func() {
			url1 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", globals.Context, "--project", globals.Project, componentName)

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--context", globals.Context, "--secure")
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			secureURL := helper.DetermineRouteURL(globals.Context)
			Expect(secureURL).To(ContainSubstring("https:"))
			helper.HttpWaitFor(secureURL, "Hello world from node.js!", 20, 1)

			stdout := helper.CmdShouldPass("odo", "url", "list", "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{secureURL, "Pushed", "true"})

			helper.CmdShouldPass("odo", "delete", "-f", "--context", globals.Context)
		})
	})

	Context("Describing urls", func() {
		It("should describe appropriate URLs and push message", func() {
			var stdout string
			url1 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", globals.Context, "--project", globals.Project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex", "--port", "8080,8000")

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--context", globals.Context)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1, "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed", url1, "odo push"})

			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1, "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})
			helper.DontMatchAllInOutput(stdout, []string{"Not Pushed", "odo push"})

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f", "--context", globals.Context)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1, "--context", globals.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url1, "odo push"})
		})
	})

	Context("when listing urls using -o json flag", func() {
		var originalDir string
		JustBeforeEach(func() {
			originalDir = helper.Getwd()
			helper.Chdir(globals.Context)
		})

		JustAfterEach(func() {
			helper.Chdir(originalDir)
		})
		It("should be able to list url in machine readable json format", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", globals.Project, "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "myurl")
			helper.CmdShouldPass("odo", "push")

			// odo url list -o json
			actualURLListJSON := helper.CmdShouldPass("odo", "url", "list", "-o", "json")
			fullURLPath := helper.DetermineRouteURL("")
			pathNoHTTP := strings.Split(fullURLPath, "//")[1]
			desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"host":"%s","protocol":"http","port":8080,"secure":false},"status":{"state": "Pushed"}}]}`, pathNoHTTP)
			Expect(desiredURLListJSON).Should(MatchJSON(actualURLListJSON))
		})

		It("should be able to list url in machine readable json format for a secure url", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", globals.Project, "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "myurl", "--secure")
			helper.CmdShouldPass("odo", "push")

			// odo url list -o json
			actualURLListJSON := helper.CmdShouldPass("odo", "url", "list", "-o", "json")
			fullURLPath := helper.DetermineRouteURL("")
			pathNoHTTP := strings.Split(fullURLPath, "//")[1]
			desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"host":"%s","protocol":"https","port":8080,"secure":true},"status":{"state": "Pushed"}}]}`, pathNoHTTP)
			Expect(desiredURLListJSON).Should(MatchJSON(actualURLListJSON))
		})
	})

	Context("when using --now flag with url create / delete", func() {
		It("should create and delete url on cluster successfully with now flag", func() {
			url1 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", globals.Context, "--project", globals.Project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex", "--port", "8080,8000")
			helper.CmdShouldPass("odo", "url", "create", url1, "--context", globals.Context, "--port", "8080", "--now")
			out1 := helper.CmdShouldPass("odo", "url", "list", "--context", globals.Context)
			helper.MatchAllInOutput(out1, []string{url1, "Pushed", url1})
			helper.DontMatchAllInOutput(out1, []string{"odo push"})
			routeURL := helper.DetermineRouteURL(globals.Context)
			// Ping said URL
			helper.HttpWaitFor(routeURL, "Node.js", 30, 1)
			helper.CmdShouldPass("odo", "url", "delete", url1, "--context", globals.Context, "--now", "-f")
			out2 := helper.CmdShouldFail("odo", "url", "list", "--context", globals.Context)
			Expect(out2).To(ContainSubstring("no URLs found"))
		})
	})
})
