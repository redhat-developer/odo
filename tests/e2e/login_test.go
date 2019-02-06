package e2e

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odoLoginE2e", func() {
	// user related constants
	const loginTestUserForNoProject = "developernoproject"
	const loginTestUserForSingleProject1 = "developersingleproject1"
	const odoTestProjectForSingleProject1 = "testproject1"
	const loginTestUserPassword = "developer"

	// variables to be used in test
	var session string
	var testUserToken string

	Describe("Check for successful login and logout", func() {
		Context("Initialize", func() {
			It("Should initialize some variables", func() {
				// Logout of current user to ensure state
				runCmd("oc logout")
			})
		})

		Context("Run login tests with no active projects, having default is also considered as not having active project", func() {
			AfterEach(func() {
				// Logout of current user to ensure state
				runCmd("oc logout")
			})

			It("Should login successfully with username and password without any projects with appropriate message", func() {
				session = runCmd(fmt.Sprintf("odo login -u %s -p %s", loginTestUserForNoProject, loginTestUserPassword))
				Expect(session).To(ContainSubstring("Login successful"))
				Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session).To(ContainSubstring("odo project create <project-name>"))
				session = runCmd("oc whoami")
				Expect(session).To(ContainSubstring(loginTestUserForNoProject))
				// One initialization needs one login, hence it happens here
				testUserToken = runCmd("oc whoami -t")
			})

			It("Should login successfully with token without any projects with appropriate message", func() {
				session = runCmd(fmt.Sprintf("odo login -t %s", testUserToken))
				Expect(session).To(ContainSubstring("Logged into"))
				Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session).To(ContainSubstring("odo project create <project-name>"))
				session = runCmd("oc whoami")
				Expect(session).To(ContainSubstring(loginTestUserForNoProject))
			})

			It("Should fail login on invalid token with appropriate message", func() {
				session = runFailCmd("odo login -t verybadtoken", 1)
				Expect(session).To(ContainSubstring("The token provided is invalid or expired"))
				runCmd(fmt.Sprintf("oc login --token %s", testUserToken))
			})
		})

		Context("Run login tests with single active project with username and password", func() {
			AfterEach(func() {
				cleanUpAfterProjects([]string{odoTestProjectForSingleProject1})
			})

			It("Should login successfully with username and password single project with appropriate message", func() {
				// Initialise for test
				runCmd(fmt.Sprintf("oc login -u %s -p %s", loginTestUserForSingleProject1, loginTestUserPassword))
				runCmd(fmt.Sprintf("odo project create %s", odoTestProjectForSingleProject1))
				runCmd("oc logout")
				session = runCmd(fmt.Sprintf("odo login -u %s -p %s", loginTestUserForSingleProject1, loginTestUserPassword))
				Expect(session).To(ContainSubstring("Login successful"))
				Expect(session).To(ContainSubstring(odoTestProjectForSingleProject1))
				session = runCmd("oc whoami")
				Expect(session).To(ContainSubstring(loginTestUserForSingleProject1))
			})
		})
	})
})

func cleanUpAfterProjects(projects []string) {
	for _, p := range projects {
		deleteProject(p)
	}
	// Logout of current user to ensure state
	runCmd("oc logout")
}

func deleteProject(project string) {
	var waitOut bool
	if len(project) > 0 {
		waitOut = waitForCmdOut(fmt.Sprintf("odo project delete -f %s", project), 10, func(out string) bool {
			return strings.Contains(out, fmt.Sprintf("Deleted project : %s", project))
		})
		Expect(waitOut).To(BeTrue())
	}
}
