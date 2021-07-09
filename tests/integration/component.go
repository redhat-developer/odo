package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

func componentTests(args ...string) {
	var oc helper.OcRunner
	var commonVar helper.CommonVar
	var appName string
	var cmpName string
	cmpNameDefault := "nodejs"

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
		appName = "app-" + helper.RandString(5)
		cmpName = "cmp-" + helper.RandString(5)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should fail if json is non-existent for a command", func() {
		output := helper.Cmd("odo", "version", "-o", "json").ShouldFail().Err()
		Expect(output).To(ContainSubstring("Machine readable output is not yet implemented for this command"))
	})

	It("should not have machine output for odo version", func() {
		output := helper.Cmd("odo", "version", "--help").ShouldPass().Out()
		Expect(output).NotTo(ContainSubstring("Specify output format, supported format: json"))
	})

	XIt("should be able to create a git component and update it from local to git", func() {
		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)...).ShouldPass()
		helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()

		helper.Cmd("odo", "update", "--git", "https://github.com/openshift/nodejs-ex.git", "--context", commonVar.Context).ShouldPass()
		// check the source location and type in the deployment config
		getSourceLocation := oc.SourceLocationDC(cmpName, appName, commonVar.Project)
		Expect(getSourceLocation).To(ContainSubstring("https://github.com/openshift/nodejs-ex"))
		getSourceType := oc.SourceTypeDC(cmpName, appName, commonVar.Project)
		Expect(getSourceType).To(ContainSubstring("git"))
	})

	XIt("should be able to update a component from git to local", func() {
		helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--git", "https://github.com/openshift/nodejs-ex", "--context", commonVar.Context, "--app", appName)...).ShouldPass()
		helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()

		// update the component config according to the git component
		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

		helper.Cmd("odo", "update", "--local", "./", "--context", commonVar.Context).ShouldPass()

		// check the source location and type in the deployment config
		getSourceLocation := oc.SourceLocationDC(cmpName, appName, commonVar.Project)
		Expect(getSourceLocation).To(ContainSubstring(""))
		getSourceType := oc.SourceTypeDC(cmpName, appName, commonVar.Project)
		Expect(getSourceType).To(ContainSubstring("local"))
	})

	XIt("should pass outside a odo directory with component name as parameter", func() {
		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpNameDefault, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...).ShouldPass()
		info := helper.LocalEnvInfo(commonVar.Context)
		Expect(info.GetApplication(), appName)
		Expect(info.GetName(), cmpNameDefault)
		helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()

		cmpListOutput := helper.Cmd("odo", append(args, "list", "--app", appName, "--project", commonVar.Project)...).ShouldPass()
		Expect(cmpListOutput).To(ContainSubstring(cmpNameDefault))

		actualDesCompJSON := helper.Cmd("odo", append(args, "describe", cmpNameDefault, "--app", appName, "--project", commonVar.Project, "-o", "json")...).ShouldPass().Out()
		valuesDescCJ := gjson.GetMany(actualDesCompJSON, "kind", "metadata.name", "spec.app", "spec.type", "status.state")
		expectedDescCJ := []string{"Component", "nodejs", "app", "nodejs", "Pushed"}
		Expect(helper.GjsonMatcher(valuesDescCJ, expectedDescCJ)).To(Equal(true))

		helper.Cmd("odo", append(args, "delete", cmpNameDefault, "--app", appName, "--project", commonVar.Project, "-f")...).ShouldPass()
	})

	It("should retain the same environment variable on multiple push", func() {
		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...).ShouldPass()
		helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()

		helper.Chdir(commonVar.Context)
		helper.Cmd("odo", "config", "set", "--env", "FOO=BAR").ShouldPass()
		helper.Cmd("odo", append(args, "push")...).ShouldPass()
		info := helper.LocalEnvInfo(commonVar.Context)
		Expect(info.GetApplication(), appName)
		Expect(info.GetName(), cmpName)
		envVars := oc.GetEnvsDevFileDeployment(cmpName, appName, commonVar.Project)
		val, ok := envVars["FOO"]
		Expect(ok).To(BeTrue())
		Expect(val).To(Equal("BAR"))
	})

	When("in context directory", func() {

		BeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})

		It("should show an error when ref flag is provided with sources except git", func() {
			outputErr := helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, cmpName, "--ref", "test")...).ShouldFail().Err()
			Expect(outputErr).To(ContainSubstring("the --ref flag is only valid for --git flag"))
		})

		It("should succeed listing catalog components", func() {
			// Since components catalog is constantly changing, we simply check to see if this command passes.. rather than checking the JSON each time.
			helper.Cmd("odo", "catalog", "list", "components", "-o", "json").ShouldPass()
		})

		It("should fail the create command as --git flag, which is specific to s2i component creation, is used without --s2i flag", func() {
			output := helper.Cmd("odo", "create", "nodejs", cmpName, "--git", "https://github.com/openshift/nodejs-ex", "--context", commonVar.Context, "--app", appName).ShouldFail().Err()
			Expect(output).Should(ContainSubstring("flag --git, requires --s2i flag to be set, when deploying S2I (Source-to-Image) components"))
		})

		It("should fail the create command as --binary flag, which is specific to s2i component creation, is used without --s2i flag", func() {
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)

			output := helper.Cmd("odo", "create", "java:8", "sb-jar-test", "--binary", filepath.Join(commonVar.Context, "sb.jar"), "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).Should(ContainSubstring("flag --binary, requires --s2i flag to be set, when deploying S2I (Source-to-Image) components"))
		})

		It("should work for s2i component from a devfile directory", func() {
			newContext := path.Join(commonVar.Context, "newContext")
			helper.MakeDir(newContext)
			helper.Chdir(newContext)
			cmpName2 := helper.RandString(6)
			helper.Cmd("odo", "create", "--starter", "nodejs").ShouldPass()
			context2 := helper.CreateNewContext()
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--context", context2, cmpName2).ShouldPass()
			output := helper.Cmd("odo", "describe", "--context", context2).ShouldPass().Out()
			Expect(output).To(ContainSubstring(fmt.Sprint("Component Name: ", cmpName2)))
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			helper.DeleteDir(context2)
		})

		It("should list the component in the same app when one is pushed and the other one is not pushed", func() {
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)...).ShouldPass()
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)
			helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()

			context2 := helper.CreateNewContext()
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", "cmp-git-2", "--project", commonVar.Project, "--context", context2, "--app", appName)...).ShouldPass()
			info = helper.LocalEnvInfo(context2)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), "cmp-git-2")
			cmpList := helper.Cmd("odo", append(args, "list", "--context", context2)...).ShouldPass().Out()
			helper.MatchAllInOutput(cmpList, []string{cmpName, "cmp-git-2", "Not Pushed", "Pushed"})

			helper.Cmd("odo", append(args, "delete", "-f", "--all", "--context", commonVar.Context)...).ShouldPass()
			helper.Cmd("odo", append(args, "delete", "-f", "--all", "--context", context2)...).ShouldPass()
			helper.DeleteDir(context2)
		})

		It("should create a local python component, push it and then delete it using --all flag", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.Cmd("odo", append(args, "create", "--s2i", "python", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...).ShouldPass()
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)
			helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()
			helper.Cmd("odo", append(args, "delete", "--context", commonVar.Context, "-f")...).ShouldPass()
			helper.Cmd("odo", append(args, "delete", "--all", "-f", "--context", commonVar.Context)...).ShouldPass()
			componentList := helper.Cmd("odo", append(args, "list", "--app", appName, "--project", commonVar.Project)...).ShouldPass().Out()
			Expect(componentList).NotTo(ContainSubstring(cmpName))
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).NotTo(ContainElement(".odo"))
		})

		It("should create a local python component, push it and then delete it using --all flag in local directory", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.Cmd("odo", append(args, "create", "--s2i", "python", cmpName, "--app", appName, "--project", commonVar.Project)...).ShouldPass()
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)
			helper.Cmd("odo", append(args, "push")...).ShouldPass()
			helper.Cmd("odo", append(args, "delete", "--all", "-f")...).ShouldPass()
			componentList := helper.Cmd("odo", append(args, "list", "--app", appName, "--project", commonVar.Project)...).ShouldPass().Out()
			Expect(componentList).NotTo(ContainSubstring(cmpName))
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).NotTo(ContainElement(".odo"))
		})

		It("should create a local python component and check for unsupported warning", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			output := helper.Cmd("odo", append(args, "create", "--s2i", "python", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Warning: python is not fully supported by odo, and it is not guaranteed to work"))
		})

		It("should create a local nodejs component and check unsupported warning hasn't occurred", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			output := helper.Cmd("odo", append(args, "create", "--s2i", "nodejs:latest", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...).ShouldPass().Out()
			Expect(output).NotTo(ContainSubstring("Warning"))
		})

		It("should create a local java component and check unsupported warning hasn't occurred", func() {
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)
			output := helper.Cmd("odo", append(args, "create", "--s2i", "java:latest", cmpName, "--project", commonVar.Project, "--context", commonVar.Context)...).ShouldPass().Out()
			Expect(output).NotTo(ContainSubstring("Warning"))
		})

		// TODO: Fix later
		XIt("should list out pushed components of different projects in json format along with path flag", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project)...).ShouldPass()
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "app")
			Expect(info.GetName(), "nodejs")
			helper.Cmd("odo", append(args, "push")...).ShouldPass()

			project2 := helper.CreateRandProject()
			context2 := helper.CreateNewContext()
			helper.Chdir(context2)
			helper.CopyExample(filepath.Join("source", "python"), context2)
			helper.Cmd("odo", append(args, "create", "--s2i", "python", "python", "--project", project2)...).ShouldPass()
			info = helper.LocalEnvInfo(context2)
			Expect(info.GetApplication(), "app")
			Expect(info.GetName(), "python")

			helper.Cmd("odo", append(args, "push")...).ShouldPass()

			actual, err := helper.Unindented(helper.Cmd("odo", append(args, "list", "-o", "json", "--path", filepath.Dir(commonVar.Context))...).ShouldPass().Out())
			Expect(err).Should(BeNil())
			helper.Chdir(commonVar.Context)
			helper.DeleteDir(context2)
			helper.DeleteProject(project2)
			// this orders the json
			expected := fmt.Sprintf(`"metadata":{"name":"nodejs","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"nodejs","sourceType": "local","ports":["8080/TCP"]}`, commonVar.Project)
			Expect(actual).Should(ContainSubstring(expected))
			// this orders the json
			expected = fmt.Sprintf(`"metadata":{"name":"python","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"python","sourceType": "local","ports":["8080/TCP"]}`, project2)
			Expect(actual).Should(ContainSubstring(expected))

		})

		When("creating a named s2i nodejs component", func() {

			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)...).ShouldPass()
				info := helper.LocalEnvInfo(commonVar.Context)
				Expect(info.GetApplication(), appName)
				Expect(info.GetName(), cmpName)
			})

			It("should list the component when it is not pushed", func() {
				cmpList := helper.Cmd("odo", append(args, "list", "--context", commonVar.Context)...).ShouldPass().Out()
				helper.MatchAllInOutput(cmpList, []string{cmpName, "Not Pushed"})
				helper.Cmd("odo", append(args, "delete", "-f", "--all", "--context", commonVar.Context)...).ShouldPass()
			})

			XIt("should list the state as unknown for disconnected cluster", func() {
				kubeconfigOrig := os.Getenv("KUBECONFIG")

				unset := func() {
					// KUBECONFIG defaults to ~/.kube/config so it can be empty in some cases.
					if kubeconfigOrig != "" {
						os.Setenv("KUBECONFIG", kubeconfigOrig)
					} else {
						os.Unsetenv("KUBECONFIG")
					}
				}

				os.Setenv("KUBECONFIG", "/no/such/path")

				defer unset()
				cmpList := helper.Cmd("odo", append(args, "list", "--context", commonVar.Context, "--v", "9")...).ShouldPass().Out()

				helper.MatchAllInOutput(cmpList, []string{cmpName, "Unknown"})
				unset()

				fmt.Printf("kubeconfig before delete %v", os.Getenv("KUBECONFIG"))
				helper.Cmd("odo", append(args, "delete", "-f", "--all", "--context", commonVar.Context)...).ShouldPass()
			})

			It("should describe the component when it is not pushed", func() {
				helper.Cmd("odo", "url", "create", "url-1", "--context", commonVar.Context).ShouldPass()
				helper.Cmd("odo", "url", "create", "url-2", "--context", commonVar.Context).ShouldPass()
				helper.Cmd("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context).ShouldPass()
				cmpDescribe := helper.Cmd("odo", append(args, "describe", "--context", commonVar.Context)...).ShouldPass().Out()
				helper.MatchAllInOutput(cmpDescribe, []string{
					cmpName,
					"nodejs",
					"url-1",
					"url-2",
					"storage-1",
				})

			})

			It("should fail to create component twice from same directory", func() {
				output := helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project)...).ShouldFail().Err()
				Expect(output).To(ContainSubstring("this directory already contains a component"))
			})

			When("odo push is executed", func() {

				BeforeEach(func() {
					helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()
				})

				It("should not list component even in new project with --project and --context at the same time", func() {
					projectList := helper.Cmd("odo", "project", "list").ShouldPass().Out()
					Expect(projectList).To(ContainSubstring(commonVar.Project))
					helper.Cmd("odo", "list", "--project", commonVar.Project, "--context", commonVar.Context).ShouldFail()
				})

				It("should list the component", func() {
					cmpList := helper.Cmd("odo", append(args, "list", "--project", commonVar.Project)...).ShouldPass().Out()
					Expect(cmpList).To(ContainSubstring(cmpName))
					actualCompListJSON := helper.Cmd("odo", append(args, "list", "--project", commonVar.Project, "-o", "json")...).ShouldPass().Out()
					valuesCList := gjson.GetMany(actualCompListJSON, "kind", "devfileComponents.0.kind", "devfileComponents.0.metadata.name", "devfileComponents.0.spec.app")
					expectedCList := []string{"List", "Component", cmpName, appName}
					Expect(helper.GjsonMatcher(valuesCList, expectedCList)).To(Equal(true))

					cmpAllList := helper.Cmd("odo", append(args, "list", "--all-apps")...).ShouldPass().Out()
					Expect(cmpAllList).To(ContainSubstring(cmpName))
				})

				It("should delete --all", func() {
					helper.Cmd("odo", append(args, "delete", "--context", commonVar.Context, "-f", "--all")...).ShouldPass()
					componentList := helper.Cmd("odo", append(args, "list", "--app", appName, "--project", commonVar.Project)...).ShouldPass().Out()
					Expect(componentList).NotTo(ContainSubstring(cmpName))
					files := helper.ListFilesInDir(commonVar.Context)
					Expect(files).NotTo(ContainElement(".odo"))
				})
			})
		})

		When("creating an s2i nodejs component with context .", func() {

			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", ".", "--app", appName)...).ShouldPass()
			})

			It("should create default named component when passed same context differently", func() {
				dir := filepath.Base(commonVar.Context)
				componentName := helper.GetLocalEnvInfoValueWithContext("Name", commonVar.Context)
				Expect(componentName).To(ContainSubstring("nodejs"))
				Expect(componentName).To(ContainSubstring(dir))

				info := helper.LocalEnvInfo(commonVar.Context)
				Expect(info.GetApplication(), appName)
				Expect(info.GetName(), componentName)

				helper.DeleteDir(filepath.Join(commonVar.Context, ".odo"))
				helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)...).ShouldPass()
				newComponentName := helper.GetLocalEnvInfoValueWithContext("Name", commonVar.Context)
				Expect(newComponentName).To(ContainSubstring("nodejs"))
				Expect(newComponentName).To(ContainSubstring(dir))
			})
		})

		When("a binary is copied in the current directory", func() {

			BeforeEach(func() {
				oc.ImportJavaIS(commonVar.Project)
				helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)

			})

			It("should not fail when --context is not set", func() {
				binaryFilePath := filepath.Join(commonVar.Context, "sb.jar")
				if runtime.GOOS == "darwin" {
					binaryFilePath = filepath.Join("/private", binaryFilePath)
				}
				helper.Cmd("odo", append(args, "create", "--s2i", "java:8", cmpName, "--project",
			helper.Cmd("odo", append(args, "create", "--s2i", "java:8", cmpName, "--project",
					commonVar.Project, "--binary", binaryFilePath)...).ShouldPass()
				info := helper.LocalEnvInfo(commonVar.Context)
				Expect(info.GetName(), cmpName)
			})

			It("should fail when --binary is not in --context folder", func() {
				newContext := helper.CreateNewContext()
				defer helper.DeleteDir(newContext)

				output := helper.Cmd("odo", append(args, "create", "--s2i", "java:8", cmpName, "--project",
					commonVar.Project, "--binary", filepath.Join(commonVar.Context, "sb.jar"), "--context", newContext)...).ShouldFail().Err()
				Expect(output).To(ContainSubstring("inside of the context directory"))
			})

			It("should be valid if path is relative and includes ../", func() {
				relativeContext := fmt.Sprintf("..%c%s", filepath.Separator, filepath.Base(commonVar.Context))
				fmt.Printf("relativeContext = %#v\n", relativeContext)
				binaryFilePath := filepath.Join(commonVar.Context, "sb.jar")
				if runtime.GOOS == "darwin" {
					binaryFilePath = filepath.Join("/private", binaryFilePath)
				}
				helper.Cmd("odo", append(args, "create", "--s2i", "java:8", cmpName, "--project",
					commonVar.Project, "--binary", binaryFilePath, "--context", relativeContext)...).ShouldPass()
				info := helper.LocalEnvInfo(relativeContext)
				Expect(info.GetApplication(), "app")
				Expect(info.GetName(), cmpName)
			})
		})
	})

	When("creating an s2i component", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpNameDefault, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...).ShouldPass()
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpNameDefault)
		})

		It("should pass inside a odo directory without component name as parameter", func() {
			helper.Cmd("odo", "url", "create", "example", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()

			// changing directory to the context directory
			helper.Chdir(commonVar.Context)
			cmpListOutput := helper.Cmd("odo", append(args, "list")...).ShouldPass().Out()
			Expect(cmpListOutput).To(ContainSubstring(cmpNameDefault))
			cmpDescribe := helper.Cmd("odo", append(args, "describe")...).ShouldPass().Out()
			helper.MatchAllInOutput(cmpDescribe, []string{cmpNameDefault, "nodejs"})

			url := helper.DetermineRouteURL(commonVar.Context)
			Expect(cmpDescribe).To(ContainSubstring(url))

			helper.Cmd("odo", append(args, "delete", "-f")...).ShouldPass()
		})

		It("should fail outside a odo directory without component name as parameter", func() {
			helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()
			// commands should fail as the component name is missing
			helper.Cmd("odo", append(args, "describe", "--app", appName, "--project", commonVar.Project)...).ShouldFail()
			helper.Cmd("odo", append(args, "delete", "-f", "--app", appName, "--project", commonVar.Project)...).ShouldFail()
		})

	})

	When("creating a named s2i component with urls and storages", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...).ShouldPass()
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)

			helper.Cmd("odo", "url", "create", "example-1", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context).ShouldPass()

			helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()

			helper.Cmd("odo", "url", "create", "example-2", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", "storage-2", "--size", "1Gi", "--path", "/data2", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()
		})

		It("should delete the component and the owned resources", func() {
			helper.Cmd("odo", append(args, "delete", "-f", "--context", commonVar.Context)...).ShouldPass()
			oc.WaitAndCheckForExistence("routes", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("dc", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("pvc", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("bc", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("is", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("service", commonVar.Project, 1)
		})

		It("should delete the component and the owned resources with wait flag", func() {
			// delete with --wait flag
			helper.Cmd("odo", append(args, "delete", "-f", "-w", "--context", commonVar.Context)...).ShouldPass()
			helper.VerifyResourcesDeleted(oc, []helper.ResourceInfo{
				{
					ResourceType: helper.ResourceTypeRoute,
					ResourceName: "example",
					Namespace:    commonVar.Project,
				},
				{
					ResourceType: helper.ResourceTypeService,
					ResourceName: "example",
					Namespace:    commonVar.Project,
				},
				{
					// verify s2i pvc is delete
					ResourceType: helper.ResourceTypePVC,
					ResourceName: "s2idata",
					Namespace:    commonVar.Project,
				},
				{
					ResourceType: helper.ResourceTypePVC,
					ResourceName: "storage-1",
					Namespace:    commonVar.Project,
				},
				{
					ResourceType: helper.ResourceTypePVC,
					ResourceName: "storage-2",
					Namespace:    commonVar.Project,
				},
				{
					ResourceType: helper.ResourceTypeDeploymentConfig,
					ResourceName: cmpName,
					Namespace:    commonVar.Project,
				},
			})
		})
	})

	When("creating a component with a numeric named context", func() {

		var contextNumeric string

		BeforeEach(func() {
			var err error
			ts := time.Now().UnixNano()
			contextNumeric, err = ioutil.TempDir("", fmt.Sprint(ts))
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			helper.DeleteDir(contextNumeric)
		})

		It("should create default named component in a directory with numeric name", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), contextNumeric)
			helper.Cmd("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", contextNumeric, "--app", appName)...).ShouldPass()
			info := helper.LocalEnvInfo(contextNumeric)
			Expect(info.GetApplication(), appName)
			helper.Cmd("odo", append(args, "push", "--context", contextNumeric, "-v4")...).ShouldPass()
		})
	})

	When("creating a component using symlink", func() {

		var symLinkPath string

		BeforeEach(func() {
			if runtime.GOOS == "windows" {
				Skip("Skipping test because for symlink creation on platform like Windows, go library needs elevated privileges.")
			}
			// create a symlink
			symLinkName := helper.RandString(10)
			helper.CreateSymLink(commonVar.Context, filepath.Join(filepath.Dir(commonVar.Context), symLinkName))
			symLinkPath = filepath.Join(filepath.Dir(commonVar.Context), symLinkName)
		})

		AfterEach(func() {
			// remove the symlink
			err := os.Remove(symLinkPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should be able to deploy a spring boot uberjar file using symlinks in all odo commands", func() {
			oc.ImportJavaIS(commonVar.Project)

			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)

			// create the component using symlink
			helper.Cmd("odo", append(args, "create", "--s2i", "java:8", cmpName, "--project",
				commonVar.Project, "--binary", filepath.Join(symLinkPath, "sb.jar"), "--context", symLinkPath)...).ShouldPass()

			// Create a URL and push without using the symlink
			helper.Cmd("odo", "url", "create", "uberjaropenjdk", "--port", "8080", "--context", symLinkPath).ShouldPass()
			info := helper.LocalEnvInfo(symLinkPath)
			Expect(info.GetApplication(), "app")
			Expect(info.GetName(), cmpName)

			helper.Cmd("odo", append(args, "push", "--context", symLinkPath)...).ShouldPass()
			routeURL := helper.DetermineRouteURL(symLinkPath)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "HTTP Booster", 300, 1)

			// Delete the component
			helper.Cmd("odo", append(args, "delete", "-f", "--context", symLinkPath)...).ShouldPass()
		})

		It("should be able to deploy a wildfly war file using symlinks in some odo commands", func() {
			helper.CopyExample(filepath.Join("binary", "java", "wildfly"), commonVar.Context)
			helper.Cmd("odo", append(args, "create", "--s2i", "wildfly", cmpName, "--project",
				commonVar.Project, "--binary", filepath.Join(symLinkPath, "ROOT.war"), "--context", symLinkPath)...).ShouldPass()

			// Create a URL
			helper.Cmd("odo", "url", "create", "warfile", "--port", "8080", "--context", commonVar.Context).ShouldPass()
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "app")
			Expect(info.GetName(), cmpName)
			helper.Cmd("odo", append(args, "push", "--context", commonVar.Context)...).ShouldPass()
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Sample", 90, 1)

			// Delete the component
			helper.Cmd("odo", append(args, "delete", "-f", "--context", commonVar.Context)...).ShouldPass()
		})
	})

	Context("convert s2i to devfile", func() {

		BeforeEach(func() {
			os.Setenv("ODO_EXPERIMENTAL", "true")
		})

		AfterEach(func() {
			os.Unsetenv("ODO_EXPERIMENTAL")
		})

		It("should convert s2i component to devfile component successfully", func() {
			urlName := "url1"
			storageName := "storage1"

			// create a s2i component using --s2i that generates a devfile
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName).ShouldPass()
			helper.Cmd("odo", "url", "create", urlName, "--port", "8080", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", storageName, "--path", "/data1", "--size", "1Gi", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// check the status of devfile component
			stdout := helper.Cmd("odo", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{cmpName, "Devfile Components", "Pushed"})

			// verify the url
			stdout = helper.Cmd("odo", "url", "list", "--context", commonVar.Context).ShouldPass().Out()

			helper.MatchAllInOutput(stdout, []string{urlName, "Pushed", "false", "route"})
			//verify storage
			stdout = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(stdout, []string{storageName, "Pushed"})

		})
	})

	When("components are not created/managed by odo", func() {
		var runner helper.CliRunner
		type compStruct struct {
			App, Name string
		}
		// This array will contain static data taken from tests/examples/manifests/dc-label.yaml and
		// tests/examples/manifests/deployment-httpd-label.yaml, tests/examples/manifests/deployment-app-label.yaml.
		// If this test breaks, check if the data here matches with that of the manifest files
		var compList = []compStruct{
			{"app", "example-deployment"},
		}

		BeforeEach(func() {
			// Create resources that are not managed by odo
			runner = helper.GetCliRunner()
			dfile := filepath.Join(commonVar.Context, "deployment-app-label.yaml")
			helper.CopyManifestFile("deployment-app-label.yaml", dfile)
			runner.Run("apply", "-f", dfile).Wait()

			// if it is openshift env, we also deploy the deploymentconfig yaml
			if !helper.IsKubernetesCluster() {
				dcfile := filepath.Join(commonVar.Context, "dc-label.yaml")
				helper.CopyManifestFile("dc-label.yaml", dcfile)
				runner.Run("apply", "-f", dcfile).Wait()
				compList = append(compList, compStruct{"app", "example-dc"})
			}
		})

		// verifyListOutput verifies if the components not managed by odo are listed
		var verifyListOutput = func(output string, componentList []compStruct) {
			Expect(output).To(ContainSubstring("Other Components running on the cluster(read-only)"))
			for _, comp := range componentList {
				Expect(output).To(ContainSubstring(comp.Name))
				Expect(output).To(ContainSubstring(comp.App))
			}
		}

		It("should list the components", func() {
			output := helper.Cmd("odo", append(args, "list")...).ShouldPass().Out()
			verifyListOutput(output, compList)
		})

		When("the component has a different app name than the default 'app'", func() {

			BeforeEach(func() {
				// Create resources that are not managed by odo
				dfile := filepath.Join(commonVar.Context, "deployment-httpd-label.yaml")
				helper.CopyManifestFile("deployment-httpd-label.yaml", dfile)
				runner.Run("apply", "-f", dfile).Wait()
			})

			It("should list the components with --all-apps flag", func() {
				output := helper.Cmd("odo", append(args, "list", "--all-apps")...).ShouldPass().Out()
				verifyListOutput(output, append(compList, compStruct{"httpd", "example-deployment-httpd"}))
			})

			It("should list the components with --app flag", func() {
				output := helper.Cmd("odo", append(args, "list", "--app", "httpd")...).ShouldPass().Out()
				verifyListOutput(output, []compStruct{{"httpd", "example-deployment-httpd"}})
			})
		})

		It("should list the components in json format with -o json flag", func() {
			output := helper.Cmd("odo", append(args, "list", "--project", commonVar.Project, "-o", "json")...).ShouldPass().Out()
			for _, comp := range compList {
				valuesCList := gjson.GetMany(output, "kind", "otherComponents.#.kind", "otherComponents.#.metadata.name", "otherComponents.#.spec.app")
				expectedCList := []string{"List", "Component", comp.Name, comp.App}
				Expect(helper.GjsonMatcher(valuesCList, expectedCList)).To(Equal(true))
			}
		})

		When("executing odo list from other project", func() {

			var otherProject string

			BeforeEach(func() {
				otherProject = commonVar.CliRunner.CreateRandNamespaceProject()
			})

			AfterEach(func() {
				helper.Cmd("odo", "project", "set", commonVar.Project).ShouldPass()
				commonVar.CliRunner.DeleteNamespaceProject(otherProject)
			})

			It("should list the components with --project flag", func() {
				output := helper.Cmd("odo", append(args, "list", "--project", commonVar.Project)...).ShouldPass().Out()
				verifyListOutput(output, compList)
			})

		})
	})
}
