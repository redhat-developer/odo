package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odoAppE2e", func() {
	// user related constants
	const loginTestUserApplicationWithoutProject1 = "odoappwithoutprojectuser1"
	const loginTestUserApplicationWithoutProject2 = "odoappwithoutprojectuser2"
	const odoTestProjectForApplicationWithoutProject1 = "odoappwithoutprojectproject1"
	const loginTestUserPassword = "developer"

	Describe("Check for failure of app creation without project, with appropriate message", func() {
		Context("Logs into new user with default as active project and tries to create application", func() {
			var currentUserToken1 string
			It("Should know who is currently logged in", func() {
				currentUserToken1 = runCmdShouldPass("oc whoami -t")
			})

			It("Should fail to create app with message", func() {
				runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", loginTestUserApplicationWithoutProject1, loginTestUserPassword))
				session := runCmdShouldFail("odo create nodejs")
				Expect(session).To(ContainSubstring("You dont have permission to project 'default' or it doesnt exist. Please create or set a different project"))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				runCmdShouldPass(fmt.Sprintf("oc login --token %s", currentUserToken1))
			})
		})
	})

	Describe("Check for failure of app creation, if user deletes current project, and creates app after, with appropriate message", func() {
		Context("Logs into user with a project, deletes it and then tries to create application", func() {
			var currentUserToken2 string
			It("Should know who is currently logged in", func() {
				currentUserToken2 = runCmdShouldPass("oc whoami -t")
			})

			It("Should failt to create app with appropriate message", func() {
				runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", loginTestUserApplicationWithoutProject2, loginTestUserPassword))
				runCmdShouldPass(fmt.Sprintf("odo project create %s", odoTestProjectForApplicationWithoutProject1))
				deleteProject(odoTestProjectForApplicationWithoutProject1)
				session := runCmdShouldFail("odo create nodejs")
				Expect(session).To(ContainSubstring(fmt.Sprintf("You dont have permission to project '%s' or it doesnt exist. Please create or set a different project", odoTestProjectForApplicationWithoutProject1)))
				Expect(session).To(ContainSubstring("odo project create|set <project_name>"))
				runCmdShouldPass(fmt.Sprintf("oc login --token %s", currentUserToken2))
			})
		})
	})
})
