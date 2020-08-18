package devfile

import (
	"path/filepath"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile log command tests", func() {
	var cmpName, projectDirPath string
	var projectDir = "/projectDir"
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()

		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		projectDirPath = commonVar.Context + projectDir
		// Devfile requires experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Verify odo log for devfile works", func() {

		It("should log run command output", func() {
			helper.MakeDir(projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			output := helper.CmdShouldPass("odo", "log")
			Expect(output).To(ContainSubstring("ODO_COMMAND_RUN"))

			/*
				Flaky Test odo log -f, see issue https://github.com/openshift/odo/issues/3809
				match, err := helper.RunCmdWithMatchOutputFromBuffer(30*time.Second, "program=devrun", "odo", "log", "-f")
				Expect(err).To(BeNil())
				Expect(match).To(BeTrue())
			*/

		})

		It("should error out if component does not exist", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, cmpName)
			helper.CmdShouldFail("odo", "log")
		})

		It("should log debug command output", func() {

			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), projectDirPath)
			helper.RenameFile("devfile-with-debugrun.yaml", "devfile.yaml")
			helper.CmdShouldPass("odo", "push", "--debug")

			output := helper.CmdShouldPass("odo", "log", "--debug")
			Expect(output).To(ContainSubstring("ODO_COMMAND_DEBUG"))

			/*
				Flaky Test odo log -f, see issue https://github.com/openshift/odo/issues/3809
				match, err := helper.RunCmdWithMatchOutputFromBuffer(30*time.Second, "program=debugrun", "odo", "log", "-f")
				Expect(err).To(BeNil())
				Expect(match).To(BeTrue())
			*/

		})

		// we do not need test for run command as odo push fails
		// if there is no run command in devfile.
		It("should give error if no debug command in devfile", func() {

			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.CmdShouldFail("odo", "log", "--debug")
		})

	})

})
