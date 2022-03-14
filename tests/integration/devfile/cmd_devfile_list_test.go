package devfile

import (
	"path"
	"path/filepath"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo list with devfile", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("a component created for deployment", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.Chdir(commonVar.Context)
		})

		AfterEach(func() {
			helper.Cmd("odo", "v2delete", "-a").ShouldPass()
		})

		It("show an odo deploy in the list", func() {

			By("should display the component as 'None' in odo list", func() {
				stdOut := helper.Cmd("odo", "list").ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring("None"))
			})

			// Fake the odo deploy image build / push passing in "echo" to PODMAN
			helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()

			By("should display the component as 'Deploy', and 'nodejs' in odo list", func() {
				stdOut := helper.Cmd("odo", "list").ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring("Deploy"))
				Expect(stdOut).To(ContainSubstring("nodejs"))
			})

		})

		It("show an odo dev in the list", func() {

			// Deploy odo dev
			session := helper.CmdRunner("odo", "dev")
			defer session.Kill()
			helper.WaitForOutputToContain("Waiting for something to change", 180, 10, session)

			By("should display the component as 'Dev', and 'nodejs' in odo list", func() {
				stdOut := helper.Cmd("odo", "list").ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring("Dev"))
				Expect(stdOut).To(ContainSubstring("nodejs"))
			})

			// Fake the odo deploy image build / push passing in "echo" to PODMAN
			stdout := helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			By("building and pushing image to registry", func() {
				Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
				Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
			})

			By("should display the component as being deployed both Dev and Deploy", func() {
				stdOut := helper.Cmd("odo", "list").ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring("Dev, Deploy"))
			})

		})

	})

	When("listing a component outside the main directory", func() {
		var deployStdout, listStdout, newContext string

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.ReplaceString("devfile.yaml", "nodejs-prj1-api-abhz", "odo-list-dir-test")

			// cd to the project directory
			// Fake the odo deploy image build / push passing in "echo" to PODMAN
			helper.Chdir(commonVar.Context)
			deployStdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()

			// cd to outside the directory (new context)
			newContext = helper.CreateNewContext()
			helper.Chdir(newContext)
			listStdout = helper.Cmd("odo", "list").ShouldPass().Out()

		})

		AfterEach(func() {
			helper.Chdir(commonVar.Context)
			helper.Cmd("odo", "v2delete", "-a").ShouldPass()
			helper.DeleteDir(newContext)
		})

		It("show an odo deploy in the list", func() {

			// Check that the fake deploy worked
			Expect(deployStdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
			Expect(deployStdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))

			// Check that odo list contains the component name when listing outside the directory
			By("should display the component in odo list even when running outside the directory", func() {
				Expect(listStdout).To(ContainSubstring("odo-list-dir-test"))
			})
		})

	})

})
