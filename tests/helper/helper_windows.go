//go:build windows
// +build windows

package helper

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/onsi/gomega/gexec"
	"golang.org/x/sys/windows"
)

func terminateProc(session *gexec.Session) error {
	pid := session.Command.Process.Pid
	dll, err := windows.LoadDLL("kernel32.dll")
	if err != nil {
		return fmt.Errorf("loading DLL: %w", err)
	}
	defer dll.Release()
	generateConsoleCtrlEvent, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return fmt.Errorf("finding GenerateConsoleCtrlEvent: %w", err)
	}
	r1, _, err := generateConsoleCtrlEvent.Call(uintptr(syscall.CTRL_BREAK_EVENT), uintptr(pid))
	if r1 == 0 {
		return fmt.Errorf("calling GenerateConsoleCtrlEvent: %w", err)
	}
	return nil
}

func setSysProcAttr(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
