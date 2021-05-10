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

type CmdWrapper struct {
	Cmd      *exec.Cmd
	program  string
	args     []string
	writer   *gexec.PrefixedWriter
	session  *gexec.Session
	timeout  time.Duration
	stopChan chan bool
	err      error
	maxRetry int
}

func Cmd(program string, args ...string) *CmdWrapper {
	prefix := fmt.Sprintf("[%s] ", filepath.Base(program))
	prefixWriter := gexec.NewPrefixedWriter(prefix, GinkgoWriter)
	command := exec.Command(program, args...)
	return &CmdWrapper{
		Cmd:     command,
		program: program,
		args:    args,
		writer:  prefixWriter,
	}
}

func (cw *CmdWrapper) Runner() *CmdWrapper {
	fmt.Fprintln(GinkgoWriter, runningCmd(cw.Cmd))
	cw.session, cw.err = gexec.Start(cw.Cmd, cw.writer, cw.writer)
	timeout := time.After(cw.timeout)
	if cw.timeout > 0 {
		select {
		case <-cw.stopChan:
			if cw.session != nil {
				if runtime.GOOS == "windows" {
					cw.session.Kill()
				} else {
					cw.session.Terminate()
				}
			}
		case <-timeout:
			if cw.session != nil {
				if runtime.GOOS == "windows" {
					cw.session.Kill()
				} else {
					cw.session.Terminate()
				}
			}
		}
	}
	return cw
}

func (cw *CmdWrapper) WithRetry(maxRetry int, intervalSeconds time.Duration) string {
	for i := 0; i < cw.maxRetry; i++ {
		fmt.Fprintf(GinkgoWriter, "try %d of %d\n", i, cw.maxRetry)

		cw.Runner()
		cw.session.Wait()
		// if exit code is 0 which means the program succeeded and hence we retry
		if cw.session.ExitCode() == 0 {
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
		} else {
			Consistently(cw.session).ShouldNot(gexec.Exit(0), runningCmd(cw.session.Command))
			return string(cw.session.Err.Contents())
		}
	}
	Fail(fmt.Sprintf("Failed after %d retries", maxRetry))
	return ""
}

func (cw *CmdWrapper) ShouldPass() *CmdWrapper {
	cw.Runner()
	Expect(cw.err).NotTo(HaveOccurred())
	Eventually(cw.session).Should(gexec.Exit(0), runningCmd(cw.session.Command))
	return cw
}

func (cw *CmdWrapper) ShouldFail() *CmdWrapper {
	cw.Runner()
	Expect(cw.err).To(HaveOccurred())
	Consistently(cw.session).ShouldNot(gexec.Exit(0), runningCmd(cw.session.Command))
	return cw
}

func (cw *CmdWrapper) WithTerminate(timeoutAfter time.Duration, stop <-chan bool) *CmdWrapper {
	cw.timeout = time.Duration(timeoutAfter) * time.Second
	cw.stopChan <- <-stop
	return cw
}

func (cw *CmdWrapper) WithTimeout(timeoutAfter time.Duration) *CmdWrapper {
	cw.timeout = time.Duration(timeoutAfter) * time.Second
	return cw
}

func (cw *CmdWrapper) WithEnv(args ...string) *CmdWrapper {
	cw.Cmd.Env = args
	return cw
}

func (cw *CmdWrapper) OutAndErr() (string, string) {
	return string(cw.session.Wait().Out.Contents()), string(cw.session.Wait().Err.Contents())
}

func (cw *CmdWrapper) Out() string {
	return string(cw.session.Wait().Out.Contents())
}
