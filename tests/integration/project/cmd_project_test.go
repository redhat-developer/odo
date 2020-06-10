package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo project command tests", func() {
	var project string
	var context string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Machine readable output tests", func() {

		It("Help for odo project list should contain machine output", func() {
			output := helper.CmdShouldPass("odo", "project", "list", "--help")
			Expect(output).To(ContainSubstring("Specify output format, supported format: json"))
		})

	})

	Context("when running help for project command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "project", "-h")
			Expect(appHelp).To(ContainSubstring("Perform project operations"))
		})
	})

	Context("Should be able to delete a project with --wait", func() {
		var projectName string
		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})

		It("--wait should work with deleting a project", func() {

			// Create the project
			helper.CmdShouldPass("odo", "project", "create", projectName)

			// Delete with --wait
			output := helper.CmdShouldPass("odo", "project", "delete", projectName, "-f", "--wait")
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
			helper.CmdShouldPass("odo", "project", "create", projectName, "-o", "json")

			actual := helper.CmdShouldPass("odo", "project", "delete", projectName, "-o", "json")
			desired := fmt.Sprintf(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","namespace":"%s","creationTimestamp":null},"message":"Deleted project : %s"}`, projectName, projectName, projectName)
			Expect(desired).Should(MatchJSON(actual))
		})
	})

	Context("when running project command app parameter in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute list along with machine readable output", func() {

			helper.WaitForCmdOut("odo", []string{"project", "list"}, 1, true, func(output string) bool {
				return strings.Contains(output, project)
			})

			// project deletion doesn't happen immediately and older projects still might exist
			// so we test subset of the string
			expected, err := helper.Unindented(`{"kind":"Project","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"` + project + `","namespace":"` + project + `","creationTimestamp":null},"spec":{},"status":{"active":true}}`)
			Expect(err).Should(BeNil())

			helper.WaitForCmdOut("odo", []string{"project", "list", "-o", "json"}, 1, true, func(output string) bool {
				listOutputJSON, err := helper.Unindented(output)
				Expect(err).Should(BeNil())
				return strings.Contains(listOutputJSON, expected)
			})
		})
	})
})
