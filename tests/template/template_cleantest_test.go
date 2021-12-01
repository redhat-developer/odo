package template

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

// following command will run tests in Describe section below in parallel (in 2 nodes)
// ginkgo -nodes=2 -focus="Example of a clean test" slowSpecThreshold=120 -randomizeAllSpecs  tests/e2e/
var _ = Describe("Example of a clean test", func() {
	// new clean project and context for each test
	var project string
	var context string

	// current directory and component name (before any test runs) so that it can be restored  after all testing is done
	var originalDir string
	var cmpName string

	BeforeEach(func() {
		// Set default timeout for Eventually assertions
		// commands like odo push, might take a long time
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "preference.yaml"))
		project = helper.CreateRandProject()

		// we will be testing components that are created from the current directory
		// switch to the clean context dir before each test
		originalDir = helper.Getwd()
		helper.Chdir(context)

	})

	AfterEach(func() {
		helper.DeleteProject(project)
		// go back to original directory after each test
		helper.Chdir(originalDir)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when --project flag is used", func() {
		JustBeforeEach(func() {
			cmpName = "nodejs"
		})

		It("create local nodejs component and push code", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.Cmd("odo", "component", "create", "nodejs", cmpName, "--project", project).ShouldPass()
			// verify that config was properly created
			info := helper.LocalEnvInfo(context)
			Expect(info.GetApplication(), "app")
			Expect(info.GetName(), cmpName)

			output := helper.Cmd("odo", "push").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})

		It("create, push and list local nodejs component", func() {
			appName := "testing"
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.Cmd("odo", "component", "create", "nodejs", cmpName, "--app", appName, "--project", project, "--context", context).ShouldPass()

			// verify that config was properly created
			info := helper.LocalEnvInfo(context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)
			helper.Cmd("odo", "push").ShouldPass()

			// list the component name
			cmpListOutput := helper.Cmd("odo", "list", "--app", appName, "--project", project).ShouldPass().Out()
			Expect(cmpListOutput).To(ContainSubstring(cmpName))
		})

	})
})
