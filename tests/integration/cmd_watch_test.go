package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo watch command tests", func() {
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

	Context("when running help for watch command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "watch", "-h")
			helper.MatchAllInOutput(appHelp, []string{"Watch for changes", "git components"})
		})
	})

	Context("when executing watch without pushing the component", func() {
		It("should fail", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context)
			output := helper.CmdShouldFail("odo", "watch", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("component does not exist. Please use `odo push` to create your component"))
		})
	})

	Context("when executing odo watch against an app that doesn't exist", func() {
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})
		It("should fail with proper error", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", "--project", commonVar.Project)
			output := helper.CmdShouldFail("odo", "watch", "--app", "dummy")
			Expect(output).To(ContainSubstring("component does not exist"))
		})
	})

	Context("when executing watch on a git source type component", func() {
		It("should fail", func() {
			helper.CmdShouldPass("odo", "create", "--context", commonVar.Context, "nodejs", "--git", "https://github.com/openshift/nodejs-ex.git")
			output := helper.CmdShouldFail("odo", "watch", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("Watch is supported by binary and local components only"))
		})
	})
})
