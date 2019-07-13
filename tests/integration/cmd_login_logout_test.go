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
	const loginTestUserForSingleProject1 = "odologinsingleproject1"
	const odoTestProjectForSingleProject1 = "odologintestproject1"
	const loginTestUserPassword = "developer"
	var session1 string
	var testUserToken1 string
	var oc helper.OcRunner
	var currentUserToken1 string

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
	})

	Context("Run login tests with no active projects, having default is also considered as not having active project", func() {
		It("Should login successfully with username and password without any projects with appropriate message", func() {
			currentUserToken1 = oc.GetToken()
			session1 = helper.CmdShouldPass("odo", "login", "-u", loginTestUserForNoProject, "-p", loginTestUserPassword)
			Expect(session1).To(ContainSubstring("Login successful"))
			Expect(session1).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
			Expect(session1).To(ContainSubstring("odo project create <project-name>"))
			session1 = oc.GetLoginUser()
			Expect(session1).To(ContainSubstring(loginTestUserForNoProject))
			// One initialization needs one login, hence it happens here
			testUserToken1 = oc.GetToken()
		})

		It("Should login successfully with token without any projects with appropriate message", func() {
			session1 = helper.CmdShouldPass("odo", "login", "-t", testUserToken1)
			Expect(session1).To(ContainSubstring("Logged into"))
			Expect(session1).To(ContainSubstring("You don't have any projects. You can try to create a new project, by running"))
			Expect(session1).To(ContainSubstring("odo project create <project-name>"))
			session1 = oc.GetLoginUser()
			Expect(session1).To(ContainSubstring(loginTestUserForNoProject))
		})

		It("Should fail login on invalid token with appropriate message", func() {
			sessionErr := helper.CmdShouldFail("odo", "login", "-t", "verybadtoken")
			Expect(sessionErr).To(ContainSubstring("The token provided is invalid or expired"))
			oc.LoginUsingToken(currentUserToken1)
		})
	})
})
