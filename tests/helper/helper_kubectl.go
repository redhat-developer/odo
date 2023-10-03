package helper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/redhat-developer/odo/pkg/labels"
)

const (
	ResourceTypeDeployment = "deployment"
	ResourceTypePod        = "pod"
	ResourceTypeJob        = "job"
	ResourceTypePVC        = "pvc"
	ResourceTypeService    = "service"
)

type KubectlRunner struct {
	// path to kubectl binary
	path string
}

// NewKubectlRunner initializes new KubectlRunner
func NewKubectlRunner(kubectlPath string) KubectlRunner {
	return KubectlRunner{
		path: kubectlPath,
	}
}

// Run kubectl with given arguments
func (kubectl KubectlRunner) Run(args ...string) *gexec.Session {
	session := CmdRunner(kubectl.path, args...)
	Eventually(session).Should(gexec.Exit(0))
	return session
}

// Exec allows generic execution of commands, returning the contents of stdout
func (kubectl KubectlRunner) Exec(podName string, projectName string, args []string, expectedSuccess *bool) (string, string) {

	cmd := []string{"exec", podName, "--namespace", projectName}

	cmd = append(cmd, args...)

	cmdWrapper := Cmd(kubectl.path, cmd...)
	if expectedSuccess == nil {
		cmdWrapper = cmdWrapper.ShouldRun()
	} else if *expectedSuccess {
		cmdWrapper = cmdWrapper.ShouldPass()
	} else {
		cmdWrapper = cmdWrapper.ShouldFail()
	}
	return cmdWrapper.OutAndErr()
}

// ExecListDir returns dir list in specified location of pod
func (kubectl KubectlRunner) ExecListDir(podName string, projectName string, dir string) string {
	stdOut := Cmd(kubectl.path, "exec", podName, "--namespace", projectName,
		"--", "ls", "-lai", dir).ShouldPass().Out()
	return stdOut
}

// CheckCmdOpInRemoteDevfilePod runs the provided command on remote component pod and returns the return value of command output handler function passed to it
func (kubectl KubectlRunner) CheckCmdOpInRemoteDevfilePod(podName string, containerName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool {
	var execOptions []string
	execOptions = []string{"exec", podName, "--namespace", prjName, "--"}
	if containerName != "" {
		execOptions = []string{"exec", podName, "-c", containerName, "--namespace", prjName, "--"}
	}
	args := append(execOptions, cmd...)
	session := CmdRunner(kubectl.path, args...)
	stdOut := string(session.Wait().Out.Contents())
	stdErr := string(session.Wait().Err.Contents())
	if stdErr != "" && session.ExitCode() != 0 {
		return checkOp(stdOut, fmt.Errorf("cmd %s failed with error %s on pod %s", cmd, stdErr, podName))
	}
	return checkOp(stdOut, nil)
}

// GetRunningPodNameByComponent executes kubectl command and returns the running pod name of a deployed
// devfile component by passing component name as a argument
func (kubectl KubectlRunner) GetRunningPodNameByComponent(compName string, namespace string) string {
	selector := fmt.Sprintf("--selector=component=%s", compName)
	stdOut := Cmd(kubectl.path, "get", ResourceTypePod, "--namespace", namespace, "--field-selector=status.phase=Running", selector, "-o", "jsonpath={.items[*].metadata.name}").ShouldPass().Out()
	return strings.TrimSpace(stdOut)
}

// GetJobNameByComponent executes kubectl command and returns the running job name
func (kubectl KubectlRunner) GetJobNameByComponent(compName string, namespace string) string {
	selector := fmt.Sprintf("--selector=app.kubernetes.io/instance=%s", compName)
	stdOut := Cmd(kubectl.path, "get", ResourceTypeJob, "--namespace", namespace, selector, "-o", "jsonpath={.items[*].metadata.name}").ShouldPass().Out()
	return strings.TrimSpace(stdOut)
}

// GetPVCSize executes kubectl command and returns the bound storage size
func (kubectl KubectlRunner) GetPVCSize(compName, storageName, namespace string) string {
	selector := fmt.Sprintf("--selector=app.kubernetes.io/storage-name=%s,app.kubernetes.io/instance=%s", storageName, compName)
	stdOut := Cmd(kubectl.path, "get", ResourceTypePVC, "--namespace", namespace, selector, "-o", "jsonpath={.items[*].spec.resources.requests.storage}").ShouldPass().Out()
	return strings.TrimSpace(stdOut)
}

// GetPodInitContainers executes kubectl command and returns the init containers of the pod
func (kubectl KubectlRunner) GetPodInitContainers(compName string, namespace string) []string {
	selector := fmt.Sprintf("--selector=component=%s", compName)
	stdOut := Cmd(kubectl.path, "get", ResourceTypePod, "--namespace", namespace, "--field-selector=status.phase=Running", selector, "-o", "jsonpath={.items[*].spec.initContainers[*].name}").ShouldPass().Out()
	return strings.Split(stdOut, " ")
}

// GetVolumeMountNamesandPathsFromContainer returns the volume name and mount path in the format name:path\n
func (kubectl KubectlRunner) GetVolumeMountNamesandPathsFromContainer(deployName string, containerName, namespace string) string {
	volumeName := Cmd(kubectl.path, "get", "deploy", deployName, "--namespace", namespace,
		"-o", "go-template="+
			"{{range .spec.template.spec.containers}}{{if eq .name \""+containerName+
			"\"}}{{range .volumeMounts}}{{.name}}{{\":\"}}{{.mountPath}}{{\"\\n\"}}{{end}}{{end}}{{end}}").ShouldPass().Out()

	return strings.TrimSpace(volumeName)
}

// GetContainerEnv returns the container env in the format name:value\n
func (kubectl KubectlRunner) GetContainerEnv(podName, containerName, namespace string) string {
	containerEnv := Cmd(kubectl.path, "get", "po", podName, "--namespace", namespace,
		"-o", "go-template="+
			"{{range .spec.containers}}{{if eq .name \""+containerName+
			"\"}}{{range .env}}{{.name}}{{\":\"}}{{.value}}{{\"\\n\"}}{{end}}{{end}}{{end}}").ShouldPass().Out()

	return strings.TrimSpace(containerEnv)
}

// WaitAndCheckForExistence wait for the given and checks if the given resource type gets deleted on the cluster
func (kubectl KubectlRunner) WaitAndCheckForExistence(resourceType, namespace string, timeoutMinutes int) bool {
	pingTimeout := time.After(time.Duration(timeoutMinutes) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout after %d minutes", timeoutMinutes))

		case <-tick:
			session := CmdRunner(kubectl.path, "get", resourceType, "--namespace", namespace)
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
func (kubectl KubectlRunner) GetServices(namespace string) string {
	session := CmdRunner(kubectl.path, "get", ResourceTypeService, "--namespace", namespace)
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	return output
}

// CreateAndSetRandNamespaceProject create and set new project
func (kubectl KubectlRunner) CreateAndSetRandNamespaceProject() string {
	projectName := GenerateProjectName()
	kubectl.createAndSetRandNamespaceProject(projectName)
	return projectName
}

func (kubectl KubectlRunner) createAndSetRandNamespaceProject(projectName string) string {
	if kubectl.HasNamespaceProject(projectName) {
		fmt.Fprintf(GinkgoWriter, "Namespace %q already exists\n", projectName)
	} else {
		fmt.Fprintf(GinkgoWriter, "Creating a new project: %s\n", projectName)
		Cmd("kubectl", "create", "namespace", projectName).ShouldPass()
	}
	Cmd("kubectl", "config", "set-context", "--current", "--namespace", projectName).ShouldPass()
	// ListNamespaceProject makes sure that project eventually appears in the list of all namespaces/projects.
	kubectl.ListNamespaceProject(projectName)
	kubectl.addConfigMapForCleanup(projectName) // add configmap for cleanup
	return projectName
}

func (kubectl KubectlRunner) SetProject(namespace string) string {
	Cmd("kubectl", "config", "set-context", "--current", "--namespace", namespace).ShouldPass()
	session := Cmd("kubectl", "get", "namespaces").ShouldPass().Out()
	Expect(session).To(ContainSubstring(namespace))
	return namespace
}

// CreateRandNamespaceProjectOfLength create new project with i as the length of the name and sets it to the current context
func (kubectl KubectlRunner) CreateAndSetRandNamespaceProjectOfLength(i int) string {
	projectName := RandString(i)
	kubectl.createAndSetRandNamespaceProject(projectName)
	return projectName
}

// DeleteNamespaceProject deletes a specified project in kubernetes cluster
func (kubectl KubectlRunner) DeleteNamespaceProject(projectName string, wait bool) {
	fmt.Fprintf(GinkgoWriter, "Deleting project: %s\n", projectName)
	Cmd("kubectl", "delete", "namespaces", projectName, "--wait="+strconv.FormatBool(wait)).ShouldPass()
}

func (kubectl KubectlRunner) GetEnvsDevFileDeployment(componentName, appName, projectName string) map[string]string {
	var mapOutput = make(map[string]string)
	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).SelectorFlag()
	output := Cmd(kubectl.path, "get", ResourceTypeDeployment, selector, "--namespace", projectName,
		"-o", "jsonpath='{range .items[0].spec.template.spec.containers[0].env[*]}{.name}:{.value}{\"\\n\"}{end}'").ShouldPass().Out()

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimPrefix(line, "'")
		splits := strings.Split(line, ":")
		name := splits[0]
		value := strings.Join(splits[1:], ":")
		mapOutput[name] = value
	}
	return mapOutput
}

func (kubectl KubectlRunner) GetAllPVCNames(namespace string) []string {
	session := CmdRunner(kubectl.path, "get", ResourceTypePVC, "--namespace", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	if output == "" {
		return []string{}
	}
	return strings.Split(output, " ")
}

// DeletePod deletes a specified pod in the namespace
func (kubectl KubectlRunner) DeletePod(podName string, namespace string) {
	Cmd(kubectl.path, "delete", ResourceTypePod, "--namespace", namespace, podName).ShouldPass()
}

// WaitAndCheckForTerminatingState waits for the given interval
// and checks if the given resource type has been deleted on the cluster or is in the terminating state
func (kubectl KubectlRunner) WaitAndCheckForTerminatingState(resourceType, namespace string, timeoutMinutes int) bool {
	return WaitAndCheckForTerminatingState(kubectl.path, resourceType, namespace, timeoutMinutes)
}

// VerifyResourceDeleted verifies if the given resource is deleted from cluster.
func (kubectl KubectlRunner) VerifyResourceDeleted(ri ResourceInfo) {
	session := CmdRunner(kubectl.path, "get", ri.ResourceType, "--namespace", ri.Namespace)
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	Expect(output).NotTo(ContainSubstring(ri.ResourceName))
}

// VerifyResourceToBeDeleted verifies if a resource if deleted, or if not, if it is marked for deletion
func (kubectl KubectlRunner) VerifyResourceToBeDeleted(ri ResourceInfo) {
	deletedOrMarkedToDelete := func() bool {
		session := CmdRunner(kubectl.path, "get", ri.ResourceType, ri.ResourceName, "--namespace", ri.Namespace, "-o", "jsonpath='{.metadata.deletionTimestamp}'")
		exit := session.Wait().ExitCode()
		if exit == 1 {
			// resources does not exist
			return true
		}
		content := session.Wait().Out.Contents()
		// resource is marked for deletion
		return len(content) > 0
	}
	Expect(deletedOrMarkedToDelete()).To(BeTrue())
}

// GetAnnotationsDeployment gets the annotations from the deployment
// belonging to the given component, app and project
func (kubectl KubectlRunner) GetAnnotationsDeployment(componentName, appName, projectName string) map[string]string {
	return GetAnnotationsDeployment(kubectl.path, componentName, appName, projectName)
}

// GetAllPodsInNs gets the list of pods in given namespace. It waits for reasonable amount of time for pods to come up
func (kubectl KubectlRunner) GetAllPodsInNs(namespace string) string {
	args := []string{"get", ResourceTypePod, "-n", namespace}
	noResourcesMsg := fmt.Sprintf("No resources found in %s namespace", namespace)
	kubectl.WaitForRunnerCmdOut(args, 1, true, func(output string) bool {
		return !strings.Contains(output, noResourcesMsg)
	}, true)
	return Cmd(kubectl.path, args...).ShouldPass().Out()
}

// GetAllPodNames gets the names of pods in given namespace
func (kubectl KubectlRunner) GetAllPodNames(namespace string) []string {
	session := CmdRunner(kubectl.path, "get", "pods", "--namespace", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	if output == "" {
		return []string{}
	}
	return strings.Split(output, " ")
}

func (kubectl KubectlRunner) PodsShouldBeRunning(project string, regex string) {
	// now verify if the pods for the operator have started
	pods := kubectl.GetAllPodsInNs(project)
	// Look for pods with specified regex
	pod := regexp.MustCompile(regex).FindString(pods)

	args := []string{"get", ResourceTypePod, pod, "-o", "template=\"{{.status.phase}}\"", "-n", project}
	kubectl.WaitForRunnerCmdOut(args, 1, true, func(output string) bool {
		return strings.Contains(output, "Running")
	})
}

// WaitForRunnerCmdOut runs "kubectl" command until it gets
// the expected output.
// It accepts 4 arguments
// args (arguments to the program)
// timeout (the time to wait for the output)
// errOnFail (flag to set if test should fail if command fails)
// check (function with output check logic)
// It times out if the command doesn't fetch the
// expected output  within the timeout period.
func (kubectl KubectlRunner) WaitForRunnerCmdOut(args []string, timeout int, errOnFail bool, check func(output string) bool, includeStdErr ...bool) bool {
	pingTimeout := time.After(time.Duration(timeout) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout after %v minutes", timeout))

		case <-tick:
			session := CmdRunner(kubectl.path, args...)
			if errOnFail {
				Eventually(session).Should(gexec.Exit(0), runningCmd(session.Command))
			} else {
				Eventually(session).Should(gexec.Exit(), runningCmd(session.Command))
			}
			session.Wait()
			output := string(session.Out.Contents())

			if len(includeStdErr) > 0 && includeStdErr[0] {
				output += "\n"
				output += string(session.Err.Contents())
			}
			if check(strings.TrimSpace(output)) {
				return true
			}
		}
	}
}

// CreateSecret takes secret name, password and the namespace where we want to create the specific secret into the cluster
func (kubectl KubectlRunner) CreateSecret(secretName, secretPass, project string) {
	Cmd(kubectl.path, "create", "secret", "generic", secretName, "--from-literal=password="+secretPass, "-n", project).ShouldPass()
}

// GetSecrets gets all the secrets belonging to the project
func (kubectl KubectlRunner) GetSecrets(project string) string {
	return GetSecrets(kubectl.path, project)
}

// GetEnvRefNames gets the ref values from the envFroms of the deployment belonging to the given data
func (kubectl KubectlRunner) GetEnvRefNames(componentName, appName, projectName string) []string {
	return GetEnvRefNames(kubectl.path, componentName, appName, projectName)
}

// GetEnvFromEntry returns envFrom entry of the deployment
func (kubectl KubectlRunner) GetEnvFromEntry(componentName string, appName string, projectName string) string {
	return GetEnvFromEntry(kubectl.path, componentName, appName, projectName)
}

// GetVolumeNamesFromDeployment gets the volumes from the deployment belonging to the given data
func (kubectl KubectlRunner) GetVolumeNamesFromDeployment(componentName, appName, projectName string) map[string]string {
	return GetVolumeNamesFromDeployment(kubectl.path, componentName, appName, projectName)
}

// add config map to the project for cleanup
func (kubectl KubectlRunner) addConfigMapForCleanup(projectName string) {
	Cmd(kubectl.path, "create", "configmap", "config-map-for-cleanup", "--from-literal", "type=testing", "--from-literal", "team=odo", "-n", projectName).ShouldPass()
}

// ScalePodToZero scales the pod of the deployment to zero.
// It waits for the pod to get deleted from the cluster before returning
func (kubectl KubectlRunner) ScalePodToZero(componentName, appName, projectName string) {
	podName := kubectl.GetRunningPodNameByComponent(componentName, projectName)
	Cmd(kubectl.path, "scale", "deploy", strings.Join([]string{componentName, appName}, "-"), "--replicas=0").ShouldPass()
	kubectl.WaitForRunnerCmdOut([]string{"get", "-n", projectName, ResourceTypePod, podName}, 1, false, func(output string) bool {
		return !strings.Contains(output, podName)
	})
}

func (kubectl KubectlRunner) GetBindableKinds() (string, string) {
	return Cmd(kubectl.path, "get", "bindablekinds", "bindable-kinds", "-ojsonpath='{.status[*].kind}'").ShouldRun().OutAndErr()
}

func (kubectl KubectlRunner) GetServiceBinding(name, projectName string) (string, string) {
	return Cmd(kubectl.path, "get", "servicebinding", name, "-n", projectName).ShouldRun().OutAndErr()
}

func (kubectl KubectlRunner) EnsureOperatorIsInstalled(partialOperatorName string) {
	WaitForCmdOut(kubectl.path, []string{"get", "csv", "-o", "jsonpath={.items[?(@.status.phase==\"Succeeded\")].metadata.name}"}, 4, true, func(output string) bool {
		return strings.Contains(output, partialOperatorName)
	})
}

func (kubectl KubectlRunner) EnsurePodIsUp(namespace, podName string) {
	WaitForCmdOut(kubectl.path, []string{"get", "pods", "-n", namespace, "-o", "jsonpath='{range .items[*]}{.metadata.name}'"}, 4, true, func(output string) bool {
		return strings.Contains(output, podName)
	})
}

func (kubectl KubectlRunner) GetNamespaceProject() string {
	return Cmd(kubectl.path, "get", "namespace").ShouldPass().Out()
}

func (kubectl KubectlRunner) HasNamespaceProject(name string) bool {
	out := Cmd(kubectl.path, "get", "namespace", name, "-o", "jsonpath={.metadata.name}").
		ShouldRun().Out()
	return strings.Contains(out, name)
}

func (kubectl KubectlRunner) ListNamespaceProject(name string) {
	Eventually(func() string {
		return Cmd(kubectl.path, "get", "ns").ShouldRun().Out()
	}, 30, 1).Should(ContainSubstring(name))
}

func (kubectl KubectlRunner) GetActiveNamespace() string {
	return Cmd(kubectl.path, "config", "view", "--minify", "-ojsonpath={..namespace}").ShouldPass().Out()
}

func (kubectl KubectlRunner) GetAllNamespaceProjects() []string {
	output := Cmd(kubectl.path, "get", "namespaces",
		"-o", "custom-columns=NAME:.metadata.name",
		"--no-headers").ShouldPass().Out()
	result, err := ExtractLines(output)
	Expect(err).ShouldNot(HaveOccurred())
	return result
}

func (kubectl KubectlRunner) GetLogs(podName string) string {
	output := Cmd(kubectl.path, "logs", podName).ShouldPass().Out()
	return output
}

func (kubectl KubectlRunner) AssertContainsLabel(kind, namespace, componentName, appName, mode, key, value string) {
	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).WithMode(mode).SelectorFlag()
	all := Cmd(kubectl.path, "get", kind, selector, "-n", namespace, "-o", "jsonpath={.items[0].metadata.labels}").ShouldPass().Out()
	Expect(all).To(ContainSubstring(fmt.Sprintf(`"%s":"%s"`, key, value)))
}

func (kubectl KubectlRunner) AssertNoContainsLabel(kind, namespace, componentName, appName, mode, key string) {
	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).WithMode(mode).SelectorFlag()
	all := Cmd(kubectl.path, "get", kind, selector, "-n", namespace, "-o", "jsonpath={.items[0].metadata.labels}").ShouldPass().Out()
	Expect(all).ToNot(ContainSubstring(fmt.Sprintf(`"%s"`, key)))
}

func (kubectl KubectlRunner) AssertNonAuthenticated() {
	// Nothing to do
}

func (kubectl KubectlRunner) GetVersion() string {
	res := Cmd(kubectl.path, "version", "--output=json").ShouldPass().Out()
	var js map[string]interface{}
	err := json.Unmarshal([]byte(res), &js)
	Expect(err).ShouldNot(HaveOccurred())
	sv := js["serverVersion"].(map[string]interface{})
	minor := sv["minor"].(string)
	major := sv["major"].(string)
	return major + "." + minor
}

func (kubectl KubectlRunner) SetLabelsOnNamespace(ns string, labelValues ...string) {
	args := []string{"label", "namespaces", ns}
	args = append(args, labelValues...)
	Cmd(kubectl.path, args...).ShouldPass()
}
