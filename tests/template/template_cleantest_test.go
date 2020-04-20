package template

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
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
	var oc helper.OcRunner

	BeforeEach(func() {
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

	AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
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

			It("create local nodejs component and push code", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context)

				//helper.CmdShouldPass("odo component create nodejs")
				helper.CmdShouldPass("odo", "component", "create", "nodejs")
				//TODO: verify that config was properly created
				helper.CmdShouldPass("odo", "push")
				//TODO: verify resources on cluster
			})

		})

		var _ = Context("when --project flag is used", func() {
			It("create local nodejs component and push code", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), context)

				helper.CmdShouldPass("odo", "component", "create", "nodejs", "--project", project)
				//TODO: verify that config was properly created
				helper.CmdShouldPass("odo", "push")
				//TODO: verify resources on cluster
			})

		})

		var _ = Context("when --context is used", func() {
			var _ = Context("when project from KUBECONFIG is used", func() {
				// Set active project for each test spec
				var _ = JustBeforeEach(func() {
					//helper.OcSwitchProject(project)
				})
				// go back to original project after each test
				var _ = JustAfterEach(func() {
					//helper.OcSwitchProject(originalProject)
				})

				It("create local nodejs component and push code", func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), context)

					helper.CmdShouldPass("odo", "component", "create", "nodejs", "nodejs", "--context", context)
					//TODO: verify that config was properly created
					helper.CmdShouldPass("odo", "push")
					//TODO: verify resources on cluster
				})

			})

			var _ = Context("when --project flag is used", func() {
				It("create local nodejs component and push code", func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), context)

					helper.CmdShouldPass("odo", "component", "create", "nodejs", "nodejs", "--project", project, "--context", context)
					//TODO: verify that config was properly created
					helper.CmdShouldPass("odo", "push")
					//TODO: verify resources on cluster
				})
			})
		})

	})
})
