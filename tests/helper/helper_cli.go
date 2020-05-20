package helper

import "github.com/onsi/gomega/gexec"

// CliRunner requires functions which are common for oc, kubectl and docker
// By abstracting these functions into an interface, it handles the cli runner and calls the functions specified to particular cluster only
type CliRunner interface {
	Run(cmd string) *gexec.Session
	ExecListDir(podName string, projectName string, dir string) string
	CheckCmdOpInRemoteDevfilePod(podName string, containerName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool
	GetRunningPodNameByComponent(compName string, namespace string) string
	GetVolumeMountNamesandPathsFromContainer(deployName string, containerName, namespace string) string
	WaitAndCheckForExistence(resourceType, namespace string, timeoutMinutes int) bool
	GetServices(namespace string) string
}

// NewCliRunner initializes new NewCliRunner with the appropriate platform
func NewCliRunner(cli string) CliRunner {
	var clirunner CliRunner
	switch cli {
	case "oc":
		clirunner = OcRunner{
			path: cli,
		}
	}

	return clirunner
}
