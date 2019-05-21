package helper

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type OcRunner struct {
	// path to oc binary
	path string
}

// NewOcRunner initializes new OcRunner
func NewOcRunner(ocPath string) OcRunner {
	return OcRunner{
		path: ocPath,
	}
}

// Run oc with given arguments
func (oc *OcRunner) Run(cmd string) *gexec.Session {
	session := CmdRunner(cmd)
	Eventually(session).Should(gexec.Exit(0))
	return session
}

// SwitchProject switch to the project
func (oc *OcRunner) SwitchProject(projectName string) {
	fmt.Fprintf(GinkgoWriter, "Switching to project : %s\n", projectName)
	session := CmdShouldPass(oc.path, "project", projectName)
	Expect(session).To(ContainSubstring(projectName))
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
		"-l", "app.kubernetes.io/component-name="+component,
		"-l", "app.kubernetes.io/name="+app,
		"-o", "jsonpath={.items[0].spec.host}")

	session.Wait()
	if session.ExitCode() == 0 {
		return string(session.Out.Contents())
	}
	return ""
}

// GetComponentRoutes run command to get the Routes in yaml format for given component
func (oc *OcRunner) GetComponentRoutes(component string, app string, project string) string {
	session := CmdRunner(oc.path, "get", "route",
		"-n", project,
		"-l", "app.kubernetes.io/component-name="+component,
		"-l", "app.kubernetes.io/name="+app,
		"-o", "yaml")

	session.Wait()
	if session.ExitCode() == 0 {
		return string(session.Out.Contents())
	}
	return ""
}

// GetComponentDC run command to get the DeploymentConfig in yaml format for given component
func (oc *OcRunner) GetComponentDC(component string, app string, project string) string {
	session := CmdRunner(oc.path, "get", "dc",
		"-n", project,
		"-l", "app.kubernetes.io/component-name="+component,
		"-l", "app.kubernetes.io/name="+app,
		"-o", "yaml")

	session.Wait()
	if session.ExitCode() == 0 {
		return string(session.Out.Contents())
	}
	return ""
}

// SourceTest checks the component-source-type and the source url in the annotation of the bc and dc
// appTestName is the name of the app
// sourceType is the type of the source of the component i.e git/binary/local
// source is the source of the component i.e gitURL or path to the directory or binary file
func (oc *OcRunner) SourceTest(appTestName string, sourceType string, source string) {
	// checking for source-type in dc
	sourceTypeInDc := CmdShouldPass(oc.path, "get", "dc", "wildfly-"+appTestName,
		"-o", "go-template='{{index .metadata.annotations \"app.kubernetes.io/component-source-type\"}}'")
	Expect(sourceTypeInDc).To(ContainSubstring(sourceType))

	// checking for source in dc
	sourceInDc := CmdShouldPass(oc.path, "get", "dc", "wildfly-"+appTestName,
		"-o", "go-template='{{index .metadata.annotations \"app.kubernetes.io/url\"}}'")
	Expect(sourceInDc).To(ContainSubstring(source))
}

// ExecListDir returns dir list in specified location of pod
func (oc *OcRunner) ExecListDir(podName string, projectName string) string {
	stdOut := CmdShouldPass(oc.path, "exec", podName, "--namespace", projectName,
		"--", "ls", "-lai", "/opt/app-root/src")
	return stdOut
}

// CheckCmdOpInRemoteCmpPod runs the provided command on remote component pod and returns the return value of command output handler function passed to it
func (oc *OcRunner) CheckCmdOpInRemoteCmpPod(cmpName string, appName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	outPodName := CmdShouldPass(oc.path, "get", "pods", "--namespace", prjName,
		"--selector=deploymentconfig="+cmpDCName,
		"-o", "jsonpath='{.items[0].metadata.name}'")
	podName := strings.Replace(outPodName, "'", "", -1)
	session := CmdRunner(oc.path, append([]string{"exec", podName, "--namespace", prjName,
		"-c", cmpDCName, "--"}, cmd...)...)
	Eventually(session).Should(gexec.Exit(0))
	stdOut := string(session.Wait().Out.Contents())
	stdErr := string(session.Wait().Err.Contents())
	if stdErr != "" {
		return checkOp(stdOut, fmt.Errorf("cmd %s failed with error %s on pod %s", cmd, stdErr, podName))
	}
	return checkOp(stdOut, nil)
}

// VerifyCmpExists verifies if component was created successfully
func (oc *OcRunner) VerifyCmpExists(cmpName string, appName string, prjName string) {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	CmdShouldPass(oc.path, "get", "dc", cmpDCName, "--namespace", prjName)
}

// VerifyAppNameOfComponent verifies app name of component
func (oc *OcRunner) VerifyAppNameOfComponent(cmpName string, appName string, namespace string) {
	session := CmdShouldPass(oc.path, "get", "dc", cmpName+"-"+appName, "--namespace", namespace,
		"--template={{.metadata.labels.app}}")
	Expect(session).To(ContainSubstring(appName))
}

// VerifyCmpName verifies the component name
func (oc *OcRunner) VerifyCmpName(cmpName string, namespace string) {
	dcName := oc.GetDcName(cmpName, namespace)
	session := CmdShouldPass(oc.path, "get", "dc", dcName,
		"--namespace", namespace,
		"-L", "app.kubernetes.io/component-name")
	Expect(session).To(ContainSubstring(cmpName))
}

// GetDcName execute oc command and returns dc name of a delopyed
// component by passing component name as a argument
func (oc *OcRunner) GetDcName(compName string, namespace string) string {
	session := CmdShouldPass(oc.path, "get", "dc", "--namespace", namespace)
	re := regexp.MustCompile(compName + `-\S+ `)
	dcName := re.FindString(session)
	return strings.TrimSpace(dcName)
}

// MaxMemory reuturns maximum memory
func (oc *OcRunner) MaxMemory(componentName string, appName string, project string) string {
	maxMemory := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'")
	return maxMemory
}

// MinMemory reuturns minimum memory
func (oc *OcRunner) MinMemory(componentName string, appName string, project string) string {
	minMemory := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'")
	return minMemory
}

// MaxCPU reuturns maximum cpu
func (oc *OcRunner) MaxCPU(componentName string, appName string, project string) string {
	maxMemory := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.limits.cpu}}{{end}}'")
	return maxMemory
}

// MinCPU reuturns maximum cpu
func (oc *OcRunner) MinCPU(componentName string, appName string, project string) string {
	minMemory := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.requests.cpu}}{{end}}'")
	return minMemory
}

// ImportJavaIsToNspace import the openjdk image which is used for jars
func (oc *OcRunner) ImportJavaIsToNspace(project string) {
	// do nothing if running on OpenShiftCI
	// java image is already present
	val, ok := os.LookupEnv("CI")
	if ok && val == "openshift" {
		return
	}

	// we need to import the openjdk image which is used for jars because it's not available by default
	CmdShouldPass(oc.path, "--request-timeout", "5m", "import-image", "java",
		"--namespace="+project, "--from=registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift:1.5",
		"--confirm")
	CmdShouldPass(oc.path, "annotate", "istag/java:latest", "--namespace="+project,
		"tags=builder", "--overwrite")
}

// EnvVarTest checks the component container env vars in the build config for git and deployment config for git/binary/local
// appTestName is the app of the app
// sourceType is the type of the source of the component i.e git/binary/local
func (oc *OcRunner) EnvVarTest(resourceName string, sourceType string, envString string) {

	if sourceType == "git" {
		// checking the values of the env vars pairs in bc
		envVars := CmdShouldPass(oc.path, "get", "bc", resourceName,
			"-o", "go-template='{{range .spec.strategy.sourceStrategy.env}}{{.name}}{{.value}}{{end}}'")
		Expect(envVars).To(Equal(envString))
	}

	// checking the values of the env vars pairs in dc
	envVars := CmdShouldPass(oc.path, "get", "dc", resourceName,
		"-o", "go-template='{{range .spec.template.spec.containers}}{{range .env}}{{.name}}{{.value}}{{end}}{{end}}'")
	Expect(envVars).To(Equal(envString))
}

// GetRunningPodNameOfComp executes oc command and returns the running pod name of a delopyed
// component by passing component name as a argument
func (oc *OcRunner) GetRunningPodNameOfComp(compName string, namespace string) string {
	stdOut := CmdShouldPass(oc.path, "get", "pods", "--namespace", namespace)
	re := regexp.MustCompile(`(` + compName + `-\S+)\s+\S+\s+Running`)
	podName := re.FindStringSubmatch(stdOut)[1]
	return strings.TrimSpace(podName)
}

// GetRoute returns route URL
func (oc *OcRunner) GetRoute(urlName string, appName string) string {
	session := CmdRunner(oc.path, "get", "routes", urlName+"-"+appName,
		"-o jsonpath={.spec.host}")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// GetToken returns current user token
func (oc *OcRunner) GetToken() string {
	session := CmdRunner(oc.path, "whoami", "-t")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// LoginUsingToken returns output after successful login
func (oc *OcRunner) LoginUsingToken(token string) string {
	session := CmdRunner(oc.path, "login", "--token", token)
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// GetLoginUser returns current user name
func (oc *OcRunner) GetLoginUser() string {
	user := CmdShouldPass(oc.path, "whoami")
	return strings.TrimSpace(user)
}

// ServiceInstanceStatus returns service instance
func (oc *OcRunner) ServiceInstanceStatus(serviceInstanceName string) string {
	serviceinstance := CmdShouldPass(oc.path, "get", "serviceinstance", serviceInstanceName,
		"-o", "go-template='{{ (index .status.conditions 0).reason}}'")
	return strings.TrimSpace(serviceinstance)
}

// GetVolumeMountName returns the name of the volume
func (oc *OcRunner) GetVolumeMountName(dcName string) string {
	volumeName := CmdShouldPass(oc.path, "get", "dc", dcName, "-o", "go-template='"+
		"{{range .spec.template.spec.containers}}"+
		"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

	return strings.TrimSpace(volumeName)
}

// GetVolumeMountPath returns the path of the volume mount
func (oc *OcRunner) GetVolumeMountPath(dcName string) string {
	volumePaths := CmdShouldPass(oc.path, "get", "dc", dcName, "-o", "go-template='"+
		"{{range .spec.template.spec.containers}}"+
		"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

	return strings.TrimSpace(volumePaths)
}

// GetEnvFromEntry returns envFrom entry
func (oc *OcRunner) GetEnvFromEntry(componentName string, appName string) string {
	envFromOut := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName,
		"-o", "jsonpath='{.spec.template.spec.containers[0].envFrom}'")
	return strings.TrimSpace(envFromOut)
}
