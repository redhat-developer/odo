package project

import (
	"strings"

	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

func HelpCommand() {
	output := helper.CmdShouldPass("odo", "project", "list", "--help")
	Expect(output).To(ContainSubstring("Specify output format, supported format: json"))
}

func GetProject() {
	projectGetJSON := helper.CmdShouldPass("odo", "project", "get", "-o", "json")
	getOutputJSON, err := helper.Unindented(projectGetJSON)
	Expect(err).Should(BeNil())
	valuesJSON := gjson.GetMany(getOutputJSON, "kind", "status.active")
	expectedJSON := []string{"Project", "true"}
	Expect(helper.GjsonMatcher(valuesJSON, expectedJSON)).To(Equal(true))
}

func HelpCommandProject() {
	projectHelp := helper.CmdShouldPass("odo", "project", "-h")
	Expect(projectHelp).To(ContainSubstring("Perform project operations"))
}

func CommandWithQFlag() {
	projectName := helper.CmdShouldPass("odo", "project", "get", "-q")
	Expect(projectName).Should(ContainSubstring(commonVar.Project))
}

func CommandWaitFlag() {
	projectName := helper.RandString(6)
	// Create the project
	helper.CmdShouldPass("odo", "project", "create", projectName)

	// Delete with --wait
	output := helper.CmdShouldPass("odo", "project", "delete", projectName, "-f", "--wait")
	Expect(output).To(ContainSubstring("Waiting for project to be deleted"))
}

func CommandDeleteWithJsonFlag() {
	projectName := helper.RandString(6)
	helper.CmdShouldPass("odo", "project", "create", projectName, "-o", "json")

	actual := helper.CmdShouldPass("odo", "project", "delete", projectName, "-o", "json")
	values := gjson.GetMany(actual, "kind", "message")
	expected := []string{"Project", "Deleted project :"}
	Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
}

func CommandWithApp() {
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
}
