package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odoLoginE2e", func() {
	// user related constants
	const loginTestUserForNoProject = "odologinnoproject"
	const loginTestUserForSingleProject1 = "odologinsingleproject1"
	const odoTestProjectForSingleProject1 = "odologintestproject1"
	const loginTestUserPassword = "developer"

	// variables to be used in test
	var session string
	var testUserToken string

	Describe("Check for successful login and logout", func() {
		Context("Initialize", func() {
			It("Should initialize some variables", func() {
				// Logout of current user to ensure state
				runCmdShouldPass("oc logout")
			})
		})

		Context("Run login tests with no active projects, having default is also considered as not having active project", func() {
			AfterEach(func() {
				// Logout of current user to ensure state
				runCmdShouldPass("oc logout")
			})

			It("Should login successfully with username and password without any projects with appropriate message", func() {
				session = runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", loginTestUserForNoProject, loginTestUserPassword))
				Expect(session).To(ContainSubstring("Login successful"))
				Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session).To(ContainSubstring("odo project create <project-name>"))
				session = runCmdShouldPass("oc whoami")
				Expect(session).To(ContainSubstring(loginTestUserForNoProject))
				// One initialization needs one login, hence it happens here
				testUserToken = runCmdShouldPass("oc whoami -t")
			})

			It("Should login successfully with token without any projects with appropriate message", func() {
				session = runCmdShouldPass(fmt.Sprintf("odo login -t %s", testUserToken))
				Expect(session).To(ContainSubstring("Logged into"))
				Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session).To(ContainSubstring("odo project create <project-name>"))
				session = runCmdShouldPass("oc whoami")
				Expect(session).To(ContainSubstring(loginTestUserForNoProject))
			})

			It("Should fail login on invalid token with appropriate message", func() {
				sessionErr := runCmdShouldFail("odo login -t verybadtoken")
				Expect(sessionErr).To(ContainSubstring("The token provided is invalid or expired"))
				runCmdShouldPass(fmt.Sprintf("oc login --token %s", testUserToken))
			})
		})

		Context("Run login tests with single active project with username and password", func() {
			AfterEach(func() {
				cleanUpAfterProjects([]string{odoTestProjectForSingleProject1})
			})

			It("Should login successfully with username and password single project with appropriate message", func() {
				// Initialise for test
				runCmdShouldPass(fmt.Sprintf("oc login -u %s -p %s", loginTestUserForSingleProject1, loginTestUserPassword))
				runCmdShouldPass(fmt.Sprintf("odo project create %s", odoTestProjectForSingleProject1))
				//make sure that project has been created
				runCmdShouldPass("oc project")
				runCmdShouldPass("oc logout")

				session = runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", loginTestUserForSingleProject1, loginTestUserPassword))
				Expect(session).To(ContainSubstring("Login successful"))
				Expect(session).To(ContainSubstring(odoTestProjectForSingleProject1))
				session = runCmdShouldPass("oc whoami")
				Expect(session).To(ContainSubstring(loginTestUserForSingleProject1))
			})
		})
	})
})
