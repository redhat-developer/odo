package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odojsonoutput", func() {

	Context("odo machine readable output", func() {
		// Basic creation
		It("Pre-Test Creation: Creating project", func() {
			odoCreateProject("json-test")
		})
		// odo app list -o json
		It("should be able to return empty list", func() {
			actual := runCmdShouldPass("odo app list -o json --project json-test")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[]}`
			Expect(desired).Should(MatchJSON(actual))
		})
		// Basic creation
		It("Pre-Test Creation Json", func() {
			runCmdShouldPass("odo create nodejs nodejs --app myapp --project json-test --git https://github.com/openshift/nodejs-ex")
			runCmdShouldPass("odo push")
		})
		// odo url list -o json
		It("should be able to list empty url list", func() {
			actual := runCmdShouldPass("odo url list -o json")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":null}`
			Expect(desired).Should(MatchJSON(actual))

		})
		// odo url create
		It("should be able to create url", func() {
			runCmdShouldPass("odo url create myurl --port 8080")
			runCmdShouldPass("odo push -v 4")
			routeURL := determineRouteURL()
			// Ping said URL
			responsePing := matchResponseSubString(routeURL, "application on OpenShift", 90, 1)
			Expect(responsePing).Should(BeTrue())
			// actual := runCmdShouldPass("odo url create myurl -o json")
			//	url := runCmdShouldPass("oc get routes myurl-myapp -o jsonpath={.spec.host}")
			//	desired := fmt.Sprintf(`{"kind":"url","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"host":"%s","protocol":"http","port":8080}}`, url)
			//	areEqual, _ := compareJSON(desired, actual)
			//	Expect(areEqual).To(BeTrue())
		})

		// odo storage create -o json
		It("should be able to create storage", func() {
			actual := runCmdShouldPass("odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi -o json")
			runCmdShouldPass("odo push")
			desired := `{"kind":"storage","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi"},"status":{"path":"/opt/app-root/src/storage/"}}`
			Expect(desired).Should(MatchJSON(actual))
		})
		// odo project list -o json
		It("should be able to list the projects", func() {
			actual := runCmdShouldPass("odo project list -o json")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"json-test","creationTimestamp":null},"spec":{"apps":["myapp"]},"status":{"active":true}},{"kind":"Project","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myproject","creationTimestamp":null},"spec":{"apps":["myapp"]},"status":{"active":false}}]}`
			Expect(desired).Should(MatchJSON(actual))

		})
		// odo app describe myapp -o json
		It("should be able to describe app", func() {
			desired := `{"kind":"app","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myapp","namespace":"json-test","creationTimestamp":null},"spec":{"components":["nodejs"]},"status":{"active":false}}`
			actual := runCmdShouldPass("odo app describe myapp -o json")
			Expect(desired).Should(MatchJSON(actual))
		})
		// odo app list -o json
		It("should be able to list the apps", func() {
			actual := runCmdShouldPass("odo app list -o json")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"app","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myapp","namespace":"json-test","creationTimestamp":null},"spec":{"components":["nodejs"]},"status":{"active":false}}]}`
			Expect(desired).Should(MatchJSON(actual))

		})
		// odo describe nodejs -o json
		It("should be able to describe component", func() {
			actual := runCmdShouldPass("odo describe nodejs -o json")
			desired := `{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex","url":["myurl"],"storage":["mystorage"]},"status":{"active":false}}`
			Expect(desired).Should(MatchJSON(actual))
		})
		// odo list -o json
		It("should be able to list components", func() {
			actual := runCmdShouldPass("odo list -o json")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex","url":["myurl"],"storage":["mystorage"]},"status":{"active":false}}]}`
			Expect(desired).Should(MatchJSON(actual))

		})
		// odo url list -o json
		It("should be able to list url", func() {
			actual := runCmdShouldPass("odo url list -o json")
			url := runCmdShouldPass("oc get routes myurl-myapp -o jsonpath={.spec.host}")
			desired := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"host":"%s","protocol":"http","port":8080}}]}`, url)
			Expect(desired).Should(MatchJSON(actual))

		})

		// odo storage list -o json
		It("should be able to list storage", func() {
			actual := runCmdShouldPass("odo storage list -o json")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"storage","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi"},"status":{"path":"/opt/app-root/src/storage/"}}]}`
			Expect(desired).Should(MatchJSON(actual))
		})
		// cleanup
		It("Cleanup", func() {
			ocDeleteProject("json-test")
			runCmdShouldPass("rm -rf .odo")
		})

	})
})
