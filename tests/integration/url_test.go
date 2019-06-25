package integration

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odoURLIntegration", func() {
	//new clean project and context for each test
	var project string
	var context string

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalProject string
	var oc helper.OcRunner

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		// Set default timeout for Eventually assertions
		// commands like odo push, might take a long time
		SetDefaultEventuallyTimeout(10 * time.Minute)

		// initialize oc runner
		// right now it uses oc binary, but we should convert it to client-go
		oc = helper.NewOcRunner("oc")
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Listing urls", func() {
		// Set active project for each test spec
		JustBeforeEach(func() {
			originalProject = oc.GetCurrentProject()
			// WARNING: this project switching makes it impossible to run this in parallel
			// it should set different KUBECONFIG before witching project
			oc.SwitchProject(project)
		})
		// go back to original project after each test
		JustAfterEach(func() {
			oc.SwitchProject(originalProject)
		})

		It("should list appropriate URLs and push message", func() {
			var stdout string
			url1 := helper.RandString(5)
			url2 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--project", project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", context)
			stdout = helper.CmdShouldFail("odo", "url", "list", "--context", context)
			Expect(stdout).To(ContainSubstring("no URLs found"))
			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--context", context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			helper.MatchAllInOutput(stdout, []string{url1, "<not created on cluster>", "Present", "create URLs", "odo push"})
			helper.CmdShouldPass("odo", "push", "--context", context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			helper.MatchAllInOutput(stdout, []string{url1, "Present"})
			helper.DontMatchAllInOutput(stdout, []string{"<not created on cluster>", "odo push"})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f", "--context", context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			helper.MatchAllInOutput(stdout, []string{url1, "Absent", "delete URLs", "odo push"})
			helper.CmdShouldPass("odo", "url", "create", url2, "--port", "8000", "--context", context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			helper.MatchAllInOutput(stdout, []string{url1, "Absent", url2, "Present", "create/delete URLs", "odo push"})
			helper.CmdShouldPass("odo", "push", "--context", context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			helper.MatchAllInOutput(stdout, []string{url2, "Present"})
			helper.DontMatchAllInOutput(stdout, []string{url1, "Absent", "odo push"})
		})
	})
})
