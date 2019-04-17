package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/e2e/helper"
)

// following command will tests in Describe section below in parallel (in 2 nodes)
// ginkgo -nodes=2 -focus="Example of a clean test" slowSpecThreshold=120 -randomizeAllSpecs  tests/e2e/
var _ = Describe("Example of a clean test", func() {
	//new clean project and context for each test
	var project string
	var context string

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalDir string
	var originalProject string

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

		var _ = Context("when project from KUBECONFIG is used", func() {
			// Set active project for each test spec
			var _ = JustBeforeEach(func() {
				originalProject = helper.OcGetCurrentProject()
				helper.OcSwitchProject(project)
			})
			// go back to original project after each test
			var _ = JustAfterEach(func() {
				helper.OcSwitchProject(originalProject)
			})

			It("create local nodejs component and push code", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context)

				helper.CmdShouldPass("odo component create nodejs")
				//TODO: verify that config was properly created
				helper.CmdShouldPass("odo push")
				//TODO: verify resources on cluster
			})

		})

		var _ = Context("when --project flag is used", func() {
			It("create local nodejs component and push code", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context)

				helper.CmdShouldPass("odo component create nodejs --project " + project + "")
				//TODO: verify that config was properly created
				helper.CmdShouldPass("odo push")
				//TODO: verify resources on cluster
			})

		})

		var _ = Context("when --context is used", func() {
			var _ = Context("when project from KUBECONFIG is used", func() {
				// Set active project for each test spec
				var _ = JustBeforeEach(func() {
					helper.OcSwitchProject(project)
				})
				// go back to original project after each test
				var _ = JustAfterEach(func() {
					helper.OcSwitchProject(originalProject)
				})

				It("create local nodejs component and push code", func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), context)

					helper.CmdShouldPass("odo component create nodejs nodejs --context " + context)
					//TODO: verify that config was properly created
					helper.CmdShouldPass("odo push")
					//TODO: verify resources on cluster
				})

			})

			var _ = Context("when --project flag is used", func() {
				It("create local nodejs component and push code", func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), context)

					helper.CmdShouldPass("odo component create nodejs nodejs --project " + project + " --context " + context)
					//TODO: verify that config was properly created
					helper.CmdShouldPass("odo push")
					//TODO: verify resources on cluster
				})
			})
		})

	})
})
