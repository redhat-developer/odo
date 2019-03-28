package helper

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
)

// CreateRandProject create new project with random name (10 letters)
// without writing to the config file (without switching project)
func OcCreateRandProject() string {
	projectName := randString(10)
	fmt.Fprintf(GinkgoWriter, "Creating a new project: %s\n", projectName)
	CmdShouldPass(fmt.Sprintf("oc new-project %s --skip-config-write", projectName))
	return projectName
}

// OcSwitchProject switch to the project
func OcSwitchProject(project string) {
	CmdShouldPass(fmt.Sprintf("oc project %s ", project))
}

// DeleteProject deletes a specified project
func OcDeleteProject(project string) {
	fmt.Fprintf(GinkgoWriter, "Deleting project: %s\n", project)
	CmdShouldPass(fmt.Sprintf("oc delete project %s --now", project))
}

// OcCurrentProject get currently active project in oc
// returns empty string if there no active project, or no access to the project
func OcCurrentProject() string {
	stdout, _, exitCode := cmdRunner("oc project -q")
	if exitCode == 0 {
		return stdout
	}
	return ""
}
