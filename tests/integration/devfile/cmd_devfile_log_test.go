package devfile

import (
	"path/filepath"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile log command tests", func() {
	var cmpName string
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

	Context("Verify odo log for devfile works", func() {

		It("should log run command output and fail for debug command", func() {

			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", commonVar.Project, cmpName, "--context", commonVar.Context)
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			output := helper.CmdShouldPass("odo", "log", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("ODO_COMMAND_RUN"))

			// It should fail for debug command as no debug command in devfile
			helper.CmdShouldFail("odo", "log", "--debug")

			/*
				Flaky Test odo log -f, see issue https://github.com/openshift/odo/issues/3809
				match, err := helper.RunCmdWithMatchOutputFromBuffer(30*time.Second, "program=devrun", "odo", "log", "-f")
				Expect(err).To(BeNil())
				Expect(match).To(BeTrue())
			*/

		})

		It("should error out if component does not exist", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, cmpName, "--context", commonVar.Context)
			helper.CmdShouldFail("odo", "log")
		})

		It("should log debug command output", func() {
			projectDir := filepath.Join(commonVar.Context, "projectDir")
			helper.CopyExample(filepath.Join("source", "web-nodejs-sample"), projectDir)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, cmpName, "--context", projectDir)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(projectDir, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDir)

			output := helper.CmdShouldPass("odo", "log", "--debug", "--context", projectDir)
			Expect(output).To(ContainSubstring("ODO_COMMAND_DEBUG"))

			/*
				Flaky Test odo log -f, see issue https://github.com/openshift/odo/issues/3809
				match, err := helper.RunCmdWithMatchOutputFromBuffer(30*time.Second, "program=debugrun", "odo", "log", "-f")
				Expect(err).To(BeNil())
				Expect(match).To(BeTrue())
			*/

		})

	})

})
