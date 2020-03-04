package helper

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func runningCmd(cmd *exec.Cmd) string {
	prog := filepath.Base(cmd.Path)
	return fmt.Sprintf("Running %s with args %v", prog, cmd.Args)
}

func CmdRunner(program string, args ...string) *gexec.Session {
	//prefix ginkgo verbose output with program name
	prefix := fmt.Sprintf("[%s] ", filepath.Base(program))
	prefixWriter := gexec.NewPrefixedWriter(prefix, GinkgoWriter)
	command := exec.Command(program, args...)
	fmt.Fprintln(GinkgoWriter, runningCmd(command))
	session, err := gexec.Start(command, prefixWriter, prefixWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

// CmdShouldPass returns stdout if command succeeds
func CmdShouldPass(program string, args ...string) string {
	session := CmdRunner(program, args...)
	Eventually(session).Should(gexec.Exit(0), runningCmd(session.Command))
	return string(session.Wait().Out.Contents())
}

// CmdShouldRunWithTimeout waits for a certain duration and then returns stdout
func CmdShouldRunWithTimeout(timeout time.Duration, program string, args ...string) string {
	session := CmdRunner(program, args...)
	time.Sleep(timeout)
	session.Terminate()
	return string(session.Out.Contents())
}

// CmdShouldRunAndTerminate waits and returns stdout after a closed signal is passed on the closed channel
func CmdShouldRunAndTerminate(timeoutAfter time.Duration, stopChan <-chan bool, program string, args ...string) string {
	session := CmdRunner(program, args...)
	timeout := time.After(timeoutAfter)
	select {
	case <-stopChan:
		if session != nil {
			if runtime.GOOS == "windows" {
				session.Kill()
			} else {
				session.Terminate()
			}
		}
	case <-timeout:
		if session != nil {
			if runtime.GOOS == "windows" {
				session.Kill()
			} else {
				session.Terminate()
			}
		}
	}

	if session == nil {
		return ""
	}

	return string(session.Out.Contents())
}

// CmdShouldFail returns stderr if command fails
func CmdShouldFail(program string, args ...string) string {
	session := CmdRunner(program, args...)
	Consistently(session).ShouldNot(gexec.Exit(0), runningCmd(session.Command))
	return string(session.Wait().Err.Contents())
}
