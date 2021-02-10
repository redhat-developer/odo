package devfile

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile create command tests", func() {
	const devfile = "devfile.yaml"
	const envFile = ".odo/env/env.yaml"
	var contextDevfile, cmpName, devfilePath string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		cmpName = helper.RandString(6)
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When executing odo create with devfile component type argument", func() {
		It("should successfully create the devfile component with valid component name", func() {
			helper.CmdShouldPass("odo", "create", "java-openliberty", cmpName)
		})

		It("should fail to create the devfile component with invalid component type", func() {
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

	Context("When executing odo create with devfile component type argument and --project flag", func() {
		It("should successfully create the devfile component", func() {
			componentNamespace := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "java-openliberty", "--project", componentNamespace)
		})

		It("should fail to create the devfile component if --project value is 'default'", func() {
			output := helper.CmdShouldFail("odo", "create", "java", "--project", "default")
			expectedString := "odo may not work as expected in the default project, please run the odo component in a non-default project"
			helper.MatchAllInOutput(output, []string{expectedString})
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
			newContext := path.Join(commonVar.Context, "newContext")
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
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(commonVar.Context, devfile))
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
				newContext := path.Join(commonVar.Context, "newContext")
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
			helper.DeleteDir(contextDevfile)
			helper.Chdir(commonVar.Context)
		})

		It("should successfully create the component and download the source", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--starter")
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
		})

		It("should successfully create the component specified with valid project and download the source", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--starter=nodejs-starter")
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
		})
	})

	Context("When executing odo create with an invalid project specified in --starter", func() {
		It("should fail with please run 'The project: invalid-project-name specified in --starter does not exist'", func() {
			invalidProjectName := "invalid-project-name"
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			output := helper.CmdShouldFail("odo", "create", "nodejs", "--starter=invalid-project-name")
			expectedString := "the project: " + invalidProjectName + " specified in --starter does not exist"
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	Context("When executing odo create using --starter with a devfile component that contains no projects", func() {
		It("should fail with please run 'no starter project found in devfile.'", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-no-starterProject.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.CmdShouldFail("odo", "create", "nodejs", "--starter")
			expectedString := "no starter project found in devfile."
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	Context("When executing odo create with git tag or git branch specified in starter project", func() {
		JustBeforeEach(func() {
			contextDevfile = helper.CreateNewContext()
			helper.Chdir(contextDevfile)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-branch.yaml"), filepath.Join(contextDevfile, "devfile.yaml"))
		})

		JustAfterEach(func() {
			helper.DeleteDir(contextDevfile)
			helper.Chdir(commonVar.Context)
		})

		It("should successfully create the component and download the source from the specified branch", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--starter")
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
		})

		It("should successfully create the component and download the source from the specified tag", func() {
			helper.ReplaceString(filepath.Join(contextDevfile, "devfile.yaml"), "revision: test-branch", "revision: 0.0.1")
			helper.CmdShouldPass("odo", "create", "nodejs", "--starter")
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
		})
	})

	Context("When executing odo create with component with no devBuild command", func() {
		It("should successfully create the devfile component and remove a dangling env file", func() {
			// Quarkus devfile has no devBuild command
			output := helper.CmdShouldPass("odo", "create", "java-quarkus")
			helper.MatchAllInOutput(output, []string{"Please use `odo push` command to create the component with source deployed"})
			helper.DeleteFile("devfile.yaml")
			out, outerr := helper.CmdShouldPassIncludeErrStream("odo", "create", "nodejs")
			helper.MatchAllInOutput(out, []string{
				"Please use `odo push` command to create the component with source deployed"})
			helper.MatchAllInOutput(outerr, []string{
				"Found a dangling env file without a devfile, overwriting it",
			})
		})
	})

	It("checks that odo push works with a devfile with now flag", func() {
		context2 := helper.CreateNewContext()
		helper.Chdir(context2)
		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context2, "devfile.yaml"))
		output := helper.CmdShouldPass("odo", "create", "--starter", "nodejs", "--now")
		Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		helper.Chdir(commonVar.OriginalWorkingDirectory)
		helper.DeleteDir(context2)
	})

	Context("When executing odo create with --s2i flag", func() {
		var newContext string
		JustBeforeEach(func() {
			newContext = path.Join(commonVar.Context, "newContext")
			devfilePath = filepath.Join(newContext, devfile)
			helper.MakeDir(newContext)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), devfilePath)
		})
		JustAfterEach(func() {
			helper.DeleteDir(newContext)
		})

		It("should fail to create the devfile component which doesn't have an s2i component of same name", func() {
			componentName := helper.RandString(6)

			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-o", "json")

			wantOutput := []string{"java-openliberty"}

			var data map[string]interface{}

			err := json.Unmarshal([]byte(output), &data)

			if err != nil {
				Expect(err).Should(BeNil())
			}
			outputBytes, err := json.Marshal(data["s2iItems"])
			if err == nil {
				output = string(outputBytes)
			}

			helper.DontMatchAllInOutput(output, wantOutput)

			outputBytes, err = json.Marshal(data["devfileItems"])
			if err == nil {
				output = string(outputBytes)
			}

			helper.MatchAllInOutput(output, wantOutput)

			helper.CmdShouldFail("odo", "create", "java-openliberty", componentName, "--s2i")
		})

		It("should fail to create the devfile component with valid file system path", func() {
			output := helper.CmdShouldFail("odo", "create", "nodejs", "--s2i", "--devfile", devfilePath)
			helper.MatchAllInOutput(output, []string{"you can't set --s2i flag as true if you want to use the devfile via --devfile flag"})
		})

		It("should fail to create the component specified with valid project and download the source", func() {
			output := helper.CmdShouldFail("odo", "create", "nodejs", "--starter=nodejs-starter", "--s2i")
			helper.MatchAllInOutput(output, []string{"you can't set --s2i flag as true if you want to use the starter via --starter flag"})
		})

		It("should fail to create the devfile component with --registry specified", func() {
			output := helper.CmdShouldFail("odo", "create", "nodejs", "--registry", "DefaultDevfileRegistry", "--s2i")
			helper.MatchAllInOutput(output, []string{"you can't set --s2i flag as true if you want to use the registry via --registry flag"})
		})

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

	Context("When executing odo create with devfile component, --starter flag and subDir has a valid value", func() {
		It("should only extract the specified path in the subDir field", func() {
			originalDir := commonVar.Context
			defer helper.Chdir(originalDir)
			contextDevfile := helper.CreateNewContext()

			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-with-subDir.yaml"), filepath.Join(contextDevfile, "devfile.yaml"))
			helper.Chdir(contextDevfile)
			helper.CmdShouldPass("odo", "create", cmpName, "--project", commonVar.Project, "--starter")

			pathsToValidate := map[string]bool{
				filepath.Join(contextDevfile, "java", "com"):                                            true,
				filepath.Join(contextDevfile, "java", "com", "example"):                                 true,
				filepath.Join(contextDevfile, "java", "com", "example", "demo"):                         true,
				filepath.Join(contextDevfile, "java", "com", "example", "demo", "DemoApplication.java"): true,
				filepath.Join(contextDevfile, "resources", "application.properties"):                    true,
			}

			pathsNotToBePresent := map[string]bool{
				filepath.Join(contextDevfile, "src"):  true,
				filepath.Join(contextDevfile, "main"): true,
			}

			found := 0
			notToBeFound := 0
			err := filepath.Walk(contextDevfile, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if ok := pathsToValidate[path]; ok {
					found++
				}

				if ok := pathsNotToBePresent[path]; ok {
					notToBeFound++
				}
				return nil
			})

			Expect(err).To(BeNil())

			Expect(found).To(Equal(len(pathsToValidate)))
			Expect(notToBeFound).To(Equal(0))

			helper.DeleteDir(contextDevfile)
		})
	})

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
