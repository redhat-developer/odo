package integration

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo watch command tests", func() {
	var project string
	var context string
	var currentWorkingDirectory string

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when running help for watch command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "watch", "-h")
			helper.MatchAllInOutput(appHelp, []string{"Watch for changes", "git components"})
		})
	})

	Context("when executing watch without pushing the component", func() {
		It("should fail", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", "--project", project, "--context", context)
			output := helper.CmdShouldFail("odo", "watch", "--context", context)
			Expect(output).To(ContainSubstring("component does not exist. Please use `odo push` to create your component"))
		})
	})

	Context("when executing odo watch against an app that doesn't exist", func() {
		JustBeforeEach(func() {
			currentWorkingDirectory = helper.Getwd()
			helper.Chdir(context)
		})
		JustAfterEach(func() {
			helper.Chdir(currentWorkingDirectory)
		})
		It("should fail with proper error", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", "--project", project)
			output := helper.CmdShouldFail("odo", "watch", "--app", "dummy")
			Expect(output).To(ContainSubstring("component does not exist"))
		})
	})

	Context("when executing watch on a git source type component", func() {
		It("should fail", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "--context", context, "nodejs", "--git", "https://github.com/openshift/nodejs-ex.git")
			output := helper.CmdShouldFail("odo", "watch", "--context", context)
			Expect(output).To(ContainSubstring("Watch is supported by binary and local components only"))
		})
	})
})
