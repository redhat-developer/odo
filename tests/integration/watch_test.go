package integration

import (
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odoWatchE2e", func() {
	var project string
	var context string

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalDir string
	//var oc helper.OcRunner

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		//oc = helper.NewOcRunner("oc")
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
	})

	//new clean project and context for each test
	var _ = Context("when component is in the current directory", func() {

		// we will be testing components that are created from the current directory
		// switch to the clean context dir before each test
		var _ = JustBeforeEach(func() {
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		// go back to original directory after each test
		var _ = JustAfterEach(func() {
			helper.Chdir(originalDir)
		})

		var _ = Context("when --project flag is used", func() {
			It("odo watch fail when component not pushed", func() {

				helper.CopyExample(filepath.Join("source", "nodejs"), context)
				helper.CmdShouldPass("odo", "component", "create", "nodejs", "--project", project)
				output := helper.CmdShouldFail("odo", "watch")
				Expect(output).To(ContainSubstring("component does not exist. Please use `odo push` to create you component"))
			})
		})
	})

})
