package devfile

import (
	"os"
	"path"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile create command tests", func() {
	const devfile = "devfile.yaml"
	const envFile = ".odo/env/env.yaml"
	var namespace string
	var context string
	var currentWorkingDirectory string
	var devfilePath string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		namespace = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(namespace)
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
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "false", "-f")
			helper.CopyExample(filepath.Join("source", "nodejs"), context)

			// Check that it will contain the experimental mode output
			experimentalOutputMsg := "Experimental mode is enabled, use at your own risk"
			Expect(helper.CmdShouldPass("odo", "create", "nodejs")).To(Not(ContainSubstring(experimentalOutputMsg)))
		})
	})

	Context("When executing odo create with devfile component type argument", func() {
		It("should successfully create the devfile component with valid component type", func() {
			helper.CmdShouldPass("odo", "create", "openLiberty")
		})

		It("should fail to create the devfile componet with invalid component type", func() {
			fakeComponentName := "fake-component"
			output := helper.CmdShouldFail("odo", "create", fakeComponentName)
			expectedString := "\"" + fakeComponentName + "\" not found"
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	Context("When executing odo create with devfile component type and component name arguments", func() {
		It("should successfully create the devfile component with valid component name", func() {
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "openLiberty", componentName)
		})

		It("should fail to create the devfile component with component name that contains invalid character", func() {
			componentName := "BAD@123"
			output := helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
			helper.MatchAllInOutput(output, []string{"Contain only lowercase alphanumeric characters or ‘-’"})
		})

		It("should fail to create the devfile component with component name that contains all numeric values", func() {
			componentName := "123456"
			output := helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
			helper.MatchAllInOutput(output, []string{"Must not contain all numeric values"})
		})

		It("should fail to create the devfile component with componet name contains more than 63 characters", func() {
			componentName := helper.RandString(64)
			output := helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
			helper.MatchAllInOutput(output, []string{"Contain at most 63 characters"})
		})
	})

	Context("When executing odo create with devfile component type argument and --project flag", func() {
		It("should successfully create the devfile component", func() {
			componentNamespace := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "openLiberty", "--project", componentNamespace)
		})
	})

	Context("When executing odo create with devfile component type argument and --registry flag", func() {
		It("should successfully create the devfile component", func() {
			componentRegistry := "DefaultDevfileRegistry"
			helper.CmdShouldPass("odo", "create", "openLiberty", "--registry", componentRegistry)
		})
	})

	Context("When executing odo create with devfile component type argument and --context flag", func() {
		It("should successfully create the devfile component in the context", func() {
			newContext := path.Join(context, "newContext")
			devfilePath = filepath.Join(newContext, devfile)
			envFilePath := filepath.Join(newContext, envFile)
			helper.MakeDir(newContext)

			helper.CmdShouldPass("odo", "create", "openLiberty", "--context", newContext)
			output := util.CheckPathExists(devfilePath)
			Expect(output).Should(BeTrue())
			output = util.CheckPathExists(envFilePath)
			Expect(output).Should(BeTrue())
			helper.DeleteDir(newContext)
		})
	})

	Context("When executing odo create with existing devfile", func() {
		Context("When devfile exists in user's working directory", func() {
			JustBeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(context, devfile))
			})

			It("should successfully create the devfile componet", func() {
				helper.CmdShouldPass("odo", "create", "nodejs")
			})

			It("should successfully create the devfile component with --devfile points to the same devfile", func() {
				helper.CmdShouldPass("odo", "create", "nodejs", "--devfile", "./devfile.yaml")
			})

			It("should fail to create the devfile component with more than 1 arguments are passed in", func() {
				helper.CmdShouldFail("odo", "create", "nodejs", "nodejs")
			})

			It("should fail to create the devfile component with --devfile points to different devfile", func() {
				helper.CmdShouldFail("odo", "create", "nodejs", "--devfile", "/path/to/file")
			})
		})

		Context("When devfile exists not in user's working directory and user specify the devfile path via --devfile", func() {
			JustBeforeEach(func() {
				newContext := path.Join(context, "newContext")
				devfilePath = filepath.Join(newContext, devfile)
				helper.MakeDir(newContext)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), devfilePath)
			})

			It("should successfully create the devfile component with valid file system path", func() {
				helper.CmdShouldPass("odo", "create", "nodejs", "--devfile", devfilePath)
			})

			It("should successfully create the devfile component with valid specifies URL path", func() {
				helper.CmdShouldPass("odo", "create", "nodejs", "--devfile", "https://raw.githubusercontent.com/elsony/devfile-registry/master/devfiles/nodejs/devfile.yaml")
			})

			It("should fail to create the devfile component with invalid file system path", func() {
				helper.CmdShouldFail("odo", "create", "nodejs", "--devfile", "#@!")
			})

			It("should fail to create the devfile component with invalid URL path", func() {
				helper.CmdShouldFail("odo", "create", "nodejs", "--devfile", "://www.example.com/")
			})

			It("should fail to create the devfile component with more than 1 arguments are passed in", func() {
				helper.CmdShouldFail("odo", "create", "nodejs", "nodejs", "--devfile", devfilePath)
			})

			It("should fail to create the devfile component with --registry specified", func() {
				helper.CmdShouldFail("odo", "create", "nodejs", "--devfile", devfilePath, "--registry", "DefaultDevfileRegistry")
			})
		})
	})

	Context("When executing odo create with devfile component and --downloadSource flag", func() {
		It("should successfully create the component and download the source", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			contextDevfile := helper.CreateNewContext()
			helper.Chdir(contextDevfile)
			helper.CmdShouldPass("odo", "create", "nodejs", "--downloadSource")
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
			helper.DeleteDir(contextDevfile)
			helper.Chdir(context)
		})
	})

	Context("When executing odo create with component with no devBuild command", func() {
		It("should successfully create the devfile component", func() {
			// Quarkus devfile has no devBuild command
			output := helper.CmdShouldPass("odo", "create", "quarkus")
			helper.MatchAllInOutput(output, []string{"Please use `odo push` command to create the component with source deployed"})
		})
	})

	Context("When executing odo create with devfile component and --downloadSource flag with a valid project", func() {
		It("should successfully create the component specified and download the source", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			contextDevfile := helper.CreateNewContext()
			helper.Chdir(contextDevfile)
			helper.CmdShouldPass("odo", "create", "nodejs", "--downloadSource=nodejs-web-app")
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
			helper.DeleteDir(contextDevfile)
			helper.Chdir(context)
		})
	})

	Context("When executing odo create with an invalid project specified in --downloadSource", func() {
		It("should fail with please run 'The project: invalid-project-name specified in --downloadSource does not exist'", func() {
			invalidProjectName := "invalid-project-name"
			output := helper.CmdShouldFail("odo", "create", "nodejs", "--downloadSource=invalid-project-name")
			expectedString := "The project: " + invalidProjectName + " specified in --downloadSource does not exist"
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	Context("When executing odo create using --downloadSource with a devfile component that contains no projects", func() {
		It("should fail with please run 'No project found in devfile component.'", func() {
			output := helper.CmdShouldFail("odo", "create", "maven", "--downloadSource")
			expectedString := "No project found in devfile component."
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	// Currently these tests need interactive mode in order to set the name of the component.
	// Once this feature is added we can change these tests.
	//Context("When executing odo create with devfile component and --downloadSource flag with github type", func() {
	//	It("should succesfully create the compoment and download the source", func() {
	//		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	//		contextDevfile := helper.CreateNewContext()
	//		helper.Chdir(contextDevfile)
	//		devfile := "devfile.yaml"
	//		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(contextDevfile, devfile))

	//		err := helper.ReplaceDevfileField(devfile, "type", "github")
	//		if err != nil {
	//			log.Error("Could not replace the entry in the devfile: " + err.Error())
	//		}
	//		helper.CmdShouldPass("odo", "create", "--downloadSource")
	//		expectedFiles := []string{"package.json", "package-lock.json", "README.MD", devfile}
	//		Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
	//		helper.DeleteDir(contextDevfile)
	//	})
	//})

	//Context("When executing odo create with devfile component and --downloadSource flag with zip type", func() {
	//	It("should create the compoment and download the source", func() {
	//		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	//		contextDevfile := helper.CreateNewContext()
	//		helper.Chdir(contextDevfile)
	//		devfile := "devfile.yaml"
	//		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(contextDevfile, devfile))
	//		err := helper.ReplaceDevfileField(devfile, "location", "https://github.com/che-samples/web-nodejs-sample/archive/master.zip")
	//		if err != nil {
	//			log.Error("Could not replace the entry in the devfile: " + err.Error())
	//		}
	//		err = helper.ReplaceDevfileField(devfile, "type", "zip")
	//		if err != nil {
	//			log.Error("Could not replace the entry in the devfile: " + err.Error())
	//		}
	//		helper.CmdShouldPass("odo", "create", "--downloadSource")
	//		expectedFiles := []string{"package.json", "package-lock.json", "README.MD", devfile}
	//		Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
	//		helper.DeleteDir(contextDevfile)
	//	})
	//})

	// Context("When executing odo create with devfile component, --downloadSource flag and sparseContextDir has a valid value", func() {
	// 	It("should only extract the specified path in the sparseContextDir field", func() {
	// 		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	// 		contextDevfile := helper.CreateNewContext()
	// 		helper.Chdir(contextDevfile)
	// 		devfile := "devfile.yaml"
	// 		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-sparseCheckoutDir"), filepath.Join(contextDevfile, devfile))
	// 		componentNamespace := helper.RandString(6)
	// 		helper.CmdShouldPass("odo", "create", "--downloadSource", "--project", componentNamespace)
	// 		expectedFiles := []string{"app.js", devfile}
	// 		Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
	// 		helper.DeleteDir(contextDevfile)
	// 	})
	// })

	// Context("When executing odo create with devfile component, --downloadSource flag and sparseContextDir has an invalid value", func() {
	// 	It("should fail and alert the user that the specified path in sparseContextDir does not exist", func() {
	// 		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	// 		contextDevfile := helper.CreateNewContext()
	// 		helper.Chdir(contextDevfile)
	// 		devfile := "devfile.yaml"
	// 		devfilePath := filepath.Join(contextDevfile, devfile)
	// 		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-sparseCheckoutDir"), devfilePath)
	// 		helper.ReplaceDevfileField(devfilePath, "sparseCheckoutDir", "/invalid/")
	// 		componentNamespace := helper.RandString(6)
	// 		output := helper.CmdShouldFail("odo", "create", "--downloadSource", "--project", componentNamespace)
	// 		expectedString := "no files were unzipped, ensure that the project repo is not empty or that sparseCheckoutDir has a valid path"
	// 		helper.MatchAllInOutput(output, []string{expectedString})
	// 		helper.DeleteDir(contextDevfile)
	// 	})
	// })
})
