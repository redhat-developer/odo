package project

import (
	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var projectName string

func ProjectTestScenario() {
	When("Machine readable output tests", func() {
		It("Help for odo project list should contain machine output", HelpCommand)
		It("should be able to get project", GetProject)
	})
	When("when running help for project command", func() {
		It("should display the help", HelpCommandProject)
	})
	When("when running get command with -q flag", func() {
		It("should display only the project name", CommandWithQFlag)
	})

	When("Should be able to delete a project with --wait", func() {

		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		It("--wait should work with deleting a project", CommandWaitFlag)
	})

	When("Delete the project with flag -o json", func() {

		JustBeforeEach(func() {
			projectName = helper.RandString(6)
		})
		It("should be able to delete project and show output in json format", CommandDeleteWithJsonFlag)
	})

	When("when running project command app parameter in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute list along with machine readable output", CommandWithApp)
	})
}
