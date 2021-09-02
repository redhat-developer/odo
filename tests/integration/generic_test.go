package integration

import (
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo generic", func() {
	var testPHPGitURL = "https://github.com/appuio/example-php-sti-helloworld"
	var testNodejsGitURL = "https://github.com/sclorg/nodejs-ex"
	var testLongURLName = "long-url-name-long-url-name-long-url-name-long-url-name-long-url-name"

	// TODO: A neater way to provide odo path. Currently we assume \
	// odo and oc in $PATH already
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

	Context("Check the help usage for odo", func() {

		It("Makes sure that we have the long-description when running odo and we dont error", func() {
			output := helper.Cmd("odo").ShouldPass().Out()
			Expect(output).To(ContainSubstring("To see a full list of commands, run 'odo --help'"))
		})

		It("Make sure we have the full description when performing odo --help", func() {
			output := helper.Cmd("odo", "--help").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Use \"odo [command] --help\" for more information about a command."))
		})

		It("Fail when entering an incorrect name for a component", func() {
			output := helper.Cmd("odo", "component", "foobar").ShouldFail().Err()
			Expect(output).To(ContainSubstring("Subcommand not found, use one of the available commands"))
		})

		It("Fail with showing help only once for incorrect command", func() {
			output := helper.Cmd("odo", "hello").ShouldFail().Err()
			Expect(strings.Count(output, "odo [flags]")).Should(Equal(1))
		})

	})

	Context("When executing catalog list without component directory", func() {
		It("should list all component catalogs", func() {
			stdOut := helper.Cmd("odo", "catalog", "list", "components").ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{"dotnet", "nginx", "php", "ruby", "wildfly"})
		})

	})

	Context("check catalog component search functionality", func() {
		It("check that a component does not exist", func() {
			componentRandomName := helper.RandString(7)
			output := helper.Cmd("odo", "catalog", "search", "component", componentRandomName).ShouldFail().Err()
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
			actual := helper.Cmd("odo", "project", "create", projectName, "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(actual, "kind", "metadata.name", "status.active")
			expected := []string{"Project", projectName, "true"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
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
			helper.Cmd("odo", "project", "create", projectName).ShouldPass()
			actual := helper.Cmd("odo", "project", "create", projectName, "-o", "json").ShouldFail().Err()
			valuesC := gjson.GetMany(actual, "kind", "message")
			expectedC := []string{"Error", "unable to create new project"}
			Expect(helper.GjsonMatcher(valuesC, expectedC)).To(Equal(true))

		})
	})

	Context("Delete the project with flag -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})

		// odo project delete foobar -o json
		It("should be able to delete project and show output in json format", func() {
			helper.Cmd("odo", "project", "create", projectName, "-o", "json").ShouldPass()

			actual := helper.Cmd("odo", "project", "delete", projectName, "-o", "json").ShouldPass().Out()
			valuesDel := gjson.GetMany(actual, "kind", "message")
			expectedDel := []string{"Status", "Deleted project"}
			Expect(helper.GjsonMatcher(valuesDel, expectedDel)).To(Equal(true))

		})
	})

	Context("creating component with an application and url", func() {
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})
		It("should create the component in default application", func() {
			helper.Cmd("odo", "create", "--s2i", "php", "testcmp", "--app", "e2e-xyzk", "--project", commonVar.Project, "--git", testPHPGitURL).ShouldPass()
			helper.Cmd("odo", "config", "set", "Ports", "8080/TCP", "-f").ShouldPass()
			helper.Cmd("odo", "push").ShouldPass()
			oc.VerifyCmpName("testcmp", commonVar.Project)
			oc.VerifyAppNameOfComponent("testcmp", "e2e-xyzk", commonVar.Project)
			helper.Cmd("odo", "app", "delete", "e2e-xyzk", "-f").ShouldPass()
		})
	})

	Context("Overwriting build timeout for git component", func() {
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})

		It("should pass to build component if the given build timeout is more than the default(300s) value", func() {
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--git", testNodejsGitURL).ShouldPass()
			helper.Cmd("odo", "preference", "set", "BuildTimeout", "600").ShouldPass()
			buildTimeout := helper.GetPreferenceValue("BuildTimeout")
			helper.MatchAllInOutput(buildTimeout, []string{"600"})
			helper.Cmd("odo", "push").ShouldPass()
		})

		It("should fail to build component if the given build timeout is pretty less(2s)", func() {
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--git", testNodejsGitURL).ShouldPass()
			helper.Cmd("odo", "preference", "set", "BuildTimeout", "2").ShouldPass()
			buildTimeout := helper.GetPreferenceValue("BuildTimeout")
			helper.MatchAllInOutput(buildTimeout, []string{"2"})
			stdOut := helper.Cmd("odo", "push").ShouldFail().Err()
			helper.MatchAllInOutput(stdOut, []string{"Failed to create component", "timeout waiting for build"})
		})
	})

	Context("should list applications in other project", func() {
		It("should be able to create a php component with application created", func() {
			helper.Cmd("odo", "create", "--s2i", "php", "testcmp", "--app", "testing", "--project", commonVar.Project, "--ref", "master", "--git", testPHPGitURL, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			currentProject := helper.CreateRandProject()
			currentAppNames := helper.Cmd("odo", "app", "list", "--project", currentProject).ShouldPass().Out()
			Expect(currentAppNames).To(ContainSubstring("There are no applications deployed in the project '" + currentProject + "'"))
			appNames := helper.Cmd("odo", "app", "list", "--project", commonVar.Project).ShouldPass().Out()
			Expect(appNames).To(ContainSubstring("testing"))
			helper.DeleteProject(currentProject)
		})
	})

	Context("when running odo push with flag --show-log", func() {
		// works
		It("should be able to push changes", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()

			// Push the changes with --show-log
			getLogging := helper.Cmd("odo", "push", "--show-log", "--context", commonVar.Context).ShouldPass().Out()
			Expect(getLogging).To(ContainSubstring("Creating Kubernetes resources for component nodejs"))
		})
	})

	Context("deploying a component with a specific image name", func() {
		It("should deploy the component", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs:latest", "testversioncmp", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "delete", "-f", "--context", commonVar.Context).ShouldPass()
		})
	})

	Context("When deleting two project one after the other", func() {
		It("should be able to delete sequentially", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()

			helper.DeleteProject(project2)
			helper.DeleteProject(project1)
		})
	})

	Context("When deleting three project one after the other in opposite order", func() {
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
			odoVersion := helper.Cmd("odo", "version").ShouldPass().Out()
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
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})
		It("should not allow creating a URL with long name", func() {
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--project", commonVar.Project).ShouldPass()
			stdOut := helper.Cmd("odo", "url", "create", testLongURLName, "--port", "8080").ShouldFail().Err()
			Expect(stdOut).To(ContainSubstring("must be shorter than 63 characters"))
		})
	})

	Context("When using cpu or memory flag with odo create", func() {
		cmpName := "nodejs"

		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
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
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				output := helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--"+testCase.paramName, testCase.paramValue, "--git", "https://github.com/sclorg/nodejs-ex.git").ShouldFail().Err()
				Expect(output).To(ContainSubstring("unknown flag: --" + testCase.paramName))
			}
		})
	})

})
