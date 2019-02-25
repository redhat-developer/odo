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

	Describe("Check for successful login and logout", func() {
		Context("Run login tests with no active projects, having default is also considered as not having active project", func() {
			// variables to be used in test
			var session1 string
			var testUserToken1 string
			var currentUserToken1 string
			It("Should know who is currently logged in", func() {
				currentUserToken1 = runCmdShouldPass("oc whoami -t")
			})
			It("Should login successfully with username and password without any projects with appropriate message", func() {
				session1 = runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", loginTestUserForNoProject, loginTestUserPassword))
				Expect(session1).To(ContainSubstring("Login successful"))
				Expect(session1).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session1).To(ContainSubstring("odo project create <project-name>"))
				session1 = runCmdShouldPass("oc whoami")
				Expect(session1).To(ContainSubstring(loginTestUserForNoProject))
				// One initialization needs one login, hence it happens here
				testUserToken1 = runCmdShouldPass("oc whoami -t")
			})

			It("Should login successfully with token without any projects with appropriate message", func() {
				session1 = runCmdShouldPass(fmt.Sprintf("odo login -t %s", testUserToken1))
				Expect(session1).To(ContainSubstring("Logged into"))
				Expect(session1).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session1).To(ContainSubstring("odo project create <project-name>"))
				session1 = runCmdShouldPass("oc whoami")
				Expect(session1).To(ContainSubstring(loginTestUserForNoProject))
			})

			It("Should fail login on invalid token with appropriate message", func() {
				sessionErr := runCmdShouldFail("odo login -t verybadtoken")
				Expect(sessionErr).To(ContainSubstring("The token provided is invalid or expired"))
				runCmdShouldPass(fmt.Sprintf("oc login --token %s", currentUserToken1))
			})
		})

		Context("Run login tests with single active project with username and password", func() {
			// variables to be used in test
			var session2 string
			var currentUserToken2 string
			It("Should know who is currently logged in", func() {
				currentUserToken2 = runCmdShouldPass("oc whoami -t")
			})

			It("Should login successfully with username and password single project with appropriate message", func() {
				// Initialise for test
				runCmdShouldPass(fmt.Sprintf("oc login -u %s -p %s", loginTestUserForSingleProject1, loginTestUserPassword))
				odoCreateProject(odoTestProjectForSingleProject1)
				session2 = runCmdShouldPass(fmt.Sprintf("odo login -u %s -p %s", loginTestUserForSingleProject1, loginTestUserPassword))
				Expect(session2).To(ContainSubstring("Login successful"))
				Expect(session2).To(ContainSubstring(odoTestProjectForSingleProject1))
				session2 = runCmdShouldPass("oc whoami")
				Expect(session2).To(ContainSubstring(loginTestUserForSingleProject1))
				cleanUpAfterProjects([]string{odoTestProjectForSingleProject1})
				runCmdShouldPass(fmt.Sprintf("oc login --token %s", currentUserToken2))
			})
		})
	})
})
