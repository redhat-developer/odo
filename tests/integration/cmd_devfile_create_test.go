package integration

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/log"
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
		context = helper.CreateNewDevfileContext()
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

	Context("When executing odo create with devfile component and --downloadSource flag", func() {
		It("should succesfully create the compoment and download the source", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			contextDevfile := helper.CreateNewDevfileContext()
			helper.Chdir(contextDevfile)
			devfile := "devfile.yaml"
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(contextDevfile, devfile))
			// TODO: Check for devfile in non-default location
			helper.CmdShouldPass("odo", "create", "nodejs", "--downloadSource", "--devfile", devfile)
			expectedFiles := []string{"package.json", "package-lock.json", "README.MD", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
			helper.DeleteDir(contextDevfile)
		})
	})

	Context("When executing odo create with devfile component and --downloadSource flag with github type", func() {
		It("should succesfully create the compoment and download the source", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			contextDevfile := helper.CreateNewDevfileContext()
			helper.Chdir(contextDevfile)
			devfile := "devfile.yaml"
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(contextDevfile, devfile))

			err := helper.ReplaceDevfileField(devfile, "type", "github")
			if err != nil {
				log.Info("Could not replace the entry in the devfile: " + err.Error())
			}
			helper.CmdShouldPass("odo", "create", "nodejs", "--downloadSource", "--devfile", devfile)
			expectedFiles := []string{"package.json", "package-lock.json", "README.MD", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
			helper.DeleteDir(contextDevfile)
		})
	})

	Context("When executing odo create with devfile component and --downloadSource flag with zip type", func() {
		It("should create the compoment and download the source", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			contextDevfile := helper.CreateNewDevfileContext()
			helper.Chdir(contextDevfile)
			devfile := "devfile.yaml"
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(contextDevfile, devfile))
			err := helper.ReplaceDevfileField(devfile, "location", "https://github.com/che-samples/web-nodejs-sample/archive/master.zip")
			if err != nil {
				log.Info("Could not replace the entry in the devfile: " + err.Error())
			}
			err = helper.ReplaceDevfileField(devfile, "type", "zip")
			if err != nil {
				log.Info("Could not replace the entry in the devfile: " + err.Error())
			}
			helper.CmdShouldPass("odo", "create", "nodejs", "--downloadSource", "--devfile", devfile, "--context", filepath.Join(context, "config.yaml"))
			expectedFiles := []string{"package.json", "package-lock.json", "README.MD", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
			helper.DeleteDir(contextDevfile)
		})
	})
})
