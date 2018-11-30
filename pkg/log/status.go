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
	"strings"

	"github.com/fatih/color"
	"github.com/redhat-developer/odo/pkg/log/fidget"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

// Spacing for logging
const suffixSpacing = " "

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
func IsTerminal(w io.Writer) bool {
	if v, ok := (w).(*os.File); ok {
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
		fmt.Fprintf(s.writer, " •   %s  ...\n", s.status)
	} else {
		s.spinner.SetSuffix(fmt.Sprintf("   %s", s.status))
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
		fmt.Fprintf(s.writer, " %s   %s\n", green("✓"), s.status)
	} else {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(s.writer, " %s   %s\n", red("✗"), s.status)
	}

	s.status = ""
}

// Namef will output the name of the component / application / project in a *bolded* manner
func Namef(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s\n", bold(fmt.Sprintf(format, a...)))
}

// Progressf will output in an appropriate "progress" manner
func Progressf(format string, a ...interface{}) {
	fmt.Printf(" %s\n", fmt.Sprintf(format, a...))
}

// Successf will output in an appropriate "progress" manner
func Successf(format string, a ...interface{}) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf(" %s%s%s\n", green("OK "), suffixSpacing, fmt.Sprintf(format, a...))
}

// Errorf will output in an appropriate "progress" manner
func Errorf(format string, a ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Printf(" %s%s%s\n", red("ERR"), suffixSpacing, fmt.Sprintf(format, a...))
}

// Error will output in an appropriate "progress" manner
func Error(a ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Printf(" %s%s%s\n", red("ERR"), suffixSpacing, fmt.Sprintln(a...))
}

// Info will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
func Info(a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s\n", bold(fmt.Sprintln(a...)))
}

// Infof will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
func Infof(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s\n", bold(fmt.Sprintf(format, a...)))
}

// Askf will print out information, but in an "Ask" way (without newline)
func Askf(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s", bold(fmt.Sprintf(format, a...)))
}

// Status creates a spinner, sets the prefix then returns it.
// Remember to use .End(bool) to stop the spin / when you're done.
// For example: defer s.End(false)
func Spinner(status string) *Status {
	s := NewStatus(os.Stdout)
	s.Start(status, IsDebug())
	return s
}

// Statusf creates a spinner, sets the prefix then returns it.
// Remember to use .End(bool) to stop the spin / when you're done.
// For example: defer s.End(false)
func Spinnerf(format string, a ...interface{}) *Status {
	s := NewStatus(os.Stdout)
	s.Start(fmt.Sprintf(format, a...), IsDebug())
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
