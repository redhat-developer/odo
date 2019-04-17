package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/e2e/helper"
)

// following command will tests in Describe section below in parallel (in 2 nodes)
// ginkgo -nodes=2 -focus="Example of a clean test" slowSpecThreshold=120 -randomizeAllSpecs  tests/e2e/
var _ = Describe("Component Tests", func() {
	//new clean project and context for each test
	var project string
	var context string

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalDir string
	// var originalProject string

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		project = helper.OcCreateRandProject()
		context = helper.CreateNewContext()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.AfterFailed()
		helper.OcDeleteProject(project)
		helper.DeleteDir(context)
	})

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
			It("create local nodejs component twice and fail", func() {

				helper.CopyExample(filepath.Join("source", "nodejs"), context)
				helper.CmdShouldPass("odo component create nodejs --project " + project + " --env key=value,key1=value1")
				output := helper.CmdShouldFail("odo component create nodejs --project " + project + " --env key=value,key1=value1")
				Expect(output).To(ContainSubstring("this directory already contains a component"))
			})

		})
	})
})
