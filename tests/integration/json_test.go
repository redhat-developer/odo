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
	var project string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		os.RemoveAll(".odo")
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

			// odo app list -o json
			actualAppListJSON := helper.CmdShouldPass("odo", "app", "list", "-o", "json")
			desiredAppListJSON := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"app","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myapp","namespace":"` + project + `","creationTimestamp":null},"spec":{"components":["nodejs"]},"status":{"active":false}}]}`
			Expect(desiredAppListJSON).Should(MatchJSON(actualAppListJSON))

			// odo app describe myapp -o json
			desiredDesAppJSON := `{"kind":"app","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myapp","namespace":"` + project + `","creationTimestamp":null},"spec":{"components":["nodejs"]},"status":{"active":false}}`
			actualDesAppJSON := helper.CmdShouldPass("odo", "app", "describe", "myapp", "-o", "json")
			Expect(desiredDesAppJSON).Should(MatchJSON(actualDesAppJSON))

			// odo component list -o json
			actualCompListJSON := helper.CmdShouldPass("odo", "list", "-o", "json")
			desiredCompListJSON := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex"},"status":{"active":false}}]}`
			Expect(desiredCompListJSON).Should(MatchJSON(actualCompListJSON))

			// odo describe component -o json
			actualDesCompJSON := helper.CmdShouldPass("odo", "describe", "nodejs", "-o", "json")
			desiredDesCompJSON := `{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex"},"status":{"active":false}}`
			Expect(desiredDesCompJSON).Should(MatchJSON(actualDesCompJSON))

			// odo storage create -o json
			actualJSONStorage := helper.CmdShouldPass("odo", "storage", "create", "mystorage", "--path=/opt/app-root/src/storage/", "--size=1Gi", "-o", "json")
			desiredJSONStorage := `{"kind":"storage","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi"},"status":{"path":"/opt/app-root/src/storage/"}}`
			Expect(desiredJSONStorage).Should(MatchJSON(actualJSONStorage))

			// odo storage list -o json
			actualSrorageList := helper.CmdShouldPass("odo", "storage", "list", "-o", "json")
			desiredSrorageList := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"storage","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi"},"status":{"path":"/opt/app-root/src/storage/"}}]}`
			Expect(desiredSrorageList).Should(MatchJSON(actualSrorageList))
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
