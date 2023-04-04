package helper

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"

	"github.com/redhat-developer/odo/pkg/labels"
)

const (
	TelemetryCaller = "TELEMETRY_CALLER"
)

func runningCmd(cmd *exec.Cmd) string {
	prog := filepath.Base(cmd.Path)
	var env []string
	for _, e := range cmd.Env {
		if strings.HasPrefix(e, "ODO_") || e == TelemetryCaller {
			env = append(env, e)
		}
	}
	return fmt.Sprintf("Running %s with args %v and odo env: %v", prog, cmd.Args, env)
}

func CmdRunner(program string, args ...string) *gexec.Session {
	// prefix ginkgo verbose output with program name
	prefix := fmt.Sprintf("[%s] ", filepath.Base(program))
	prefixWriter := gexec.NewPrefixedWriter(prefix, GinkgoWriter)
	command := exec.Command(program, args...)
	setSysProcAttr(command)
	fmt.Fprintln(GinkgoWriter, runningCmd(command))
	session, err := gexec.Start(command, prefixWriter, prefixWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

// WaitForOutputToContain waits for the session stdout output to contain a particular substring;
// if the session exits, it checks for the substring and returns early
func WaitForOutputToContain(substring string, timeoutInSeconds int, intervalInSeconds int, session *gexec.Session) {
	Eventually(func() string {
		if session.ExitCode() != -1 {
			Expect(string(session.Out.Contents())).To(ContainSubstring(substring), "session exited, but substring not found")
		}
		contents := string(session.Out.Contents())
		return contents
	}, timeoutInSeconds, intervalInSeconds).Should(ContainSubstring(substring))

}

func WaitForOutputToContainOne(substrings []string, timeoutInSeconds int, intervalInSeconds int, session *gexec.Session) {

	matchers := make([]types.GomegaMatcher, 0, len(substrings))
	for _, substring := range substrings {
		matchers = append(matchers, ContainSubstring(substring))
	}
	Eventually(func() string {
		if session.ExitCode() != -1 {
			Expect(string(session.Out.Contents())).To(SatisfyAny(matchers...), "session exited, but substring not found")
		}
		contents := string(session.Out.Contents())
		return contents
	}, timeoutInSeconds, intervalInSeconds).Should(SatisfyAny(matchers...))
}

// WaitForErroutToContain waits for the session stdout output to contain a particular substring
// if the session exits, it checks for the substring and returns early
func WaitForErroutToContain(substring string, timeoutInSeconds int, intervalInSeconds int, session *gexec.Session) {

	Eventually(func() string {
		if session.ExitCode() != -1 {
			Expect(string(session.Err.Contents())).To(ContainSubstring(substring), "session exited, but substring not found")
		}
		contents := string(session.Err.Contents())
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
	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).SelectorFlag()
	output := Cmd(path, "get", "deployment", selector, "--namespace", projectName,
		"-o", "go-template='{{ range $k, $v := (index .items 0).metadata.annotations}}{{$k}}:{{$v}}{{\"\\n\"}}{{end}}'").ShouldPass().Out()

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

// GetSecrets gets all the secrets belonging to the project
func GetSecrets(path, project string) string {
	session := CmdRunner(path, "get", "secrets", "--namespace", project)
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	return output
}

// GetEnvRefNames gets the ref values from the envFroms of the deployment belonging to the given data
func GetEnvRefNames(path, componentName, appName, projectName string) []string {
	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).SelectorFlag()
	output := Cmd(path, "get", "deployment", selector, "--namespace", projectName,
		"-o", "jsonpath='{range .items[0].spec.template.spec.containers[0].envFrom[*]}{.secretRef.name}{\"\\n\"}{end}'").ShouldPass().Out()

	var result []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimPrefix(line, "'")
		result = append(result, strings.TrimSpace(line))
	}
	return result
}

// GetEnvFromEntry returns envFrom entry of the deployment
func GetEnvFromEntry(path string, componentName string, appName string, projectName string) string {
	envFromOut := Cmd(path, "get", "deployment", componentName+"-"+appName, "--namespace", projectName,
		"-o", "jsonpath='{.spec.template.spec.containers[0].envFrom}'").ShouldPass().Out()
	return strings.TrimSpace(envFromOut)
}

// GetVolumeNamesFromDeployment gets the volumes from the deployment belonging to the given data
func GetVolumeNamesFromDeployment(path, componentName, appName, projectName string) map[string]string {
	var mapOutput = make(map[string]string)
	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).SelectorFlag()
	output := Cmd(path, "get", "deployment", selector, "--namespace", projectName,
		"-o", "jsonpath='{range .items[0].spec.template.spec.volumes[*]}{.name}{\":\"}{.persistentVolumeClaim.claimName}{\"\\n\"}{end}'").ShouldPass().Out()

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimPrefix(line, "'")
		splits := strings.Split(line, ":")
		if splits[0] == "" {
			continue
		}
		name := splits[0]

		// if there is no persistent volume claim for the volume
		// we mark it as emptyDir
		value := "emptyDir"
		if len(splits) > 1 && splits[1] != "" {
			value = splits[1]
		}
		mapOutput[name] = value
	}
	return mapOutput
}
