package e2e

import (
	"encoding/json"
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odojsonoutput", func() {

	Context("odo machine readable output", func() {
		// // Basic creation
		It("Pre-Test Creation", func() {
			runCmd("odo project create json-test")
			runCmd("odo app create myapp")
			runCmd("odo create nodejs nodejs --git https://github.com/openshift/nodejs-ex")
			runCmd("odo url create myurl")
			runCmd("odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi")

		})
		// odo app describe myapp -o json
		It("should be able to describe app", func() {
			desired := `{"kind":"app","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myapp","namespace":"json-test","creationTimestamp":null},"spec":{"components":["nodejs"]},"status":{"active":true}}`
			actual := runCmd("odo app describe myapp -o json")
			areEqual, _ := compareJSON(desired, actual)
			Expect(areEqual).To(BeTrue())
		})
		// odo app list -o json
		It("should be able to list the apps", func() {
			actual := runCmd("odo app list -o json")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"app","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myapp","namespace":"json-test","creationTimestamp":null},"spec":{"components":["nodejs"]},"status":{"active":true}}]}`
			areEqual, _ := compareJSON(desired, actual)
			Expect(areEqual).To(BeTrue())

		})
		// odo describe nodejs -o json
		It("should be able to describe component", func() {
			actual := runCmd("odo describe nodejs -o json")
			desired := `{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex","url":["myurl"],"storage":["mystorage"]},"status":{"active":true}}`
			areEqual, _ := compareJSON(desired, actual)
			Expect(areEqual).To(BeTrue())
		})
		// odo list -o json
		It("should be able to list components", func() {
			actual := runCmd("odo list -o json")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"Component","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"nodejs","creationTimestamp":null},"spec":{"type":"nodejs","source":"https://github.com/openshift/nodejs-ex","url":["myurl"],"storage":["mystorage"]},"status":{"active":true}}]}`
			areEqual, _ := compareJSON(desired, actual)
			Expect(areEqual).To(BeTrue())

		})
		// odo url list -o json
		It("should be able to list url", func() {
			actual := runCmd("odo url list -o json")
			url := runCmd("oc get routes myurl-myapp -o jsonpath={.spec.host}")
			desired := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"myurl","creationTimestamp":null},"spec":{"path":"%s","port":8080}}]}`, url)
			areEqual, _ := compareJSON(desired, actual)
			Expect(areEqual).To(BeTrue())

		})
		// odo storage list -o json
		It("should be able to list storage", func() {
			actual := runCmd("odo storage list -o json")
			desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1aplha1","metadata":{},"items":[{"kind":"Storage","apiVersion":"odo.openshift.io/v1alpha1","metadata":{"name":"mystorage","creationTimestamp":null},"spec":{"size":"1Gi","path":"/opt/app-root/src/storage/"},"status":{"mounted":true}}]}`

			areEqual, _ := compareJSON(desired, actual)
			Expect(areEqual).To(BeTrue())

		})
		// cleanup
		It("Cleanup", func() {
			runCmd("odo project delete json-test -f")
		})

	})
})

func compareJSON(desired, actual string) (bool, error) {
	var o1, o2 interface{}
	err := json.Unmarshal([]byte(actual), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(desired), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string :: %s", err.Error())
	}
	return reflect.DeepEqual(o1, o2), nil

}
