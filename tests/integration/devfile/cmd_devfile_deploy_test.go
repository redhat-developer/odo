package devfile

import (
	segment "github.com/redhat-developer/odo/pkg/segment/context"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile deploy command tests", func() {

	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("directory is empty", func() {

		BeforeEach(func() {
			Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
		})

		It("should error", func() {
			output := helper.Cmd("odo", "deploy").ShouldFail().Err()
			Expect(output).To(ContainSubstring("this command cannot run in an empty directory"))

		})
	})

	When("using a devfile.yaml containing a deploy command", func() {
		// from devfile
		cmpName := "nodejs-prj1-api-abhz"
		deploymentName := "my-component"
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
		})

		When("running odo deploy", func() {
			var stdout string
			BeforeEach(func() {
				stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
				// An ENV file should have been created indicating current namespace
				Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeTrue())
				helper.FileShouldContainSubstring(".odo/env/env.yaml", "Project: "+commonVar.Project)
			})
			It("should succeed", func() {
				By("building and pushing image to registry", func() {
					Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
					Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
				})
				By("deploying a deployment with the built image", func() {
					out := commonVar.CliRunner.Run("get", "deployment", deploymentName, "-n", commonVar.Project, "-o", `jsonpath="{.spec.template.spec.containers[0].image}"`).Wait().Out.Contents()
					Expect(out).To(ContainSubstring("quay.io/unknown-account/myimage"))
				})
			})

			It("should run odo dev successfully", func() {
				_, _, _, _, err := helper.StartDevMode()
				Expect(err).ToNot(HaveOccurred())
			})

			When("deleting previous deployment and switching kubeconfig to another namespace", func() {
				var otherNS string
				BeforeEach(func() {
					helper.Cmd("odo", "delete", "component", "--name", cmpName, "-f").ShouldPass()
					output := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Err.Contents()
					Expect(string(output)).To(ContainSubstring("No resources found in " + commonVar.Project + " namespace."))

					otherNS = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
				})

				AfterEach(func() {
					commonVar.CliRunner.DeleteNamespaceProject(otherNS)
				})

				It("should run odo deploy on initial namespace", func() {
					helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()

					output := commonVar.CliRunner.Run("get", "deployment").Err.Contents()
					Expect(string(output)).To(ContainSubstring("No resources found in " + otherNS + " namespace."))

					output = commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
					Expect(string(output)).To(ContainSubstring(deploymentName))
				})

			})
		})
	})

	When("using a devfile.yaml containing two deploy commands", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-two-deploy-commands.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
		})
		It("should run odo deploy", func() {
			stdout := helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			By("building and pushing image to registry", func() {
				Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
				Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
			})
			By("deploying a deployment with the built image", func() {
				out := commonVar.CliRunner.Run("get", "deployment", "my-component", "-n", commonVar.Project, "-o", `jsonpath="{.spec.template.spec.containers[0].image}"`).Wait().Out.Contents()
				Expect(out).To(ContainSubstring("quay.io/unknown-account/myimage"))
			})
		})
	})

	When("recording telemetry data", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
			helper.EnableTelemetryDebug()
			helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
		})
		AfterEach(func() {
			helper.ResetTelemetry()
		})
		It("should record the telemetry data correctly", func() {
			td := helper.GetTelemetryDebugData()
			Expect(td.Event).To(ContainSubstring("odo deploy"))
			Expect(td.Properties.Success).To(BeTrue())
			Expect(td.Properties.Error == "").To(BeTrue())
			Expect(td.Properties.ErrorType == "").To(BeTrue())
			Expect(td.Properties.CmdProperties[segment.ComponentType]).To(ContainSubstring("nodejs"))
			Expect(td.Properties.CmdProperties[segment.Language]).To(ContainSubstring("javascript"))
			Expect(td.Properties.CmdProperties[segment.ProjectType]).To(ContainSubstring("nodejs"))
		})
	})
})
