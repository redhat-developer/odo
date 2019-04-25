/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
	This package is a FORK of https://github.com/kubernetes-sigs/kind/blob/master/pkg/log/status.go
	See above license
*/

// Package log contains logging related functionality
package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/openshift/odo/pkg/log/fidget"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

// Spacing for logging
const suffixSpacing = "  "
const prefixSpacing = " "

// Status is used to track ongoing status in a CLI, with a nice loading spinner
// when attached to a terminal
type Status struct {
	spinner *fidget.Spinner
	status  string
	writer  io.Writer
}

// NewStatus creates a new default Status
func NewStatus(w io.Writer) *Status {
	spin := fidget.NewSpinner(w)
	s := &Status{
		spinner: spin,
		writer:  w,
	}
	return s
}

// StatusFriendlyWriter is used to wrap another Writer to make it toggle the
// status spinner before and after writes so that they do not collide
type StatusFriendlyWriter struct {
	status *Status
	inner  io.Writer
}

var _ io.Writer = &StatusFriendlyWriter{}

func (ww *StatusFriendlyWriter) Write(p []byte) (n int, err error) {
	ww.status.spinner.Stop()
	_, err = ww.inner.Write([]byte("\r"))
	if err != nil {
		return n, err
	}
	n, err = ww.inner.Write(p)
	ww.status.spinner.Start()
	return n, err
}

// WrapWriter returns a StatusFriendlyWriter for w
func (s *Status) WrapWriter(w io.Writer) io.Writer {
	return &StatusFriendlyWriter{
		status: s,
		inner:  w,
	}
}

// MaybeWrapWriter returns a StatusFriendlyWriter for w IFF w and spinner's
// output are a terminal, otherwise it returns w
func (s *Status) MaybeWrapWriter(w io.Writer) io.Writer {
	if IsTerminal(s.writer) && IsTerminal(w) {
		return s.WrapWriter(w)
	}
	return w
}

// IsTerminal returns true if the writer w is a terminal
// This function is modified if we are running within Windows..
// as Golang's built-in "IsTerminal" command only works on UNIX-based systems:
// https://github.com/golang/crypto/blob/master/ssh/terminal/util.go#L5
func IsTerminal(w io.Writer) bool {
	if runtime.GOOS == "windows" {
		return true
	} else if v, ok := (w).(*os.File); ok {
		return terminal.IsTerminal(int(v.Fd()))
	}
	return false
}

// Start starts a new phase of the status, if attached to a terminal
// there will be a loading spinner with this status
func (s *Status) Start(status string, debug bool) {
	s.End(true)
	// set new status
	isTerm := IsTerminal(s.writer)
	s.status = status

	// If we are in debug mode, don't spin!
	if !isTerm || debug {
		fmt.Fprintf(s.writer, prefixSpacing+getSpacingString()+suffixSpacing+"%s  ...\n", s.status)
	} else {
		s.spinner.SetPrefix(prefixSpacing)
		s.spinner.SetSuffix(fmt.Sprintf(suffixSpacing+"%s", s.status))
		s.spinner.Start()
	}

}

// End completes the current status, ending any previous spinning and
// marking the status as success or failure
func (s *Status) End(success bool) {
	if s.status == "" {
		return
	}

	isTerm := IsTerminal(s.writer)
	if isTerm {
		s.spinner.Stop()
		fmt.Fprint(s.writer, "\r")
	}

	if success {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(s.writer, prefixSpacing+"%s"+suffixSpacing+"%s\n", green(getSuccessString()), s.status)
	} else {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(s.writer, prefixSpacing+"%s"+suffixSpacing+"%s\n", red(getErrString()), s.status)
	}

	s.status = ""
}

// Namef will output the name of the component / application / project in a *bolded* manner
func Namef(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(GetStdout(), "%s\n", bold(fmt.Sprintf(format, a...)))
}

// Progressf will output in an appropriate "progress" manner
func Progressf(format string, a ...interface{}) {
	fmt.Fprintf(GetStdout(), " %s%s\n", prefixSpacing, fmt.Sprintf(format, a...))
}

// Success will output in an appropriate "success" manner
func Success(a ...interface{}) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(GetStdout(), "%s%s%s%s", prefixSpacing, green(getSuccessString()), suffixSpacing, fmt.Sprintln(a...))
}

// Successf will output in an appropriate "progress" manner
func Successf(format string, a ...interface{}) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(GetStdout(), "%s%s%s%s\n", prefixSpacing, green(getSuccessString()), suffixSpacing, fmt.Sprintf(format, a...))
}

// Errorf will output in an appropriate "progress" manner
func Errorf(format string, a ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Fprintf(GetStderr(), " %s%s%s\n", red(getErrString()), suffixSpacing, fmt.Sprintf(format, a...))
}

// Error will output in an appropriate "progress" manner
func Error(a ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Fprintf(GetStderr(), "%s%s%s%s", prefixSpacing, red(getErrString()), suffixSpacing, fmt.Sprintln(a...))
}

// Info will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
func Info(a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(GetStdout(), "%s", bold(fmt.Sprintln(a...)))
}

// Infof will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
func Infof(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(GetStdout(), "%s\n", bold(fmt.Sprintf(format, a...)))
}

// Askf will print out information, but in an "Ask" way (without newline)
func Askf(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(GetStdout(), "%s", bold(fmt.Sprintf(format, a...)))
}

// Spinner creates a spinner, sets the prefix then returns it.
// Remember to use .End(bool) to stop the spin / when you're done.
// For example: defer s.End(false)
func Spinner(status string) *Status {
	s := NewStatus(GetStdout())
	s.Start(status, IsDebug())
	return s
}

// Spinnerf creates a spinner, sets the prefix then returns it.
// Remember to use .End(bool) to stop the spin / when you're done.
// For example: defer s.End(false)
// for situations where spinning isn't viable (debug)
func Spinnerf(format string, a ...interface{}) *Status {
	s := NewStatus(GetStdout())
	s.Start(fmt.Sprintf(format, a...), IsDebug())
	return s
}

// SpinnerNoSpin is the same as the "Spinner" function but forces no spinning
func SpinnerNoSpin(status string) *Status {
	s := NewStatus(os.Stdout)
	s.Start(status, true)
	return s
}

// IsDebug returns true if we are debugging (-v is set to anything but 0)
func IsDebug() bool {

	flag := pflag.Lookup("v")

	if flag != nil {
		return !strings.Contains(pflag.Lookup("v").Value.String(), "0")
	}

	return false
}

// GetStdout gets the appropriate stdout from the OS. If it's Linux, it will use
// the go-colorable library in order to fix any and all color ASCII issues.
// TODO: Test needs to be added once we get Windows testing available on TravisCI / CI platform.
func GetStdout() io.Writer {
	if runtime.GOOS == "windows" {
		return colorable.NewColorableStdout()
	}
	return os.Stdout
}

// GetStderr gets the appropriate stderrfrom the OS. If it's Linux, it will use
// the go-colorable library in order to fix any and all color ASCII issues.
// TODO: Test needs to be added once we get Windows testing available on TravisCI / CI platform.
func GetStderr() io.Writer {
	if runtime.GOOS == "windows" {
		return colorable.NewColorableStderr()
	}
	return os.Stderr
}

// getErrString returns a certain string based upon the OS.
// Some Windows terminals do not support unicode and must use ASCII.
// TODO: Test needs to be added once we get Windows testing available on TravisCI / CI platform.
func getErrString() string {
	if runtime.GOOS == "windows" {
		return "X"
	}
	return "✗"
}

// getSuccessString returns a certain string based upon the OS.
// Some Windows terminals do not support unicode and must use ASCII.
// TODO: Test needs to be added once we get Windows testing available on TravisCI / CI platform.
func getSuccessString() string {
	if runtime.GOOS == "windows" {
		return "V"
	}
	return "✓"
}

// getSpacingString returns a certain string based upon the OS.
// Some Windows terminals do not support unicode and must use ASCII.
// TODO: Test needs to be added once we get Windows testing available on TravisCI / CI platform.
func getSpacingString() string {
	if runtime.GOOS == "windows" {
		return "-"
	}
	return "•"
}
