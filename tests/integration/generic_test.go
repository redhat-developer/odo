package integration

import (
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
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

	When("running odo --help", func() {
		It("retuns full help contents including usage, examples, commands, utility commands, component shortcuts, and flags sections", func() {
			output := helper.Cmd("odo", "--help").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Usage:"))
			Expect(output).To(ContainSubstring("Examples:"))
			Expect(output).To(ContainSubstring("Commands:"))
			Expect(output).To(ContainSubstring("Utility Commands:"))
			Expect(output).To(ContainSubstring("Component Shortcuts:"))
			Expect(output).To(ContainSubstring("Flags:"))
		})

	})

	When("running odo without subcommand and flags", func() {
		It("a short vesion of help contents is returned, an error is not expected", func() {
			output := helper.Cmd("odo").ShouldPass().Out()
			Expect(output).To(ContainSubstring("To see a full list of commands, run 'odo --help'"))
		})
	})

	When("using an invalid subcommand", func() {
		It("an error message and help including Usage, Examples, Available commands, and flags is returned", func() {
			output := helper.Cmd("odo", "hello").ShouldFail().Err()
			Expect(output).To(ContainSubstring("Usage:"))
			Expect(output).To(ContainSubstring("Examples:"))
			Expect(output).To(ContainSubstring("Commands:"))
			Expect(output).To(ContainSubstring("Flags:"))
		})
	})

	When("trying to create a component with an invalid component name", func() {
		It("Fail when entering an incorrect name for a component", func() {
			output := helper.Cmd("odo", "component", "foobar").ShouldFail().Err()
			Expect(output).To(ContainSubstring("Error: Subcommand not found, use one of the available commands:"))
			Expect(output).To(ContainSubstring("Usage:"))
			Expect(output).To(ContainSubstring("Examples:"))
			Expect(output).To(ContainSubstring("Available Commands:"))
			Expect(output).To(ContainSubstring("Flags:"))
			Expect(output).To(ContainSubstring("Additional Flags:"))
		})
	})

	When("executing catalog list without component directory", func() {
		It("should list all devfile components", func() {
			stdOut := helper.Cmd("odo", "catalog", "list", "components").ShouldPass().Out()
			helper.MatchAllInOutput(stdOut, []string{"Devfile", "nodejs", "python", "php", "go", "java"})
		})
	})

	When("searching the catalog for a non existing component", func() {
		It("searches for the component and returs message indicating that no component matched the query", func() {
			componentRandomName := helper.RandString(7)
			output := helper.Cmd("odo", "catalog", "search", "component", componentRandomName).ShouldFail().Err()
			Expect(output).To(ContainSubstring("no component matched the query: " + componentRandomName))
		})
	})

	// Test machine readable output
	When("creating an application using -o json flag", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		JustAfterEach(func() {
			helper.DeleteProject(projectName)
		})
		It("should create the application and show output in json format", func() {
			actual := helper.Cmd("odo", "project", "create", projectName, "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(actual, "kind", "metadata.name", "status.active")
			expected := []string{"Project", projectName, "true"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})
		It("should fail if trying to create twice along with proper machine readable output", func() {
			helper.Cmd("odo", "project", "create", projectName).ShouldPass()
			actual := helper.Cmd("odo", "project", "create", projectName, "-o", "json").ShouldFail().Err()
			valuesC := gjson.GetMany(actual, "kind", "message")
			expectedC := []string{"Error", "unable to create new project"}
			Expect(helper.GjsonMatcher(valuesC, expectedC)).To(Equal(true))

		})
	})

	When("deleting a project with flag -o json", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		It("should delete the project and show output in json format", func() {
			helper.Cmd("odo", "project", "create", projectName, "-o", "json").ShouldPass()
			actual := helper.Cmd("odo", "project", "delete", projectName, "-o", "json").ShouldPass().Out()
			valuesDel := gjson.GetMany(actual, "kind", "message")
			expectedDel := []string{"Status", "Deleted project"}
			Expect(helper.GjsonMatcher(valuesDel, expectedDel)).To(Equal(true))
		})

	})

	When("deleting two project one after the other", func() {
		It("should be able to delete them sequentially", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()

			helper.DeleteProject(project1)
			helper.DeleteProject(project2)
		})
	})

	When("deleting several projects", func() {
		It("should be able to delete them in any order", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()
			project3 := helper.CreateRandProject()

			helper.DeleteProject(project2)
			helper.DeleteProject(project1)
			helper.DeleteProject(project3)

		})
	})

	When("executing odo version command", func() {
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
