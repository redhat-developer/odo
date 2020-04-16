package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile create command tests", func() {
	var project string
	var context string
	var currentWorkingDirectory string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Enabling experimental preference should show a disclaimer", func() {
		It("checks that the experimental warning appears for create", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			helper.CopyExample(filepath.Join("source", "nodejs"), context)

			// Check that it will contain the experimental mode output
			experimentalOutputMsg := "Experimental mode is enabled, use at your own risk"
			Expect(helper.CmdShouldPass("odo", "create", "nodejs")).To(ContainSubstring(experimentalOutputMsg))

		})

		It("checks that the experimental warning does *not* appear when Experimental is set to false for create", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "false")
			helper.CopyExample(filepath.Join("source", "nodejs"), context)

			// Check that it will contain the experimental mode output
			experimentalOutputMsg := "Experimental mode is enabled, use at your own risk"
			Expect(helper.CmdShouldPass("odo", "create", "nodejs")).To(Not(ContainSubstring(experimentalOutputMsg)))
		})
	})

	Context("When executing odo create with devfile component type argument", func() {
		It("should successfully create the devfile component", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			helper.CmdShouldPass("odo", "create", "openLiberty")
		})
	})

	Context("When executing odo create with devfile component type and component name arguments", func() {
		It("should successfully create the devfile component", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "openLiberty", componentName)
		})
	})

	Context("When executing odo create with devfile component type argument and --project flag", func() {
		It("should successfully create the devfile component", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			componentNamespace := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "openLiberty", "--project", componentNamespace)
		})
	})

	Context("When executing odo create with devfile component name that contains unsupported character", func() {
		It("should failed with component name is not valid and prompt supported character", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			componentName := "BAD@123"
			output := helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
			helper.MatchAllInOutput(output, []string{"Contain only lowercase alphanumeric characters or ‘-’"})
		})
	})

	Context("When executing odo create with devfile component name that contains all numeric values", func() {
		It("should failed with component name is not valid and prompt container name must not contain all numeric values", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			componentName := "123456"
			output := helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
			helper.MatchAllInOutput(output, []string{"Must not contain all numeric values"})
		})
	})

	Context("When executing odo create with devfile component name that contains more than 63 characters ", func() {
		It("should failed with component name is not valid and prompt container name contains at most 63 characters", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			componentName := helper.RandString(64)
			output := helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
			helper.MatchAllInOutput(output, []string{"Contain at most 63 characters"})
		})
	})

	Context("When executing odo create with an invalid devfile component", func() {
		It("should fail with please run 'odo catalog list components'", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			fakeComponentName := "fake-component"
			output := helper.CmdShouldFail("odo", "create", fakeComponentName)
			expectedString := "\"" + fakeComponentName + "\" not found"
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})
})
