package e2e

import (
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
func waitForEqualCmd(cmd string, expOut string) bool {

	return waitForCmdOut(cmd, 1, func(output string) bool {
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

// waitForEqualCmd calls the waitForCmdOut function to wait and check if the output is equal to the given string within 5 mins
// cmd is the command to run
// expOut is the expected output
func waitForServiceCreateCmd(cmd string, status string) bool {

	return waitForCmdOut(cmd, 5, func(output string) bool {
		return output == status
	})
}

func pollNonRetCmdStdOutForString(cmdStr string, timeout time.Duration, check func(output string) bool) (bool, error) {
	var cmd *exec.Cmd

	cmdStrParts := strings.Split(cmdStr, " ")
	cmdName := cmdStrParts[0]
	if len(cmdStrParts) > 1 {
		cmdStrParts = cmdStrParts[1:]
		cmd = exec.Command(cmdName, cmdStrParts...)
	} else {
		cmd = exec.Command(cmdName)
	}

	pingTimeout := time.After(timeout)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}

	buff := make([]byte, 1048576)
	tick := time.Tick(1 * time.Second)

	if err := cmd.Start(); err != nil {
		return false, err
	}

	for {
		select {
		case <-pingTimeout:
			Fail("Timeout out after " + string(timeout) + " minutes")
		case <-tick:
			n, err := stdout.Read(buff)
			if err != nil {
				return false, err
			}
			if n > 0 {
				s := string(buff[:n])
				if check(s) {
					if err := cmd.Process.Kill(); err != nil {
						return true, err
					}
					return true, nil
				}
			}
		}
	}
}
