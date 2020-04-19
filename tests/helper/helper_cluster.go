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

type ClusterRunner struct {
	// path to cluster binary
	path string
}

// NewClusterRunner initializes new ClusterRunner
func NewClusterRunner(clusterPath string) ClusterRunner {
	return ClusterRunner{
		path: clusterPath,
	}
}

// Run cluster with given arguments
func (cluster *ClusterRunner) Run(cmd string) *gexec.Session {
	session := CmdRunner(cmd)
	Eventually(session).Should(gexec.Exit(0))
	return session
}

// SwitchProject switch to the project
func (cluster *ClusterRunner) SwitchProject(projectName string) {
	fmt.Fprintf(GinkgoWriter, "Switching to project : %s\n", projectName)
	session := CmdShouldPass(cluster.path, "project", projectName)
	Expect(session).To(ContainSubstring(projectName))
}

// GetCurrentProject get currently active project in cluster
// returns empty string if there no active project, or no access to the project
func (cluster *ClusterRunner) GetCurrentProject() string {
	session := CmdRunner(cluster.path, "project", "-q")
	session.Wait()
	if session.ExitCode() == 0 {
		return strings.TrimSpace(string(session.Out.Contents()))
	}
	return ""
}

// GetFirstURL returns the url of the first Route that it can find for given component
func (cluster *ClusterRunner) GetFirstURL(component string, app string, project string) string {
	session := CmdRunner(cluster.path, "get", "route",
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
func (cluster *ClusterRunner) GetComponentRoutes(component string, app string, project string) string {
	session := CmdRunner(cluster.path, "get", "route",
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
func (cluster *ClusterRunner) GetComponentDC(component string, app string, project string) string {
	session := CmdRunner(cluster.path, "get", "dc",
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
func (cluster *ClusterRunner) SourceTest(appTestName string, sourceType string, source string) {
	// checking for source-type in dc
	sourceTypeInDc := CmdShouldPass(cluster.path, "get", "dc", "wildfly-"+appTestName,
		"-o", "go-template='{{index .metadata.annotations \"app.kubernetes.io/component-source-type\"}}'")
	Expect(sourceTypeInDc).To(ContainSubstring(sourceType))

	// checking for source in dc
	sourceInDc := CmdShouldPass(cluster.path, "get", "dc", "wildfly-"+appTestName,
		"-o", "go-template='{{index .metadata.annotations \"app.openshift.io/vcs-uri\"}}'")
	Expect(sourceInDc).To(ContainSubstring(source))
}

// ExecListDir returns dir list in specified location of pod
func (cluster *ClusterRunner) ExecListDir(podName string, projectName string, dir string) string {
	stdOut := CmdShouldPass(cluster.path, "exec", podName, "--namespace", projectName,
		"--", "ls", "-lai", dir)
	return stdOut
}

// CheckCmdOpInRemoteCmpPod runs the provided command on remote component pod and returns the return value of command output handler function passed to it
func (cluster *ClusterRunner) CheckCmdOpInRemoteCmpPod(cmpName string, appName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	outPodName := CmdShouldPass(cluster.path, "get", "pods", "--namespace", prjName,
		"--selector=deploymentconfig="+cmpDCName,
		"-o", "jsonpath='{.items[0].metadata.name}'")
	podName := strings.Replace(outPodName, "'", "", -1)
	session := CmdRunner(cluster.path, append([]string{"exec", podName, "--namespace", prjName,
		"-c", cmpDCName, "--"}, cmd...)...)
	stdOut := string(session.Wait().Out.Contents())
	stdErr := string(session.Wait().Err.Contents())
	if stdErr != "" {
		return checkOp(stdOut, fmt.Errorf("cmd %s failed with error %s on pod %s", cmd, stdErr, podName))
	}
	return checkOp(stdOut, nil)
}

// CheckCmdOpInRemoteDevfilePod runs the provided command on remote component pod and returns the return value of command output handler function passed to it
func (cluster *ClusterRunner) CheckCmdOpInRemoteDevfilePod(podName string, containerName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool {
	var execOptions []string
	execOptions = []string{"exec", podName, "--namespace", prjName, "--"}
	if containerName != "" {
		execOptions = []string{"exec", podName, "-c", containerName, "--namespace", prjName, "--"}
	}
	args := append(execOptions, cmd...)
	session := CmdRunner(cluster.path, args...)
	stdOut := string(session.Wait().Out.Contents())
	stdErr := string(session.Wait().Err.Contents())
	if stdErr != "" {
		return checkOp(stdOut, fmt.Errorf("cmd %s failed with error %s on pod %s", cmd, stdErr, podName))
	}
	return checkOp(stdOut, nil)
}

// VerifyCmpExists verifies if component was created successfully
func (cluster *ClusterRunner) VerifyCmpExists(cmpName string, appName string, prjName string) {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	CmdShouldPass(cluster.path, "get", "dc", cmpDCName, "--namespace", prjName)
}

// VerifyLabelExistsOfComponent verifies app name of component
func (cluster *ClusterRunner) VerifyLabelExistsOfComponent(cmpName string, namespace string, labelName string) {
	dcName := cluster.GetDcName(cmpName, namespace)
	session := CmdShouldPass(cluster.path, "get", "dc", dcName, "--namespace", namespace,
		"--template={{.metadata.labels}}")
	Expect(session).To(ContainSubstring(labelName))
}

// VerifyAppNameOfComponent verifies app name of component
func (cluster *ClusterRunner) VerifyAppNameOfComponent(cmpName string, appName string, namespace string) {
	session := CmdShouldPass(cluster.path, "get", "dc", cmpName+"-"+appName, "--namespace", namespace,
		"--template={{.metadata.labels.app}}")
	Expect(session).To(ContainSubstring(appName))
}

// VerifyCmpName verifies the component name
func (cluster *ClusterRunner) VerifyCmpName(cmpName string, namespace string) {
	dcName := cluster.GetDcName(cmpName, namespace)
	session := CmdShouldPass(cluster.path, "get", "dc", dcName,
		"--namespace", namespace,
		"-L", "app.kubernetes.io/instance")
	Expect(session).To(ContainSubstring(cmpName))
}

// GetDcName execute cluster command and returns dc name of a delopyed
// component by passing component name as a argument
func (cluster *ClusterRunner) GetDcName(compName string, namespace string) string {
	session := CmdShouldPass(cluster.path, "get", "dc", "--namespace", namespace)
	re := regexp.MustCompile(compName + `-\S+ `)
	dcName := re.FindString(session)
	return strings.TrimSpace(dcName)
}

// DescribeDc execute cluster command and returns dc describe as a string
// by passing dcname and namespace as arguments
func (cluster *ClusterRunner) DescribeDc(dcName string, namespace string) string {
	describeInfo := CmdShouldPass(cluster.path, "describe", "dc/"+dcName, "-n", namespace)
	return strings.TrimSpace(describeInfo)
}

// GetDcPorts returns the ports of the component
func (cluster *ClusterRunner) GetDcPorts(componentName string, appName string, project string) string {
	ports := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.ports}}{{end}}'")
	return ports
}

// MaxMemory returns maximum memory
func (cluster *ClusterRunner) MaxMemory(componentName string, appName string, project string) string {
	maxMemory := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.limits.memory}}{{end}}'")
	return maxMemory
}

// MinMemory returns minimum memory
func (cluster *ClusterRunner) MinMemory(componentName string, appName string, project string) string {
	minMemory := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.requests.memory}}{{end}}'")
	return minMemory
}

// MaxCPU returns maximum cpu
func (cluster *ClusterRunner) MaxCPU(componentName string, appName string, project string) string {
	maxCPU := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.limits.cpu}}{{end}}'")
	return maxCPU
}

// MinCPU returns minimum cpu
func (cluster *ClusterRunner) MinCPU(componentName string, appName string, project string) string {
	minCPU := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{range.spec.template.spec.containers}}{{.resources.requests.cpu}}{{end}}'")
	return minCPU
}

// SourceTypeDC returns the source type from the deployment config
func (cluster *ClusterRunner) SourceTypeDC(componentName string, appName string, project string) string {
	sourceType := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{index .metadata.annotations \"app.kubernetes.io/component-source-type\"}}'")
	return sourceType
}

// SourceTypeBC returns the source type from the build config
func (cluster *ClusterRunner) SourceTypeBC(componentName string, appName string, project string) string {
	sourceType := CmdShouldPass(cluster.path, "get", "bc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{.spec.source.type}}'")
	return sourceType
}

// SourceLocationDC returns the source location from the deployment config
func (cluster *ClusterRunner) SourceLocationDC(componentName string, appName string, project string) string {
	sourceLocation := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{index .metadata.annotations \"app.openshift.io/vcs-uri\"}}'")
	return sourceLocation
}

// SourceLocationBC returns the source location from the build config
func (cluster *ClusterRunner) SourceLocationBC(componentName string, appName string, project string) string {
	sourceLocation := CmdShouldPass(cluster.path, "get", "bc", componentName+"-"+appName, "--namespace", project,
		"-o", "go-template='{{index .spec.source.git \"uri\"}}'")
	return sourceLocation
}

// checkForImageStream checks if there is a ImageStram with name and tag in openshift namespace
func (cluster *ClusterRunner) checkForImageStream(name string, tag string) bool {
	// first check if there is ImageStream with given name
	names := CmdShouldPass(cluster.path, "get", "is", "-n", "openshift",
		"-o", "jsonpath='{range .items[*]}{.metadata.name}{\"\\n\"}{end}'")
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
		tags := CmdShouldPass(cluster.path, "get", "is", name, "-n", "openshift",
			"-o", "jsonpath='{range .spec.tags[*]}{.name}{\"\\n\"}{end}'")
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
func (cluster *ClusterRunner) ImportImageFromRegistry(registry, image, cmpType, project string) {
	CmdShouldPass(cluster.path, "--request-timeout", "5m", "import-image", cmpType, "--namespace="+project, "--from="+filepath.Join(registry, image), "--confirm")
	CmdShouldPass(cluster.path, "annotate", filepath.Join("istag", cmpType), "--namespace="+project, "tags=builder", "--overwrite")

}

// ImportJavaIS import the openjdk image which is used for jars
func (cluster *ClusterRunner) ImportJavaIS(project string) {
	// if ImageStram already exists, no need to do anything
	if cluster.checkForImageStream("java", "8") {
		return
	}

	// we need to import the openjdk image which is used for jars because it's not available by default
	CmdShouldPass(cluster.path, "--request-timeout", "5m", "import-image", "java:8",
		"--namespace="+project, "--from=registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift:1.5",
		"--confirm")
	CmdShouldPass(cluster.path, "annotate", "istag/java:8", "--namespace="+project,
		"tags=builder", "--overwrite")
}

// ImportDotnet20IS import the dotnet image
func (cluster *ClusterRunner) ImportDotnet20IS(project string) {
	// if ImageStram already exists, no need to do anything
	if cluster.checkForImageStream("dotnet", "2.0") {
		return
	}

	// we need to import the openjdk image which is used for jars because it's not available by default
	CmdShouldPass(cluster.path, "--request-timeout", "5m", "import-image", "dotnet:2.0",
		"--namespace="+project, "--from=registry.centos.org/dotnet/dotnet-20-centos7",
		"--confirm")
	CmdShouldPass(cluster.path, "annotate", "istag/dotnet:2.0", "--namespace="+project,
		"tags=builder", "--overwrite")
}

// EnvVarTest checks the component container env vars in the build config for git and deployment config for git/binary/local
// appTestName is the app of the app
// sourceType is the type of the source of the component i.e git/binary/local
func (cluster *ClusterRunner) EnvVarTest(resourceName string, sourceType string, envString string) {

	if sourceType == "git" {
		// checking the values of the env vars pairs in bc
		envVars := CmdShouldPass(cluster.path, "get", "bc", resourceName,
			"-o", "go-template='{{range .spec.strategy.sourceStrategy.env}}{{.name}}{{.value}}{{end}}'")
		Expect(envVars).To(Equal(envString))
	}

	// checking the values of the env vars pairs in dc
	envVars := CmdShouldPass(cluster.path, "get", "dc", resourceName,
		"-o", "go-template='{{range .spec.template.spec.containers}}{{range .env}}{{.name}}{{.value}}{{end}}{{end}}'")
	Expect(envVars).To(Equal(envString))
}

// GetRunningPodNameOfComp executes cluster command and returns the running pod name of a delopyed
// component by passing component name as a argument
func (cluster *ClusterRunner) GetRunningPodNameOfComp(compName string, namespace string) string {
	stdOut := CmdShouldPass(cluster.path, "get", "pods", "--namespace", namespace, "--show-labels")
	re := regexp.MustCompile(`(` + compName + `-\S+)\s+\S+\s+Running.*deploymentconfig=` + compName)
	podName := re.FindStringSubmatch(stdOut)[1]
	return strings.TrimSpace(podName)
}

// GetRunningPodNameByComponent executes cluster command and returns the running pod name of a delopyed
// devfile component by passing component name as a argument
func (cluster *ClusterRunner) GetRunningPodNameByComponent(compName string, namespace string) string {
	stdOut := CmdShouldPass(cluster.path, "get", "pods", "--namespace", namespace, "--show-labels")
	re := regexp.MustCompile(`(` + compName + `-\S+)\s+\S+\s+Running.*component=` + compName)
	podName := re.FindStringSubmatch(stdOut)[1]
	return strings.TrimSpace(podName)
}

// GetRoute returns route URL
func (cluster *ClusterRunner) GetRoute(urlName string, appName string) string {
	session := CmdRunner(cluster.path, "get", "routes", urlName+"-"+appName,
		"-o jsonpath={.spec.host}")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// GetToken returns current user token
func (cluster *ClusterRunner) GetToken() string {
	session := CmdRunner(cluster.path, "whoami", "-t")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// LoginUsingToken returns output after successful login
func (cluster *ClusterRunner) LoginUsingToken(token string) string {
	session := CmdRunner(cluster.path, "login", "--token", token)
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Wait().Out.Contents()))
}

// GetLoginUser returns current user name
func (cluster *ClusterRunner) GetLoginUser() string {
	user := CmdShouldPass(cluster.path, "whoami")
	return strings.TrimSpace(user)
}

// ServiceInstanceStatus returns service instance
func (cluster *ClusterRunner) ServiceInstanceStatus(serviceInstanceName string) string {
	serviceinstance := CmdShouldPass(cluster.path, "get", "serviceinstance", serviceInstanceName,
		"-o", "go-template='{{ (index .status.conditions 0).reason}}'")
	return strings.TrimSpace(serviceinstance)
}

// GetVolumeMountNamesandPathsFromContainer returns the volume name and mount path in the format name:path\n
func (cluster *ClusterRunner) GetVolumeMountNamesandPathsFromContainer(deployName string, containerName, namespace string) string {
	volumeName := CmdShouldPass(cluster.path, "get", "deploy", deployName, "--namespace", namespace,
		"-o", "go-template="+
			"{{range .spec.template.spec.containers}}{{if eq .name \""+containerName+
			"\"}}{{range .volumeMounts}}{{.name}}{{\":\"}}{{.mountPath}}{{\"\\n\"}}{{end}}{{end}}{{end}}")

	return strings.TrimSpace(volumeName)
}

// GetVolumeMountName returns the name of the volume
func (cluster *ClusterRunner) GetVolumeMountName(dcName string, namespace string) string {
	volumeName := CmdShouldPass(cluster.path, "get", "dc", dcName, "--namespace", namespace,
		"-o", "go-template='"+
			"{{range .spec.template.spec.containers}}"+
			"{{range .volumeMounts}}{{.name}}{{end}}{{end}}'")

	return strings.TrimSpace(volumeName)
}

// GetVolumeMountPath returns the path of the volume mount
func (cluster *ClusterRunner) GetVolumeMountPath(dcName string, namespace string) string {
	volumePaths := CmdShouldPass(cluster.path, "get", "dc", dcName, "--namespace", namespace,
		"-o", "go-template='"+
			"{{range .spec.template.spec.containers}}"+
			"{{range .volumeMounts}}{{.mountPath}} {{end}}{{end}}'")

	return strings.TrimSpace(volumePaths)
}

// GetEnvFromEntry returns envFrom entry
func (cluster *ClusterRunner) GetEnvFromEntry(componentName string, appName string, projectName string) string {
	envFromOut := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", projectName,
		"-o", "jsonpath='{.spec.template.spec.containers[0].envFrom}'")
	return strings.TrimSpace(envFromOut)
}

// GetEnvs returns all env variables in deployment config
func (cluster *ClusterRunner) GetEnvs(componentName string, appName string, projectName string) map[string]string {
	var mapOutput = make(map[string]string)

	output := CmdShouldPass(cluster.path, "get", "dc", componentName+"-"+appName, "--namespace", projectName,
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
func (cluster *ClusterRunner) WaitForDCRollout(dcName string, project string, timeout time.Duration) {
	session := CmdRunner(cluster.path, "rollout", "status",
		"-w",
		"-n", project,
		"dc", dcName)

	Eventually(session).Should(gexec.Exit(0), runningCmd(session.Command))
	session.Wait(timeout)
}

// WaitAndCheckForExistence wait for the given and checks if the given resource type gets deleted on the cluster
func (cluster *ClusterRunner) WaitAndCheckForExistence(resourceType, namespace string, timeoutMinutes int) bool {
	pingTimeout := time.After(time.Duration(timeoutMinutes) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout out after %v minutes", timeoutMinutes))

		case <-tick:
			session := CmdRunner(cluster.path, "get", resourceType, "--namespace", namespace)
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
func (cluster *ClusterRunner) GetServices(namespace string) string {
	session := CmdRunner(cluster.path, "get", "services", "--namespace", namespace)
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	return output
}

// VerifyResourceDeleted verifies if the given resource is deleted from cluster
func (cluster *ClusterRunner) VerifyResourceDeleted(resourceType, resourceName, namespace string) {
	session := CmdRunner(cluster.path, "get", resourceType, "--namespace", namespace)
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	Expect(output).NotTo(ContainSubstring(resourceName))
}
