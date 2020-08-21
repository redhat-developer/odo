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
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Listing urls", func() {
		It("should list appropriate URLs and push message", func() {
			var stdout string
			url1 := helper.RandString(5)
			url2 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex", "--port", "8080,8000")
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			stdout = helper.CmdShouldFail("odo", "url", "list", "--context", commonVar.Context)
			Expect(stdout).To(ContainSubstring("no URLs found"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--context", commonVar.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed", url1, "odo push"})

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})
			helper.DontMatchAllInOutput(stdout, []string{"Not Pushed", "odo push"})

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f", "--context", commonVar.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url1, "odo push"})

			helper.CmdShouldPass("odo", "url", "create", url2, "--port", "8000", "--context", commonVar.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url2, "Not Pushed", "odo push"})
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url2, "Pushed"})
			helper.DontMatchAllInOutput(stdout, []string{url1, "Not Pushed", "odo push"})
		})

		It("should create a secure URL", func() {
			url1 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project, componentName)

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--context", commonVar.Context, "--secure")

			stdout := helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed", "true"})

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			secureURL := helper.DetermineRouteURL(commonVar.Context)
			Expect(secureURL).To(ContainSubstring("https:"))
			helper.HttpWaitFor(secureURL, "Hello world from node.js!", 20, 1)

			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{secureURL, "Pushed", "true"})

			helper.CmdShouldPass("odo", "delete", "-f", "--context", commonVar.Context)
		})
	})

	Context("Describing urls", func() {
		It("should describe appropriate URLs and push message", func() {
			var stdout string
			url1 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex", "--port", "8080,8000")

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--context", commonVar.Context)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1, "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed", url1, "odo push"})

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1, "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})
			helper.DontMatchAllInOutput(stdout, []string{"Not Pushed", "odo push"})

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f", "--context", commonVar.Context)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1, "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url1, "odo push"})
		})

		It("should be able to describe a url in CLI format and machine readable json format for a secure url", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", commonVar.Project, "--git", "https://github.com/openshift/nodejs-ex", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "url", "create", "myurl", "--secure", "--context", commonVar.Context)

			actualURLDescribeJSON := helper.CmdShouldPass("odo", "url", "describe", "myurl", "-o", "json", "--context", commonVar.Context)
			desiredURLDescribeJSON := `{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{ "name": "myurl","creationTimestamp": null},"spec":{"port": 8080,"secure": true,"path": "/", "kind": "route"},"status": {"state": "Not Pushed"}}`
			Expect(desiredURLDescribeJSON).Should(MatchJSON(actualURLDescribeJSON))

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// odo url describe -o json
			actualURLDescribeJSON = helper.CmdShouldPass("odo", "url", "describe", "myurl", "-o", "json", "--context", commonVar.Context)
			// get the route URL
			fullURLPath := helper.DetermineRouteURL(commonVar.Context)
			pathNoHTTP := strings.Split(fullURLPath, "//")[1]
			desiredURLDescribeJSON = fmt.Sprintf(`{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{ "name": "myurl","creationTimestamp": null},"spec":{"host":"%s","protocol": "https","port": 8080,"secure": true, "path": "/", "kind": "route"},"status": {"state": "Pushed"}}`, pathNoHTTP)
			Expect(desiredURLDescribeJSON).Should(MatchJSON(actualURLDescribeJSON))
		})
	})

	Context("when listing urls using -o json flag", func() {
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})
		It("should be able to list url in machine readable json format", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", commonVar.Project, "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "myurl")
			helper.CmdShouldPass("odo", "push")

			// odo url list -o json
			actualURLListJSON := helper.CmdShouldPass("odo", "url", "list", "-o", "json")
			fullURLPath := helper.DetermineRouteURL("")
			pathNoHTTP := strings.Split(fullURLPath, "//")[1]
			desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"host":"%s","protocol":"http","port":8080,"secure":false,"path": "/", "kind": "route"},"status":{"state": "Pushed"}}]}`, pathNoHTTP)
			Expect(desiredURLListJSON).Should(MatchJSON(actualURLListJSON))
		})

		It("should be able to list url in machine readable json format for a secure url", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", commonVar.Project, "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "myurl", "--secure")
			actualURLListJSON := helper.CmdShouldPass("odo", "url", "list", "-o", "json")
			desiredURLListJSON := `{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"port":8080,"secure":true,"path": "/","kind": "route"},"status":{"state": "Not Pushed"}}]}`
			Expect(desiredURLListJSON).Should(MatchJSON(actualURLListJSON))

			helper.CmdShouldPass("odo", "push")

			// odo url list -o json
			actualURLListJSON = helper.CmdShouldPass("odo", "url", "list", "-o", "json")
			fullURLPath := helper.DetermineRouteURL("")
			pathNoHTTP := strings.Split(fullURLPath, "//")[1]
			desiredURLListJSON = fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"host":"%s","protocol":"https","port":8080,"secure":true,"path": "/", "kind": "route"},"status":{"state": "Pushed"}}]}`, pathNoHTTP)
			Expect(desiredURLListJSON).Should(MatchJSON(actualURLListJSON))
		})
	})

	Context("when using --now flag with url create / delete", func() {
		It("should create and delete url on cluster successfully with now flag", func() {
			url1 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex", "--port", "8080,8000")
			helper.CmdShouldPass("odo", "url", "create", url1, "--context", commonVar.Context, "--port", "8080", "--now")
			out1 := helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(out1, []string{url1, "Pushed", url1})
			helper.DontMatchAllInOutput(out1, []string{"odo push"})
			routeURL := helper.DetermineRouteURL(commonVar.Context)
			// Ping said URL
			helper.HttpWaitFor(routeURL, "Node.js", 30, 1)
			helper.CmdShouldPass("odo", "url", "delete", url1, "--context", commonVar.Context, "--now", "-f")
			out2 := helper.CmdShouldFail("odo", "url", "list", "--context", commonVar.Context)
			Expect(out2).To(ContainSubstring("no URLs found"))
		})
	})
})
