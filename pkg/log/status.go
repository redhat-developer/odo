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
	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/openshift/odo/pkg/log/fidget"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

// Spacing for logging
const suffixSpacing = "  "
const prefixSpacing = " "

var mu sync.Mutex

// Status is used to track ongoing status in a CLI, with a nice loading spinner
// when attached to a terminal
type Status struct {
	spinner       *fidget.Spinner
	status        string
	warningStatus string
	writer        io.Writer
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

// WarningStatus puts a warning status within the spinner and then updates the current status
func (s *Status) WarningStatus(status string) {
	s.warningStatus = status
	s.updateStatus()
}

// Updates the status and makes sure that if the previous status was longer, it
// "clears" the rest of the message.
func (s *Status) updateStatus() {
	mu.Lock()
	if s.warningStatus != "" {
		yellow := color.New(color.FgYellow).SprintFunc()

		// Determine the warning size, so that we can calculate its length and use that length as padding parameter
		warningSubstring := fmt.Sprintf(" [%s %s]", yellow(getWarningString()), yellow(s.warningStatus))

		// Combine suffix and spacing, then resize them
		newSuffix := fmt.Sprintf(suffixSpacing+"%s", s.status)
		newSuffix = truncateSuffixIfNeeded(newSuffix, s.writer, len(warningSubstring))

		// Combine the warning and non-warning text (since we don't want to truncate the warning text)
		s.spinner.SetSuffix(fmt.Sprintf("%s%s", newSuffix, warningSubstring))
	} else {
		newSuffix := fmt.Sprintf(suffixSpacing+"%s", s.status)
		s.spinner.SetSuffix(truncateSuffixIfNeeded(newSuffix, s.writer, 0))
	}
	mu.Unlock()
}

// Start starts a new phase of the status, if attached to a terminal
// there will be a loading spinner with this status
func (s *Status) Start(status string, debug bool) {
	s.End(true)

	// set new status
	isTerm := IsTerminal(s.writer)
	s.status = status

	// If we are in debug mode, don't spin!
	// In under no circumstances do we output if we're using -o json.. to
	// to avoid parsing errors.
	if !IsJSON() {
		if !isTerm || debug {
			fmt.Fprintf(s.writer, prefixSpacing+getSpacingString()+suffixSpacing+"%s  ...\n", s.status)
		} else {
			s.spinner.SetPrefix(prefixSpacing)
			newSuffix := fmt.Sprintf(suffixSpacing+"%s", s.status)
			s.spinner.SetSuffix(truncateSuffixIfNeeded(newSuffix, s.writer, 0))
			s.spinner.Start()
		}
	}
}

// truncateSuffixIfNeeded returns a represention of the 'suffix' parameter that fits within the terminal
// (including the extra space occupied by the padding parameter).
func truncateSuffixIfNeeded(suffix string, w io.Writer, padding int) string {

	terminalWidth := getTerminalWidth(w)
	if terminalWidth == nil {
		return suffix
	}

	// Additional padding to account for animation widget on lefthand side, and to avoid getting too close to the righthand terminal edge
	const additionalPadding = 10

	maxWidth := *terminalWidth - padding - additionalPadding

	// For VERY small terminals, or very large padding, just return the suffix unmodified
	if maxWidth <= 20 {
		return suffix
	}

	// If we are compliant, return the unmodified suffix...
	if len(suffix) <= maxWidth {
		return suffix
	}

	// Otherwise truncate down to the desired length and append '...'
	abbrevSuffix := "..."
	maxWidth -= len(abbrevSuffix) // maxWidth is necessarily >20 at this point

	// len(suffix) is necessarily >= maxWidth at this point
	suffix = suffix[:maxWidth] + abbrevSuffix

	return suffix
}

func getTerminalWidth(w io.Writer) *int {

	if runtime.GOOS != "windows" {

		if v, ok := (w).(*os.File); ok {
			w, _, err := terminal.GetSize(int(v.Fd()))
			if err == nil {
				return &w
			}
		}

	}

	return nil
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
		if !IsJSON() {
			fmt.Fprint(s.writer, "\r")
		}
	}

	if !IsJSON() {
		if success {
			// Clear the warning (unneeded now)
			s.WarningStatus("")
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Fprintf(s.writer, prefixSpacing+"%s"+suffixSpacing+"%s [%s]\n", green(getSuccessString()), s.status, s.spinner.TimeSpent())
		} else {
			red := color.New(color.FgRed).SprintFunc()
			if s.warningStatus != "" {
				fmt.Fprintf(s.writer, prefixSpacing+"%s"+suffixSpacing+"%s [%s] [%s]\n", red(getErrString()), s.status, s.spinner.TimeSpent(), s.warningStatus)
			} else {
				fmt.Fprintf(s.writer, prefixSpacing+"%s"+suffixSpacing+"%s [%s]\n", red(getErrString()), s.status, s.spinner.TimeSpent())
			}
		}
	}

	s.status = ""
}

// Namef will output the name of the component / application / project in a *bolded* manner
func Namef(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s\n", bold(fmt.Sprintf(format, a...)))
	}
}

// Progressf will output in an appropriate "progress" manner
func Progressf(format string, a ...interface{}) {
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), " %s%s\n", prefixSpacing, fmt.Sprintf(format, a...))
	}
}

// Success will output in an appropriate "success" manner
func Success(a ...interface{}) {
	green := color.New(color.FgGreen).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s%s%s%s", prefixSpacing, green(getSuccessString()), suffixSpacing, fmt.Sprintln(a...))
	}
}

// Successf will output in an appropriate "progress" manner
func Successf(format string, a ...interface{}) {
	green := color.New(color.FgGreen).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s%s%s%s\n", prefixSpacing, green(getSuccessString()), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Warningf will output in an appropriate "warning" manner
func Warningf(format string, a ...interface{}) {
	yellow := color.New(color.FgYellow).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStderr(), " %s%s%s\n", yellow(getWarningString()), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Swarningf (like Sprintf) will return a string in the "warning" manner
func Swarningf(format string, a ...interface{}) string {
	yellow := color.New(color.FgYellow).SprintFunc()
	return fmt.Sprintf(" %s%s%s", yellow(getWarningString()), suffixSpacing, fmt.Sprintf(format, a...))
}

// Experimental will output in an appropriate "progress" manner
func Experimental(a ...interface{}) {
	yellow := color.New(color.FgYellow).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s\n", yellow(fmt.Sprintln(a...)))
	}
}

// Warning will output in an appropriate "progress" manner
func Warning(a ...interface{}) {
	yellow := color.New(color.FgYellow).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStderr(), "%s%s%s%s", prefixSpacing, yellow(getWarningString()), suffixSpacing, fmt.Sprintln(a...))
	}
}

// Errorf will output in an appropriate "progress" manner
func Errorf(format string, a ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStderr(), " %s%s%s\n", red(getErrString()), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Error will output in an appropriate "progress" manner
func Error(a ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStderr(), "%s%s%s%s", prefixSpacing, red(getErrString()), suffixSpacing, fmt.Sprintln(a...))
	}
}

// Italic will simply print out information on a new italic line
func Italic(a ...interface{}) {
	italic := color.New(color.Italic).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s", italic(fmt.Sprintln(a...)))
	}
}

// Italicf will simply print out information on a new italic line
// this is **normally** used as a way to describe what's next within odo.
func Italicf(format string, a ...interface{}) {
	italic := color.New(color.Italic).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s\n", italic(fmt.Sprintf(format, a...)))
	}
}

// Info will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
func Info(a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s", bold(fmt.Sprintln(a...)))
	}
}

// Infof will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
func Infof(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s\n", bold(fmt.Sprintf(format, a...)))
	}
}

// Describef will print out the first variable as BOLD and then the second not..
// this is intended to be used with `odo describe` and other outputs that list
// a lot of information
func Describef(title string, format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s%s\n", bold(title), fmt.Sprintf(format, a...))
	}
}

// Askf will print out information, but in an "Ask" way (without newline)
func Askf(format string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s", bold(fmt.Sprintf(format, a...)))
	}
}

// Spinner creates a spinner, sets the prefix then returns it.
// Remember to use .End(bool) to stop the spin / when you're done.
// For example: defer s.End(false)
func Spinner(status string) *Status {
	return ExplicitSpinner(status, false)
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
	return ExplicitSpinner(status, true)
}

// ExplicitSpinner creates a spinner that can or not spin based on the value of the preventSpinning parameter
func ExplicitSpinner(status string, preventSpinning bool) *Status {
	doNotSpin := true
	if !preventSpinning {
		doNotSpin = IsDebug()
	}
	s := NewStatus(GetStdout())
	s.Start(status, doNotSpin)
	return s
}

// IsJSON returns true if we are in machine output mode..
// under NO circumstances should we output any logging.. as we are only outputting json
func IsJSON() bool {

	flag := pflag.Lookup("o")
	if flag != nil && flag.Changed {
		return strings.Contains(pflag.Lookup("o").Value.String(), "json")
	}

	return false
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

// getWarningString returns a certain string based upon the OS.
// Some Windows terminals do not support unicode and must use ASCII.
// TODO: Test needs to be added once we get Windows testing available on TravisCI / CI platform.
func getWarningString() string {
	if runtime.GOOS == "windows" {
		return "!"
	}
	return "⚠"
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
