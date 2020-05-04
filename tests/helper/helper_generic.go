package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// RandString returns a random string of given length
func RandString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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
			Fail(fmt.Sprintf("Timeout out after %v minutes", timeout))

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

	cmdStrParts := strings.Split(cmdStr, " ")
	cmdName := cmdStrParts[0]
	fmt.Println("Running command: ", cmdStrParts)
	if len(cmdStrParts) > 1 {
		cmdStrParts = cmdStrParts[1:]
		cmd = exec.Command(cmdName, cmdStrParts...)
	} else {
		cmd = exec.Command(cmdName)
	}
	cmd.Stdout = &buf

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
			Fail("Timeout out after " + string(timeout) + " minutes")
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

// GetUserHomeDir gets the user home directory
func GetUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	Expect(err).NotTo(HaveOccurred())
	return homeDir
}
