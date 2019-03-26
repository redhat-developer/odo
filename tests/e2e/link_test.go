package e2e

import (
	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
)

var _ = Describe("odoLinkE2e", func() {
	// Uncomment when link commands are made to use config file and context flag
	/*
		projName := generateTimeBasedName("odolnk")
		const appTestName = "testing"

		// Create a separate project for Java
		Context("create separate project", func() {
			It("should create a new test project", func() {
				session := runCmdShouldPass("odo project create " + projName)
				Expect(session).To(ContainSubstring(projName))
				waitForCmdOut("odo project set "+projName, 4, false, func(output string) bool {
					return strings.Contains(output, "Already on project : "+projName)
				})
				runCmdShouldPass("odo app create " + appTestName)
			})
		})

		Context("odo link/unlink handling between components and service", func() {

			It("create a frontend and backend application", func() {
				runCmdShouldPass("odo create nodejs frontend")
				runCmdShouldPass("odo create python backend")

				cmpList := runCmdShouldPass("odo list")
				Expect(cmpList).To(ContainSubstring("frontend"))
				Expect(cmpList).To(ContainSubstring("backend"))
			})

			It("reports error when using wrong port", func() {
				outputErr := runCmdShouldFail("odo link backend --component frontend --port 1234")
				Expect(outputErr).To(ContainSubstring("8080"))
			})

			It("link the frontend application to the backend", func() {
				runCmdShouldPass("odo link backend --component frontend")

				// ensure that the proper envFrom entry was created
				envFromOutput :=
					runCmdShouldPass("oc get dc frontend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
				Expect(envFromOutput).To(ContainSubstring("backend"))
			})

			It("describe on the frontend should show the linked backend component", func() {
				describeOutput := runCmdShouldPass("odo describe frontend")

				// ensure that the output contains the component and port
				Expect(describeOutput).To(ContainSubstring("backend"))
				Expect(describeOutput).To(ContainSubstring("8080"))
			})

			It("link should fail when linking to the same component again", func() {
				outputErr := runCmdShouldFail("odo link backend --component frontend")
				Expect(outputErr).To(ContainSubstring("been linked"))
			})

			It("should be able to create a service", func() {
				runCmdShouldPass("odo service create mysql-persistent")

				waitForCmdOut("oc get serviceinstance -o name", 1, true, func(output string) bool {
					return strings.Contains(output, "mysql-persistent")
				})
			})

			It("app describe should show the mysql service", func() {
				describeOutput := runCmdShouldPass("odo app describe")

				// ensure that the output contains the service
				Expect(describeOutput).To(ContainSubstring("mysql-persistent"))
			})

			It("should link backend to service", func() {
				runCmdShouldPass("odo link mysql-persistent --wait-for-target --component backend")

				// ensure that the proper envFrom entry was created
				envFromOutput :=
					runCmdShouldPass("oc get dc backend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
				Expect(envFromOutput).To(ContainSubstring("mysql-persistent"))
			})

			It("link should fail when linking to the same service again", func() {
				outputErr := runCmdShouldFail("odo link mysql-persistent --component backend")
				Expect(outputErr).To(ContainSubstring("been linked"))
			})

			It("describe on the backend should show the linked mysql service", func() {
				describeOutput := runCmdShouldPass("odo describe backend")

				// ensure that the output contains the service
				Expect(describeOutput).To(ContainSubstring("mysql-persistent"))
			})

			It("delete the service", func() {
				runCmdShouldPass("odo service delete mysql-persistent -f")

				// ensure that the backend no longer has an envFrom value
				backendEnvFromOutput :=
					runCmdShouldPass("oc get dc backend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
				Expect(backendEnvFromOutput).To(BeEmpty())

				// ensure that the frontend envFrom was not changed
				frontEndEnvFromOutput :=
					runCmdShouldPass("oc get dc frontend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
				Expect(frontEndEnvFromOutput).To(ContainSubstring("backend"))
			})

			It("unlink the backend from the frontend", func() {
				runCmdShouldPass("odo unlink backend --component frontend")

				// ensure that the proper envFrom entry was created
				envFromOutput :=
					runCmdShouldPass("oc get dc frontend-testing -o jsonpath='{.spec.template.spec.containers[0].envFrom}'")
				Expect(envFromOutput).To(BeEmpty())
			})
		})

		// Delete the project
		Context("delete delete", func() {
			It("should delete test project", func() {
				session := runCmdShouldPass("odo project delete " + projName + " -f")
				Expect(session).To(ContainSubstring(projName))
			})
		})
	*/
})
