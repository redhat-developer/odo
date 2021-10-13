package devfile

import (
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

	var _ = BeforeEach(func() {
		cmpName = helper.RandString(6)
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	// checkNodeJSDirContent checks if the required nodejs files are present in the context directory after odo create
	var checkNodeJSDirContent = func(contextDir string) {
		expectedFiles := []string{"package.json", "package-lock.json", "README.md", devfile}
		Expect(helper.VerifyFilesExist(contextDir, expectedFiles)).To(Equal(true))
	}

	It("should check that .odo/env exists in gitignore", func() {
		helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
		ignoreFilePath := filepath.Join(commonVar.Context, ".gitignore")
		helper.FileShouldContainSubstring(ignoreFilePath, filepath.Join(".odo", "env"))
	})

	It("should successfully create the devfile component with valid component name", func() {
		helper.Cmd("odo", "create", "java-openliberty", cmpName).ShouldPass()

		By("checking that component name and language is set correctly in the devfile", func() {
			metadata := helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
			Expect(metadata.Name).To(BeEquivalentTo(cmpName))
			Expect(metadata.Language).To(ContainSubstring("java"))
		})
	})

	It("should fail to create the devfile component with invalid component type", func() {
		fakeComponentName := "fake-component"
		output := helper.Cmd("odo", "create", fakeComponentName).ShouldFail().Err()
		expectedString := "component type \"" + fakeComponentName + "\" is not supported"
		Expect(output).To(ContainSubstring(expectedString))
	})

	It("should successfully create the devfile component with --project flag", func() {
		componentNamespace := helper.RandString(6)
		helper.Cmd("odo", "create", "java-openliberty", "--project", componentNamespace).ShouldPass()
		fileContents, err := helper.ReadFile(filepath.Join(commonVar.Context, ".odo/env/env.yaml"))
		Expect(err).To(BeNil())
		Expect(fileContents).To(ContainSubstring(componentNamespace))
	})

	When("odo create is executed with the --registry flag", func() {
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

	When("odo create is executed with the --context flag", func() {
		var newContext, envFilePath string
		BeforeEach(func() {
			newContext = filepath.Join(commonVar.Context, "newContext")
			helper.MakeDir(newContext)
			devfilePath = filepath.Join(newContext, devfile)
			envFilePath = filepath.Join(newContext, envFile)
			helper.CopyExample(filepath.Join("source", "nodejs"), newContext)
		})

		AfterEach(func() {
			helper.DeleteDir(newContext)
		})

		It("should successfully create the devfile component in the context", func() {
			helper.Cmd("odo", "create", "nodejs", "--context", newContext).ShouldPass()

			By("checking the devfile and env file exists", func() {
				Expect(util.CheckPathExists(devfilePath)).Should(BeTrue())
				Expect(util.CheckPathExists(envFilePath)).Should(BeTrue())
			})

			By("checking the auto generated name is displayed", func() {
				output := helper.Cmd("odo", "env", "view", "--context", newContext, "-o", "json").ShouldPass().Out()
				value := gjson.Get(output, "spec.name")
				Expect(strings.TrimSpace(value.String())).To(ContainSubstring(strings.TrimSpace("nodejs-" + filepath.Base(strings.ToLower(newContext)))))
			})
		})

		It("should successfully create the devfile component and download the source in the context when used with --starter flag", func() {
			helper.Cmd("odo", "create", "nodejs", "--starter", "nodejs-starter", "--context", newContext).ShouldPass()
			checkNodeJSDirContent(newContext)
		})

		It("should successfully create the devfile component and show json output", func() {
			output := helper.Cmd("odo", "create", "nodejs", "--context", newContext, "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "metadata.name", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Not Pushed"})).To(Equal(true))
		})

		It("should successfully create and push the devfile component with --now and show json output", func() {
			output := helper.Cmd("odo", "create", "nodejs", "--starter", "nodejs-starter", "--context", newContext, "-o", "json", "--now").ShouldPass().Out()
			checkNodeJSDirContent(newContext)
			helper.MatchAllInOutput(output, []string{"Pushed", "nodejs", "Component"})
		})

		It("should successfully create the devfile component and show json output for non connected cluster", func() {
			output := helper.Cmd("odo", "create", "nodejs", "--context", newContext, "-o", "json").WithEnv("KUBECONFIG=/no/such/path", "GLOBALODOCONFIG="+os.Getenv("GLOBALODOCONFIG")).ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "metadata.name", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Unknown"})).To(Equal(true))
		})

		When("the cluster is unreachable", func() {
			var newKubeConfigPath string
			BeforeEach(func() {
				path := os.Getenv("KUBECONFIG")

				// read the contents from the kubeconfig and replace the server entries
				reg := regexp.MustCompile(`server: .*`)
				kubeConfigContents, err := helper.ReadFile(path)
				Expect(err).To(BeNil())
				kubeConfigContents = reg.ReplaceAllString(kubeConfigContents, "server: https://not-reachable.com:443")

				// write to a new file which will be used as the new kubeconfig
				newKubeConfigPath = filepath.Join(commonVar.Context, "newKUBECONFIG")
				newKubeConfig, err := os.Create(newKubeConfigPath)
				Expect(err).To(BeNil())
				defer newKubeConfig.Close()

				_, err = newKubeConfig.WriteString(kubeConfigContents)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(newKubeConfigPath)
			})

			It("should successfully create the devfile component and show json output", func() {
				output := helper.Cmd("odo", "create", "nodejs", "--context", newContext, "-o", "json").WithEnv("KUBECONFIG="+newKubeConfigPath, "GLOBALODOCONFIG="+os.Getenv("GLOBALODOCONFIG")).ShouldPass().Out()
				values := gjson.GetMany(output, "kind", "metadata.name", "status.state")
				Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Unknown"})).To(Equal(true))
			})
		})
	})

	When("odo create is executed with the --now flag", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		})

		It("checks that odo push works with a devfile with now flag", func() {
			output := helper.Cmd("odo", "create", "nodejs", "--now").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})
	})

	When("odo create is executed with the --starter flag", func() {
		BeforeEach(func() {
			contextDevfile = helper.CreateNewContext()
			helper.Chdir(contextDevfile)
			devfilePath = filepath.Join(contextDevfile, devfile)
		})

		AfterEach(func() {
			helper.DeleteDir(contextDevfile)
			helper.Chdir(commonVar.Context)
		})

		It("should successfully create the component and download the source", func() {
			helper.Cmd("odo", "create", "nodejs", "--starter", "nodejs-starter").ShouldPass()
			checkNodeJSDirContent(contextDevfile)
		})

		It("should fail to create the component when an invalid starter project is specified", func() {
			invalidProjectName := "invalid-project-name"
			output := helper.Cmd("odo", "create", "nodejs", "--starter=invalid-project-name").ShouldFail().Err()
			expectedString := "the project: " + invalidProjectName + " specified in --starter does not exist"
			helper.MatchAllInOutput(output, []string{expectedString, "available projects", "nodejs-starter"})
		})

		When("the starter project has git tag or git branch specified", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-branch.yaml"), devfilePath)
			})

			It("should successfully create the component and download the source from the specified branch", func() {
				helper.Cmd("odo", "create", "nodejs", "--starter", "nodejs-starter").ShouldPass()
				checkNodeJSDirContent(contextDevfile)
			})

			It("should successfully create the component and download the source from the specified tag", func() {
				helper.ReplaceString(devfilePath, "revision: test-branch", "revision: 0.0.1")
				helper.Cmd("odo", "create", "nodejs", "--starter", "nodejs-starter").ShouldPass()
				checkNodeJSDirContent(contextDevfile)
			})
		})

		When("the starter project has subDir", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-with-subDir.yaml"), devfilePath)
				helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			})

			It("should successfully create the component and extract the project in the specified subDir path", func() {
				var found, notToBeFound int
				helper.Cmd("odo", "create", "--project", commonVar.Project, "--starter", "springbootproject").ShouldPass()
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
			})
		})
	})

	When("devfile exists in the working directory", func() {
		BeforeEach(func() {
			devfilePath = filepath.Join(commonVar.Context, devfile)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), devfilePath)
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

		It("should fail to create the devfile component", func() {
			By("passing more than 1 arguments", func() {
				helper.Cmd("odo", "create", "nodejs", "mynode").ShouldFail()

			})
			By("invalid value to the --devfile flag", func() {
				helper.Cmd("odo", "create", "nodejs", "--devfile", "/path/to/file").ShouldFail()
			})

			By("creating the devfile component multiple times", func() {
				helper.Cmd("odo", "create", "nodejs").ShouldPass()
				output := helper.Cmd("odo", "create", "nodejs").ShouldFail().Err()
				Expect(output).To(ContainSubstring("this directory already contains a component"))
			})
		})
	})

	When("devfile does not exist in the working directory and user specifies the devfile path via --devfile", func() {
		BeforeEach(func() {
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

		It("should fail to create the devfile component", func() {
			By("using an invalid file system path", func() {
				errOut := helper.Cmd("odo", "create", "nodejs", "--devfile", "@123!").ShouldFail().Err()
				Expect(errOut).To(ContainSubstring("the devfile path you specify is invalid"))
			})
			By("using an invalid URL path", func() {
				errOut := helper.Cmd("odo", "create", "nodejs", "--devfile", "://www.example.com/").ShouldFail().Err()
				Expect(errOut).To(ContainSubstring("the devfile path you specify is invalid"))
			})

			By("passing more than 1 arguments", func() {
				errOut := helper.Cmd("odo", "create", "nodejs", "nodejs", "--devfile", devfilePath).ShouldFail().Err()
				Expect(errOut).To(ContainSubstring("accepts between 0 and 1 arg when using existing devfile, received 2"))
			})

			By("using --registry flag", func() {
				errOut := helper.Cmd("odo", "create", "nodejs", "--devfile", devfilePath, "--registry", "DefaultDevfileRegistry").ShouldFail().Err()
				Expect(errOut).To(ContainSubstring("you can't specify registry via --registry if you want to use the devfile that is specified via --devfile"))
			})
		})
	})

	When("a dangling env file exists in the working directory", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", "java-quarkus").ShouldPass()
			helper.DeleteFile("devfile.yaml")
		})
		It("should successfully create a devfile component and remove the dangling env file", func() {
			out, outerr := helper.Cmd("odo", "create", "nodejs").ShouldPass().OutAndErr()
			helper.MatchAllInOutput(out, []string{
				"Please use `odo push` command to create the component with source deployed"})
			helper.MatchAllInOutput(outerr, []string{
				"Found a dangling env file without a devfile, overwriting it",
			})
		})
	})
	When("creating a component using .devfile.yaml", func() {
		stdout := ""
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, ".devfile.yaml"))
			stdout = helper.Cmd("odo", "create", cmpName, "--project", commonVar.Project).ShouldPass().Out()
		})

		It("should successfully create a devfile component", func() {
			Expect(stdout).To(ContainSubstring("Please use `odo push` command to create the component with source deployed"))
		})
	})
})
