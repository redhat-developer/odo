package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile test command tests", func() {
	var cmpName string
	var sourcePath = "/projects"
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should error out on devfile flag", func() {
		helper.Cmd("odo", "test", "--devfile", "invalid.yaml", "--context", commonVar.Context).ShouldFail()
	})

	When("Create a nodejs component", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", "--context", commonVar.Context, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})

		It("should show error if component is not pushed", func() {
			output := helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("error occurred while getting the pod: pod not found for the selector"))
		})

		When("doing odo push", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})
			It("should error out if a non-existent command or a command from wrong group is specified", func() {
				By("should error out if a non-existent command", func() {
					output := helper.Cmd("odo", "test", "--test-command", "invalidcmd", "--context", commonVar.Context).ShouldFail().Err()
					Expect(output).To(ContainSubstring("not found in the devfile"))
				})
				By("a command from wrong group is specified", func() {
					output := helper.Cmd("odo", "test", "--test-command", "devrun", "--context", commonVar.Context).ShouldFail().Err()
					Expect(output).To(ContainSubstring("command devrun is of group run in devfile.yaml"))
				})
			})

			It("Should run test command successfully with only one default specified", func() {
				output := helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{"Executing test1 command", "mkdir test1"})

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
				Expect(output).To(ContainSubstring("test1"))
			})

			It("Should run test command successfully with test-command specified", func() {
				output := helper.Cmd("odo", "test", "--test-command", "test2", "--context", commonVar.Context).ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{"Executing test2 command", "mkdir test2"})

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
				Expect(output).To(ContainSubstring("test2"))
			})

			It("Should run composite test command successfully", func() {
				output := helper.Cmd("odo", "test", "--test-command", "compositetest", "--context", commonVar.Context).ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{"Executing test1 command", "mkdir test1", "Executing test2 command", "mkdir test2"})

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
				helper.MatchAllInOutput(output, []string{"test1", "test2"})
			})
		})

		When("removing default command from devfile", func() {
			BeforeEach(func() {
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "isDefault: true", "")
			})

			When("doing odo push", func() {
				output := ""
				BeforeEach(func() {
					output = helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldFail().Err()
				})
				It("push should fail", func() {
					Expect(output).To(ContainSubstring("command group test warning - there should be exactly one default command, currently there is no default command"))
				})
				It("should show error if devfile has no default test command", func() {
					output := helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldFail().Err()
					Expect(output).To(ContainSubstring("command group test warning - there should be exactly one default command, currently there is no default command"))
				})
			})
		})

		When("using devfile that doesn't contains group of kind \"test\" and doing odo push", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})
			It("should show error if no test group is defined", func() {
				output := helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldFail().Err()
				Expect(output).To(ContainSubstring("the command group of kind \"test\" is not found in the devfile"))
			})
		})

		When("devfile has multiple default test command", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			})

			It("should show error if devfile has multiple default test command", func() {
				By("should fail on odo push", func() {
					output := helper.Cmd("odo", "push", "--build-command", "firstbuild", "--run-command", "secondrun", "--context", commonVar.Context).ShouldFail().Err()
					Expect(output).To(ContainSubstring("command group test error - there should be exactly one default command, currently there are multiple default commands"))
				})
				By("should fail on odo test", func() {
					output := helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldFail().Err()
					Expect(output).To(ContainSubstring("command group test error - there should be exactly one default command, currently there are multiple default commands"))
				})
			})
		})
	})

})
