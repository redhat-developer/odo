package helper

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

type OcRunner struct {
	// path to oc binary
	path string
}

// New initializes new OcRunner
func NewOcRunner(ocPath string) OcRunner {
	return OcRunner{
		path: ocPath,
	}
}

// Run oc with given arguments
func (oc *OcRunner) Run(args ...string) *gexec.Session {
	session := CmdRunner(oc.path, args...)
	Eventually(session).Should(gexec.Exit(0))
	return session
}

// CreateRandProject create new project with random name (10 letters)
// without writing to the config file (without switching project)
func (oc *OcRunner) CreateRandProject() string {
	projectName := randString(10)
	fmt.Fprintf(GinkgoWriter, "Creating a new project: %s\n", projectName)
	session := CmdRunner(oc.path, "new-project", projectName, "--skip-config-write")

	Eventually(session).Should(gexec.Exit(0))
	Eventually(session).Should(gbytes.Say("created on server"))

	return projectName
}

// SwitchProject switch to the project
func (oc *OcRunner) SwitchProject(projectName string) {
	session := CmdRunner(oc.path, "project", projectName)
	Eventually(session).Should(gexec.Exit(0))
}

// DeleteProject deletes a specified project
func (oc *OcRunner) DeleteProject(projectName string) {
	fmt.Fprintf(GinkgoWriter, "Deleting project: %s\n", projectName)
	session := CmdRunner(oc.path, "delete", "project", projectName, "--now")
	Eventually(session).Should(gexec.Exit(0))
}

// GetCurrentProject get currently active project in oc
// returns empty string if there no active project, or no access to the project
func (oc *OcRunner) GetCurrentProject() string {
	session := CmdRunner(oc.path, "project", "-q")
	session.Wait()
	if session.ExitCode() == 0 {
		return strings.TrimSpace(string(session.Out.Contents()))
	}
	return ""
}

// GetFirstURL returns the url of the first Route that it can find for given component
func (oc *OcRunner) GetFirstURL(component string, app string, project string) string {
	session := CmdRunner(oc.path, "get", "route",
		"-n", project,
		"-l", "app.kubernetes.io/instance="+component,
		"-l", "app.kubernetes.io/part-of="+app,
		"-o", "jsonpath={.items[0].spec.host}")

	session.Wait()
	if session.ExitCode() == 0 {
		return string(session.Out.Contents())
	}
	return ""
}

// GetComponentRoute run command to get the Routes in yaml format for given component
func (oc *OcRunner) GetComponentRoutes(component string, app string, project string) string {
	session := CmdRunner(oc.path, "get", "route",
		"-n", project,
		"-l", "app.kubernetes.io/instance="+component,
		"-l", "app.kubernetes.io/part-of="+app,
		"-o", "yaml")

	Eventually(session).Should(gexec.Exit(0))

	return string(session.Wait().Out.Contents())
}

// GetComponentDC run command to get the DeploymentConfig in yaml format for given component
func (oc *OcRunner) GetComponentDC(component string, app string, project string) string {
	session := CmdRunner(oc.path, "get", "dc",
		"-n", project,
		"-l", "app.kubernetes.io/instance="+component,
		"-l", "app.kubernetes.io/part-of="+app,
		"-o", "yaml")

	Eventually(session).Should(gexec.Exit(0))

	return string(session.Wait().Out.Contents())
}
