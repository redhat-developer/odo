package devfile

import (
	"path/filepath"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile delete command tests", func() {
	var componentName string

	var globals helper.Globals

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEach()

		componentName = helper.RandString(6)

		helper.Chdir(globals.Context)

		// Devfile commands require experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeach(globals)

	})

	Context("when devfile delete command is executed", func() {

		It("should delete the component created from the devfile and also the owned resources", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io")

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project)

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "--project", globals.Project, "-f")

			globals.CliRunner.WaitAndCheckForExistence("deployments", globals.Project, 1)
			globals.CliRunner.WaitAndCheckForExistence("pods", globals.Project, 1)
			globals.CliRunner.WaitAndCheckForExistence("services", globals.Project, 1)
			globals.CliRunner.WaitAndCheckForExistence("ingress", globals.Project, 1)
		})
	})

	Context("when devfile delete command is executed with all flag", func() {

		It("should delete the component created from the devfile and also the env and odo folders and the odo-index-file.json file", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", globals.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), globals.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(globals.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", globals.Project)

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--context", globals.Context)

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "--project", globals.Project, "-f", "--all")

			globals.CliRunner.WaitAndCheckForExistence("deployments", globals.Project, 1)

			files := helper.ListFilesInDir(globals.Context)
			Expect(files).To(Not(ContainElement(".odo")))
		})
	})
})
