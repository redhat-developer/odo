package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	//. "github.com/Benjamintf1/unmarshalledmatchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odojsonoutput", func() {
	var project string
	var context string
	var originalDir string

	// This is run after every Spec (It)
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	// Test machine readable output
	Context("Pass on creation: odo project create $PROJECT -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		JustAfterEach(func() {
			helper.DeleteProject(projectName)
		})

		// odo project create foobar -o json
		It("should be able to create project and show output in json format", func() {
			actual := helper.CmdShouldPass("odo", "project", "create", projectName, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"message":"Project '%s' is ready for use"}`, projectName, projectName, projectName)
			Expect(desired).Should(MatchJSON(actual))
		})

	})

	Context("Fail on double creation: odo project create $PROJECT -o json", func() {

		// odo project create foobar -o json (x2)
		It("should fail saying that there is already an existing project", func() {
			actual := helper.CmdShouldFail("odo", "project", "create", project, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Error","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"creationTimestamp":null},"message":"unable to create new project: unable to create new project %s: project.project.openshift.io \"%s\" already exists"}`, project, project)
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
		JustBeforeEach(func() {
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.Chdir(originalDir)
		})
		It("Shows storage, list, app in json format", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", project, "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "push")

			// odo component list -o json
			actualCompListJSON := helper.CmdShouldPass("odo", "list", "-o", "json")
			desiredCompListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null,"namespace":"%s"},"spec":{"type":"nodejs","app": "myapp","source":"https://github.com/openshift/nodejs-ex"},"status":{"state":"Pushed"}}]}`, project)
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			// odo describe component -o json
			actualDesCompJSON := helper.CmdShouldPass("odo", "describe", "nodejs", "-o", "json")
			desiredDesCompJSON := fmt.Sprintf(`{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null,"namespace":"%s"},"spec":{"type":"nodejs","app": "myapp","source":"https://github.com/openshift/nodejs-ex","ports": ["8080/TCP"]},"status":{"state":"Pushed"}}`, project)
			Expect(desiredDesCompJSON).Should(MatchJSON(actualDesCompJSON))

			// odo list -o json --path .
			desiredJSONPathFlag := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null,"namespace":"%s"},"spec":{"type":"nodejs","app": "myapp","source":"https://github.com/openshift/nodejs-ex","ports": ["8080/TCP"]},"status":{"context":"%s","state":"Pushed"}}]}`, project, strings.TrimSpace(context))
			actualJSONPathFlag := helper.CmdShouldPass("odo", "list", "-o", "json", "--path", context)
			Expect(desiredJSONPathFlag).Should(MatchJSON(actualJSONPathFlag))

		})
	})

	Context("odo machine readable output on project nodejs is deployed", func() {
		JustBeforeEach(func() {
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.Chdir(originalDir)
		})
		It("should be able to list url", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--app", "myapp", "--project", project, "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "myurl")
			helper.CmdShouldPass("odo", "push")

			// odo url list -o json
			actualURLListJSON := helper.CmdShouldPass("odo", "url", "list", "-o", "json")
			fullURLPath := helper.DetermineRouteURL("")
			pathNoHTTP := strings.Split(fullURLPath, "//")[1]
			desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"host":"%s","protocol":"http","port":8080},"status":{"state": "Pushed"}}]}`, pathNoHTTP)
			Expect(desiredURLListJSON).Should(MatchJSON(actualURLListJSON))

			// odo project list -o json
			// json output varies in CI and locally
			// actualProjectListJSON := helper.CmdShouldPass("odo", "project", "list", "-o", "json")
			// desiredProjectListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myproject","creationTimestamp":null},"spec":{"apps":["myapp"]},"status":{"active":false}},{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"apps":["myapp"]},"status":{"active":true}}]}`, project)
			// Expect(desiredProjectListJSON).Should(MatchJSON(actualProjectListJSON))

		})
	})
})
