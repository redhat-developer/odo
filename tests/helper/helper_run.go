package helper

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/component/labels"
)

func runningCmd(cmd *exec.Cmd) string {
	prog := filepath.Base(cmd.Path)
	return fmt.Sprintf("Running %s with args %v", prog, cmd.Args)
}

func CmdRunner(program string, args ...string) *gexec.Session {
	//prefix ginkgo verbose output with program name
	prefix := fmt.Sprintf("[%s] ", filepath.Base(program))
	prefixWriter := gexec.NewPrefixedWriter(prefix, GinkgoWriter)
	command := exec.Command(program, args...)
	fmt.Fprintln(GinkgoWriter, runningCmd(command))
	session, err := gexec.Start(command, prefixWriter, prefixWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

// CmdShouldPass returns stdout if command succeeds
func CmdShouldPass(program string, args ...string) string {
	session := CmdRunner(program, args...)
	Eventually(session).Should(gexec.Exit(0), runningCmd(session.Command))
	return string(session.Wait().Out.Contents())
}

// CmdShouldPassIncludeErrStream returns stdout and stderr if command succeeds
func CmdShouldPassIncludeErrStream(program string, args ...string) (string, string) {
	session := CmdRunner(program, args...)
	Eventually(session).Should(gexec.Exit(0), runningCmd(session.Command))
	stdout := string(session.Wait().Out.Contents())
	stderr := string(session.Wait().Err.Contents())
	return stdout, stderr
}

// CmdShouldRunWithTimeout waits for a certain duration and then returns stdout
func CmdShouldRunWithTimeout(timeout time.Duration, program string, args ...string) string {
	session := CmdRunner(program, args...)
	time.Sleep(timeout)
	if runtime.GOOS == "windows" {
		session.Kill()
	} else {
		session.Terminate()
	}
	return string(session.Out.Contents())
}

// CmdShouldRunAndTerminate waits and returns stdout after a closed signal is passed on the closed channel
func CmdShouldRunAndTerminate(timeoutAfter time.Duration, stopChan <-chan bool, program string, args ...string) string {
	session := CmdRunner(program, args...)
	timeout := time.After(timeoutAfter)
	select {
	case <-stopChan:
		if session != nil {
			if runtime.GOOS == "windows" {
				session.Kill()
			} else {
				session.Terminate()
			}
		}
	case <-timeout:
		if session != nil {
			if runtime.GOOS == "windows" {
				session.Kill()
			} else {
				session.Terminate()
			}
		}
	}

	if session == nil {
		return ""
	}

	return string(session.Out.Contents())
}

// CmdShouldFail returns stderr if command fails
func CmdShouldFail(program string, args ...string) string {
	session := CmdRunner(program, args...)
	Consistently(session).ShouldNot(gexec.Exit(0), runningCmd(session.Command))
	return string(session.Wait().Err.Contents())
}

// CmdShouldFailWithRetry runs a command and checks if it fails, if it doesn't then it retries
func CmdShouldFailWithRetry(maxRetry, intervalSeconds int, program string, args ...string) string {
	for i := 0; i < maxRetry; i++ {
		fmt.Fprintf(GinkgoWriter, "try %d of %d\n", i, maxRetry)

		session := CmdRunner(program, args...)
		session.Wait()
		// if exit code is 0 which means the program succeeded and hence we retry
		if session.ExitCode() == 0 {
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
		} else {
			Consistently(session).ShouldNot(gexec.Exit(0), runningCmd(session.Command))
			return string(session.Err.Contents())
		}
	}
	Fail(fmt.Sprintf("Failed after %d retries", maxRetry))
	return ""

}

// WaitForOutputToContain waits for for the session stdout output to contain a particular substring
func WaitForOutputToContain(substring string, timeoutInSeconds int, intervalInSeconds int, session *gexec.Session) {

	Eventually(func() string {
		contents := string(session.Out.Contents())
		return contents
	}, timeoutInSeconds, intervalInSeconds).Should(ContainSubstring(substring))

}

// WaitAndCheckForTerminatingState waits for the given interval
// and checks if the given resource type has been deleted on the cluster or is in the terminating state
// path is the path to the program's binary
func WaitAndCheckForTerminatingState(path, resourceType, namespace string, timeoutMinutes int) bool {
	pingTimeout := time.After(time.Duration(timeoutMinutes) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout after %d minutes", timeoutMinutes))

		case <-tick:
			session := CmdRunner(path, "get", resourceType, "--namespace", namespace)
			Eventually(session).Should(gexec.Exit(0))
			// https://github.com/kubernetes/kubectl/issues/847
			outputStdErr := string(session.Wait().Err.Contents())
			outputStdOut := string(session.Wait().Out.Contents())

			// if the resource gets deleted before the check, we won't get the `terminating` state output
			// thus we also check and exit when the resource has been deleted on the cluster.
			if strings.Contains(strings.ToLower(outputStdErr), "no resources found") || strings.Contains(strings.ToLower(outputStdOut), "terminating") {
				return true
			}
		}
	}
}

// GetAnnotationsDeployment gets the annotations from the deployment
// belonging to the given component, app and project
func GetAnnotationsDeployment(path, componentName, appName, projectName string) map[string]string {
	var mapOutput = make(map[string]string)

	selector := fmt.Sprintf("--selector=%s=%s,%s=%s", labels.ComponentLabel, componentName, applabels.ApplicationLabel, appName)
	output := CmdShouldPass(path, "get", "deployment", selector, "--namespace", projectName,
		"-o", "go-template='{{ range $k, $v := (index .items 0).metadata.annotations}}{{$k}}:{{$v}}{{\"\\n\"}}{{end}}'")

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimPrefix(line, "'")
		splits := strings.Split(line, ":")
		if len(splits) < 2 {
			continue
		}
		name := splits[0]
		value := strings.Join(splits[1:], ":")
		mapOutput[name] = value
	}
	return mapOutput
}
