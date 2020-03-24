package integration

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile url command tests", func() {
	var project string
	var context string
	var currentWorkingDirectory string

	oc = helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
		context = helper.CreateNewDevfileContext()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Listing urls", func() {
		It("should list appropriate URLs and push message", func() {
			var stdout string
			url1 := helper.RandString(5)
			// url2 := helper.RandString(5)
			host := helper.RandString(5) + ".com"
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)
			stdout = helper.CmdShouldFail("odo", "url", "list", "--context", context)
			Expect(stdout).To(ContainSubstring("no URLs found"))

			stdout = helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080")
			Expect(stdout).To(ContainSubstring("is not exposed"))

			stdout = helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090")
			Expect(stdout).To(ContainSubstring("host must be provided"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host)

			// helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project)
			// stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			// helper.MatchAllInOutput(stdout, []string{url1, "Pushed"})
			// helper.DontMatchAllInOutput(stdout, []string{"Not Pushed", "odo push"})

			// helper.CmdShouldPass("odo", "url", "delete", url1, "-f", "--context", context)
			// stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			// helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url1, "odo push"})

			// helper.CmdShouldPass("odo", "url", "create", url2, "--port", "8000", "--context", context)
			// stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			// helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", url2, "Not Pushed", "odo push"})
			// helper.CmdShouldPass("odo", "push", "--context", context)
			// stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			// helper.MatchAllInOutput(stdout, []string{url2, "Pushed"})
			// helper.DontMatchAllInOutput(stdout, []string{url1, "Not Pushed", "odo push"})
		})
	})

})
