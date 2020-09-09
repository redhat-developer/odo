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
	var namespace, context, contextDevfile, cmpName, currentWorkingDirectory, devfilePath, originalKubeconfig string

	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("When executing odo create with devfile component type argument", func() {
		It("should successfully create the devfile component", func() {
			helper.CmdShouldPass("odo", "create", "java-openliberty")
		})

		It("should fail to create the devfile componet with invalid component type", func() {
			fakeComponentName := "fake-component"
			output := helper.CmdShouldFail("odo", "create", fakeComponentName)
			var expectedString string
			if os.Getenv("KUBERNETES") == "true" {
				expectedString = "component type not found"
			} else {
				expectedString = "component type \"" + fakeComponentName + "\" not found"
			}
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	Context("When executing odo create with devfile component type and component name arguments", func() {
		It("should successfully create the devfile component with valid component name", func() {
			helper.CmdShouldPass("odo", "create", "java-openliberty", cmpName)
		})

		It("should fail to create the devfile component with component name that contains invalid character", func() {
			componentName := "BAD@123"
			output := helper.CmdShouldFail("odo", "create", "java-openliberty", componentName)
			helper.MatchAllInOutput(output, []string{"Contain only lowercase alphanumeric characters or ‘-’"})
		})

		It("should fail to create the devfile component with component name that contains all numeric values", func() {
			componentName := "123456"
			output := helper.CmdShouldFail("odo", "create", "java-openliberty", componentName)
			helper.MatchAllInOutput(output, []string{"Must not contain all numeric values"})
		})

		It("should fail to create the devfile component with componet name contains more than 63 characters", func() {
			componentName := helper.RandString(64)
			output := helper.CmdShouldFail("odo", "create", "java-openliberty", componentName)
			helper.MatchAllInOutput(output, []string{"Contain at most 63 characters"})
		})
	})

	Context("When executing odo create with component type argument and --s2i flag", func() {

		JustBeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("Skipping test because s2i image is not supported on Kubernetes cluster")
			}
		})

		componentType := "nodejs"

		It("should successfully create the localconfig component", func() {
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", componentType), context)
			helper.CmdShouldPass("odo", "create", componentType, componentName, "--s2i")
			helper.ValidateLocalCmpExist(context, "Type,nodejs", "Name,"+componentName, "Application,app")
			helper.CmdShouldPass("odo", "push", "--context", context, "-v4")

			// clean up
			helper.CmdShouldPass("odo", "app", "delete", "app", "-f")
			helper.CmdShouldFail("odo", "app", "delete", "app", "-f")
			helper.CmdShouldFail("odo", "delete", componentName, "-f")

		})

		It("should successfully create the localconfig component with --git flag", func() {
			componentName := "cmp-git"
			helper.CmdShouldPass("odo", "create", componentType, "--git", "https://github.com/openshift/nodejs-ex", "--context", context, "--s2i", "true", "--app", "testing")
			helper.CmdShouldPass("odo", "push", "--context", context, "-v4")

			// clean up
			helper.CmdShouldPass("odo", "app", "delete", "testing", "-f")
			helper.CmdShouldFail("odo", "app", "delete", "testing", "-f")
			helper.CmdShouldFail("odo", "delete", componentName, "-f")
		})

		It("should fail to create the devfile component which doesn't have an s2i component of same name", func() {
			helper.CmdShouldFail("odo", "create", "java-openliberty", cmpName, "--s2i")
		})

		It("should fail the create command as --git flag, which is specific to s2i component creation, is used without --s2i flag", func() {
			output := helper.CmdShouldFail("odo", "create", "nodejs", "cmp-git", "--git", "https://github.com/openshift/nodejs-ex", "--context", context, "--app", "testing")
			Expect(output).Should(ContainSubstring("flag --git, requires --s2i flag to be set, when deploying S2I (Source-to-Image) components."))
		})

		It("should fail the create command as --binary flag, which is specific to s2i component creation, is used without --s2i flag", func() {
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), context)

			output := helper.CmdShouldFail("odo", "create", "java:8", "sb-jar-test", "--binary", filepath.Join(context, "sb.jar"), "--context", context)
			Expect(output).Should(ContainSubstring("flag --binary, requires --s2i flag to be set, when deploying S2I (Source-to-Image) components."))
		})
	})

	Context("When executing odo create with devfile component type argument and --project flag", func() {
		It("should successfully create the devfile component", func() {
			componentNamespace := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "java-openliberty", "--project", componentNamespace)
		})
	})

	Context("When executing odo create with devfile component type argument and --registry flag", func() {
		It("should successfully create the devfile component if specified registry is valid", func() {
			componentRegistry := "DefaultDevfileRegistry"
			helper.CmdShouldPass("odo", "create", "java-openliberty", "--registry", componentRegistry)
		})

		It("should fail to create the devfile component if specified registry is invalid", func() {
			componentRegistry := "fake"
			output := helper.CmdShouldFail("odo", "create", "java-openliberty", "--registry", componentRegistry)
			helper.MatchAllInOutput(output, []string{"registry fake doesn't exist, please specify a valid registry via --registry"})
		})
	})

	Context("When executing odo create with devfile component type argument and --context flag", func() {
		It("should successfully create the devfile component in the context", func() {
			newContext := path.Join(context, "newContext")
			devfilePath = filepath.Join(newContext, devfile)
			envFilePath := filepath.Join(newContext, envFile)
			helper.MakeDir(newContext)

			helper.CmdShouldPass("odo", "create", "java-openliberty", "--context", newContext)
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
				fileIsEmpty, err := helper.FileIsEmpty("./devfile.yaml")
				Expect(err).Should(BeNil())
				Expect(fileIsEmpty).Should(BeFalse())
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
				helper.CmdShouldPass("odo", "create", "nodejs", "--devfile", "https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/devfile.yaml")
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

	Context("When executing odo create with devfile component and --starter flag", func() {
		JustBeforeEach(func() {
			contextDevfile = helper.CreateNewContext()
			helper.Chdir(contextDevfile)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(contextDevfile, "devfile.yaml"))
		})

		JustAfterEach(func() {
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
			helper.DeleteDir(contextDevfile)
			helper.Chdir(context)
		})

		It("should successfully create the component and download the source", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--starter")
		})

		It("should successfully create the component specified with valid project and download the source", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--starter=nodejs-starter")
		})
	})

	Context("When executing odo create with an invalid project specified in --starter", func() {
		It("should fail with please run 'The project: invalid-project-name specified in --starter does not exist'", func() {
			invalidProjectName := "invalid-project-name"
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			output := helper.CmdShouldFail("odo", "create", "nodejs", "--starter=invalid-project-name")
			expectedString := "the project: " + invalidProjectName + " specified in --starter does not exist"
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	Context("When executing odo create using --starter with a devfile component that contains no projects", func() {
		It("should fail with please run 'no starter project found in devfile.'", func() {
			output := helper.CmdShouldFail("odo", "create", "java-maven", "--starter")
			expectedString := "no starter project found in devfile."
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	Context("When executing odo create with component with no devBuild command", func() {
		It("should successfully create the devfile component", func() {
			// Quarkus devfile has no devBuild command
			output := helper.CmdShouldPass("odo", "create", "java-quarkus")
			helper.MatchAllInOutput(output, []string{"Please use `odo push` command to create the component with source deployed"})
		})
	})

	It("checks that odo push works with a devfile with now flag", func() {
		originalDir := helper.Getwd()
		context2 := helper.CreateNewContext()
		helper.Chdir(context2)
		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context2, "devfile.yaml"))
		output := helper.CmdShouldPass("odo", "create", "--starter", "nodejs", "--now")
		Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		helper.Chdir(originalDir)
		helper.DeleteDir(context2)
	})

	// Currently these tests need interactive mode in order to set the name of the component.
	// Once this feature is added we can change these tests.
	//Context("When executing odo create with devfile component and --downloadSource flag with github type", func() {
	//	It("should successfully create the component and download the source", func() {
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
	//	It("should create the component and download the source", func() {
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
