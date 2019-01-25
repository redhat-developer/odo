package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odoLinkE2e", func() {

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

	Context("odo link/unlink handling between components and service", func() {

		It("create a frontend and backend application", func() {
			runCmd("odo create nodejs frontend")
			runCmd("odo create python backend")

			cmpList := runCmd("odo list")
			Expect(cmpList).To(ContainSubstring("frontend"))
			Expect(cmpList).To(ContainSubstring("backend"))
		})

		It("reports error when using wrong port", func() {
			output := runFailCmd("odo link backend --component frontend --port 1234", 1)
			Expect(output).To(ContainSubstring("8080"))
		})

		It("link the frontend application to the backend", func() {
			runCmd("odo link backend --component frontend")

			// ensure that the proper envFrom entry was created
			envFromOutput :=
				runCmd("oc get dc frontend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
			Expect(envFromOutput).To(ContainSubstring("backend"))
		})

		It("describe should show the linked service", func() {
			describeOutput := runCmd("odo describe frontend")

			// ensure that the output contains the component and port
			Expect(describeOutput).To(ContainSubstring("backend"))
			Expect(describeOutput).To(ContainSubstring("8080"))
		})

		It("link should fail when linking to the same component again", func() {
			output := runFailCmd("odo link backend --component frontend", 1)
			Expect(output).To(ContainSubstring("been linked"))
		})

		It("should be able to create a service", func() {
			runCmd("odo service create mysql-persistent")

			waitForCmdOut("oc get serviceinstance -o name", 1, func(output string) bool {
				return strings.Contains(output, "mysql-persistent")
			})
		})

		It("app describe should show the mysql service", func() {
			describeOutput := runCmd("odo app describe")

			// ensure that the output contains the service
			Expect(describeOutput).To(ContainSubstring("mysql-persistent"))
		})

		It("should link backend to service", func() {
			runCmd("odo link mysql-persistent -w --component backend")

			// ensure that the proper envFrom entry was created
			envFromOutput :=
				runCmd("oc get dc backend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
			Expect(envFromOutput).To(ContainSubstring("mysql-persistent"))
		})

		It("link should fail when linking to the same service again", func() {
			output := runFailCmd("odo link mysql-persistent --component backend", 1)
			Expect(output).To(ContainSubstring("been linked"))
		})

		It("describe should show the linked service", func() {
			describeOutput := runCmd("odo describe backend")

			// ensure that the output contains the service
			Expect(describeOutput).To(ContainSubstring("mysql-persistent"))
		})

		It("delete the service", func() {
			runCmd("odo service delete mysql-persistent -f")

			// ensure that the backend no longer has an envFrom value
			backendEnvFromOutput :=
				runCmd("oc get dc backend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
			Expect(backendEnvFromOutput).To(BeEmpty())

			// ensure that the frontend envFrom was not changed
			frontEndEnvFromOutput :=
				runCmd("oc get dc frontend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
			Expect(frontEndEnvFromOutput).To(ContainSubstring("backend"))
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
