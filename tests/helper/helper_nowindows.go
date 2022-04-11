//go:build !windows
// +build !windows

package helper

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"
)

func terminateProc(session *gexec.Session) error {
	session.Interrupt()
	return nil
}

func setSysProcAttr(command *exec.Cmd) {}
