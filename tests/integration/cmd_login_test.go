package integration

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odoLoginE2e", func() {
	// user related constants
	const loginTestUserForNoProject = "odologinnoproject"
	const loginTestUserPassword = "developer"
	var session string
	var testUserToken string
	var oc helper.OcRunner

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
	})

	Context("when running help for login command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "login", "-h")
			Expect(appHelp).To(ContainSubstring("Login to cluster"))
		})
	})

	Context("when login using parameter username, password, token", func() {
		It("login should successful", func() {
			session = helper.CmdShouldPass("odo", "login", "-u", loginTestUserForNoProject, "-p", loginTestUserPassword)
			Expect(session).To(ContainSubstring("Login successful"))
			Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
			Expect(session).To(ContainSubstring("odo project create <project-name>"))
			session = oc.GetLoginUser()
			Expect(session).To(ContainSubstring(loginTestUserForNoProject))
			// One initialization needs one login, hence it happens here
			testUserToken = oc.GetToken()
			session = helper.CmdShouldPass("odo", "login", "-t", testUserToken)
			Expect(session).To(ContainSubstring("Logged into"))
			Expect(session).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
			Expect(session).To(ContainSubstring("odo project create <project-name>"))
			session = oc.GetLoginUser()
			Expect(session).To(ContainSubstring(loginTestUserForNoProject))
			sessionErr := helper.CmdShouldFail("odo", "login", "-t", "verybadtoken")
			Expect(sessionErr).To(ContainSubstring("The token provided is invalid or expired"))
		})
	})
})
