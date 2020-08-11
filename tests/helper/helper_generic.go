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
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/openshift/odo/pkg/util"
)

// RandString returns a random string of given length
func RandString(n int) string {
	return util.GenerateRandomString(n)
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

// WatchNonRetCmdStdOut run odo watch and get the cmdSTDOUT output into buffer.
// startIndicatorFunc sets true and startSimulationCh starts, when buffer contain "Waiting for something to change"
// check function checks for the changes into the buffer
func WatchNonRetCmdStdOut(cmdStr string, timeout time.Duration, check func(output string) bool, startSimulationCh chan bool, startIndicatorFunc func(output string) bool) (bool, error) {
	var cmd *exec.Cmd
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	var cmdStrParts []string
	// interestingly "odo watch  --context" ( observe the 2 spaces between watch and --context ) becomes ["odo", "watch", "", "--context"] which falls
	// apart with Error: unknown command "" for "odo watch" on cobra when used "cobra.NoArgs".
	tmpParts := strings.Split(cmdStr, " ")
	for i := 0; i < len(tmpParts); i++ {
		trimedPart := strings.TrimSpace(tmpParts[i])
		if trimedPart == "" {
			continue
		}
		cmdStrParts = append(cmdStrParts, trimedPart)
	}
	cmdName := cmdStrParts[0]
	_, err := fmt.Fprintln(GinkgoWriter, "Running command: ", cmdStrParts)
	Expect(err).To(BeNil())
	if len(cmdStrParts) > 1 {
		cmdStrParts = cmdStrParts[1:]
		cmd = exec.Command(cmdName, cmdStrParts...)
	} else {
		cmd = exec.Command(cmdName)
	}
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
				_, err := fmt.Fprintln(GinkgoWriter, "Output from stdout:")
				Expect(err).To(BeNil())
				_, err = fmt.Fprintln(GinkgoWriter, buf.String())
				Expect(err).To(BeNil())
			}
			errBufStr := errBuf.String()
			if errBufStr != "" {
				_, err = fmt.Fprintln(GinkgoWriter, "Output from stderr:")
				Expect(err).To(BeNil())
				_, err = fmt.Fprintln(GinkgoWriter, errBufStr)
				Expect(err).To(BeNil())
			}
			Fail(fmt.Sprintf("Timeout after %.2f minutes", timeout.Minutes()))
		case <-ticker.C:
			if !startedFileModification && startIndicatorFunc(buf.String()) {
				startedFileModification = true
				startSimulationCh <- true
			}
			if check(buf.String()) {
				if err := cmd.Process.Kill(); err != nil {
					return true, err
				}
				return true, nil
			}
		}
	}
}

// RunCmdWithMatchOutputFromBuffer starts the command, and command stdout is attached to buffer.
// we read data from buffer line by line, and if expected string is matcher it reutrns true
// It is diffrent from WaitforCmdOut which gives stdout in one go usinf session.Out.Contents()
// for commands like odo log -f which streams continuous data and does not terminate by their own
// we need to read the stream data from buffer.
func RunCmdWithMatchOutputFromBuffer(timeoutAfter time.Duration, matchString, program string, args ...string) (bool, error) {
	var buf bytes.Buffer

	command := exec.Command(program, args...)
	command.Stdout = &buf

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

	// if we do not expected output timeout.
	for {
		select {
		case <-timeoutCh:
			return false, errors.New("Timeout waiting for the conditon")
		case <-matchOutputCh:
			return true, nil
		case <-errorCh:
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
	if os.Getenv("KUBERNETES") == "true" {
		return NewKubectlRunner("kubectl")
	}
	return NewOcRunner("oc")
}

// Suffocate the string by removing all the space from it ;-)
func Suffocate(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\t", "")
}
