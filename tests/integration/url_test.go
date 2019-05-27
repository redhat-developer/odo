package integration

import (
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
	var originalDir string
	var originalProject string
	var oc helper.OcRunner

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
		originalDir = helper.Getwd()
		originalProject = oc.GetCurrentProject()
		helper.Chdir(context)
		oc.SwitchProject(project)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.Chdir(originalDir)
		oc.SwitchProject(originalProject)
		helper.DeleteProject(project)
		helper.DeleteDir(context)
	})

	Context("Listing urls", func() {
		It("should list appropriate URLs and push message", func() {
			var stdout string
			url1 := helper.RandString(5)
			url2 := helper.RandString(5)
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", project, componentName, "--ref", "master", "--git", "https://github.com/openshift/nodejs-ex")
			helper.CmdShouldPass("odo", "push")
			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))
			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080")
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{url1, "<not created on cluster>", "Present", "create URLs", "odo push"})
			helper.CmdShouldPass("odo", "push")
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{url1, "Present"})
			helper.DontMatchAllInOutput(stdout, []string{"<not created on cluster>", "odo push"})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{url1, "Absent", "delete URLs", "odo push"})
			helper.CmdShouldPass("odo", "url", "create", url2, "--port", "8000")
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{url1, "Absent", url2, "Present", "create/delete URLs", "odo push"})
			helper.CmdShouldPass("odo", "push")
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{url2, "Present"})
			helper.DontMatchAllInOutput(stdout, []string{url1, "Absent", "odo push"})
		})
	})
})
