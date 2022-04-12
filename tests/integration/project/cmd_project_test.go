package project

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo project command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should display the help of project command", func() {
		projectHelp := helper.Cmd("odo", "project", "-h").ShouldPass().Out()
		Expect(projectHelp).To(ContainSubstring("Perform project operations"))
	})

	It("should display only the project name when running command with -q flag", func() {
		projectName := helper.Cmd("odo", "project", "get", "-q").ShouldPass().Out()
		Expect(projectName).Should(Equal(commonVar.Project))
	})

	It("should list current empty project", func() {
		helper.WaitForCmdOut("odo", []string{"project", "list"}, 1, true, func(output string) bool {
			return strings.Contains(output, commonVar.Project)
		})
	})

	When("creating a new project", func() {
		var projectName string

		BeforeEach(func() {
			projectName = "cmd-project-" + helper.RandString(6)
			helper.Cmd("odo", "project", "create", projectName).ShouldPass()
		})

		It("should delete a project with --wait", func() {
			output := helper.Cmd("odo", "project", "delete", projectName, "-f", "--wait").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Waiting for project to be deleted"))
		})
	})

	When("user is logged out", func() {
		var token string
		BeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("Openshift specific scenario.")
			}
			ocRunner := helper.NewOcRunner("oc")
			token = ocRunner.GetToken()
			helper.Cmd("odo", "logout").ShouldPass()
		})
		AfterEach(func() {
			helper.Cmd("odo", "login", "--token", token).ShouldPass()
		})
		It("should show login message when setting project and not login", func() {
			err := helper.Cmd("odo", "project", "set", "something").ShouldFail().Err()
			Expect(err).To(ContainSubstring("Unauthorized to access the cluster"))
		})
		It("should show login message when deleting project and not login", func() {
			err := helper.Cmd("odo", "project", "delete", "something").ShouldFail().Err()
			Expect(err).To(ContainSubstring("Unauthorized to access the cluster"))
		})
	})
})
