package project

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo project command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Machine readable output tests", func() {

		It("Help for odo project list should contain machine output", func() {
			output := helper.Cmd("odo", "project", "list", "--help").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Specify output format, supported format: json"))
		})

		It("should be able to get project", func() {
			projectGetJSON := helper.Cmd("odo", "project", "get", "-o", "json").ShouldPass().Out()
			getOutputJSON, err := helper.Unindented(projectGetJSON)
			Expect(err).Should(BeNil())
			valuesJSON := gjson.GetMany(getOutputJSON, "kind", "status.active")
			expectedJSON := []string{"Project", "true"}
			Expect(helper.GjsonMatcher(valuesJSON, expectedJSON)).To(Equal(true))

		})

	})

	Context("when running help for project command", func() {
		It("should display the help", func() {
			projectHelp := helper.Cmd("odo", "project", "-h").ShouldPass().Out()
			Expect(projectHelp).To(ContainSubstring("Perform project operations"))
		})
	})

	Context("when running get command with -q flag", func() {
		It("should display only the project name", func() {
			projectName := helper.Cmd("odo", "project", "get", "-q").ShouldPass().Out()
			Expect(projectName).Should(ContainSubstring(commonVar.Project))
		})
	})

	// Uncomment via https://github.com/openshift/odo/issues/2117 fix
	// Context("odo machine readable output on empty project", func() {
	// 	It("should be able to list current project", func() {
	// 		projectListJSON := helper.CmdShouldPass("odo", "project", "list", "-o", "json")
	// 		listOutputJSON, err := helper.Unindented(projectListJSON)
	// 		Expect(err).Should(BeNil())
	// 		partOfProjectListJSON, err := helper.Unindented(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"` + project + `","namespace":"` + project + `","creationTimestamp":null},"spec":{},"status":{"active":true}}`)
	// 		Expect(err).Should(BeNil())
	// 		Expect(listOutputJSON).To(ContainSubstring(partOfProjectListJSON))
	// 	})
	// })

	Context("Should be able to delete a project with --wait", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})

		It("--wait should work with deleting a project", func() {

			// Create the project
			helper.Cmd("odo", "project", "create", projectName).ShouldPass()

			// Delete with --wait
			output := helper.Cmd("odo", "project", "delete", projectName, "-f", "--wait").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Waiting for project to be deleted"))

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
			values := gjson.GetMany(actual, "kind", "message")
			expected := []string{"Project", "Deleted project :"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))

		})
	})

	Context("when running project command app parameter in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute list along with machine readable output", func() {

			helper.WaitForCmdOut("odo", []string{"project", "list"}, 1, true, func(output string) bool {
				return strings.Contains(output, commonVar.Project)
			})

			// project deletion doesn't happen immediately and older projects still might exist
			// so we test subset of the string
			expected, err := helper.Unindented(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"` + commonVar.Project + `","namespace":"` + commonVar.Project + `","creationTimestamp":null},"spec":{},"status":{"active":true}}`)
			Expect(err).Should(BeNil())

			helper.WaitForCmdOut("odo", []string{"project", "list", "-o", "json"}, 1, true, func(output string) bool {
				listOutputJSON, err := helper.Unindented(output)
				Expect(err).Should(BeNil())
				return strings.Contains(listOutputJSON, expected)
			})
		})
	})
})
