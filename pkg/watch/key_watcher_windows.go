package watch

import (
	"golang.org/x/sys/windows"
)

// makeRaw is forked from golang.org/x/term
// Differences are:
// - ctrl-c is enabled
func enableCharInput(fd int) error {
	var st uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &st); err != nil {
		return err
	}
	raw := st &^ (windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT)
	if err := windows.SetConsoleMode(windows.Handle(fd), raw); err != nil {
		return err
	}
	return nil
}
