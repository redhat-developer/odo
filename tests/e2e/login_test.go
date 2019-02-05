package e2e

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odoLoginE2e", func() {
	// user related constants
	const loginTestBaseUser = "testdeveloper"
	const loginTestUserPassword = "developer"

	// variables to be used in test
	var session string
	var testUserToken string
	var testUser string

	Describe("Check for successful login and logout", func() {
		Context("Initialize", func() {
			It("Should initialize some variables", func() {
				// Logout of current user to ensure state
				runCmd("oc logout")
			})
		})

		Context("Run login tests with no active projects, having default is also considered as not having active project", func() {
			AfterEach(func() {
				// Log out of whoever is logged in
				runCmd("oc logout")
			})

			It("Should login successfully with username and password without any projects with appropriate message", func() {
				session = runCmd(fmt.Sprintf("odo login -u %s -p %s", loginTestBaseUser, loginTestUserPassword))
				Expect(session).To(ContainSubstring("Login successful"))
				Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session).To(ContainSubstring("odo project create <project-name>"))
				session = runCmd("oc whoami")
				Expect(session).To(ContainSubstring(loginTestBaseUser))
				// One initialization needs one login, hence it happens here
				testUserToken = runCmd("oc whoami -t")
			})

			It("Should login successfully with token without any projects with appropriate message", func() {
				session = runCmd(fmt.Sprintf("odo login -t %s", testUserToken))
				Expect(session).To(ContainSubstring("Logged into"))
				Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session).To(ContainSubstring("odo project create <project-name>"))
				session = runCmd("oc whoami")
				Expect(session).To(ContainSubstring(loginTestBaseUser))
			})

			It("Should fail login on invalid token with appropriate message", func() {
				session = runFailCmd("odo login -t verybadtoken", 1)
				Expect(session).To(ContainSubstring("The token provided is invalid or expired"))
				runCmd(fmt.Sprintf("oc login -t %s", testUserToken))
			})
		})

		Context("Run login tests with single active project", func() {
			AfterEach(func() {
				runCmd("oc logout")
			})

			It("Should login successfully with username and password single project with appropriate message", func() {
				// Initialise for test
				testUser = fmt.Sprintf("%s%s", loginTestBaseUser, "1")
				runCmd(fmt.Sprintf("oc login -u %s -p %s", testUser, loginTestUserPassword))
				runCmd("oc new-project testproject1")
				runCmd("oc logout")
				session = runCmd("odo login -u %s -p %s")
				Expect(session).To(ContainSubstring("Login successful"))
				Expect(session).To(ContainSubstring("testproject1"))
				session = runCmd("oc whoami -t")
				Expect(session).To(ContainSubstring("testproject1"))
				deleteProject("testproject1")
			})
		})
	})
})

func deleteProject(project string) {
	var waitOut bool
	if len(project) > 0 {
		waitOut = waitForCmdOut(fmt.Sprintf("odo project delete %s", project), 10, func(out string) bool {
			return strings.Contains(out, fmt.Sprintf("Deleted project : %s", project))
		})
		Expect(waitOut).To(BeTrue())
	}
}
