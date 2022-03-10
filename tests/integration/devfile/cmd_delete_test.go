package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo delete command tests", func() {
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("a component is bootstrapped", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "project", "set", commonVar.Project).ShouldPass()
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-deploy.yaml")).ShouldPass()
		})
		When("the component is deployed in DEV mode", func() {
			BeforeEach(func() {
				session := helper.CmdRunner("odo", "dev")
				defer session.Kill()
				helper.WaitForOutputToContain("Waiting for something to change", 180, 10, session)

				list := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
				Expect(list).To(ContainSubstring(cmpName))
			})

			When("the component is deleted using its name and namespace from another directory", func() {
				var out string
				BeforeEach(func() {
					otherDir := filepath.Join(commonVar.Context, "tmp")
					helper.MakeDir(otherDir)
					helper.Chdir(otherDir)
					out = helper.Cmd("odo", "delete", "component", "--name", cmpName, "--namespace", commonVar.Project, "-f").ShouldPass().Out()
				})

				It("should have deleted the component", func() {
					By("listing the resource to delete", func() {
						Expect(out).To(ContainSubstring("Deployment: " + cmpName))
					})
					By("deleting the deployment", func() {
						list := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
						Expect(list).To(Not(ContainSubstring(cmpName)))
					})
				})
			})
		})

		When("the component is deployed in DEPLOY mode", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()

				list := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
				Expect(list).To(ContainSubstring("my-component"))

				commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project, "-o", "yaml")
			})

			When("the component is deleted using its name and namespace from another directory", func() {
				var out string
				BeforeEach(func() {
					otherDir := filepath.Join(commonVar.Context, "tmp")
					helper.MakeDir(otherDir)
					helper.Chdir(otherDir)
					out = helper.Cmd("odo", "delete", "component", "--name", cmpName, "--namespace", commonVar.Project, "-f").ShouldPass().Out()
				})

				It("should have deleted the component", func() {
					By("listing the resource to delete", func() {
						Expect(out).To(ContainSubstring("Deployment: my-component"))
					})
					By("deleting the deployment", func() {
						list := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
						Expect(list).To(Not(ContainSubstring("my-component")))
					})
				})
			})
		})
	})
})
