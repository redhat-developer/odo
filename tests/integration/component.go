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

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Generic machine readable output tests", func() {

		It("Command should fail if json is non-existent for a command", func() {
			output := helper.CmdShouldFail("odo", "version", "-o", "json")
			Expect(output).To(ContainSubstring("Machine readable output is not yet implemented for this command"))
		})

		It("Help for odo version should not contain machine output", func() {
			output := helper.CmdShouldPass("odo", "version", "--help")
			Expect(output).NotTo(ContainSubstring("Specify output format, supported format: json"))
		})

	})

	Context("Creating component", func() {
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})

		It("should create but not list component even in new project with --project and --context at the same time", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")...)

			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "testing")
			Expect(info.GetName(), "cmp-git")
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context, "-v4")...)
			projectList := helper.CmdShouldPass("odo", "project", "list")
			Expect(projectList).To(ContainSubstring(commonVar.Project))
			helper.CmdShouldFail("odo", "list", "--project", commonVar.Project, "--context", commonVar.Context)
		})

		// works
		It("should create default named component when passed same context differently", func() {
			dir := filepath.Base(commonVar.Context)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", ".", "--app", "testing")...)
			componentName := helper.GetLocalEnvInfoValueWithContext("Name", commonVar.Context)
			Expect(componentName).To(ContainSubstring("nodejs"))
			Expect(componentName).To(ContainSubstring(dir))

			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "testing")
			Expect(info.GetName(), componentName)

			helper.DeleteDir(filepath.Join(commonVar.Context, ".odo"))
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")...)
			newComponentName := helper.GetLocalEnvInfoValueWithContext("Name", commonVar.Context)
			Expect(newComponentName).To(ContainSubstring("nodejs"))
			Expect(newComponentName).To(ContainSubstring(dir))
		})

		It("should show an error when ref flag is provided with sources except git", func() {
			outputErr := helper.CmdShouldFail("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "cmp-git", "--ref", "test")...)
			Expect(outputErr).To(ContainSubstring("the --ref flag is only valid for --git flag"))
		})

		It("create component twice fails from same directory", func() {
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project)...)
			output := helper.CmdShouldFail("odo", append(args, "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project)...)
			Expect(output).To(ContainSubstring("this directory already contains a component"))
		})

		// TODO: Fix later
		// It("should list out pushed components of different projects in json format along with path flag", func() {
		// 	helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		// 	helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project)...)
		// 	info := helper.LocalEnvInfo(commonVar.Context)
		// 	Expect(info.GetApplication(), "app")
		// 	Expect(info.GetName(), "nodejs")
		// 	helper.CmdShouldPass("odo", append(args, "push")...)

		// 	project2 := helper.CreateRandProject()
		// 	context2 := helper.CreateNewContext()
		// 	helper.Chdir(context2)
		// 	helper.CopyExample(filepath.Join("source", "python"), context2)
		// 	helper.CmdShouldPass("odo", append(args, "create", "--s2i", "python", "python", "--project", project2)...)
		// 	info = helper.LocalEnvInfo(context2)
		// 	Expect(info.GetApplication(), "app")
		// 	Expect(info.GetName(), "python")

		// 	helper.CmdShouldPass("odo", append(args, "push")...)

		// 	actual, err := helper.Unindented(helper.CmdShouldPass("odo", append(args, "list", "-o", "json", "--path", filepath.Dir(commonVar.Context))...))
		// 	Expect(err).Should(BeNil())
		// 	helper.Chdir(commonVar.Context)
		// 	helper.DeleteDir(context2)
		// 	helper.DeleteProject(project2)
		// 	// this orders the json
		// 	expected := fmt.Sprintf(`"metadata":{"name":"nodejs","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"nodejs","sourceType": "local","ports":["8080/TCP"]}`, commonVar.Project)
		// 	Expect(actual).Should(ContainSubstring(expected))
		// 	// this orders the json
		// 	expected = fmt.Sprintf(`"metadata":{"name":"python","namespace":"%s","creationTimestamp":null},"spec":{"app":"app","type":"python","sourceType": "local","ports":["8080/TCP"]}`, project2)
		// 	Expect(actual).Should(ContainSubstring(expected))

		// })

		It("should list the component", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")...)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "testing")
			Expect(info.GetName(), "cmp-git")
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			cmpList := helper.CmdShouldPass("odo", append(args, "list", "--project", commonVar.Project)...)
			Expect(cmpList).To(ContainSubstring("cmp-git"))
			actualCompListJSON := helper.CmdShouldPass("odo", append(args, "list", "--project", commonVar.Project, "-o", "json")...)
			valuesCList := gjson.GetMany(actualCompListJSON, "kind", "devfileComponents.0.kind", "devfileComponents.0.metadata.name", "devfileComponents.0.spec.app")
			expectedCList := []string{"List", "Component", "cmp-git", "testing"}
			Expect(helper.GjsonMatcher(valuesCList, expectedCList)).To(Equal(true))

			cmpAllList := helper.CmdShouldPass("odo", append(args, "list", "--all-apps")...)
			Expect(cmpAllList).To(ContainSubstring("cmp-git"))
			helper.CmdShouldPass("odo", append(args, "delete", "cmp-git", "-f")...)
		})

		It("should list the component when it is not pushed", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")...)

			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "testing")
			Expect(info.GetName(), "cmp-git")
			cmpList := helper.CmdShouldPass("odo", append(args, "list", "--context", commonVar.Context)...)
			helper.MatchAllInOutput(cmpList, []string{"cmp-git", "Not Pushed"})
			helper.CmdShouldPass("odo", append(args, "delete", "-f", "--all", "--context", commonVar.Context)...)
		})

		// It("should list the state as unknown for disconnected cluster", func() {
		// 	helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		// 	helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")...)
		// 	info := helper.LocalEnvInfo(commonVar.Context)
		// 	Expect(info.GetApplication(), "testing")
		// 	Expect(info.GetName(), "cmp-git")
		// 	kubeconfigOrig := os.Getenv("KUBECONFIG")

		// 	unset := func() {
		// 		// KUBECONFIG defaults to ~/.kube/config so it can be empty in some cases.
		// 		if kubeconfigOrig != "" {
		// 			os.Setenv("KUBECONFIG", kubeconfigOrig)
		// 		} else {
		// 			os.Unsetenv("KUBECONFIG")
		// 		}
		// 	}

		// 	os.Setenv("KUBECONFIG", "/no/such/path")

		// 	defer unset()
		// 	cmpList := helper.CmdShouldPass("odo", append(args, "list", "--context", commonVar.Context, "--v", "9")...)

		// 	helper.MatchAllInOutput(cmpList, []string{"cmp-git", "Unknown"})
		// 	unset()

		// 	fmt.Printf("kubeconfig before delete %v", os.Getenv("KUBECONFIG"))
		// 	helper.CmdShouldPass("odo", append(args, "delete", "-f", "--all", "--context", commonVar.Context)...)
		// })

		It("should describe the component when it is not pushed", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")...)
			helper.CmdShouldPass("odo", "url", "create", "url-1", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "url", "create", "url-2", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "testing")
			Expect(info.GetName(), "cmp-git")
			cmpDescribe := helper.CmdShouldPass("odo", append(args, "describe", "--context", commonVar.Context)...)
			helper.MatchAllInOutput(cmpDescribe, []string{
				"cmp-git",
				"nodejs",
				"url-1",
				"url-2",
				"storage-1",
			})

		})

		It("checks that odo describe works for s2i component from a devfile directory", func() {
			newContext := path.Join(commonVar.Context, "newContext")
			helper.MakeDir(newContext)
			helper.Chdir(newContext)
			cmpName2 := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "--starter", "nodejs")
			context2 := helper.CreateNewContext()
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--context", context2, cmpName2)
			output := helper.CmdShouldPass("odo", "describe", "--context", context2)
			Expect(output).To(ContainSubstring(fmt.Sprint("Component Name: ", cmpName2)))
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			helper.DeleteDir(context2)
		})

		It("should list the component in the same app when one is pushed and the other one is not pushed", func() {
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")...)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "testing")
			Expect(info.GetName(), "cmp-git")
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			context2 := helper.CreateNewContext()
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git-2", "--project", commonVar.Project, "--context", context2, "--app", "testing")...)
			info = helper.LocalEnvInfo(context2)
			Expect(info.GetApplication(), "testing")
			Expect(info.GetName(), "cmp-git-2")
			cmpList := helper.CmdShouldPass("odo", append(args, "list", "--context", context2)...)
			helper.MatchAllInOutput(cmpList, []string{"cmp-git", "cmp-git-2", "Not Pushed", "Pushed"})

			helper.CmdShouldPass("odo", append(args, "delete", "-f", "--all", "--context", commonVar.Context)...)
			helper.CmdShouldPass("odo", append(args, "delete", "-f", "--all", "--context", context2)...)
			helper.DeleteDir(context2)
		})

		It("should succeed listing catalog components", func() {
			// Since components catalog is constantly changing, we simply check to see if this command passes.. rather than checking the JSON each time.
			helper.CmdShouldPass("odo", "catalog", "list", "components", "-o", "json")
		})

		It("binary component should not fail when --context is not set", func() {
			oc.ImportJavaIS(commonVar.Project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)
			// Was failing due to https://github.com/openshift/odo/issues/1969
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "java:8", "sb-jar-test", "--project",
				commonVar.Project, "--binary", filepath.Join(commonVar.Context, "sb.jar"))...)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetName(), "sb-jar-test")
		})

		It("binary component should fail when --binary is not in --context folder", func() {
			oc.ImportJavaIS(commonVar.Project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)

			newContext := helper.CreateNewContext()
			defer helper.DeleteDir(newContext)

			output := helper.CmdShouldFail("odo", append(args, "create", "--s2i", "java:8", "sb-jar-test", "--project",
				commonVar.Project, "--binary", filepath.Join(commonVar.Context, "sb.jar"), "--context", newContext)...)
			Expect(output).To(ContainSubstring("inside of the context directory"))
		})

		It("binary component is valid if path is relative and includes ../", func() {
			oc.ImportJavaIS(commonVar.Project)
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)

			relativeContext := fmt.Sprintf("..%c%s", filepath.Separator, filepath.Base(commonVar.Context))
			fmt.Printf("relativeContext = %#v\n", relativeContext)

			if runtime.GOOS == "darwin" {
				helper.CmdShouldPass("odo", append(args, "create", "--s2i", "java:8", "sb-jar-test", "--project",
					commonVar.Project, "--binary", filepath.Join("/private", commonVar.Context, "sb.jar"), "--context", relativeContext)...)
			} else {
				helper.CmdShouldPass("odo", append(args, "create", "--s2i", "java:8", "sb-jar-test", "--project",
					commonVar.Project, "--binary", filepath.Join(commonVar.Context, "sb.jar"), "--context", relativeContext)...)
			}
			info := helper.LocalEnvInfo(relativeContext)
			Expect(info.GetApplication(), "app")
			Expect(info.GetName(), "sb-jar-test")
		})

		It("should fail the create command as --git flag, which is specific to s2i component creation, is used without --s2i flag", func() {
			output := helper.CmdShouldFail("odo", "create", "nodejs", "cmp-git", "--git", "https://github.com/openshift/nodejs-ex", "--context", commonVar.Context, "--app", "testing")
			Expect(output).Should(ContainSubstring("flag --git, requires --s2i flag to be set, when deploying S2I (Source-to-Image) components"))
		})

		It("should fail the create command as --binary flag, which is specific to s2i component creation, is used without --s2i flag", func() {
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)

			output := helper.CmdShouldFail("odo", "create", "java:8", "sb-jar-test", "--binary", filepath.Join(commonVar.Context, "sb.jar"), "--context", commonVar.Context)
			Expect(output).Should(ContainSubstring("flag --binary, requires --s2i flag to be set, when deploying S2I (Source-to-Image) components"))
		})
	})

	Context("when component is in the current directory and --project flag is used", func() {

		appName := "app"
		componentName := "my-component"

		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})

		It("create local nodejs component twice and fail", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "--env", "key=value,key1=value1")...)
			output := helper.CmdShouldFail("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "--env", "key=value,key1=value1")...)
			Expect(output).To(ContainSubstring("this directory already contains a component"))
		})

		It("creates and pushes local nodejs component and then deletes --all", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", componentName, "--app", appName, "--project", commonVar.Project, "--env", "key=value,key1=value1")...)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), componentName)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)
			helper.CmdShouldPass("odo", append(args, "delete", "--context", commonVar.Context, "-f", "--all")...)
			componentList := helper.CmdShouldPass("odo", append(args, "list", "--app", appName, "--project", commonVar.Project)...)
			Expect(componentList).NotTo(ContainSubstring(componentName))
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).NotTo(ContainElement(".odo"))
		})

		It("creates a local python component, pushes it and then deletes it using --all flag", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "python", componentName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), componentName)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)
			helper.CmdShouldPass("odo", append(args, "delete", "--context", commonVar.Context, "-f")...)
			helper.CmdShouldPass("odo", append(args, "delete", "--all", "-f", "--context", commonVar.Context)...)
			componentList := helper.CmdShouldPass("odo", append(args, "list", "--app", appName, "--project", commonVar.Project)...)
			Expect(componentList).NotTo(ContainSubstring(componentName))
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).NotTo(ContainElement(".odo"))
		})

		It("creates a local python component, pushes it and then deletes it using --all flag in local directory", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "python", componentName, "--app", appName, "--project", commonVar.Project)...)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), componentName)
			helper.CmdShouldPass("odo", append(args, "push")...)
			helper.CmdShouldPass("odo", append(args, "delete", "--all", "-f")...)
			componentList := helper.CmdShouldPass("odo", append(args, "list", "--app", appName, "--project", commonVar.Project)...)
			Expect(componentList).NotTo(ContainSubstring(componentName))
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).NotTo(ContainElement(".odo"))
		})

		It("creates a local python component and check for unsupported warning", func() {
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			output := helper.CmdShouldPass("odo", append(args, "create", "--s2i", "python", componentName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			Expect(output).To(ContainSubstring("Warning: python is not fully supported by odo, and it is not guaranteed to work"))
		})

		It("creates a local nodejs component and check unsupported warning hasn't occurred", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			output := helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs:latest", componentName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			Expect(output).NotTo(ContainSubstring("Warning"))
		})

		It("creates a local java component and check unsupported warning hasn't occurred", func() {
			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)
			output := helper.CmdShouldPass("odo", append(args, "create", "--s2i", "java:latest", componentName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			Expect(output).NotTo(ContainSubstring("Warning"))
		})
	})

	// devfile doesn't support odo update command
	// Context("odo component updating", func() {

	// 	It("should be able to create a git component and update it from local to git", func() {
	// 		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
	// 		helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")...)
	// 		helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

	// 		helper.CmdShouldPass("odo", "update", "--git", "https://github.com/openshift/nodejs-ex.git", "--context", commonVar.Context)
	// 		// check the source location and type in the deployment config
	// 		getSourceLocation := oc.SourceLocationDC("cmp-git", "testing", commonVar.Project)
	// 		Expect(getSourceLocation).To(ContainSubstring("https://github.com/openshift/nodejs-ex"))
	// 		getSourceType := oc.SourceTypeDC("cmp-git", "testing", commonVar.Project)
	// 		Expect(getSourceType).To(ContainSubstring("git"))
	// 	})

	// 	It("should be able to update a component from git to local", func() {
	// 		helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "cmp-git", "--project", commonVar.Project, "--git", "https://github.com/openshift/nodejs-ex", "--context", commonVar.Context, "--app", "testing")...)
	// 		helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

	// 		// update the component config according to the git component
	// 		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

	// 		helper.CmdShouldPass("odo", "update", "--local", "./", "--context", commonVar.Context)

	// 		// check the source location and type in the deployment config
	// 		getSourceLocation := oc.SourceLocationDC("cmp-git", "testing", commonVar.Project)
	// 		Expect(getSourceLocation).To(ContainSubstring(""))
	// 		getSourceType := oc.SourceTypeDC("cmp-git", "testing", commonVar.Project)
	// 		Expect(getSourceType).To(ContainSubstring("local"))
	// 	})
	// })

	Context("odo component delete, list and describe", func() {
		appName := "app"
		cmpName := "nodejs"

		It("should pass inside a odo directory without component name as parameter", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			helper.CmdShouldPass("odo", "url", "create", "example", "--context", commonVar.Context)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			// changing directory to the context directory
			helper.Chdir(commonVar.Context)
			cmpListOutput := helper.CmdShouldPass("odo", append(args, "list")...)
			Expect(cmpListOutput).To(ContainSubstring(cmpName))
			cmpDescribe := helper.CmdShouldPass("odo", append(args, "describe")...)
			helper.MatchAllInOutput(cmpDescribe, []string{cmpName, "nodejs"})

			url := helper.DetermineRouteURL(commonVar.Context)
			Expect(cmpDescribe).To(ContainSubstring(url))

			helper.CmdShouldPass("odo", append(args, "delete", "-f")...)
		})

		It("should fail outside a odo directory without component name as parameter", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			// commands should fail as the component name is missing
			helper.CmdShouldFail("odo", append(args, "describe", "--app", appName, "--project", commonVar.Project)...)
			helper.CmdShouldFail("odo", append(args, "delete", "-f", "--app", appName, "--project", commonVar.Project)...)
		})

		// issue https://github.com/openshift/odo/issues/4451
		// It("should pass outside a odo directory with component name as parameter", func() {
		// 	helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		// 	helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
		// 	info := helper.LocalEnvInfo(commonVar.Context)
		// 	Expect(info.GetApplication(), appName)
		// 	Expect(info.GetName(), cmpName)
		// 	helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

		// 	cmpListOutput := helper.CmdShouldPass("odo", append(args, "list", "--app", appName, "--project", commonVar.Project)...)
		// 	Expect(cmpListOutput).To(ContainSubstring(cmpName))

		// 	actualDesCompJSON := helper.CmdShouldPass("odo", append(args, "describe", cmpName, "--app", appName, "--project", commonVar.Project, "-o", "json")...)
		// 	valuesDescCJ := gjson.GetMany(actualDesCompJSON, "kind", "metadata.name", "spec.app", "spec.type", "status.state")
		// 	expectedDescCJ := []string{"Component", "nodejs", "app", "nodejs", "Pushed"}
		// 	Expect(helper.GjsonMatcher(valuesDescCJ, expectedDescCJ)).To(Equal(true))

		// 	helper.CmdShouldPass("odo", append(args, "delete", cmpName, "--app", appName, "--project", commonVar.Project, "-f")...)
		// })
	})

	Context("when running odo push multiple times, check for existence of environment variables", func() {
		It("should should retain the same environment variable on multiple push", func() {
			componentName := helper.RandString(6)
			appName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", componentName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			helper.Chdir(commonVar.Context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "FOO=BAR")
			helper.CmdShouldPass("odo", append(args, "push")...)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), componentName)
			envVars := oc.GetEnvsDevFileDeployment(componentName, commonVar.Project)
			val, ok := envVars["FOO"]
			Expect(ok).To(BeTrue())
			Expect(val).To(Equal("BAR"))
		})
	})

	Context("Creating component with numeric named context", func() {
		var contextNumeric string
		JustBeforeEach(func() {
			var err error
			ts := time.Now().UnixNano()
			contextNumeric, err = ioutil.TempDir("", fmt.Sprint(ts))
			Expect(err).ToNot(HaveOccurred())
		})
		JustAfterEach(func() {
			helper.DeleteDir(contextNumeric)
		})

		It("should create default named component in a directory with numeric name", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), contextNumeric)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", contextNumeric, "--app", "testing")...)
			info := helper.LocalEnvInfo(contextNumeric)
			Expect(info.GetApplication(), "testing")
			helper.CmdShouldPass("odo", append(args, "push", "--context", contextNumeric, "-v4")...)
		})
	})

	Context("Creating component using symlink", func() {
		var symLinkPath string

		JustBeforeEach(func() {
			if runtime.GOOS == "windows" {
				Skip("Skipping test because for symlink creation on platform like Windows, go library needs elevated privileges.")
			}
			// create a symlink
			symLinkName := helper.RandString(10)
			helper.CreateSymLink(commonVar.Context, filepath.Join(filepath.Dir(commonVar.Context), symLinkName))
			symLinkPath = filepath.Join(filepath.Dir(commonVar.Context), symLinkName)
		})
		JustAfterEach(func() {
			// remove the symlink
			err := os.Remove(symLinkPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should be able to deploy a spring boot uberjar file using symlinks in all odo commands", func() {
			oc.ImportJavaIS(commonVar.Project)

			helper.CopyExample(filepath.Join("binary", "java", "openjdk"), commonVar.Context)

			// create the component using symlink
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "java:8", "sb-jar-test", "--project",
				commonVar.Project, "--binary", filepath.Join(symLinkPath, "sb.jar"), "--context", symLinkPath)...)

			// Create a URL and push without using the symlink
			helper.CmdShouldPass("odo", "url", "create", "uberjaropenjdk", "--port", "8080", "--context", symLinkPath)
			info := helper.LocalEnvInfo(symLinkPath)
			Expect(info.GetApplication(), "app")
			Expect(info.GetName(), "sb-jar-test")

			helper.CmdShouldPass("odo", append(args, "push", "--context", symLinkPath)...)
			routeURL := helper.DetermineRouteURL(symLinkPath)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "HTTP Booster", 300, 1)

			// Delete the component
			helper.CmdShouldPass("odo", append(args, "delete", "sb-jar-test", "-f", "--context", symLinkPath)...)
		})

		It("Should be able to deploy a wildfly war file using symlinks in some odo commands", func() {
			helper.CopyExample(filepath.Join("binary", "java", "wildfly"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "wildfly", "javaee-war-test", "--project",
				commonVar.Project, "--binary", filepath.Join(symLinkPath, "ROOT.war"), "--context", symLinkPath)...)

			// Create a URL
			helper.CmdShouldPass("odo", "url", "create", "warfile", "--port", "8080", "--context", commonVar.Context)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), "app")
			Expect(info.GetName(), "javaee-war-test")
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)
			routeURL := helper.DetermineRouteURL(commonVar.Context)

			// Ping said URL
			helper.HttpWaitFor(routeURL, "Sample", 90, 1)

			// Delete the component
			helper.CmdShouldPass("odo", append(args, "delete", "javaee-war-test", "-f", "--context", commonVar.Context)...)
		})
	})

	Context("odo component delete should clean owned resources", func() {
		appName := helper.RandString(5)
		cmpName := helper.RandString(5)

		It("should delete the component and the owned resources", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			helper.CmdShouldPass("odo", "url", "create", "example-1", "--context", commonVar.Context)

			helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			helper.CmdShouldPass("odo", "url", "create", "example-2", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-2", "--size", "1Gi", "--path", "/data2", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			helper.CmdShouldPass("odo", append(args, "delete", "-f", "--context", commonVar.Context)...)

			oc.WaitAndCheckForExistence("routes", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("dc", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("pvc", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("bc", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("is", commonVar.Project, 1)
			oc.WaitAndCheckForExistence("service", commonVar.Project, 1)
		})

		It("should delete the component and the owned resources with wait flag", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "create", "--s2i", "nodejs", cmpName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)...)
			helper.CmdShouldPass("odo", "url", "create", "example-1", "--context", commonVar.Context)

			helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context)
			info := helper.LocalEnvInfo(commonVar.Context)
			Expect(info.GetApplication(), appName)
			Expect(info.GetName(), cmpName)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			helper.CmdShouldPass("odo", "url", "create", "example-2", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-2", "--size", "1Gi", "--path", "/data2", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", append(args, "push", "--context", commonVar.Context)...)

			// delete with --wait flag
			helper.CmdShouldPass("odo", append(args, "delete", "-f", "-w", "--context", commonVar.Context)...)

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

	Context("convert s2i to devfile", func() {

		JustBeforeEach(func() {
			os.Setenv("ODO_EXPERIMENTAL", "true")
		})

		JustAfterEach(func() {
			os.Unsetenv("ODO_EXPERIMENTAL")
		})

		It("should convert s2i component to devfile component successfully", func() {
			cmpName := "mynodejs"
			appName := "app"
			urlName := "url1"
			storageName := "storage1"

			// create a s2i component using --s2i that generates a devfile
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
			helper.CmdShouldPass("odo", "url", "create", urlName, "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "storage", "create", storageName, "--path", "/data1", "--size", "1Gi", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// check the status of devfile component
			stdout := helper.CmdShouldPass("odo", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{cmpName, "Devfile Components", "Pushed"})

			// verify the url
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)

			helper.MatchAllInOutput(stdout, []string{urlName, "Pushed", "false", "route"})
			//verify storage
			stdout = helper.CmdShouldPass("odo", "storage", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{storageName, "Pushed"})

		})
	})

	Context("when components are not created/managed by odo", func() {
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

		JustBeforeEach(func() {
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
		JustAfterEach(func() {
			// Relying on the project deletion to delete resources not managed by odo
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
			output := helper.CmdShouldPass("odo", append(args, "list")...)
			verifyListOutput(output, compList)
		})

		Context("The component has a different app name than the default 'app'", func() {
			JustBeforeEach(func() {
				// Create resources that are not managed by odo
				dfile := filepath.Join(commonVar.Context, "deployment-httpd-label.yaml")
				helper.CopyManifestFile("deployment-httpd-label.yaml", dfile)
				runner.Run("apply", "-f", dfile).Wait()
			})
			JustAfterEach(func() {
				// relying on project deletion to delete the resources not managed by odo
			})

			It("should list the components with --all-apps flag", func() {
				output := helper.CmdShouldPass("odo", append(args, "list", "--all-apps")...)
				verifyListOutput(output, append(compList, compStruct{"httpd", "example-deployment-httpd"}))
			})

			It("should list the components with --app flag", func() {
				output := helper.CmdShouldPass("odo", append(args, "list", "--app", "httpd")...)
				verifyListOutput(output, []compStruct{{"httpd", "example-deployment-httpd"}})
			})
		})

		It("should list the components in json format with -o json flag", func() {
			output := helper.CmdShouldPass("odo", append(args, "list", "--project", commonVar.Project, "-o", "json")...)
			for _, comp := range compList {
				valuesCList := gjson.GetMany(output, "kind", "otherComponents.#.kind", "otherComponents.#.metadata.name", "otherComponents.#.spec.app")
				expectedCList := []string{"List", "Component", comp.Name, comp.App}
				Expect(helper.GjsonMatcher(valuesCList, expectedCList)).To(Equal(true))
			}
		})

		When("executing odo list from other project", func() {
			var otherProject string
			JustBeforeEach(func() {
				otherProject = commonVar.CliRunner.CreateRandNamespaceProject()
			})
			JustAfterEach(func() {
				helper.CmdShouldPass("odo", "project", "set", commonVar.Project)
				commonVar.CliRunner.DeleteNamespaceProject(otherProject)
			})

			It("should list the components with --project flag", func() {
				output := helper.CmdShouldPass("odo", append(args, "list", "--project", commonVar.Project)...)
				verifyListOutput(output, compList)
			})

		})
	})

}
