package e2e

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"fmt"
	"os/exec"
	"time"
)

func runCmd(cmdS string) string {
	cmd := exec.Command("/bin/sh", "-c", cmdS)
	fmt.Fprintf(GinkgoWriter, "Running command: %s\n", cmdS)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

	// wait for the command execution to complete
	<-session.Exited
	Expect(session.ExitCode()).To(Equal(0))
	Expect(err).NotTo(HaveOccurred())

	return string(session.Out.Contents())
}

// runFailCmd runs a failing command
// and returns the stdout
func runFailCmd(cmdS string) string {
	cmd := exec.Command("/bin/sh", "-c", cmdS)
	fmt.Fprintf(GinkgoWriter, "Running command: %s\n", cmdS)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

	// wait for the command execution to complete
	Expect(session.ExitCode()).To(Equal(1))
	Expect(err).NotTo(HaveOccurred())

	return string(session.Out.Contents())
}

// waitForCmdOut runs a command until it gets
// the expected output.
// It accepts 2 arguments, cmd (command to be run)
// timeout (the time to wait for the output)
// check (function with output check logic)
// It times out if the command doesn't fetch the
// expected output  within the timeout period.
func waitForCmdOut(cmd string, timeout int, check func(output string) bool) bool {

	pingTimeout := time.After(time.Duration(timeout) * time.Minute)
	tick := time.Tick(time.Second)

	for {
		select {
		case <-pingTimeout:
			Fail(fmt.Sprintf("Timeout out after %v minutes", timeout))

		case <-tick:
			out, err := exec.Command("/bin/sh", "-c", cmd).Output()
			if err != nil {
				Fail(err.Error())
			}

			if check(string(out)) {
				return true
			}
		}
	}

}

// waitForEqualCmd calls the waitForCmdOut function to wait and check if the output is equal to the given string within 1 min
// cmd is the command to run
// expOut is the expected output
func waitForEqualCmd(cmd string, expOut string, timeout int) bool {

	return waitForCmdOut(cmd, timeout, func(output string) bool {
		return output == expOut
	})
}

// waitForEqualCmd calls the waitForCmdOut function to wait and check if the output is not equal to the given string within 1 min
// cmd is the command to run
// expOut is the expected output which should not be contained in the output string
func waitForDeleteCmd(cmd string, object string) bool {

	return waitForCmdOut(cmd, 5, func(output string) bool {
		return !strings.Contains(output, object)
	})
}

// waitForServiceStatusCmd calls the waitForCmdOut function to wait and check if the output is equal to the given string within 10 mins
// cmd is the command to run
// expOut is the expected output
func waitForServiceStatusCmd(cmd string, status string) bool {

	return waitForCmdOut(cmd, 10, func(output string) bool {
		return output == status
	})
}

func pollNonRetCmdStdOutForString(cmdStr string, timeout time.Duration, check func(output string) bool, startSimulationCh chan bool, startIndicatorFunc func(output string) bool) (bool, error) {
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

	ticker := time.NewTicker(time.Millisecond)
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
