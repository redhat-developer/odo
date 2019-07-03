package integration

import (
	"fmt"
	"os"
	"strings"
	"time"

	//. "github.com/Benjamintf1/unmarshalledmatchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odojsonoutput", func() {
	var project, tmpDir string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
		tmpDir = helper.CreateNewContext()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		os.RemoveAll(".odo")
		helper.DeleteDir(tmpDir)
	})

	Context("odo machine create project json output", func() {
		// odo project create newprojectjson -o json
		It("should be able to create project and show output in json format", func() {
			actual := helper.CmdShouldPass("odo", "project", "create", "newprojectjson", "-o", "json")
			desired := `{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"newprojectjson","creationTimestamp":null},"spec":{"apps":null},"status":{"active":false}}`
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	Context("odo machine readable output on empty project", func() {
		//https://github.com/openshift/odo/issues/1708
		//odo project list -o json
		/*It("should be able to return project list", func() {
			actualProjectListJSON := helper.CmdShouldPass("odo", "project", "list", "-o", "json")
			desiredProjectListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myproject","creationTimestamp":null},"spec":{"apps":null},"status":{"active":false}},{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"apps":null},"status":{"active":true}}]}`, project)
			Expect(desiredProjectListJSON).Should(MatchUnorderedJSON(actualProjectListJSON, WithUnorderedListKeys("items")))
		})*/

		// odo app list -o json
		It("should be able to return empty list", func() {
			actual := helper.CmdShouldPass("odo", "app", "list", "-o", "json", "--project", project)
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[]}`
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	Context("odo machine readable output on project nodejs is deployed", func() {
		It("Shows storage, list, app in json format", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", project, "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "push")

			// odo component list -o json
			actualCompListJSON := helper.CmdShouldPass("odo", "list", "-o", "json")
			desiredCompListJSON := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex"},"status":{"state":"Pushed"}}]}`
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			// odo describe component -o json
			actualDesCompJSON := helper.CmdShouldPass("odo", "describe", "nodejs", "-o", "json")
			desiredDesCompJSON := `{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex"},"status":{"state":"Pushed"}}`
			Expect(desiredDesCompJSON).Should(MatchJSON(actualDesCompJSON))

			// odo list -o json --path .
			pwd := helper.Getwd()
			desired := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex"},"status":{"context":"%s","state":"Pushed"}}]}`, strings.TrimSpace(pwd))
			helper.Chdir(tmpDir)
			actual := helper.CmdShouldPass("odo", "list", "-o", "json", "--path", pwd)
			helper.Chdir(pwd)
			Expect(desired).Should(MatchJSON(actual))

		})
	})

	Context("odo machine readable output on project nodejs is deployed", func() {
		It("should be able to list url", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", project, "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "myurl", "--port", "8080")
			helper.CmdShouldPass("odo", "push")

			// odo url list -o json
			actualURLListJSON := helper.CmdShouldPass("odo", "url", "list", "-o", "json")
			fullURLPath := helper.DetermineRouteURL("")
			pathNoHTTP := strings.Split(fullURLPath, "//")[1]
			desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"host":"%s","protocol":"http","port":8080}}]}`, pathNoHTTP)
			Expect(desiredURLListJSON).Should(MatchJSON(actualURLListJSON))

			// odo project list -o json
			// json output varies in CI and locally
			// actualProjectListJSON := helper.CmdShouldPass("odo", "project", "list", "-o", "json")
			// desiredProjectListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myproject","creationTimestamp":null},"spec":{"apps":["myapp"]},"status":{"active":false}},{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"apps":["myapp"]},"status":{"active":true}}]}`, project)
			// Expect(desiredProjectListJSON).Should(MatchJSON(actualProjectListJSON))

		})
	})
})
