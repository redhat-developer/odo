package helper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"

	dfutil "github.com/devfile/library/pkg/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// RandString returns a random string of given length
func RandString(n int) string {
	return dfutil.GenerateRandomString(n)
}

// WaitForCmdOut runs a command until it gets
// the expected output.
// It accepts 5 arguments, program (program to be run)
// args (arguments to the program)
// timeoutInMinutes (the time to wait for the output)
// errOnFail (flag to set if test should fail if command fails)
// check (function with output check logic)
// It times out if the command doesn't fetch the
// expected output  within the timeout period.
func WaitForCmdOut(program string, args []string, timeoutInMinutes int, errOnFail bool, check func(output string) bool, includeStdErr ...bool) bool {
	pingTimeout := time.After(time.Duration(timeoutInMinutes) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout after %v minutes", timeoutInMinutes))

		case <-tick:
			session := CmdRunner(program, args...)
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

// MatchAllInOutput ensures all strings are in output
func MatchAllInOutput(output string, tomatch []string) {
	for _, i := range tomatch {
		Expect(output).To(ContainSubstring(i))
	}
}

// DontMatchAllInOutput ensures all strings are not in output
func DontMatchAllInOutput(output string, tonotmatch []string) {
	for _, i := range tonotmatch {
		Expect(output).ToNot(ContainSubstring(i))
	}
}

// Unindented returns the unindented version of the jsonStr passed to it
func Unindented(jsonStr string) (string, error) {
	var tmpMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &tmpMap)
	if err != nil {
		return "", err
	}

	obj, err := json.Marshal(tmpMap)
	if err != nil {
		return "", err
	}
	return string(obj), err
}

// ExtractLines returns all lines of the given `output` string
func ExtractLines(output string) ([]string, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// FindFirstElementIndexByPredicate returns the index of the first element in `slice` that satisfies the given `predicate`.
func FindFirstElementIndexByPredicate(slice []string, predicate func(string) bool) (int, bool) {
	for i, s := range slice {
		if predicate(s) {
			return i, true
		}
	}
	return 0, false
}

// FindFirstElementIndexMatchingRegExp returns the index of the first element in `slice` that contains any match of
// the given regular expression `regularExpression`.
func FindFirstElementIndexMatchingRegExp(slice []string, regularExpression string) (int, bool) {
	return FindFirstElementIndexByPredicate(slice, func(s string) bool {
		matched, err := regexp.MatchString(regularExpression, s)
		Expect(err).To(BeNil(), func() string {
			return fmt.Sprintf("regular expression error: %v", err)
		})
		return matched
	})
}

// GetUserHomeDir gets the user home directory
func GetUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	Expect(err).NotTo(HaveOccurred())
	return homeDir
}

// LocalKubeconfigSet sets the KUBECONFIG to the temporary config file
func LocalKubeconfigSet(context string) {
	originalKubeCfg := os.Getenv("KUBECONFIG")
	if originalKubeCfg == "" {
		homeDir := GetUserHomeDir()
		originalKubeCfg = filepath.Join(homeDir, ".kube", "config")
	}
	copyKubeConfigFile(originalKubeCfg, filepath.Join(context, "config"))
}

// GetCliRunner gets the running cli against Kubernetes or OpenShift
func GetCliRunner() CliRunner {
	if IsKubernetesCluster() {
		return NewKubectlRunner("kubectl")
	}
	return NewOcRunner("oc")
}

// IsJSON returns true if a string is in json format
func IsJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

type CommonVar struct {
	// Project is new clean project/namespace for each test
	Project string
	// Context is a new temporary directory
	Context string
	// ConfigDir is a new temporary directory
	ConfigDir string
	// CliRunner is program command (oc or kubectl runner) according to cluster
	CliRunner CliRunner
	// original values to get restored after the test is done
	OriginalWorkingDirectory string
	OriginalKubeconfig       string
	// Ginkgo test realted
	testFileName string
	testCase     string
	testFailed   bool
	testDuration float64
}

const SetupClusterTrue = true
const SetupClusterFalse = false

// CommonBeforeEach is common function runs before every test Spec (It)
// returns CommonVar values that are used within the test script
func CommonBeforeEach(setupCluster bool) CommonVar {
	SetDefaultEventuallyTimeout(10 * time.Minute)
	SetDefaultConsistentlyDuration(30 * time.Second)

	commonVar := CommonVar{}
	commonVar.Context = CreateNewContext()
	commonVar.ConfigDir = CreateNewContext()
	commonVar.OriginalKubeconfig = os.Getenv("KUBECONFIG")
	commonVar.CliRunner = GetCliRunner()
	LocalKubeconfigSet(commonVar.ConfigDir)
	if setupCluster {
		commonVar.Project = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
	}
	commonVar.OriginalWorkingDirectory = Getwd()
	os.Setenv("GLOBALODOCONFIG", filepath.Join(commonVar.ConfigDir, "preference.yaml"))
	// Set ConsentTelemetry to false so that it does not prompt to set a preference value
	cfg, _ := preference.NewClient()
	err := cfg.SetConfiguration(preference.ConsentTelemetrySetting, "false")
	Expect(err).To(BeNil())
	// Use ephemeral volumes (emptyDir) in tests to make test faster
	err = cfg.SetConfiguration(preference.EphemeralSetting, "true")
	Expect(err).To(BeNil())
	SetDefaultDevfileRegistryAsStaging()
	return commonVar
}

// CommonAfterEach is common function that cleans up after every test Spec (It)
func CommonAfterEach(commonVar CommonVar) {
	// Get details, including test filename, test case name, test result, and test duration for each test spec and adds it to local testResults.txt file
	// Ginkgo test related variables
	commonVar.testFileName = CurrentSpecReport().ContainerHierarchyLocations[0].FileName
	commonVar.testCase = CurrentSpecReport().FullText()
	commonVar.testFailed = CurrentSpecReport().Failed()
	commonVar.testDuration = CurrentSpecReport().RunTime.Seconds()

	var prNum string
	var resultsRow string
	prNum = os.Getenv("GIT_PR_NUMBER")
	passedOrFailed := "PASSED"
	if commonVar.testFailed {
		passedOrFailed = "FAILED"
	}
	clusterType := "OCP"
	if IsKubernetesCluster() {
		clusterType = "KUBERNETES"
	}
	testDate := strings.Split(time.Now().Format(time.RFC3339), "T")[0]
	resultsRow = prNum + "," + testDate + "," + clusterType + "," + commonVar.testFileName + "," + commonVar.testCase + "," + passedOrFailed + "," + strconv.FormatFloat(commonVar.testDuration, 'E', -1, 64) + "\n"
	testResultsFile := filepath.Join("/", "tmp", "testResults.txt")
	if runtime.GOOS == "windows" {
		testResultsFile = filepath.Join(os.Getenv("TEMP"), "testResults.txt")
	}
	f, err := os.OpenFile(testResultsFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println("Error when opening file: ", err)
	} else {
		_, err = f.WriteString(resultsRow)
		if err != nil {
			fmt.Println("Error when writing to file: ", err)
		}
		if err = f.Close(); err != nil {
			fmt.Println("Error when closing file: ", err)
		}
	}

	if commonVar.Project != "" {
		// delete the random project/namespace created in CommonBeforeEach
		commonVar.CliRunner.DeleteNamespaceProject(commonVar.Project, false)
	}
	// restores the original kubeconfig and working directory
	Chdir(commonVar.OriginalWorkingDirectory)
	err = os.Setenv("KUBECONFIG", commonVar.OriginalKubeconfig)
	Expect(err).NotTo(HaveOccurred())

	// delete the temporary context directory
	DeleteDir(commonVar.Context)
	DeleteDir(commonVar.ConfigDir)

	os.Unsetenv("GLOBALODOCONFIG")
}

// JsonPathContentIs expects that the content of the path to equal value
func JsonPathContentIs(json string, path string, value string) {
	result := gjson.Get(json, path)
	Expect(result.String()).To(Equal(value), fmt.Sprintf("content of path %q should be %q but is %q", path, value, result.String()))
}

// JsonPathContentContain expects that the content of the path to contain value
func JsonPathContentContain(json string, path string, value string) {
	result := gjson.Get(json, path)
	Expect(result.String()).To(ContainSubstring(value), fmt.Sprintf("content of path %q should contain %q but is %q", path, value, result.String()))
}

// JsonPathDoesNotExist expects that the content of the path does not exist in the JSON string
func JsonPathDoesNotExist(json string, path string) {
	result := gjson.Get(json, path)
	Expect(result.Exists()).To(BeFalse(),
		fmt.Sprintf("content should not contain %q but is %q", path, result.String()))
}

func JsonPathContentIsValidUserPort(json string, path string) {
	result := gjson.Get(json, path)
	intVal, err := strconv.Atoi(result.String())
	Expect(err).ToNot(HaveOccurred())
	Expect(intVal).To(SatisfyAll(
		BeNumerically(">=", 1024),
		BeNumerically("<=", 65535),
	))
}

// SetProjectName sets projectNames based on the name of the test file name (without path and replacing _ with -), line number of current ginkgo execution, and a random string of 3 letters
func SetProjectName() string {
	//Get current test filename and remove file path, file extension and replace undescores with hyphens
	currGinkgoTestFileName := strings.Replace(strings.Split(strings.Split(CurrentSpecReport().
		ContainerHierarchyLocations[0].FileName, "/")[len(strings.Split(CurrentSpecReport().ContainerHierarchyLocations[0].FileName, "/"))-1], ".")[0], "_", "-", -1)
	currGinkgoTestLineNum := fmt.Sprint(CurrentSpecReport().LineNumber())
	projectName := currGinkgoTestFileName + currGinkgoTestLineNum + RandString(3)
	return projectName
}

// RunTestSpecs defines a common way how test specs in test suite are executed
func RunTestSpecs(t *testing.T, description string) {
	os.Setenv(segment.TrackingConsentEnv, "no")
	RegisterFailHandler(Fail)
	RunSpecs(t, description)
}

func IsKubernetesCluster() bool {
	return os.Getenv("KUBERNETES") == "true"
}

type ResourceInfo struct {
	ResourceType string
	ResourceName string
	Namespace    string
}

func SetDefaultDevfileRegistryAsStaging() {
	const registryName string = "DefaultDevfileRegistry"
	addRegistryURL := "https://registry.stage.devfile.io"
	proxy := os.Getenv("DEVFILE_PROXY")
	if proxy != "" {
		addRegistryURL = "http://" + proxy
	}
	Cmd("odo", "preference", "remove", "registry", registryName, "-f").ShouldPass()
	Cmd("odo", "preference", "add", "registry", registryName, addRegistryURL).ShouldPass()
}
