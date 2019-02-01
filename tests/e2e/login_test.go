package e2e

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odoLoginE2e", func() {
	// user related constants
	const loginTestUser = "testdeveloper"
	const loginTestUserPassword = "testdeveloper"
	const odoTestProject1 = "testproject1"

	// Comand related constants
	const baseOdoLoginCommand = "odo login"
	const baseOdoProjectCreate = "odo project create"
	const baseOdoProjectDelete = "odo project delete -f "
	const ocWhoamiCommand = "oc whoami"
	const ocTokenCommand = "oc whoami -t"

	// variables to be used in test
	var session string
	var backToCurrentUserCommand string
	var testUserLoginCommand string
	var testUserLoginCommandWithToken string
	var testUserLoginFailCommandWithToken string
	var testUserCreateProject1Command string

	Describe("Check for successful login and logout", func() {
		Context("Initialize", func() {
			It("Should initialize some variables", func() {
				// Save currently logged in users token, so we can get back to that context after being done
				t := runCmd(ocTokenCommand)
				backToCurrentUserCommand = fmt.Sprintf("%s -t %s", baseOdoLoginCommand, t)
				testUserLoginCommand = fmt.Sprintf("%s -u %s -p %s", baseOdoLoginCommand, loginTestUser, loginTestUserPassword)
				testUserCreateProject1Command = fmt.Sprintf("%s %s", baseOdoProjectCreate, odoTestProject1)
				testUserLoginFailCommandWithToken = fmt.Sprintf("%s -t verybadtoken", baseOdoLoginCommand)
			})
		})

		Context("Run login tests", func() {
			AfterEach(func() {
				t := runCmd("oc projects -q")
				projects := strings.Split(t, "\n")
				for _, p := range projects {
					if len(p) > 0 {
						runCmd(fmt.Sprintf("%s %s", baseOdoProjectDelete, p))
						time.Sleep(100)
					}
				}
				runCmd(backToCurrentUserCommand)
			})

			It("Should login successfully with username and password without any projects with appropriate message", func() {
				session = runCmd(testUserLoginCommand)
				Expect(session).To(ContainSubstring("Login successful"))
				Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session).To(ContainSubstring("odo project create <project-name>"))
				session = runCmd(ocWhoamiCommand)
				Expect(session).To(ContainSubstring(loginTestUser))
				token := runCmd(ocTokenCommand)
				// One initialization needs one login, hence it happens here
				testUserLoginCommandWithToken = fmt.Sprintf("%s -t %s", baseOdoLoginCommand, token)
			})

			It("Should login successfully with token without any projects with appropriate message", func() {
				session = runCmd(testUserLoginCommandWithToken)
				Expect(session).To(ContainSubstring("Logged into"))
				Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
				Expect(session).To(ContainSubstring("odo project create <project-name>"))
				session = runCmd(ocWhoamiCommand)
				Expect(session).To(ContainSubstring(loginTestUser))
			})

			It("Should login successfully with username and password single project with appropriate message", func() {
				runCmd(testUserLoginCommand)
				runCmd(testUserCreateProject1Command)
				runCmd(backToCurrentUserCommand)
				session = runCmd(testUserLoginCommand)
				Expect(session).To(ContainSubstring("Login successful"))
				Expect(session).To(ContainSubstring(odoTestProject1))
				session = runCmd(ocWhoamiCommand)
				Expect(session).To(ContainSubstring(loginTestUser))
			})

			It("Should fail login on invalid token with appropriate message", func() {
				session = runFailCmd(testUserLoginFailCommandWithToken, 1)
				Expect(session).To(ContainSubstring("The token provided is invalid or expired"))
				runCmd(testUserLoginCommand)
			})
		})
	})
})
