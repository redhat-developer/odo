package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/tests/helper/reporter"

	dfutil "github.com/devfile/library/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tidwall/gjson"
)

// RandString returns a random string of given length
func RandString(n int) string {
	return dfutil.GenerateRandomString(n)
}

// WaitForCmdOut runs a command until it gets
// the expected output.
// It accepts 5 arguments, program (program to be run)
// args (arguments to the program)
// timeout (the time to wait for the output)
// errOnFail (flag to set if test should fail if command fails)
// check (function with output check logic)
// It times out if the command doesn't fetch the
// expected output  within the timeout period.
func WaitForCmdOut(program string, args []string, timeout int, errOnFail bool, check func(output string) bool, includeStdErr ...bool) bool {
	pingTimeout := time.After(time.Duration(timeout) * time.Minute)
	// this is a test package so time.Tick() is acceptable
	// nolint
	tick := time.Tick(time.Second)
	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout after %v minutes", timeout))

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
			if check(strings.TrimSpace(string(output))) {
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

// ExtractSubString extracts substring from output, beginning at start and before end
func ExtractSubString(output, start, end string) string {
	i := strings.Index(output, start)
	if i >= 0 {
		j := strings.Index(output[i:], end)
		if j >= 0 {
			return output[i : i+j]
		}
	}
	return ""
}

// WatchNonRetCmdStdOut runs an 'odo watch' command and stores the process' stdout output into buffer.
// - startIndicatorFunc should check stdout output and return true when simulation is ready to begin (for example, buffer contains "Waiting for something to change")
// - startSimulationCh will be sent a 'true' when startIndicationFunc first returns true, at which point files/directories should be created by associated goroutine
// - success function is passed stdout buffer, and should return if the test conditions have passes
func WatchNonRetCmdStdOut(cmdStr string, timeout time.Duration, success func(output string) bool, startSimulationCh chan bool, startIndicatorFunc func(output string) bool) (bool, error) {
	var cmd *exec.Cmd
	var buf bytes.Buffer
	var errBuf bytes.Buffer

	cmdStrParts := strings.Fields(cmdStr)

	fmt.Fprintln(GinkgoWriter, "Running command: ", cmdStrParts)

	cmd = exec.Command(cmdStrParts[0], cmdStrParts[1:]...)

	cmd.Stdout = &buf
	cmd.Stderr = &errBuf

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeoutCh := make(chan bool)
	go func() {
		time.Sleep(timeout)
		timeoutCh <- true
	}()

	if err := cmd.Start(); err != nil {
		return false, err
	}
	startedFileModification := false
	for {
		select {
		case <-timeoutCh:
			if buf.String() != "" {
				_, err := fmt.Fprintln(GinkgoWriter, "Output from stdout ["+cmdStr+"]:")
				Expect(err).To(BeNil())
				_, err = fmt.Fprintln(GinkgoWriter, buf.String())
				Expect(err).To(BeNil())
			}
			errBufStr := errBuf.String()
			if errBufStr != "" {
				_, err := fmt.Fprintln(GinkgoWriter, "Output from stderr ["+cmdStr+"]:")
				Expect(err).To(BeNil())
				_, err = fmt.Fprintln(GinkgoWriter, errBufStr)
				Expect(err).To(BeNil())
			}
			Fail(fmt.Sprintf("Timeout after %.2f minutes", timeout.Minutes()))
		case <-ticker.C: // Every 10 seconds...

			// If we have not yet begun file modification, query the parameter function to see if we should, do so if true
			if !startedFileModification && startIndicatorFunc(buf.String()) {
				startedFileModification = true
				startSimulationCh <- true
			}
			// Call success(...) to determine if stdout contains expected text, exit if true
			if success(buf.String()) {
				if err := cmd.Process.Kill(); err != nil {
					return true, err
				}
				return true, nil
			}
		}
	}
}

// RunCmdWithMatchOutputFromBuffer starts the command, and command stdout is attached to buffer.
// we read data from buffer line by line, and if expected string is matched it returns true
// It is different from WaitforCmdOut which gives stdout in one go using session.Out.Contents()
// for commands like odo log -f which streams continuous data and does not terminate by their own
// we need to read the stream data from buffer.
func RunCmdWithMatchOutputFromBuffer(timeoutAfter time.Duration, matchString, program string, args ...string) (bool, error) {
	var buf, errBuf bytes.Buffer

	command := exec.Command(program, args...)
	command.Stdout = &buf
	command.Stderr = &errBuf

	timeoutCh := time.After(timeoutAfter)
	matchOutputCh := make(chan bool)
	errorCh := make(chan error)

	_, err := fmt.Fprintln(GinkgoWriter, runningCmd(command))
	if err != nil {
		return false, err
	}

	err = command.Start()
	if err != nil {
		return false, err
	}

	// go routine which is reading data from buffer until expected string matched
	go func() {
		for {
			line, err := buf.ReadString('\n')
			if err != nil && err != io.EOF {
				errorCh <- err
			}
			if len(line) > 0 {
				_, err = fmt.Fprintln(GinkgoWriter, line)
				if err != nil {
					errorCh <- err
				}
				if strings.Contains(line, matchString) {
					matchOutputCh <- true
				}
			}
		}
	}()

	for {
		select {
		case <-timeoutCh:
			fmt.Fprintln(GinkgoWriter, errBuf.String())
			return false, errors.New("Timeout waiting for the condition")
		case <-matchOutputCh:
			return true, nil
		case <-errorCh:
			fmt.Fprintln(GinkgoWriter, errBuf.String())
			return false, <-errorCh
		}
	}

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

// Suffocate the string by removing all the space from it ;-)
func Suffocate(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\t", ""), "\n", "")
}

// IsJSON returns true if a string is in json format
func IsJSON(s string) bool {
	var js map[string]interface{}
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

// CommonBeforeEach is common function runs before every test Spec (It)
// returns CommonVar values that are used within the test script
func CommonBeforeEach() CommonVar {
	SetDefaultEventuallyTimeout(10 * time.Minute)
	SetDefaultConsistentlyDuration(30 * time.Second)

	commonVar := CommonVar{}
	commonVar.Context = CreateNewContext()
	commonVar.ConfigDir = CreateNewContext()
	commonVar.OriginalKubeconfig = os.Getenv("KUBECONFIG")
	commonVar.CliRunner = GetCliRunner()
	LocalKubeconfigSet(commonVar.ConfigDir)
	commonVar.Project = commonVar.CliRunner.CreateRandNamespaceProject()
	commonVar.OriginalWorkingDirectory = Getwd()
	os.Setenv("GLOBALODOCONFIG", filepath.Join(commonVar.ConfigDir, "preference.yaml"))
	// Set ConsentTelemetry to false so that it does not prompt to set a preference value
	cfg, _ := preference.NewClient()
	err := cfg.SetConfiguration(preference.ConsentTelemetrySetting, "false")
	Expect(err).To(BeNil())
	SetDefaultDevfileRegistryAsStaging()
	// Ginkgo test related variables
	commonVar.testFileName = strings.Replace(CurrentGinkgoTestDescription().FileName[strings.LastIndex(CurrentGinkgoTestDescription().FileName, "/")+1:strings.LastIndex(CurrentGinkgoTestDescription().FileName, ".")], "_", "-", -1) + ".go"
	commonVar.testCase = CurrentGinkgoTestDescription().FullTestText
	commonVar.testFailed = CurrentGinkgoTestDescription().Failed
	commonVar.testDuration = CurrentGinkgoTestDescription().Duration.Seconds()
	return commonVar
}

// CommonAfterEach is common function that cleans up after every test Spec (It)
func CommonAfterEach(commonVar CommonVar) {
	// Get details, including test result for each test spec and adds it to local testResults.txt file
	var prNum string
	var K8SorOcp string
	var resultsRow string
	prNum = os.Getenv("GIT_PR_NUMBER")
	K8SorOcp = os.Getenv("KUBERNETES")
	passedOrFailed := "PASSED"
	if commonVar.testFailed {
		passedOrFailed = "FAILED"
	}
	clusterType := "OCP"
	if K8SorOcp == "KUBERNETES" {
		clusterType = "KUBERNETES"
	}
	now := time.Now()
	y, m, d := now.Date()
	testDate := strconv.Itoa(y) + "-" + strconv.Itoa(int(m)) + "-" + strconv.Itoa(d)
	resultsRow = prNum + ", " + testDate + ", " + clusterType + ", " + commonVar.testFileName + ", " + commonVar.testCase + ", " + passedOrFailed + ", " + strconv.FormatFloat(commonVar.testDuration, 'E', -1, 64) + "\n"
	testResultsFile := filepath.Join("/", "tmp", "testResults.txt")

	f, err := os.OpenFile(testResultsFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println("Error: ", err)
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(resultsRow); err != nil {
		fmt.Println("Error: ", err)
		panic(err)
	}

	f.Close()

	// delete the random project/namespace created in CommonBeforeEach
	commonVar.CliRunner.DeleteNamespaceProject(commonVar.Project)

	// restores the original kubeconfig and working directory
	Chdir(commonVar.OriginalWorkingDirectory)
	err = os.Setenv("KUBECONFIG", commonVar.OriginalKubeconfig)
	Expect(err).NotTo(HaveOccurred())

	// delete the temporary context directory
	DeleteDir(commonVar.Context)
	DeleteDir(commonVar.ConfigDir)

	os.Unsetenv("GLOBALODOCONFIG")
}

// GjsonMatcher validates if []results from gjson.GetMany match the expected values for each json path requested
// Values is an array of results returned by the gjson.GetMany function
// Expected is an array of strings, defines the expected values for each of the paths requested in the gjson.GetMany function
// For documentation about gjson see https://github.com/tidwall/gjson#get-multiple-values-at-once
func GjsonMatcher(values []gjson.Result, expected []string) bool {
	matched := 0
	for i, v := range values {
		if strings.Contains(v.String(), expected[i]) {
			matched++
		}
	}
	numVars := len(expected)
	return matched == numVars
}

// GjsonExactMatcher validates if []results from gjson.GetMany match the expected values for each json path requested
// Values is an array of results returned by the gjson.GetMany function
// Expected is an array of strings, defines the expected values for each of the paths requested in the gjson.GetMany function
// For documentation about gjson see https://github.com/tidwall/gjson#get-multiple-values-at-once
func GjsonExactMatcher(values []gjson.Result, expected []string) bool {
	matched := 0
	for i, v := range values {
		if v.String() == expected[i] {
			matched++
		}
	}
	numVars := len(expected)
	return matched == numVars
}

//SetProjectName sets projectNames based on the neame of the test file name (withouth path and replacing _ with -), line number of current ginkgo execution, and a random string of 3 letters
func SetProjectName() string {
	//Get current test filename and remove file path, file extension and replace undescores with hyphens
	currGinkgoTestFileName := strings.Replace(CurrentGinkgoTestDescription().FileName[strings.LastIndex(CurrentGinkgoTestDescription().FileName, "/")+1:strings.LastIndex(CurrentGinkgoTestDescription().FileName, ".")], "_", "-", -1)
	currGinkgoTestLineNum := strconv.Itoa(CurrentGinkgoTestDescription().LineNumber)
	projectName := currGinkgoTestFileName + currGinkgoTestLineNum + RandString(3)
	return projectName
}

// RunTestSpecs defines a common way how test specs in test suite are executed
func RunTestSpecs(t *testing.T, description string) {
	os.Setenv("ODO_DISABLE_TELEMETRY", "true")
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, description, []Reporter{reporter.JunitReport(t, "../../reports/")})
}

func IsKubernetesCluster() bool {
	return os.Getenv("KUBERNETES") == "true"
}

type ResourceInfo struct {
	ResourceType string
	ResourceName string
	Namespace    string
}

func VerifyResourcesDeleted(runner CliRunner, resources []ResourceInfo) {
	for _, item := range resources {
		runner.VerifyResourceDeleted(item)
	}
}

func VerifyResourcesToBeDeleted(runner CliRunner, resources []ResourceInfo) {
	for _, item := range resources {
		runner.VerifyResourceToBeDeleted(item)
	}
}

func SetDefaultDevfileRegistryAsStaging() {
	const registryName string = "DefaultDevfileRegistry"
	const addRegistryURL string = "https://registry.stage.devfile.io"
	Cmd("odo", "preference", "registry", "update", registryName, addRegistryURL, "-f").ShouldPass()
}

// CopyAndCreate copies required source code and devfile to the given context directory, and creates a component
func CopyAndCreate(sourcePath, devfilePath, contextDir string) {
	CopyExample(sourcePath, contextDir)
	CopyExampleDevFile(devfilePath, filepath.Join(contextDir, "devfile.yaml"))
	Cmd("odo", "create", "--context", contextDir).ShouldPass()
}
