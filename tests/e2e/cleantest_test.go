package e2e

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/e2e/helper"
)

//  current directory and project (before eny test is run) so it can restored  after all testing is done
var originalDir string
var originalProject string

// BeforeSuite is run before whole test suite
// there can be only one BeforeSuite
var _ = BeforeSuite(func() {
	// seed pseudo-random generator
	rand.Seed(time.Now().UTC().UnixNano())

	// this needs to bedeclared separately
	// if := is used in the next line it won't save originalDir to global variable
	var err error
	originalDir, err = os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	originalProject = helper.OcCurrentProject()
})

// AfterSuite is run before whole test suite
// there can be only one AfterSuite
var _ = AfterSuite(func() {
	if originalDir != "" {
		helper.Chdir(originalDir)
	}
	if originalProject != "" {
		helper.OcSwitchProject(originalProject)
	}
})

var _ = Describe("Example of a clean test", func() {
	var project string
	var context string

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
		helper.OcDeleteProject(project)
		helper.DeleteDir(context)
	})

	var _ = Context("when component is in the current directory", func() {

		// we will be testing components that are created from the current directory
		// switch to the clean context dir before each test
		var _ = JustBeforeEach(func() {
			helper.Chdir(context)
		})

		// go back to original directory after each test
		var _ = JustAfterEach(func() {
			helper.Chdir(originalDir)
		})

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
