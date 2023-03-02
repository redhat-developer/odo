package helper

import "github.com/onsi/gomega/gexec"

// CliRunner requires functions which are common for oc and kubectl
// By abstracting these functions into an interface, it handles the cli runner and calls the functions specified to particular cluster only
type CliRunner interface {
	Run(args ...string) *gexec.Session
	ExecListDir(podName string, projectName string, dir string) string
	// Exec executes the command in the specified pod and project/namespace.
	// If expectedSuccess is nil, the command is just supposed to run, with no assertion on its exit code.
	// If *expectedSuccess is true, the command exit code is expected to be 0.
	// If *expectedSuccess is false, the command exit code is expected to be non-zero.
	Exec(podName string, projectName string, args []string, expectedSuccess *bool) (string, string)
	CheckCmdOpInRemoteDevfilePod(podName string, containerName string, prjName string, cmd []string, checkOp func(cmdOp string, err error) bool) bool
	GetRunningPodNameByComponent(compName string, namespace string) string
	GetVolumeMountNamesandPathsFromContainer(deployName string, containerName, namespace string) string
	WaitAndCheckForExistence(resourceType, namespace string, timeoutMinutes int) bool
	GetServices(namespace string) string
	CreateAndSetRandNamespaceProject() string
	CreateAndSetRandNamespaceProjectOfLength(i int) string
	SetProject(namespace string) string

	// DeleteNamespaceProject deletes the specified namespace or project, optionally waiting until it is gone
	DeleteNamespaceProject(projectName string, wait bool)

	DeletePod(podName string, projectName string)
	GetAllNamespaceProjects() []string
	GetNamespaceProject() string

	// HasNamespaceProject returns whether the specified namespace or project exists in the cluster
	HasNamespaceProject(name string) bool
	// ListNamespaceProject checks if the namespace is present in the list of namespaces
	ListNamespaceProject(name string)

	GetActiveNamespace() string
	GetEnvsDevFileDeployment(componentName, appName, projectName string) map[string]string
	GetEnvRefNames(componentName, appName, projectName string) []string
	GetPVCSize(compName, storageName, namespace string) string
	GetAllPVCNames(namespace string) []string
	GetPodInitContainers(compName, namespace string) []string
	GetContainerEnv(podName, containerName, namespace string) string
	WaitAndCheckForTerminatingState(resourceType, namespace string, timeoutMinutes int) bool
	VerifyResourceDeleted(ri ResourceInfo)
	VerifyResourceToBeDeleted(ri ResourceInfo)
	GetAnnotationsDeployment(cmpName, appName, projectName string) map[string]string
	GetAllPodsInNs(namespace string) string
	WaitForRunnerCmdOut(args []string, timeout int, errOnFail bool, check func(output string) bool, includeStdErr ...bool) bool
	PodsShouldBeRunning(project string, regex string)
	CreateSecret(secretName, secretPass, project string)
	GetSecrets(project string) string
	GetEnvFromEntry(componentName string, appName string, projectName string) string
	GetVolumeNamesFromDeployment(componentName, appName, projectName string) map[string]string
	ScalePodToZero(componentName, appName, projectName string)
	GetAllPodNames(namespace string) []string
	EnsureOperatorIsInstalled(partialOperatorName string)
	GetBindableKinds() (string, string)
	GetServiceBinding(name, projectName string) (string, string)
	GetLogs(podName string) string
	AssertContainsLabel(kind, namespace, componentName, appName, mode, key, value string)
	AssertNoContainsLabel(kind, namespace, componentName, appName, mode, key string)
	EnsurePodIsUp(namespace, podName string)
	AssertNonAuthenticated()
	GetVersion() string
}
