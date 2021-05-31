package project

import (
	. "github.com/onsi/ginkgo"
)

func ProjectTestScenario() {
	When("Executing Machine readable output tests", func() {
		It("Help for odo project list should contain machine output", HelpCommand)
		It("should be able to get project", GetProject)
		It("should display the help", HelpCommandProject)
		It("should display only the project name", CommandWithQFlag)
		It("--wait should work with deleting a project", CommandWaitFlag)
		It("should be able to delete project and show output in json format", CommandDeleteWithJsonFlag)
	})

	When("when running project command app parameter in directory that doesn't contain .odo config directory", func() {
		It("should successfully execute list along with machine readable output", CommandWithApp)
	})
}
