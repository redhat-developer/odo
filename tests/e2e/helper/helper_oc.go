package helper

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
)

const CmdWaitTimeOut time.Duration = 5 * time.Minute

// CreateRandProject create new project with random name (10 letters)
// without writing to the config file (without switching project)
func OcCreateRandProject() string {
	projectName := randString(10)
	fmt.Fprintf(GinkgoWriter, "Creating a new project: %s\n", projectName)
	CmdShouldPass(fmt.Sprintf("oc new-project %s --skip-config-write", projectName))
	waitForCmdOut("oc get projects", CmdWaitTimeOut, true, func(output string) bool {
		return strings.Contains(output, projectName)
	})
	return projectName
}

// OcSwitchProject switch to the project
func OcSwitchProject(project string) {
	CmdShouldPass(fmt.Sprintf("oc project %s ", project))
	waitForCmdOut("oc project -q", CmdWaitTimeOut, true, func(output string) bool {
		return (strings.Compare(strings.TrimSpace(output), project) == 0)
	})
}

// DeleteProject deletes a specified project
func OcDeleteProject(project string) {
	fmt.Fprintf(GinkgoWriter, "Deleting project: %s\n", project)
	CmdShouldPass(fmt.Sprintf("oc delete project %s --now", project))
	waitForCmdOut("oc get projects", CmdWaitTimeOut, true, func(output string) bool {
		return !strings.Contains(output, project)
	})
}

// OcCurrentProject get currently active project in oc
// returns empty string if there no active project, or no access to the project
func OcGetCurrentProject() string {
	stdout, _, exitCode := cmdRunner("oc project -q")
	if exitCode == 0 {
		return stdout
	}
	return ""
}

// CheckCmdOpInRemoteCmpPod runs the provided command on remote component pod and returns the return value of command output handler function passed to it
func CheckCmdOpInRemoteCmpPod(cmpName string, appName string, cmd string, checkOp func(cmdOp string, err error) bool) bool {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	podName := CmdShouldPass(fmt.Sprintf("oc get pods --selector=\"deploymentconfig=%s\" -o jsonpath='{.items[0].metadata.name}'", cmpDCName))
	remoteCmpPodExecCmdStr := fmt.Sprintf("oc exec %s -c %s -- %s;exit", podName, cmpDCName, cmd)
	stdout, stderr, exitcode := cmdRunner(remoteCmpPodExecCmdStr)
	if exitcode != 0 || stderr != "" {
		return checkOp(stdout, fmt.Errorf("cmd %s failed with error %s on pod %s", cmd, stderr, podName))
	}
	return checkOp(stdout, nil)
}

// VerifyCmpExists verifies if component was created successfully
func VerifyCmpExists(cmpName string, appName string) {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	CmdShouldPass(fmt.Sprintf("oc get dc %s", cmpDCName))
}

// waitForCmdOut runs a command until it gets
// the expected output.
// It accepts 4 arguments, cmd (command to be run)
// timeout (the time to wait for the output)
// errOnFail (flag to set if test should fail if command fails)
// check (function with output check logic)
// It times out if the command doesn't fetch the
// expected output  within the timeout period.
func waitForCmdOut(cmd string, timeout time.Duration, errOnFail bool, check func(output string) bool) bool {

	pingTimeout := time.After(timeout)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout out after %v minutes", timeout))

		case <-tick:
			out, err := exec.Command("/bin/sh", "-c", cmd).Output()
			if err != nil && errOnFail {
				fmt.Fprintf(GinkgoWriter, "Command (%s) output: %s\n", cmd, out)
				Fail(err.Error())
			}

			if check(strings.TrimSpace(string(out))) {
				return true
			}
		}
	}

}
