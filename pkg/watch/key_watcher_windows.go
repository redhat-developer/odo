package watch

import (
	"golang.org/x/sys/windows"
)

// enableCharInput is inspired from Windows implementation of MakeRaw in golang.org/x/term
// It enables the treatment of input stream char by char instead of line by line
// See https://docs.microsoft.com/en-us/windows/console/setconsolemode for reference
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
