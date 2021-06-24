package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
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
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex", "--port", "8080,8000").ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldFail().Err()
			Expect(stdout).To(ContainSubstring("no URLs found"))

			helper.Cmd("odo", "url", "create", url1, "--port", "8080", "--context", commonVar.Context).ShouldPass()
			stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed", url1, "odo push", "<provided by cluster>"})

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})
			helper.DontMatchAllInOutput(stdout, []string{"Not Pushed", "odo push"})

			helper.Cmd("odo", "url", "delete", url1, "-f", "--context", commonVar.Context).ShouldPass()
			stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url1, "odo push"})

			helper.Cmd("odo", "url", "create", url2, "--port", "8000", "--context", commonVar.Context).ShouldPass()
			stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url2, "Not Pushed", "odo push"})
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{url2, "Pushed"})
			helper.DontMatchAllInOutput(stdout, []string{url1, "Not Pushed", "odo push"})
		})

		It("should create a secure URL", func() {
			url1 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project, componentName).ShouldPass()

			helper.Cmd("odo", "url", "create", url1, "--port", "8080", "--context", commonVar.Context, "--secure").ShouldPass()

			stdout := helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed", "true"})

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			secureURLs := helper.DetermineRouteURLs(commonVar.Context)
			Expect(secureURLs).To(ContainElement(ContainSubstring("https")))
			helper.HttpWaitFor(secureURLs[0], "Hello world from node.js!", 20, 1)

			stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{secureURLs[0], "Pushed", "true"})

			helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()
		})
	})

	Context("when listing urls using -o json flag", func() {
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})
		It("should be able to list url in machine readable json format", func() {
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--app", "myapp", "--project", commonVar.Project, "--git", "https://github.com/openshift/nodejs-ex").ShouldPass()
			helper.Cmd("odo", "url", "create", "myurl").ShouldPass()
			helper.Cmd("odo", "push").ShouldPass()

			// odo url list -o json
			actualURLListJSON := helper.Cmd("odo", "url", "list", "-o", "json").ShouldPass().Out()
			valuesURLL := gjson.GetMany(actualURLListJSON, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.kind", "items.0.status.state")
			expectedURLL := []string{"List", "url", "myurl", "route", "Pushed"}
			Expect(helper.GjsonMatcher(valuesURLL, expectedURLL)).To(Equal(true))

		})

		It("should be able to list url in machine readable json format for a secure url", func() {
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--app", "myapp", "--project", commonVar.Project, "--git", "https://github.com/openshift/nodejs-ex").ShouldPass()
			helper.Cmd("odo", "url", "create", "myurl", "--secure").ShouldPass()
			actualURLListJSON := helper.Cmd("odo", "url", "list", "-o", "json").ShouldPass().Out()
			valuesURLLJ := gjson.GetMany(actualURLListJSON, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.port", "items.0.status.state")
			expectedURLLJ := []string{"List", "url", "myurl", "8080", "Not Pushed"}
			Expect(helper.GjsonMatcher(valuesURLLJ, expectedURLLJ)).To(Equal(true))

			helper.Cmd("odo", "push").ShouldPass()

			// odo url list -o json
			actualURLListJSON = helper.Cmd("odo", "url", "list", "-o", "json").ShouldPass().Out()
			valuesURLLJP := gjson.GetMany(actualURLListJSON, "kind", "items.0.kind", "items.0.metadata.name", "items.0.spec.port", "items.0.spec.secure", "items.0.status.state")
			expectedURLLJP := []string{"List", "url", "myurl", "8080", "true", "Pushed"}
			Expect(helper.GjsonMatcher(valuesURLLJP, expectedURLLJP)).To(Equal(true))

		})
	})

	Context("when using --now flag with url create / delete", func() {
		It("should create and delete url on cluster successfully with now flag", func() {
			url1 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex", "--port", "8080,8000").ShouldPass()
			helper.Cmd("odo", "url", "create", url1, "--context", commonVar.Context, "--port", "8080", "--now").ShouldPass()
			out1 := helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(out1, []string{url1, "Pushed", url1})
			helper.DontMatchAllInOutput(out1, []string{"odo push"})
			routeURL := helper.DetermineRouteURL(commonVar.Context)
			// Ping said URL
			helper.HttpWaitFor(routeURL, "Node.js", 30, 1)
			helper.Cmd("odo", "url", "delete", url1, "--context", commonVar.Context, "--now", "-f").ShouldPass()
			out2 := helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldFail().Err()
			Expect(out2).To(ContainSubstring("no URLs found"))
		})
	})
})
