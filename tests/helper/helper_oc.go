package helper

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
func (oc OcRunner) Run(cmd string) *gexec.Session {
	session := CmdRunner(cmd)
	Eventually(session).Should(gexec.Exit(0))
	return session
}

// SwitchProject switch to the project
func (oc OcRunner) SwitchProject(projectName string) {
	fmt.Fprintf(GinkgoWriter, "Switching to project : %s\n", projectName)
	session := CmdShouldPass(oc.path, "project", projectName)
	Expect(session).To(ContainSubstring(projectName))
}

// GetCurrentProject get currently active project in oc
// returns empty string if there no active project, or no access to the project
func (oc OcRunner) GetCurrentProject() string {
	session := CmdRunner(oc.path, "project", "-q")
	session.Wait()
	if session.ExitCode() == 0 {
		return strings.TrimSpace(string(session.Out.Contents()))
	}
	return ""
}

// GetCurrentServerURL retrieves the URL of the server we're currently connected to
// returns empty if not connected or an error occurred
func (oc OcRunner) GetCurrentServerURL() string {
	session := CmdRunner(oc.path, "project")
	session.Wait()
	if session.ExitCode() == 0 {
		output := strings.TrimSpace(string(session.Out.Contents()))
		// format is: Using project "<namespace>" on server "<url>".
		a := strings.Split(output, "\"")
		return a[len(a)-2] // last entry is ".", we need the one before that
	}
	return ""
}

// GetFirstURL returns the url of the first Route that it can find for given component
func (oc OcRunner) GetFirstURL(component string, app string, project string) string {
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

// GetComponentRoutes run command to get the Routes in yaml format for given component
func (oc OcRunner) GetComponentRoutes(component string, app string, project string) string {
	session := CmdRunner(oc.path, "get", "route",
		"-n", project,
		"-l", "app.kubernetes.io/instance="+component,
		"-l", "app.kubernetes.io/part-of="+app,
		"-o", "yaml")

	session.Wait()
	if session.ExitCode() == 0 {
		return string(session.Out.Contents())
	}
	return ""
}

// GetComponentDC run command to get the DeploymentConfig in yaml format for given component
func (oc OcRunner) GetComponentDC(component string, app string, project string) string {
	session := CmdRunner(oc.path, "get", "dc",
		"-n", project,
		"-l", "app.kubernetes.io/instance="+component,
		"-l", "app.kubernetes.io/part-of="+app,
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
func (oc OcRunner) SourceTest(appTestName string, sourceType string, source string) {
	// checking for source-type in dc
	sourceTypeInDc := CmdShouldPass(oc.path, "get", "dc", "wildfly-"+appTestName,
		"-o", "go-template='{{index .metadata.annotations \"app.kubernetes.io/component-source-type\"}}'")
	Expect(sourceTypeInDc).To(ContainSubstring(sourceType))

	// checking for source in dc
	sourceInDc := CmdShouldPass(oc.path, "get", "dc", "wildfly-"+appTestName,
		"-o", "go-template='{{index .metadata.annotations \"app.openshift.io/vcs-uri\"}}'")
	Expect(sourceInDc).To(ContainSubstring(source))
}

// ExecListDir returns dir list in specified location of pod
func (oc OcRunner) ExecListDir(podName string, projectName string, dir string) string {
	stdOut := CmdShouldPass(oc.path, "exec", podName, "--namespace", projectName,
		"--", "ls", "-lai", dir)
	return stdOut
}

// CheckCmdOpInRemoteCmpPod runs the provided command on remote component pod and returns the return value of command output handler function passed to it
func (oc OcRunner) CheckCmdOpInRemoteCmpPod(cmpName string, appName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	outPodName := CmdShouldPass(oc.path, "get", "pods", "--namespace", prjName,
		"--selector=deploymentconfig="+cmpDCName,
		"-o", "jsonpath='{.items[0].metadata.name}'")
	podName := strings.Replace(outPodName, "'", "", -1)
	session := CmdRunner(oc.path, append([]string{"exec", podName, "--namespace", prjName,
		"-c", cmpDCName, "--"}, cmd...)...)
	stdOut := string(session.Wait().Out.Contents())
	stdErr := string(session.Wait().Err.Contents())
	if stdErr != "" && session.ExitCode() != 0 {
		return checkOp(stdOut, fmt.Errorf("cmd %s failed with error %s on pod %s", cmd, stdErr, podName))
	}
	return checkOp(stdOut, nil)
}

// CheckCmdOpInRemoteDevfilePod runs the provided command on remote component pod and returns the return value of command output handler function passed to it
func (oc OcRunner) CheckCmdOpInRemoteDevfilePod(podName string, containerName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool {
	var execOptions []string
	execOptions = []string{"exec", podName, "--namespace", prjName, "--"}
	if containerName != "" {
		execOptions = []string{"exec", podName, "-c", containerName, "--namespace", prjName, "--"}
	}
	args := append(execOptions, cmd...)
	session := CmdRunner(oc.path, args...)
	stdOut := string(session.Wait().Out.Contents())
	stdErr := string(session.Wait().Err.Contents())
	if stdErr != "" && session.ExitCode() != 0 {
		return checkOp(stdOut, fmt.Errorf("cmd %s failed with error %s on pod %s", cmd, stdErr, podName))
	}
	return checkOp(stdOut, nil)
}

// VerifyCmpExists verifies if component was created successfully
func (oc OcRunner) VerifyCmpExists(cmpName string, appName string, prjName string) {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	CmdShouldPass(oc.path, "get", "dc", cmpDCName, "--namespace", prjName)
}

// VerifyLabelExistsOfComponent verifies app name of component
func (oc OcRunner) VerifyLabelExistsOfComponent(cmpName string, namespace string, labelName string) {
	dcName := oc.GetDcName(cmpName, namespace)
	session := CmdShouldPass(oc.path, "get", "dc", dcName, "--namespace", namespace,
		"--template={{.metadata.labels}}")
	Expect(session).To(ContainSubstring(labelName))
}

// VerifyAppNameOfComponent verifies app name of component
func (oc OcRunner) VerifyAppNameOfComponent(cmpName string, appName string, namespace string) {
	session := CmdShouldPass(oc.path, "get", "dc", cmpName+"-"+appName, "--namespace", namespace,
		"--template={{.metadata.labels.app}}")
	Expect(session).To(ContainSubstring(appName))
}

// VerifyCmpName verifies the component name
func (oc OcRunner) VerifyCmpName(cmpName string, namespace string) {
	dcName := oc.GetDcName(cmpName, namespace)
	session := CmdShouldPass(oc.path, "get", "dc", dcName,
		"--namespace", namespace,
		"-L", "app.kubernetes.io/instance")
	Expect(session).To(ContainSubstring(cmpName))
}

// GetDcName execute oc command and returns dc name of a delopyed
// component by passing component name as a argument
func (oc OcRunner) GetDcName(compName string, namespace string) string {
	session := CmdShouldPass(oc.path, "get", "dc", "--namespace", namespace)
	re := regexp.MustCompile(compName + `-\S+ `)
	dcName := re.FindString(session)
	return strings.TrimSpace(dcName)
}

// DescribeDc execute oc command and returns dc describe as a string
// by passing dcname and namespace as arguments
func (oc OcRunner) DescribeDc(dcName string, namespace string) string {
	describeInfo := CmdShouldPass(oc.path, "describe", "dc/"+dcName, "-n", namespace)
	return strings.TrimSpace(describeInfo)
}

// GetDcPorts returns the ports of the component
func (oc OcRunner) GetDcPorts(componentName string, appName string, project string) string {
	ports := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.ports}}{{end}}'")
	return ports
}

// MaxMemory returns maximum memory
func (oc OcRunner) MaxMemory(componentName string, appName string, project string) string {
	maxMemory := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'")
	return maxMemory
}

// MinMemory returns minimum memory
func (oc OcRunner) MinMemory(componentName string, appName string, project string) string {
	minMemory := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'")
	return minMemory
}

// MaxCPU returns maximum cpu
func (oc OcRunner) MaxCPU(componentName string, appName string, project string) string {
	maxCPU := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.limits.cpu}}{{end}}'")
	return maxCPU
}

// MinCPU returns minimum cpu
func (oc OcRunner) MinCPU(componentName string, appName string, project string) string {
	minCPU := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.requests.cpu}}{{end}}'")
	return minCPU
}

// SourceTypeDC returns the source type from the deployment config
func (oc OcRunner) SourceTypeDC(componentName string, appName string, project string) string {
	sourceType := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{index .metadata.annotations \"app.kubernetes.io/component-source-type\"}}'")
	return sourceType
}

// SourceTypeBC returns the source type from the build config
func (oc OcRunner) SourceTypeBC(componentName string, appName string, project string) string {
	sourceType := CmdShouldPass(oc.path, "get", "bc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{.spec.source.type}}'")
	return sourceType
}

// SourceLocationDC returns the source location from the deployment config
func (oc OcRunner) SourceLocationDC(componentName string, appName string, project string) string {
	sourceLocation := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{index .metadata.annotations \"app.openshift.io/vcs-uri\"}}'")
	return sourceLocation
}

// SourceLocationBC returns the source location from the build config
func (oc OcRunner) SourceLocationBC(componentName string, appName string, project string) string {
	sourceLocation := CmdShouldPass(oc.path, "get", "bc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{index .spec.source.git \"uri\"}}'")
	return sourceLocation
}

// CheckForImageStream checks if there is a ImageStram with name and tag in the specified namespace
func (oc OcRunner) CheckForImageStream(namespace string, name string, tag string) bool {
	// first check if there is ImageStream with given name
	names := strings.Trim(CmdShouldPass(oc.path, "get", "is", "-n", namespace,
		"-o", "jsonpath='{range .items[*]}{.metadata.name}{\"\\n\"}{end}'"), "'")
	scanner := bufio.NewScanner(strings.NewReader(names))
	namePresent := false
	for scanner.Scan() {
		if scanner.Text() == name {
			namePresent = true
		}
	}
	tagPresent := false
	// if there is a ImageStream check if there is a given tag
	if namePresent {
		tags := strings.Trim(CmdShouldPass(oc.path, "get", "is", name, "-n", namespace,
			"-o", "jsonpath='{range .status.tags[*]}{.tag}{\"\\n\"}{end}'"), "'")
		scanner := bufio.NewScanner(strings.NewReader(tags))
		for scanner.Scan() {
			if scanner.Text() == tag {
				tagPresent = true
			}
		}
	}

	if tagPresent {
		return true
	}
	return false
}

// ImportImageFromRegistry import the required image of the respective component type from the specified registry
func (oc OcRunner) ImportImageFromRegistry(registry, image, cmpType, project string) {
	CmdShouldPass(oc.path, "--request-timeout", "5m", "import-image", cmpType, "--namespace="+project, "--from="+filepath.Join(registry, image), "--confirm")
	CmdShouldPass(oc.path, "annotate", filepath.Join("istag", cmpType), "--namespace="+project, "tags=builder", "--overwrite")

}

// ImportJavaIS import the openjdk image which is used for jars
func (oc OcRunner) ImportJavaIS(project string) {
	// if ImageStram already exists, no need to do anything
	if oc.CheckForImageStream("openshift", "java", "8") {
		return
	}

	// we need to import the openjdk image which is used for jars because it's not available by default
	CmdShouldPass(oc.path, "--request-timeout", "5m", "import-image", "java:8",
		"--namespace="+project, "--from=registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift:1.5",
		"--confirm")
	CmdShouldPass(oc.path, "annotate", "istag/java:8", "--namespace="+project,
		"tags=builder", "--overwrite")
}

// ImportDotnet20IS import the dotnet image
func (oc OcRunner) ImportDotnet20IS(project string) {
	// if ImageStram already exists, no need to do anything
	if oc.CheckForImageStream("openshift", "dotnet", "2.0") {
		return
	}

	// we need to import the openjdk image which is used for jars because it's not available by default
	CmdShouldPass(oc.path, "--request-timeout", "5m", "import-image", "dotnet:2.0",
		"--namespace="+project, "--from=registry.centos.org/dotnet/dotnet-20-centos7",
		"--confirm")
	CmdShouldPass(oc.path, "annotate", "istag/dotnet:2.0", "--namespace="+project,
		"tags=builder", "--overwrite")
}

// EnvVarTest checks the component container env vars in the build config for git and deployment config for git/binary/local
// appTestName is the app of the app
// sourceType is the type of the source of the component i.e git/binary/local
func (oc OcRunner) EnvVarTest(resourceName string, sourceType string, envString string) {

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
func (oc OcRunner) GetRunningPodNameOfComp(compName string, namespace string) string {
	stdOut := CmdShouldPass(oc.path, "get", "pods", "--namespace", namespace, "--show-labels")
	re := regexp.MustCompile(`(` + compName + `-\S+)\s+\S+\s+Running.*deploymentconfig=` + compName)
	podName := re.FindStringSubmatch(stdOut)[1]
	return strings.TrimSpace(podName)
}

// GetRunningPodNameByComponent executes oc command and returns the running pod name of a delopyed
// devfile component by passing component name as a argument
func (oc OcRunner) GetRunningPodNameByComponent(compName string, namespace string) string {
	stdOut := CmdShouldPass(oc.path, "get", "pods", "--namespace", namespace, "--show-labels")
	re := regexp.MustCompile(`(` + compName + `-\S+)\s+\S+\s+Running.*component=` + compName)
	podName := re.FindStringSubmatch(stdOut)[1]
	return strings.TrimSpace(podName)
}

// GetPVCSize executes oc command and returns the bound storage size
func (oc OcRunner) GetPVCSize(compName, storageName, namespace string) string {
	stdOut := CmdShouldPass(oc.path, "get", "pvc", "--namespace", namespace, "--show-labels")
	re := regexp.MustCompile(storageName + `-\S+\s+Bound\s+\S+\s+(\S+).*component=` + compName + `,storage-name=` + storageName)
	storageSize := re.FindStringSubmatch(stdOut)[1]
	return strings.TrimSpace(storageSize)
}

// GetRoute returns route URL
func (oc OcRunner) GetRoute(urlName string, appName string) string {
	session := CmdRunner(oc.path, "get", "routes", urlName+"-"+appName,
		"-o jsonpath={.spec.host}")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// GetToken returns current user token
func (oc OcRunner) GetToken() string {
	session := CmdRunner(oc.path, "whoami", "-t")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// LoginUsingToken returns output after successful login
func (oc OcRunner) LoginUsingToken(token string) string {
	session := CmdRunner(oc.path, "login", "--token", token)
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// GetLoginUser returns current user name
func (oc OcRunner) GetLoginUser() string {
	user := CmdShouldPass(oc.path, "whoami")
	return strings.TrimSpace(user)
}

// ServiceInstanceStatus returns service instance
func (oc OcRunner) ServiceInstanceStatus(serviceInstanceName string) string {
	serviceinstance := CmdShouldPass(oc.path, "get", "serviceinstance", serviceInstanceName,
		"-o", "go-template='{{ (index .status.conditions 0).reason}}'")
	return strings.TrimSpace(serviceinstance)
}

// GetVolumeMountNamesandPathsFromContainer returns the volume name and mount path in the format name:path\n
func (oc OcRunner) GetVolumeMountNamesandPathsFromContainer(deployName string, containerName, namespace string) string {
	volumeName := CmdShouldPass(oc.path, "get", "deploy", deployName, "--namespace", namespace,
		"-o", "go-template="+
			"{{range .spec.template.spec.containers}}{{if eq .name \""+containerName+
			"\"}}{{range .volumeMounts}}{{.name}}{{\":\"}}{{.mountPath}}{{\"\\n\"}}{{end}}{{end}}{{end}}")

	return strings.TrimSpace(volumeName)
}

// GetVolumeMountName returns the name of the volume
func (oc OcRunner) GetVolumeMountName(dcName string, namespace string) string {
	volumeName := CmdShouldPass(oc.path, "get", "dc", dcName, "--namespace", namespace,
		"-o", "go-template='"+
			"{{range .spec.template.spec.containers}}"+
			"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

	return strings.TrimSpace(volumeName)
}

// GetVolumeMountPath returns the path of the volume mount
func (oc OcRunner) GetVolumeMountPath(dcName string, namespace string) string {
	volumePaths := CmdShouldPass(oc.path, "get", "dc", dcName, "--namespace", namespace,
		"-o", "go-template='"+
			"{{range .spec.template.spec.containers}}"+
			"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

	return strings.TrimSpace(volumePaths)
}

// GetEnvFromEntry returns envFrom entry
func (oc OcRunner) GetEnvFromEntry(componentName string, appName string, projectName string) string {
	envFromOut := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", projectName,
		"-o", "jsonpath='{.spec.template.spec.containers[0].envFrom}'")
	return strings.TrimSpace(envFromOut)
}

// GetEnvs returns all env variables in deployment config
func (oc OcRunner) GetEnvs(componentName string, appName string, projectName string) map[string]string {
	var mapOutput = make(map[string]string)

	output := CmdShouldPass(oc.path, "get", "dc", componentName+"-"+appName, "--namespace", projectName,
		"-o", "jsonpath='{range .spec.template.spec.containers[0].env[*]}{.name}:{.value}{\"\\n\"}{end}'")

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimPrefix(line, "'")
		splits := strings.Split(line, ":")
		name := splits[0]
		value := strings.Join(splits[1:], ":")
		mapOutput[name] = value
	}
	return mapOutput
}

func (oc OcRunner) GetEnvsDevFileDeployment(componentName string, projectName string) map[string]string {
	var mapOutput = make(map[string]string)

	output := CmdShouldPass(oc.path, "get", "deployment", componentName, "--namespace", projectName,
		"-o", "jsonpath='{range .spec.template.spec.containers[0].env[*]}{.name}:{.value}{\"\\n\"}{end}'")

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimPrefix(line, "'")
		splits := strings.Split(line, ":")
		name := splits[0]
		value := strings.Join(splits[1:], ":")
		mapOutput[name] = value
	}
	return mapOutput
}

// WaitForDCRollout wait for DeploymentConfig to finish active rollout
// timeout is a maximum wait time in seconds
func (oc OcRunner) WaitForDCRollout(dcName string, project string, timeout time.Duration) {
	session := CmdRunner(oc.path, "rollout", "status",
		"-w",
		"-n", project,
		"dc", dcName)

	Eventually(session).Should(gexec.Exit(0), runningCmd(session.Command))
	session.Wait(timeout)
}

// WaitAndCheckForExistence wait for the given and checks if the given resource type gets deleted on the cluster
func (oc OcRunner) WaitAndCheckForExistence(resourceType, namespace string, timeoutMinutes int) bool {
	pingTimeout := time.After(time.Duration(timeoutMinutes) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout after %d minutes", timeoutMinutes))

		case <-tick:
			session := CmdRunner(oc.path, "get", resourceType, "--namespace", namespace)
			Eventually(session).Should(gexec.Exit(0))
			// https://github.com/kubernetes/kubectl/issues/847
			output := string(session.Wait().Err.Contents())

			if strings.Contains(strings.ToLower(output), "no resources found") {
				return true
			}
		}
	}
}

// GetServices gets services on the cluster
func (oc OcRunner) GetServices(namespace string) string {
	session := CmdRunner(oc.path, "get", "services", "--namespace", namespace)
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	return output
}

// VerifyResourceDeleted verifies if the given resource is deleted from cluster
func (oc OcRunner) VerifyResourceDeleted(resourceType, resourceName, namespace string) {
	session := CmdRunner(oc.path, "get", resourceType, "--namespace", namespace)
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	Expect(output).NotTo(ContainSubstring(resourceName))
}

// CreateRandNamespaceProject create new project with random name in oc cluster (10 letters)
func (oc OcRunner) CreateRandNamespaceProject() string {
	projectName := RandString(10)
	fmt.Fprintf(GinkgoWriter, "Creating a new project: %s\n", projectName)
	session := CmdShouldPass("odo", "project", "create", projectName, "-w", "-v4")
	Expect(session).To(ContainSubstring("New project created"))
	Expect(session).To(ContainSubstring(projectName))
	return projectName
}

// DeleteNamespaceProject deletes a specified project in oc cluster
func (oc OcRunner) DeleteNamespaceProject(projectName string) {
	fmt.Fprintf(GinkgoWriter, "Deleting project: %s\n", projectName)
	session := CmdShouldPass("odo", "project", "delete", projectName, "-f")
	Expect(session).To(ContainSubstring("Deleted project : " + projectName))
}
