package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	"time"
)

var _ = Describe("odo-link-e2e", func() {

	var t = strconv.FormatInt(time.Now().Unix(), 10)
	var projName = fmt.Sprintf("odolnk-%s", t)
	const appTestName = "testing"

	// Create a separate project for Java
	Context("create separate project", func() {
		It("should create a new test project", func() {
			session := runCmd("odo project create " + projName)
			Expect(session).To(ContainSubstring(projName))
			runCmd("odo app create " + appTestName)
		})
	})

	Context("odo link/unlink handling between components", func() {

		It("create a frontend and backend application", func() {
			runCmd("odo create nodejs frontend")
			runCmd("odo create python backend")

			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("frontend"))
			Expect(cmpList).To(ContainSubstring("backend"))
		})

		It("link the frontend application to the backend", func() {
			runCmd("odo link backend --component frontend")

			// ensure that the proper envFrom entry was created
			envFromOutput :=
				runCmd("oc get dc frontend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
			Expect(envFromOutput).To(ContainSubstring("backend"))
		})

		It("unlink the backend from the frontend", func() {
			runCmd("odo unlink backend --component frontend")

			// ensure that the proper envFrom entry was created
			envFromOutput :=
				runCmd("oc get dc frontend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
			Expect(envFromOutput).To(BeEmpty())
		})
	})

	// Delete the project
	Context("delete delete", func() {
		It("should delete test project", func() {
			session := runCmd("odo project delete " + projName + " -f")
			Expect(session).To(ContainSubstring(projName))
		})
	})
})
