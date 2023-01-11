package helper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type CmdWrapper struct {
	Cmd             *exec.Cmd
	program         string
	args            []string
	writer          *gexec.PrefixedWriter
	session         *gexec.Session
	timeout         time.Duration
	intervalSeconds time.Duration
	stopChan        chan bool
	err             error
	maxRetry        int
	pass            bool
}

func Cmd(program string, args ...string) *CmdWrapper {
	prefix := fmt.Sprintf("[%s] ", filepath.Base(program))
	prefixWriter := gexec.NewPrefixedWriter(prefix, GinkgoWriter)
	command := exec.Command(program, args...)
	setSysProcAttr(command)
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
	} else if cw.maxRetry > 0 {
		cw.session.Wait()
		// cw.pass is to check if it is used with ShouldPass Or ShouldFail
		if !cw.pass {
			// we retry on success because the user has set “ShouldFail” as true
			// if exit code is 0 which means the program succeeded and hence we retry
			if cw.session.ExitCode() == 0 {
				time.Sleep(cw.intervalSeconds * time.Second)
				cw.maxRetry = cw.maxRetry - 1
				cw.Runner()
			}
			return cw
		} else {
			// if exit code is not 0 which means the program Failed and hence we retry
			if cw.session.ExitCode() != 0 {
				time.Sleep(cw.intervalSeconds * time.Second)
				cw.maxRetry = cw.maxRetry - 1
				cw.Runner()
			}
			return cw
		}
	}
	return cw
}

func (cw *CmdWrapper) WithRetry(maxRetry int, intervalSeconds time.Duration) *CmdWrapper {
	cw.maxRetry = maxRetry
	cw.intervalSeconds = intervalSeconds
	return cw
}

func (cw *CmdWrapper) ShouldPass() *CmdWrapper {
	cw.pass = true
	cw.Runner()
	Expect(cw.err).NotTo(HaveOccurred())
	Eventually(cw.session).Should(gexec.Exit(0), runningCmd(cw.session.Command))
	return cw
}

func (cw *CmdWrapper) ShouldFail() *CmdWrapper {
	cw.pass = false
	cw.Runner()
	Consistently(cw.session).ShouldNot(gexec.Exit(0), runningCmd(cw.session.Command))
	return cw
}

func (cw *CmdWrapper) ShouldRun() *CmdWrapper {
	cw.Runner()
	return cw
}

func (cw *CmdWrapper) WithTerminate(timeoutAfter time.Duration, stop chan bool) *CmdWrapper {
	cw.timeout = timeoutAfter * time.Second
	cw.stopChan = stop
	return cw
}

func (cw *CmdWrapper) WithTimeout(timeoutAfter time.Duration) *CmdWrapper {
	cw.timeout = timeoutAfter * time.Second
	return cw
}

func (cw *CmdWrapper) WithWorkingDir(dir string) *CmdWrapper {
	cw.Cmd.Dir = dir
	return cw
}

func (cw *CmdWrapper) WithEnv(args ...string) *CmdWrapper {
	cw.Cmd.Env = args
	return cw
}

func (cw *CmdWrapper) AddEnv(args ...string) *CmdWrapper {
	if cw.Cmd.Env == nil {
		cw.Cmd.Env = append(cw.Cmd.Env, os.Environ()...)
	}
	cw.Cmd.Env = append(cw.Cmd.Env, args...)
	return cw
}

func (cw *CmdWrapper) OutAndErr() (string, string) {
	return string(cw.session.Wait().Out.Contents()), string(cw.session.Wait().Err.Contents())
}

func (cw *CmdWrapper) Out() string {
	return string(cw.session.Wait().Out.Contents())
}

func (cw *CmdWrapper) Err() string {
	return string(cw.session.Wait().Err.Contents())
}
