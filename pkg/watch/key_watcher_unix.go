//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris zos

package watch

import (
	"golang.org/x/sys/unix"
)

// enableCharInput is inspired from the Unix implementation of MakeRaw in golang.org/x/term
// It enables the treatment of input stream char by char instead of line by line
// See https://man7.org/linux/man-pages/man3/termios.3.html for reference
func enableCharInput(fd int) error {
	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return err
	}
	termios.Lflag &^= unix.ICANON
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, termios); err != nil {
		return err
	}

	return nil
}
