package helper

import (
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// and returns the stdout, stderr and exitcode
func cmdRunner(cmdS string) (string, string, int) {
	//TODO: this needs to be os independent
	cmd := exec.Command("/bin/sh", "-c", cmdS)
	fmt.Fprintf(GinkgoWriter, "Running command: %s\n", cmdS)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

	// wait for the command execution to complete
	<-session.Exited

	Expect(err).NotTo(HaveOccurred())

	stdout := string(session.Out.Contents())
	stderr := string(session.Err.Contents())
	exitCode := session.ExitCode()

	//fmt.Fprintf(GinkgoWriter, "Result: \n stdout: %s\n stderr:%s \n exitcode: %d \n", stdout, stderr, exitCode)

	return stdout, stderr, exitCode
}

// CmdShouldPass command needs to retrun 0 as en exit code
// returns just stderr
func CmdShouldPass(cmd string) string {
	stdout, _, exitcode := cmdRunner(cmd)
	Expect(exitcode).To(Equal(0))
	return strings.TrimSpace(stdout)
}

// CmdShouldFail command needs to return non 0 as en exit code
// returns just stderr
func CmdShouldFail(cmd string) string {
	_, stderr, exitcode := cmdRunner(cmd)
	Expect(exitcode).NotTo(Equal(0))
	return strings.TrimSpace(stderr)
}
