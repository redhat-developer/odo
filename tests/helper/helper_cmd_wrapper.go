package helper

import (
	"fmt"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type CmdWrapper struct {
	Cmd                *exec.Cmd
	program            string
	args               []string
	includeErrorStream bool
	writer             *gexec.PrefixedWriter
	session            *gexec.Session
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

func (cw *CmdWrapper) ShouldPass() *CmdWrapper {
	fmt.Fprintln(GinkgoWriter, runningCmd(cw.Cmd))
	session, err := gexec.Start(cw.Cmd, cw.writer, cw.writer)
	Expect(err).NotTo(HaveOccurred())
	cw.session = session
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
