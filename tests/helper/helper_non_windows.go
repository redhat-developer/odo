//go:build !windows
// +build !windows

package helper

import "github.com/onsi/gomega/gexec"

func terminateProc(session *gexec.Session) error {
	session.Interrupt()
	return nil
}
