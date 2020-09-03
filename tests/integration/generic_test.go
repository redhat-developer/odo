package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo generic", func() {
	// TODO: A neater way to provide odo path. Currently we assume \
	// odo and oc in $PATH already.
	var project string
	var context string
	var originalDir string
	var oc helper.OcRunner
	var testPHPGitURL = "https://github.com/appuio/example-php-sti-helloworld"
	var testNodejsGitURL = "https://github.com/sclorg/nodejs-ex"
	var testLongURLName = "long-url-name-long-url-name-long-url-name-long-url-name-long-url-name"

	BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
	})

	Context("Check the help usage for odo", func() {

		It("Makes sure that we have the long-description when running odo and we dont error", func() {
			output := helper.CmdShouldPass("odo")
			Expect(output).To(ContainSubstring("To see a full list of commands, run 'odo --help'"))
		})

		It("Make sure we have the full description when performing odo --help", func() {
			output := helper.CmdShouldPass("odo", "--help")
			Expect(output).To(ContainSubstring("Use \"odo [command] --help\" for more information about a command."))
		})

		It("Fail when entering an incorrect name for a component", func() {
			output := helper.CmdShouldFail("odo", "component", "foobar")
			Expect(output).To(ContainSubstring("Subcommand not found, use one of the available commands"))
		})

		It("Fail with showing help only once for incorrect command", func() {
			output := helper.CmdShouldFail("odo", "hello")
			Expect(strings.Count(output, "odo [flags]")).Should(Equal(1))
		})

	})

	Context("When executing catalog list without component directory", func() {
		It("should list all component catalogs", func() {
			stdOut := helper.CmdShouldPass("odo", "catalog", "list", "components")
			helper.MatchAllInOutput(stdOut, []string{"dotnet", "nginx", "php", "ruby", "wildfly"})
		})

	})

	Context("check catalog component search functionality", func() {
		It("check that a component does not exist", func() {
			componentRandomName := helper.RandString(7)
			output := helper.CmdShouldFail("odo", "catalog", "search", "component", componentRandomName)
			Expect(output).To(ContainSubstring("no component matched the query: " + componentRandomName))
		})
	})

	// Test machine readable output
	Context("when creating project -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		JustAfterEach(func() {
			helper.DeleteProject(projectName)
		})

		// odo project create foobar -o json
		It("should be able to create project and show output in json format", func() {
			actual := helper.CmdShouldPass("odo", "project", "create", projectName, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"message":"Project '%s' is ready for use"}`, projectName, projectName, projectName)
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	Context("Creating same project twice with flag -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		JustAfterEach(func() {
			helper.DeleteProject(projectName)
		})
		// odo project create foobar -o json (x2)
		It("should fail along with proper machine readable output", func() {
			helper.CmdShouldPass("odo", "project", "create", projectName)
			actual := helper.CmdShouldFail("odo", "project", "create", projectName, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Error","apiVersion":"odo.dev/v1alpha1","metadata":{"creationTimestamp":null},"message":"unable to create new project: unable to create new project %s: project.project.openshift.io \"%s\" already exists"}`, projectName, projectName)
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	Context("Delete the project with flag -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})

		// odo project delete foobar -o json
		It("should be able to delete project and show output in json format", func() {
			helper.CmdShouldPass("odo", "project", "create", projectName, "-o", "json")

			actual := helper.CmdShouldPass("odo", "project", "delete", projectName, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"message":"Deleted project : %s"}`, projectName, projectName, projectName)
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	Context("creating component with an application and url", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should create the component in default application", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "php", "testcmp", "--app", "e2e-xyzk", "--project", project, "--git", testPHPGitURL)
			helper.CmdShouldPass("odo", "config", "set", "Ports", "8080/TCP", "-f")
			helper.CmdShouldPass("odo", "push")
			oc.VerifyCmpName("testcmp", project)
			oc.VerifyAppNameOfComponent("testcmp", "e2e-xyzk", project)
			helper.CmdShouldPass("odo", "app", "delete", "e2e-xyzk", "-f")
		})
	})

	Context("Overwriting build timeout for git component", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should pass to build component if the given build timeout is more than the default(300s) value", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "nodejs", "--project", project, "--git", testNodejsGitURL)
			helper.CmdShouldPass("odo", "preference", "set", "BuildTimeout", "600")
			buildTimeout := helper.GetPreferenceValue("BuildTimeout")
			helper.MatchAllInOutput(buildTimeout, []string{"600"})
			helper.CmdShouldPass("odo", "push")
		})

		It("should fail to build component if the given build timeout is pretty less(2s)", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "nodejs", "--project", project, "--git", testNodejsGitURL)
			helper.CmdShouldPass("odo", "preference", "set", "BuildTimeout", "2")
			buildTimeout := helper.GetPreferenceValue("BuildTimeout")
			helper.MatchAllInOutput(buildTimeout, []string{"2"})
			stdOut := helper.CmdShouldFail("odo", "push")
			helper.MatchAllInOutput(stdOut, []string{"Failed to create component", "timeout waiting for build"})
		})
	})

	Context("should list applications in other project", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should be able to create a php component with application created", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "php", "testcmp", "--app", "testing", "--project", project, "--ref", "master", "--git", testPHPGitURL, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			currentProject := helper.CreateRandProject()
			currentAppNames := helper.CmdShouldPass("odo", "app", "list", "--project", currentProject)
			Expect(currentAppNames).To(ContainSubstring("There are no applications deployed in the project '" + currentProject + "'"))
			appNames := helper.CmdShouldPass("odo", "app", "list", "--project", project)
			Expect(appNames).To(ContainSubstring("testing"))
			helper.DeleteProject(currentProject)
		})
	})

	Context("when running odo push with flag --show-log", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.RemoveAll(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should be able to push changes", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "nodejs", "--project", project, "--context", context)

			// Push the changes with --show-log
			getLogging := helper.CmdShouldPass("odo", "push", "--show-log", "--context", context)
			Expect(getLogging).To(ContainSubstring("Building component"))
		})
	})

	Context("deploying a component with a specific image name", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.RemoveAll(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should deploy the component", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs:latest", "testversioncmp", "--project", project, "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "delete", "-f", "--context", context+"/nodejs-ex")
		})
	})

	Context("When deleting two project one after the other", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		})

		JustAfterEach(func() {
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should be able to delete sequentially", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()

			helper.DeleteProject(project2)
			helper.DeleteProject(project1)
		})
	})

	Context("When deleting three project one after the other in opposite order", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		})

		JustAfterEach(func() {
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should be able to delete", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()
			project3 := helper.CreateRandProject()

			helper.DeleteProject(project1)
			helper.DeleteProject(project2)
			helper.DeleteProject(project3)

		})
	})

	Context("when executing odo version command", func() {
		It("should show the version of odo major components including server login URL", func() {
			odoVersion := helper.CmdShouldPass("odo", "version")
			reOdoVersion := regexp.MustCompile(`^odo\s*v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?\s*\(\w+\)`)
			odoVersionStringMatch := reOdoVersion.MatchString(odoVersion)
			rekubernetesVersion := regexp.MustCompile(`Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+((-\w+\.[0-9]+)?\+\w+)?`)
			kubernetesVersionStringMatch := rekubernetesVersion.MatchString(odoVersion)
			Expect(odoVersionStringMatch).Should(BeTrue())
			Expect(kubernetesVersionStringMatch).Should(BeTrue())
			serverURL := oc.GetCurrentServerURL()
			Expect(odoVersion).Should(ContainSubstring("Server: " + serverURL))
		})
	})

	Context("prevent the user from creating invalid URLs", func() {
		var originalDir string

		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should not allow creating a URL with long name", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", project)
			stdOut := helper.CmdShouldFail("odo", "url", "create", testLongURLName, "--port", "8080")
			Expect(stdOut).To(ContainSubstring("must be shorter than 63 characters"))
		})
	})

	Context("when component's deployment config is deleted with oc", func() {
		var componentRandomName string

		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			componentRandomName = helper.RandString(6)
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
			os.RemoveAll(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should delete all OpenShift objects except the component's imagestream", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", componentRandomName, "--project", project)
			helper.CmdShouldPass("odo", "push")

			// Delete the deployment config using oc delete
			dc := oc.GetDcName(componentRandomName, project)
			helper.CmdShouldPass("oc", "delete", "--wait", "dc", dc, "--namespace", project)

			// insert sleep because it takes a few seconds to delete *all*
			// objects owned by DC but we should be able to check if a service
			// got deleted in a second.
			time.Sleep(1 * time.Second)

			// now check if the service owned by the DC exists. Service name is
			// same as DC name for a given component.
			stdOut := helper.CmdShouldFail("oc", "get", "svc", dc, "--namespace", project)
			Expect(stdOut).To(ContainSubstring("NotFound"))

			// ensure that the image stream still exists
			helper.CmdShouldPass("oc", "get", "is", dc, "--namespace", project)
		})
	})

	Context("When using cpu or memory flag with odo create", func() {
		var originalDir string
		cmpName := "nodejs"

		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		JustAfterEach(func() {
			helper.DeleteProject(project)
			helper.Chdir(originalDir)
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})

		It("should not allow using any memory or cpu flag", func() {

			cases := []struct {
				paramName  string
				paramValue string
			}{
				{
					paramName:  "cpu",
					paramValue: "0.4",
				},
				{
					paramName:  "mincpu",
					paramValue: "0.2",
				},
				{
					paramName:  "maxcpu",
					paramValue: "0.4",
				},
				{
					paramName:  "memory",
					paramValue: "200Mi",
				},
				{
					paramName:  "minmemory",
					paramValue: "100Mi",
				},
				{
					paramName:  "maxmemory",
					paramValue: "200Mi",
				},
			}
			for _, testCase := range cases {
				helper.CopyExample(filepath.Join("source", "nodejs"), context)
				output := helper.CmdShouldFail("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", project, "--context", context, "--"+testCase.paramName, testCase.paramValue)
				Expect(output).To(ContainSubstring("unknown flag: --" + testCase.paramName))
			}
		})
	})

})
