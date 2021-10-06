package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
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

	Context("Should show proper errors", func() {

		// used ";" as consolidating symbol as this spec covers multiple scenerios
		It("should show error if component is not pushed; should error out if a non-existent command or a command from wrong group is specified", func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--context", commonVar.Context, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml")).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

			output := helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("error occurred while getting the pod: pod not found for the selector"))

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			output = helper.Cmd("odo", "test", "--test-command", "invalidcmd", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("not found in the devfile"))

			output = helper.Cmd("odo", "test", "--test-command", "devrun", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("command devrun is of group run in devfile.yaml"))
		})

		It("should show error if no test group is defined", func() {
			helper.Cmd("odo", "create", "--context", commonVar.Context, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			output := helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldFail().Err()

			Expect(output).To(ContainSubstring("the command group of kind \"test\" is not found in the devfile"))
		})

		It("should show error if devfile has no default test command", func() {
			helper.Cmd("odo", "create", "--context", commonVar.Context, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "isDefault: true", "")
			output := helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("command group test error - there should be exactly one default command, currently there is no default command"))
			output = helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("command group test error - there should be exactly one default command, currently there is no default command"))
		})

		It("should show error if devfile has multiple default test command", func() {
			helper.Cmd("odo", "create", "--context", commonVar.Context, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-defaults.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			output := helper.Cmd("odo", "push", "--build-command", "firstbuild", "--run-command", "secondrun", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("command group test error - there should be exactly one default command, currently there is more than one default command"))
			output = helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("command group test error - there should be exactly one default command, currently there is more than one default command"))
		})

		It("should error out on devfile flag", func() {
			helper.Cmd("odo", "test", "--devfile", "invalid.yaml", "--context", commonVar.Context).ShouldFail()
		})
	})

	Context("Should run test command successfully", func() {

		It("Should run test command successfully with only one default specified", func() {
			helper.Cmd("odo", "create", "--context", commonVar.Context, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml")).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			output := helper.Cmd("odo", "test", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{"Executing test1 command", "mkdir test1"})

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
			Expect(output).To(ContainSubstring("test1"))
		})

		It("Should run test command successfully with test-command specified", func() {
			helper.Cmd("odo", "create", "--context", commonVar.Context, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml")).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			output := helper.Cmd("odo", "test", "--test-command", "test2", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{"Executing test2 command", "mkdir test2"})

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
			Expect(output).To(ContainSubstring("test2"))
		})

		It("Should run composite test command successfully", func() {
			helper.Cmd("odo", "create", "--context", commonVar.Context, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-testgroup.yaml")).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			output := helper.Cmd("odo", "test", "--test-command", "compositetest", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{"Executing test1 command", "mkdir test1", "Executing test2 command", "mkdir test2"})

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
			helper.MatchAllInOutput(output, []string{"test1", "test2"})
		})
	})

})
