package integration

import (
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/v2/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo generic", func() {
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
			helper.MatchAllInOutput(stdOut, []string{"nodejs", "python", "php", "go", "java"})
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

})
