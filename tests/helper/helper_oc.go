package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/redhat-developer/odo/pkg/labels"
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
func (oc OcRunner) Run(args ...string) *gexec.Session {
	session := CmdRunner(oc.path, args...)
	Eventually(session).Should(gexec.Exit(0))
	return session
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

// ExecListDir returns dir list in specified location of pod
func (oc OcRunner) ExecListDir(podName string, projectName string, dir string) string {
	stdOut := Cmd(oc.path, "exec", podName, "--namespace", projectName,
		"--", "ls", "-lai", dir).ShouldPass().Out()
	return stdOut
}

// Exec allows generic execution of commands, returning the contents of stdout
func (oc OcRunner) Exec(podName string, projectName string, args ...string) string {

	cmd := []string{"exec", podName, "--namespace", projectName}

	cmd = append(cmd, args...)

	stdOut := Cmd(oc.path, cmd...).ShouldPass().Out()
	return stdOut
}

// CheckCmdOpInRemoteCmpPod runs the provided command on remote component pod and returns the return value of command output handler function passed to it
func (oc OcRunner) CheckCmdOpInRemoteCmpPod(cmpName string, appName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool {
	cmpDCName := fmt.Sprintf("%s-%s", cmpName, appName)
	outPodName := Cmd(oc.path, "get", "pods", "--namespace", prjName,
		"--selector=deploymentconfig="+cmpDCName,
		"-o", "jsonpath='{.items[0].metadata.name}'").ShouldPass().Out()
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

// GetRunningPodNameOfComp executes oc command and returns the running pod name of a deployed
// component by passing component name as a argument
func (oc OcRunner) GetRunningPodNameOfComp(compName string, namespace string) string {
	stdOut := Cmd(oc.path, "get", "pods", "--namespace", namespace, "--show-labels").ShouldPass().Out()
	re := regexp.MustCompile(`(` + compName + `-\S+)\s+\S+\s+Running.*deploymentconfig=` + compName)
	podName := re.FindStringSubmatch(stdOut)[1]
	return strings.TrimSpace(podName)
}

// GetRunningPodNameByComponent executes oc command and returns the running pod name of a deployed
// devfile component by passing component name as a argument
func (oc OcRunner) GetRunningPodNameByComponent(compName string, namespace string) string {
	selector := fmt.Sprintf("--selector=component=%s", compName)
	stdOut := Cmd(oc.path, "get", "pods", "--namespace", namespace, selector, "-o", "jsonpath={.items[*].metadata.name}").ShouldPass().Out()
	return strings.TrimSpace(stdOut)
}

// GetPVCSize executes oc command and returns the bound storage size
func (oc OcRunner) GetPVCSize(compName, storageName, namespace string) string {
	selector := fmt.Sprintf("--selector=app.kubernetes.io/storage-name=%s,app.kubernetes.io/instance=%s", storageName, compName)
	stdOut := Cmd(oc.path, "get", "pvc", "--namespace", namespace, selector, "-o", "jsonpath={.items[*].spec.resources.requests.storage}").ShouldPass().Out()
	return strings.TrimSpace(stdOut)
}

// GetPodInitContainers executes oc command and returns the init containers of the pod
func (oc OcRunner) GetPodInitContainers(compName string, namespace string) []string {
	selector := fmt.Sprintf("--selector=component=%s", compName)
	stdOut := Cmd(oc.path, "get", "pods", "--namespace", namespace, selector, "-o", "jsonpath={.items[*].spec.initContainers[*].name}").ShouldPass().Out()
	return strings.Split(stdOut, " ")
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
	user := Cmd(oc.path, "whoami").ShouldPass().Out()
	return strings.TrimSpace(user)
}

// ServiceInstanceStatus returns service instance
func (oc OcRunner) ServiceInstanceStatus(serviceInstanceName string) string {
	serviceinstance := Cmd(oc.path, "get", "serviceinstance", serviceInstanceName,
		"-o", "go-template='{{ (index .status.conditions 0).reason}}'").ShouldPass().Out()
	return strings.TrimSpace(serviceinstance)
}

// GetVolumeMountNamesandPathsFromContainer returns the volume name and mount path in the format name:path\n
func (oc OcRunner) GetVolumeMountNamesandPathsFromContainer(deployName string, containerName, namespace string) string {
	volumeName := Cmd(oc.path, "get", "deploy", deployName, "--namespace", namespace,
		"-o", "go-template="+
			"{{range .spec.template.spec.containers}}{{if eq .name \""+containerName+
			"\"}}{{range .volumeMounts}}{{.name}}{{\":\"}}{{.mountPath}}{{\"\\n\"}}{{end}}{{end}}{{end}}").ShouldPass().Out()

	return strings.TrimSpace(volumeName)
}

// GetContainerEnv returns the container env in the format name:value\n
func (oc OcRunner) GetContainerEnv(podName, containerName, namespace string) string {
	containerEnv := Cmd(oc.path, "get", "po", podName, "--namespace", namespace,
		"-o", "go-template="+
			"{{range .spec.containers}}{{if eq .name \""+containerName+
			"\"}}{{range .env}}{{.name}}{{\":\"}}{{.value}}{{\"\\n\"}}{{end}}{{end}}{{end}}").ShouldPass().Out()

	return strings.TrimSpace(containerEnv)
}

// GetEnvFromEntry returns envFrom entry of the deployment
func (oc OcRunner) GetEnvFromEntry(componentName string, appName string, projectName string) string {
	return GetEnvFromEntry(oc.path, componentName, appName, projectName)
}

func (oc OcRunner) GetEnvsDevFileDeployment(componentName, appName, projectName string) map[string]string {
	var mapOutput = make(map[string]string)

	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).SelectorFlag()
	output := Cmd(oc.path, "get", "deployment", selector, "--namespace", projectName,
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

// GetEnvRefNames gets the ref values from the envFroms of the deployment belonging to the given data
func (oc OcRunner) GetEnvRefNames(componentName, appName, projectName string) []string {
	return GetEnvRefNames(oc.path, componentName, appName, projectName)
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
func (oc OcRunner) VerifyResourceDeleted(ri ResourceInfo) {
	session := CmdRunner(oc.path, "get", ri.ResourceType, "--namespace", ri.Namespace)
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	Expect(output).NotTo(ContainSubstring(ri.ResourceName))
}

func (oc OcRunner) VerifyResourceToBeDeleted(ri ResourceInfo) {
	deletedOrMarkedToDelete := func() bool {
		session := CmdRunner(oc.path, "get", ri.ResourceType, ri.ResourceName, "--namespace", ri.Namespace, "-o", "jsonpath='{.metadata.deletionTimestamp}'")
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

// CreateAndSetRandNamespaceProject create and set new project
func (oc OcRunner) CreateAndSetRandNamespaceProject() string {
	projectName := GetProjectName()
	oc.createAndSetRandNamespaceProject(projectName)
	return projectName
}

// CreateAndSetRandNamespaceProjectOfLength creates a new project with name of length i and sets it to the current context
func (oc OcRunner) CreateAndSetRandNamespaceProjectOfLength(i int) string {
	projectName := RandString(i)
	oc.createAndSetRandNamespaceProject(projectName)
	return projectName
}

func (oc OcRunner) createAndSetRandNamespaceProject(projectName string) string {
	fmt.Fprintf(GinkgoWriter, "Creating a new project: %s\n", projectName)
	session := Cmd(oc.path, "new-project", projectName).ShouldPass().Out()
	Expect(session).To(ContainSubstring(projectName))
	oc.addConfigMapForCleanup(projectName)
	return projectName
}

func (oc OcRunner) SetProject(namespace string) string {
	fmt.Fprintf(GinkgoWriter, "Setting project: %s\n", namespace)
	Cmd("odo", "set", "project", namespace).ShouldPass()
	return namespace
}

// DeleteNamespaceProject deletes a specified project in oc cluster
func (oc OcRunner) DeleteNamespaceProject(projectName string, wait bool) {
	fmt.Fprintf(GinkgoWriter, "Deleting project: %s\n", projectName)
	Cmd(oc.path, "delete", "project", projectName, "--wait="+strconv.FormatBool(wait)).ShouldPass()
}

func (oc OcRunner) GetAllPVCNames(namespace string) []string {
	session := CmdRunner(oc.path, "get", "pvc", "--namespace", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	if output == "" {
		return []string{}
	}
	return strings.Split(output, " ")
}

// DeletePod deletes a specified pod in the namespace
func (oc OcRunner) DeletePod(podName string, namespace string) {
	Cmd(oc.path, "delete", "pod", "--namespace", namespace, podName).ShouldPass()
}

// GetAllPodsInNs gets the list of pods in given namespace. It waits for reasonable amount of time for pods to come up
func (oc OcRunner) GetAllPodsInNs(namespace string) string {
	args := []string{"get", "pods", "-n", namespace}
	noResourcesMsg := fmt.Sprintf("No resources found in %s namespace", namespace)
	oc.WaitForRunnerCmdOut(args, 1, true, func(output string) bool {
		return !strings.Contains(output, noResourcesMsg)
	}, true)
	return Cmd(oc.path, args...).ShouldPass().Out()
}

// GetAllPodNames gets the names of pods in given namespace
func (oc OcRunner) GetAllPodNames(namespace string) []string {
	session := CmdRunner(oc.path, "get", "pods", "--namespace", namespace, "-o", "jsonpath={.items[*].metadata.name}")
	Eventually(session).Should(gexec.Exit(0))
	output := string(session.Wait().Out.Contents())
	if output == "" {
		return []string{}
	}
	return strings.Split(output, " ")
}

// StatFileInPod returns stat result of filepath in pod of given component, in a given app, in a given project.
// It also strips access time information as it vaires accross file systems/kernel configs, and we are not interested
// in it anyway
func (oc OcRunner) StatFileInPod(cmpName, appName, project, filepath string) string {
	var result string
	oc.CheckCmdOpInRemoteCmpPod(
		cmpName,
		appName,
		project,
		[]string{"stat", filepath},
		func(cmdOp string, err error) bool {
			// strip out access info as
			// 1. Touching a file (such as running it in a script) modifies access times. This gives wrong value on mounts without noatime
			// 2. We are not interested in Access info anyway.
			re := regexp.MustCompile("(?m)[\r\n]+^.*Access.*$")
			result = re.ReplaceAllString(cmdOp, "")
			return true
		},
	)
	return result
}

// WaitAndCheckForTerminatingState waits for the given interval
// and checks if the given resource type has been deleted on the cluster or is in the terminating state
func (oc OcRunner) WaitAndCheckForTerminatingState(resourceType, namespace string, timeoutMinutes int) bool {
	return WaitAndCheckForTerminatingState(oc.path, resourceType, namespace, timeoutMinutes)
}

// GetAnnotationsDeployment gets the annotations from the deployment
// belonging to the given component, app and project
func (oc OcRunner) GetAnnotationsDeployment(componentName, appName, projectName string) map[string]string {
	return GetAnnotationsDeployment(oc.path, componentName, appName, projectName)
}

func (oc OcRunner) PodsShouldBeRunning(project string, regex string) {
	// now verify if the pods for the operator have started
	pods := oc.GetAllPodsInNs(project)
	// Look for pods with specified regex
	pod := regexp.MustCompile(regex).FindString(pods)
	args := []string{"get", "pods", pod, "-o", "template=\"{{.status.phase}}\"", "-n", project}
	oc.WaitForRunnerCmdOut(args, 8, true, func(output string) bool {
		return strings.Contains(output, "Running")
	})
}

// WaitForRunnerCmdOut runs "oc" command until it gets
// the expected output.
// It accepts 4 arguments
// args (arguments to the program)
// timeout (the time to wait for the output)
// errOnFail (flag to set if test should fail if command fails)
// check (function with output check logic)
// It times out if the command doesn't fetch the
// expected output  within the timeout period.
func (oc OcRunner) WaitForRunnerCmdOut(args []string, timeout int, errOnFail bool, check func(output string) bool, includeStdErr ...bool) bool {
	pingTimeout := time.After(time.Duration(timeout) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout after %v minutes", timeout))

		case <-tick:
			session := CmdRunner(oc.path, args...)
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
func (oc OcRunner) CreateSecret(secretName, secretPass, project string) {
	Cmd(oc.path, "create", "secret", "generic", secretName, "--from-literal=password="+secretPass, "-n", project).ShouldPass()
}

// GetSecrets gets all the secrets belonging to the project
func (oc OcRunner) GetSecrets(project string) string {
	return GetSecrets(oc.path, project)
}

// GetVolumeNamesFromDeployment gets the volumes from the deployment belonging to the given data
func (oc OcRunner) GetVolumeNamesFromDeployment(componentName, appName, projectName string) map[string]string {
	return GetVolumeNamesFromDeployment(oc.path, componentName, appName, projectName)
}

// AddSecret adds pull-secret to the namespace, for e2e-test
func (oc OcRunner) AddSecret(comvar CommonVar) {

	clusterType := os.Getenv("CLUSTER_TYPE")
	if clusterType == "PSI" || clusterType == "IBM" {

		token := oc.doAsAdmin(clusterType)

		yaml := Cmd(oc.path, "get", "secret", "pull-secret", "-n", "openshift-config", "-o", "yaml").ShouldPass().Out()
		newYaml := strings.Replace(yaml, "openshift-config", comvar.Project, -1)
		filename := fmt.Sprint(RandString(4), ".yaml")
		newYamlinByte := []byte(newYaml)
		err := ioutil.WriteFile(filename, newYamlinByte, 0600)
		if err != nil {
			fmt.Println(err)
		}
		Cmd(oc.path, "apply", "-f", filename).ShouldPass()
		os.Remove(filename)
		oc.doAsDeveloper(token, clusterType)
	}

}

// doAsAdmin logins as admin to perform some task that requires admin privileges
func (oc OcRunner) doAsAdmin(clusterType string) string {
	// save token for developer
	token := oc.GetToken()
	if clusterType == "PSI" || clusterType == "IBM" {

		adminToken := os.Getenv("IBMC_OCLOGIN_APIKEY")
		if adminToken != "" {
			ibmcloudAdminToken := os.Getenv("IBMC_ADMIN_LOGIN_APIKEY")
			cluster := os.Getenv("IBMC_OCP47_SERVER")
			// login ibmcloud
			Cmd("ibmcloud", "login", "--apikey", ibmcloudAdminToken, "-r", "eu-de", "-g", "Developer-CI-and-QE")
			// login as admin in cluster
			Cmd(oc.path, "login", "--token=", adminToken, "--server=", cluster)
		} else {
			pass := os.Getenv("OCP4X_KUBEADMIN_PASSWORD")
			cluster := os.Getenv("OCP4X_API_URL")
			// login as kubeadmin
			Cmd(oc.path, "login", "-u", "kubeadmin", "-p", pass, cluster).ShouldPass()
		}
	}
	return token
}

// doAsDeveloper logins as developer to perform some task
func (oc OcRunner) doAsDeveloper(token, clusterType string) {

	if clusterType == "IBM" {
		ibmcloudDeveloperToken := os.Getenv("IBMC_DEVELOPER_LOGIN_APIKEY")
		Cmd("ibmcloud", "login", "--apikey", ibmcloudDeveloperToken, "-r", "eu-de", "-g", "Developer-CI-and-QE")
		// login as developer using token
	}
	oc.LoginUsingToken(token)
}

// add config map to the project for cleanup
func (oc OcRunner) addConfigMapForCleanup(projectName string) {
	Cmd(oc.path, "create", "configmap", "config-map-for-cleanup", "--from-literal", "type=testing", "--from-literal", "team=odo", "-n", projectName).ShouldPass()
}

func (oc OcRunner) Logout() {
	Cmd(oc.path, "logout")
}

// ScalePodToZero scales the pod of the deployment to zero.
// It waits for the pod to get deleted from the cluster before returning
func (oc OcRunner) ScalePodToZero(componentName, appName, projectName string) {
	podName := oc.GetRunningPodNameByComponent(componentName, projectName)
	Cmd(oc.path, "scale", "deploy", strings.Join([]string{componentName, appName}, "-"), "--replicas=0").ShouldPass()
	oc.WaitForRunnerCmdOut([]string{"get", "-n", projectName, "pod", podName}, 1, false, func(output string) bool {
		return !strings.Contains(output, podName)
	})
}

func (oc OcRunner) GetBindableKinds() (string, string) {
	return Cmd(oc.path, "get", "bindablekinds", "bindable-kinds", "-ojsonpath={.status[*].kind}").ShouldRun().OutAndErr()
}

func (oc OcRunner) GetServiceBinding(name, projectName string) (string, string) {
	return Cmd(oc.path, "get", "servicebinding", name, "-n", projectName).ShouldRun().OutAndErr()
}

func (oc OcRunner) EnsureOperatorIsInstalled(partialOperatorName string) {
	WaitForCmdOut(oc.path, []string{"get", "csv", "-o", "jsonpath={.items[?(@.status.phase==\"Succeeded\")].metadata.name}"}, 4, true, func(output string) bool {
		return strings.Contains(output, partialOperatorName)
	})
}

func (oc OcRunner) GetNamespaceProject() string {
	return Cmd(oc.path, "get", "project").ShouldPass().Out()
}

func (oc OcRunner) HasNamespaceProject(name string) bool {
	out := Cmd(oc.path, "get", "project", name, "-o", "jsonpath={.metadata.name}").
		ShouldRun().Out()
	return strings.Contains(out, name)
}

func (oc OcRunner) ListNamespaceProject(name string) {
	Eventually(func() string {
		return Cmd(oc.path, "get", "project").ShouldRun().Out()
	}, 30, 1).Should(ContainSubstring(name))
}

func (oc OcRunner) GetActiveNamespace() string {
	return Cmd(oc.path, "config", "view", "--minify", "-ojsonpath={..namespace}").ShouldPass().Out()
}

func (oc OcRunner) GetAllNamespaceProjects() []string {
	output := Cmd(oc.path, "get", "projects",
		"-o", "custom-columns=NAME:.metadata.name",
		"--no-headers").ShouldPass().Out()
	result, err := ExtractLines(output)
	Expect(err).ShouldNot(HaveOccurred())
	return result
}

func (oc OcRunner) GetLogs(podName string) string {
	output := Cmd(oc.path, "logs", podName).ShouldPass().Out()
	return output
}

func (oc OcRunner) AssertContainsLabel(kind, namespace, componentName, appName, mode, key, value string) {
	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).WithMode(mode).SelectorFlag()
	all := Cmd(oc.path, "get", kind, selector, "-n", namespace, "-o", "jsonpath={.items[0].metadata.labels}").ShouldPass().Out()
	Expect(all).To(ContainSubstring(fmt.Sprintf(`"%s":"%s"`, key, value)))
}

func (oc OcRunner) AssertNoContainsLabel(kind, namespace, componentName, appName, mode, key string) {
	selector := labels.Builder().WithComponentName(componentName).WithAppName(appName).WithMode(mode).SelectorFlag()
	all := Cmd(oc.path, "get", kind, selector, "-n", namespace, "-o", "jsonpath={.items[0].metadata.labels}").ShouldPass().Out()
	Expect(all).ToNot(ContainSubstring(fmt.Sprintf(`"%s"`, key)))
}

func (oc OcRunner) EnsurePodIsUp(namespace, podName string) {
	WaitForCmdOut(oc.path, []string{"get", "pods", "-n", namespace, "-o", "jsonpath='{range .items[*]}{.metadata.name}'"}, 4, true, func(output string) bool {
		return strings.Contains(output, podName)
	})
}

func (oc OcRunner) AssertNonAuthenticated() {
	Cmd(oc.path, "whoami").ShouldFail()
}
