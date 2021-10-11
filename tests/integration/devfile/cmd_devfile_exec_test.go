package devfile

import (
	"path/filepath"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile exec command tests", func() {
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

	When("a component is created", func() {

		BeforeEach(func() {
			helper.Cmd("odo", "create", cmpName, "--context", commonVar.Context, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

		})

		It("should error out", func() {
			By("exec on a non deployed component", func() {
				err := helper.Cmd("odo", "exec", "--context", commonVar.Context, "--", "touch", "/projects/blah.js").ShouldFail().Err()
				Expect(err).To(ContainSubstring("doesn't exist on the cluster"))
			})
			By("exec with invalid devfile flag", func() {
				err := helper.Cmd("odo", "exec", "--context", commonVar.Context, "--devfile", "invalid.yaml", "--", "touch", "/projects/blah.js").ShouldFail().Err()
				Expect(err).To(ContainSubstring("unknown flag: --devfile"))
			})

			// TODO(feloy): Uncomment when https://github.com/openshift/odo/issues/5012 is fixed
			//	By("exec from a context with no component", func() {
			//		err := helper.Cmd("odo", "exec", "--context", "/tmp", "--", "touch", "/projects/blah.js").ShouldFail().Err()
			//		Expect(err).To(ContainSubstring("the current directory does not contain an odo component"))
			//	})
		})

		When("odo push is executed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})

			It("should execute the given command successfully in the container", func() {
				helper.Cmd("odo", "exec", "--context", commonVar.Context, "--", "touch", "/projects/blah.js").ShouldPass()
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				listDir := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				Expect(listDir).To(ContainSubstring("blah.js"))
			})

			It("should error out when no command is given by the user", func() {
				output := helper.Cmd("odo", "exec", "--context", commonVar.Context, "--").ShouldFail().Err()
				Expect(output).To(ContainSubstring("no command was given"))
			})

			It("should error out when an invalid command is given by the user", func() {
				output := helper.Cmd("odo", "exec", "--context", commonVar.Context, "--", "invalidCommand").ShouldFail().Err()
				Expect(output).To(ContainSubstring("executable file not found in $PATH"))
			})
		})

	})
})
