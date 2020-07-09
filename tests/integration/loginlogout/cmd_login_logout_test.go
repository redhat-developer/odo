package integration

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo login and logout command tests", func() {
	// user related constants
	const loginTestUserForNoProject = "odologinnoproject"
	const loginTestUserPassword = "developer"
	var session1 string
	var testUserToken string
	var oc helper.OcRunner
	var currentUserToken string

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		oc = helper.NewOcRunner("oc")
	})

	Context("when running help for login command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "login", "-h")
			Expect(appHelp).To(ContainSubstring("Login to cluster"))
		})
	})

	Context("when running help for logout command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "logout", "-h")
			Expect(appHelp).To(ContainSubstring("Log out of the current OpenShift session"))
		})
	})

	Context("when running login tests", func() {
		JustBeforeEach(func() {
			// To make sure that the script is using the developer login session
			helper.CmdShouldPass("odo", "login", "-u", "developer", "-p", "developer")
		})
		It("should successful with correct credentials and fails with incorrect token", func() {
			// Current user login token
			currentUserToken = oc.GetToken()

			// Login successful without any projects with appropriate message
			session1 = helper.CmdShouldPass("odo", "login", "-u", loginTestUserForNoProject, "-p", loginTestUserPassword)
			Expect(session1).To(ContainSubstring("Login successful"))
			Expect(session1).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
			Expect(session1).To(ContainSubstring("odo project create <project-name>"))
			session1 = oc.GetLoginUser()
			Expect(session1).To(ContainSubstring(loginTestUserForNoProject))

			// odologinnoproject user login token
			testUserToken = oc.GetToken()

			// Login successful with token without any projects with appropriate message
			session1 = helper.CmdShouldPass("odo", "login", "-t", testUserToken)
			Expect(session1).To(ContainSubstring("Logged into"))
			Expect(session1).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
			Expect(session1).To(ContainSubstring("odo project create <project-name>"))
			session1 = oc.GetLoginUser()
			Expect(session1).To(ContainSubstring(loginTestUserForNoProject))

			// Login fails on invalid token with appropriate message
			sessionErr := helper.CmdShouldFail("odo", "login", "-t", "verybadtoken")
			Expect(sessionErr).To(ContainSubstring("The token provided is invalid or expired"))

			// loging back to current user
			helper.CmdShouldPass("odo", "login", "--token", currentUserToken)
		})
	})
})
