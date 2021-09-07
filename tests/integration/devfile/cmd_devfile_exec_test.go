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
		helper.SetDefaultDevfileRegistryAsStaging()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("a component is created", func() {

		BeforeEach(func() {
			args := []string{"create", "nodejs", cmpName, "--context", commonVar.Context}
			helper.Cmd("odo", args...).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

		})

		It("should error out when a component is not present or when a devfile flag is used", func() {
			args := []string{"exec", "--context", commonVar.Context}
			args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
			helper.Cmd("odo", args...).ShouldFail()

			args = []string{"exec", "--context", commonVar.Context, "--devfile", "invalid.yaml"}
			args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
			helper.Cmd("odo", args...).ShouldFail()
		})

		When("odo push is executed", func() {
			BeforeEach(func() {
				args := []string{"push", "--context", commonVar.Context}
				helper.Cmd("odo", args...).ShouldPass()
			})

			It("should execute the given command successfully in the container", func() {
				args := []string{"exec", "--context", commonVar.Context}
				args = append(args, []string{"--", "touch", "/projects/blah.js"}...)
				helper.Cmd("odo", args...).ShouldPass()

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				listDir := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				Expect(listDir).To(ContainSubstring("blah.js"))
			})

			It("should error out when no command is given by the user", func() {
				args := []string{"exec", "--context", commonVar.Context}
				args = append(args, "--")
				output := helper.Cmd("odo", args...).ShouldFail().Err()

				Expect(output).To(ContainSubstring("no command was given"))
			})

			It("should error out when a invalid command is given by the user", func() {
				args := []string{"exec", "--context", commonVar.Context}
				args = append(args, "--", "invalidCommand")
				output := helper.Cmd("odo", args...).ShouldFail().Out()

				Expect(output).To(ContainSubstring("executable file not found in $PATH"))
			})
		})

	})
})
