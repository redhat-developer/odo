package devfile

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"

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

	Context("when .gitignore file exists", func() {
		It("checks that .odo/env exists in gitignore", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			ignoreFilePath := filepath.Join(commonVar.Context, ".gitignore")

			helper.FileShouldContainSubstring(ignoreFilePath, filepath.Join(".odo", "env"))

		})
	})

	Context("When executing odo create with devfile component type argument", func() {
		It("should successfully create the devfile component with valid component name", func() {
			helper.Cmd("odo", "create", "java-openliberty", cmpName).ShouldPass()
		})

		It("should fail to create the devfile component with invalid component type", func() {
			fakeComponentName := "fake-component"
			output := helper.Cmd("odo", "create", fakeComponentName).ShouldFail().Err()
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
			helper.Cmd("odo", "create", "java-openliberty", "--project", componentNamespace).ShouldPass()
			fileContents, err := helper.ReadFile(filepath.Join(commonVar.Context, ".odo/env/env.yaml"))
			Expect(err).To(BeNil())
			Expect(fileContents).To(ContainSubstring(componentNamespace))
		})

	})

	Context("When executing odo create with devfile component type argument and --registry flag", func() {
		It("should successfully create the devfile component if specified registry is valid", func() {
			componentRegistry := "DefaultDevfileRegistry"
			helper.Cmd("odo", "create", "java-openliberty", "--registry", componentRegistry).ShouldPass()
		})

		It("should fail to create the devfile component if specified registry is invalid", func() {
			componentRegistry := "fake"
			output := helper.Cmd("odo", "create", "java-openliberty", "--registry", componentRegistry).ShouldFail().Err()
			helper.MatchAllInOutput(output, []string{"registry fake doesn't exist, please specify a valid registry via --registry"})
		})
	})

	Context("When executing odo create with devfile component type argument and --context flag", func() {
		var newContext, envFilePath string
		JustBeforeEach(func() {
			newContext = filepath.Join(commonVar.Context, "newContext")
			devfilePath = filepath.Join(newContext, devfile)
			helper.MakeDir(newContext)
		})

		JustAfterEach(func() {
			helper.DeleteDir(newContext)
		})

		It("should successfully create the devfile component in the context", func() {
			envFilePath = filepath.Join(newContext, envFile)
			helper.Cmd("odo", "create", "java-openliberty", "--context", newContext).ShouldPass()
			output := util.CheckPathExists(devfilePath)
			Expect(output).Should(BeTrue())
			output = util.CheckPathExists(envFilePath)
			Expect(output).Should(BeTrue())
		})

		It("should successfully create the devfile component and download the source when used with --starter flag", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(devfilePath))
			helper.Cmd("odo", "create", "nodejs", "--starter", "--context", newContext).ShouldPass()
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(newContext, expectedFiles)).To(Equal(true))
		})

		It("should successfully create the devfile component with auto generated name", func() {
			helper.Cmd("odo", "create", "nodejs", "--context", newContext).ShouldPass()
			output := helper.Cmd("odo", "env", "view", "--context", newContext, "-o", "json").ShouldPass().Out()
			value := gjson.Get(output, "spec.name")
			Expect(strings.TrimSpace(value.String())).To(ContainSubstring(strings.TrimSpace("nodejs-" + filepath.Base(strings.ToLower(newContext)))))
		})

		It("should successfully create the devfile component and show json output for working cluster", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(devfilePath))
			output := helper.Cmd("odo", "create", "nodejs", "--context", newContext, "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "metadata.name", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Not Pushed"})).To(Equal(true))
		})

		It("should successfully create and push the devfile component and show json output for working cluster", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(devfilePath))
			output := helper.Cmd("odo", "create", "nodejs", "--starter", "--context", newContext, "-o", "json", "--now").ShouldPass().Out()
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(newContext, expectedFiles)).To(Equal(true))
			helper.MatchAllInOutput(output, []string{"Pushed", "nodejs", "Component"})
		})

		It("should successfully create the devfile component and show json output for non connected cluster", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(devfilePath))
			cmd := helper.Cmd("odo", "create", "nodejs", "--context", newContext, "-o", "json")
			output := cmd.WithEnv("KUBECONFIG=/no/such/path", "GLOBALODOCONFIG="+os.Getenv("GLOBALODOCONFIG")).ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "metadata.name", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Unknown"})).To(Equal(true))
		})

		It("should successfully create the devfile component and show json output for a unreachable cluster", func() {

			path := os.Getenv("KUBECONFIG")

			// read the contents from the kubeconfig and replace the server entries
			reg := regexp.MustCompile(`server: .*`)
			kubeConfigContents, err := helper.ReadFile(path)
			Expect(err).To(BeNil())
			kubeConfigContents = reg.ReplaceAllString(kubeConfigContents, "server: https://not-reachable.com:443")

			// write to a new file which will be used as the new kubeconfig
			newKubeConfigPath := filepath.Join(commonVar.Context, "newKUBECONFIG")
			newKubeConfig, err := os.Create(newKubeConfigPath)
			Expect(err).To(BeNil())
			_, err = newKubeConfig.WriteString(kubeConfigContents)
			Expect(err).To(BeNil())

			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(devfilePath))
			cmd := helper.Cmd("odo", "create", "nodejs", "--context", newContext, "-o", "json")
			output := cmd.WithEnv("KUBECONFIG="+newKubeConfigPath, "GLOBALODOCONFIG="+os.Getenv("GLOBALODOCONFIG")).ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "metadata.name", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Unknown"})).To(Equal(true))

			err = os.Remove(newKubeConfigPath)
			Expect(err).To(BeNil())
		})
	})

	Context("When executing odo create with existing devfile", func() {
		Context("When devfile exists in user's working directory", func() {
			JustBeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(commonVar.Context, devfile))
			})

			It("should successfully create the devfile component", func() {
				helper.Cmd("odo", "create", "nodejs").ShouldPass()
			})

			It("should successfully create the devfile component with --devfile points to the same devfile", func() {
				helper.Cmd("odo", "create", "nodejs", "--devfile", "./devfile.yaml").ShouldPass()
				fileIsEmpty, err := helper.FileIsEmpty("./devfile.yaml")
				Expect(err).Should(BeNil())
				Expect(fileIsEmpty).Should(BeFalse())
			})

			It("should fail to create the devfile component with more than 1 arguments are passed in", func() {
				helper.Cmd("odo", "create", "nodejs", "nodejs").ShouldFail()
			})

			It("should fail to create the devfile component with --devfile points to different devfile", func() {
				helper.Cmd("odo", "create", "nodejs", "--devfile", "/path/to/file").ShouldFail()
			})

			It("should fail when we create the devfile component multiple times", func() {
				helper.Cmd("odo", "create", "nodejs").ShouldPass()
				output := helper.Cmd("odo", "create", "nodejs").ShouldFail().Err()
				Expect(output).To(ContainSubstring("this directory already contains a component"))
			})
		})

		Context("Testing Create for OpenShift specific scenarios", func() {
			JustBeforeEach(func() {
				if os.Getenv("KUBERNETES") == "true" {
					Skip("This is a OpenShift specific scenario, skipping")
				}
			})

			It("should fail when we create the devfile or s2i component multiple times", func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), filepath.Join(commonVar.Context, devfile))
				helper.Cmd("odo", "create", "nodejs").ShouldPass()
				output := helper.Cmd("odo", "create", "nodejs", "--s2i").ShouldFail().Err()
				Expect(output).To(ContainSubstring("this directory already contains a component"))
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
				helper.Cmd("odo", "create", "nodejs", "--devfile", devfilePath).ShouldPass()
			})

			It("should successfully create the devfile component with valid specifies URL path", func() {
				helper.Cmd("odo", "create", "nodejs", "--devfile", "https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/devfile.yaml").ShouldPass()
			})

			It("should fail to create the devfile component with invalid file system path", func() {
				helper.Cmd("odo", "create", "nodejs", "--devfile", "#@!").ShouldFail()
			})

			It("should fail to create the devfile component with invalid URL path", func() {
				helper.Cmd("odo", "create", "nodejs", "--devfile", "://www.example.com/").ShouldFail()
			})

			It("should fail to create the devfile component with more than 1 arguments are passed in", func() {
				helper.Cmd("odo", "create", "nodejs", "nodejs", "--devfile", devfilePath).ShouldFail()
			})

			It("should fail to create the devfile component with --registry specified", func() {
				helper.Cmd("odo", "create", "nodejs", "--devfile", devfilePath, "--registry", "DefaultDevfileRegistry").ShouldFail()
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
			helper.Cmd("odo", "create", "nodejs", "--starter").ShouldPass()
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
		})

		It("should successfully create the component specified with valid project and download the source", func() {
			helper.Cmd("odo", "create", "nodejs", "--starter=nodejs-starter").ShouldPass()
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
		})
	})

	Context("When executing odo create with an invalid project specified in --starter", func() {
		It("should fail with please run 'The project: invalid-project-name specified in --starter does not exist'", func() {
			invalidProjectName := "invalid-project-name"
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			output := helper.Cmd("odo", "create", "nodejs", "--starter=invalid-project-name").ShouldFail().Err()
			expectedString := "the project: " + invalidProjectName + " specified in --starter does not exist"
			helper.MatchAllInOutput(output, []string{expectedString})
		})
	})

	Context("When executing odo create using --starter with a devfile component that contains no projects", func() {
		It("should fail with please run 'no starter project found in devfile.'", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-no-starterProject.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "create", "nodejs", "--starter").ShouldFail().Err()
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
			helper.Cmd("odo", "create", "nodejs", "--starter").ShouldPass()
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
		})

		It("should successfully create the component and download the source from the specified tag", func() {
			helper.ReplaceString(filepath.Join(contextDevfile, "devfile.yaml"), "revision: test-branch", "revision: 0.0.1")
			helper.Cmd("odo", "create", "nodejs", "--starter").ShouldPass()
			expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
			Expect(helper.VerifyFilesExist(contextDevfile, expectedFiles)).To(Equal(true))
		})
	})

	Context("When executing odo create with component with no devBuild command", func() {
		It("should successfully create the devfile component and remove a dangling env file", func() {
			// Quarkus devfile has no devBuild command
			output := helper.Cmd("odo", "create", "java-quarkus").ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{"Please use `odo push` command to create the component with source deployed"})
			helper.DeleteFile("devfile.yaml")
			out, outerr := helper.Cmd("odo", "create", "nodejs").ShouldPass().OutAndErr()
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
		output := helper.Cmd("odo", "create", "--starter", "nodejs", "--now").ShouldPass().Out()
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

			output := helper.Cmd("odo", "catalog", "list", "components", "-o", "json").ShouldPass().Out()

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

			helper.Cmd("odo", "create", "java-openliberty", componentName, "--s2i").ShouldFail().Err()
		})

		It("should fail to create the devfile component with valid file system path", func() {
			output := helper.Cmd("odo", "create", "nodejs", "--s2i", "--devfile", devfilePath).ShouldFail().Err()
			helper.MatchAllInOutput(output, []string{"you can't set --s2i flag as true if you want to use the devfile via --devfile flag"})
		})

		It("should fail to create the component specified with valid project and download the source", func() {
			output := helper.Cmd("odo", "create", "nodejs", "--starter=nodejs-starter", "--s2i").ShouldFail().Err()
			helper.MatchAllInOutput(output, []string{"you can't set --s2i flag as true if you want to use the starter via --starter flag"})
		})

		It("should fail to create the devfile component with --registry specified", func() {
			output := helper.Cmd("odo", "create", "nodejs", "--registry", "DefaultDevfileRegistry", "--s2i").ShouldFail().Err()
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
			helper.Cmd("odo", "create", cmpName, "--project", commonVar.Project, "--starter").ShouldPass()

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
