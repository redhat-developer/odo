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
	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/redhat-developer/odo/pkg/log/fidget"
	"github.com/redhat-developer/odo/pkg/version"
)

// Spacing for logging
const suffixSpacing = "  "
const prefixSpacing = " "

var mu sync.Mutex
var colors = []color.Attribute{color.FgRed, color.FgGreen, color.FgYellow, color.FgBlue, color.FgMagenta, color.FgCyan, color.FgWhite}
var colorCounter = 0

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
		return term.IsTerminal(int(v.Fd()))
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

// truncateSuffixIfNeeded returns a representation of the 'suffix' parameter that fits within the terminal
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
			w, _, err := term.GetSize(int(v.Fd()))
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

		time := ""
		if s.spinner.TimeSpent() != "" {
			time = fmt.Sprintf("[%s]", s.spinner.TimeSpent())
		}

		if success {
			// Clear the warning (unneeded now)
			s.WarningStatus("")
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Fprintf(s.writer, prefixSpacing+"%s"+suffixSpacing+"%s %s\n", green(getSuccessString()), s.status, time)
		} else {
			red := color.New(color.FgRed).SprintFunc()
			if s.warningStatus != "" {
				fmt.Fprintf(s.writer, prefixSpacing+"%s"+suffixSpacing+"%s %s [%s]\n", red(getErrString()), s.status, time, s.warningStatus)
			} else {
				fmt.Fprintf(s.writer, prefixSpacing+"%s"+suffixSpacing+"%s %s\n", red(getErrString()), s.status, time)
			}
		}
	}

	s.status = ""
}

// EndWithStatus is similar to End, but lets the user specify a custom message/status while ending
func (s *Status) EndWithStatus(status string, success bool) {
	if status == "" {
		return
	}
	s.status = status
	s.End(success)
}

// Printf will output in an appropriate "information" manner; for e.g.
// • <message>
func Printf(format string, a ...interface{}) {
	if !IsJSON() {
		fmt.Fprintf(GetStdout(), "%s%s%s%s\n", prefixSpacing, getSpacingString(), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Fprintf will output in an appropriate "information" manner; for e.g.
// • <message>
func Fprintf(w io.Writer, format string, a ...interface{}) {
	if !IsJSON() {
		fmt.Fprintf(w, "%s%s%s%s\n", prefixSpacing, getSpacingString(), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Println will output a new line when applicable
func Println() {
	if !IsJSON() {
		fmt.Fprintln(GetStdout())
	}
}

// Fprintln will output a new line when applicable
func Fprintln(w io.Writer) {
	if !IsJSON() {
		fmt.Fprintln(w)
	}
}

// Success will output in an appropriate "success" manner
// ✓  <message>
func Success(a ...interface{}) {
	if !IsJSON() {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(GetStdout(), "%s%s%s%s", prefixSpacing, green(getSuccessString()), suffixSpacing, fmt.Sprintln(a...))
	}
}

// Successf will output in an appropriate "progress" manner
//
//	✓  <message>
func Successf(format string, a ...interface{}) {
	if !IsJSON() {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(GetStdout(), "%s%s%s%s\n", prefixSpacing, green(getSuccessString()), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Warning will output in an appropriate "progress" manner
//
//	⚠ <message>
func Warning(a ...interface{}) {
	Fwarning(GetStderr(), a...)
}

// Fwarning will output in an appropriate "progress" manner in out writer
//
//	⚠ <message>
func Fwarning(out io.Writer, a ...interface{}) {
	if !IsJSON() {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Fprintf(out, "%s%s%s%s", prefixSpacing, yellow(getWarningString()), suffixSpacing, fmt.Sprintln(a...))
	}
}

// Warningf will output in an appropriate "warning" manner
//
//	⚠ <message>
func Warningf(format string, a ...interface{}) {
	if !IsJSON() {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Fprintf(GetStderr(), " %s%s%s\n", yellow(getWarningString()), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Fwarningf will output in an appropriate "warning" manner
//
//	⚠ <message>
func Fwarningf(w io.Writer, format string, a ...interface{}) {
	if !IsJSON() {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Fprintf(w, " %s%s%s\n", yellow(getWarningString()), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Fsuccess will output in an appropriate "progress" manner in out writer
//
//	✓ <message>
func Fsuccess(out io.Writer, a ...interface{}) {
	if !IsJSON() {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(out, "%s%s%s%s", prefixSpacing, green(getSuccessString()), suffixSpacing, fmt.Sprintln(a...))
	}
}

// DisplayExperimentalWarning displays the experimental mode warning message.
func DisplayExperimentalWarning() {
	if !IsJSON() {
		yellow := color.New(color.FgYellow).SprintFunc()
		h := "============================================================================"
		fmt.Fprintln(GetStdout(), yellow(fmt.Sprintf(`%[1]s
%s Experimental mode enabled. Use at your own risk.
More details on https://odo.dev/docs/user-guides/advanced/experimental-mode
%[1]s
`, h, getWarningString())))
	}
}

// Title Prints the logo as well as the first line being BLUE (indicator of the command information);
// the second line is optional and provides information in regard to what is being run.
// The last line displays information about the current odo version.
//
//	 __
//	/  \__     **First line**
//	\__/  \    Second line
//	/  \__/    odo version: <VERSION>
//	\__/
func Title(firstLine, secondLine string) {
	if !IsJSON() {
		fmt.Fprint(GetStdout(), Stitle(firstLine, secondLine))
	}
}

// Stitle is the same as Title but returns the string instead
func Stitle(firstLine, secondLine string) string {
	var versionMsg string
	if version.VERSION != "" {
		versionMsg = "odo version: " + version.VERSION
	}
	if version.GITCOMMIT != "" {
		versionMsg += " (" + version.GITCOMMIT + ")"
	}
	return StitleWithVersion(firstLine, secondLine, versionMsg)
}

// StitleWithVersion is the same as Stitle, but it allows to customize the version message line
func StitleWithVersion(firstLine, secondLine, versionLine string) string {
	blue := color.New(color.FgBlue).SprintFunc()
	return fmt.Sprintf(`  __
 /  \__     %s
 \__/  \    %s
 /  \__/    %s
 \__/%s`, blue(firstLine), secondLine, versionLine, "\n")
}

// Sectionf outputs a title in BLUE and underlined for separating a section (such as building a container, deploying files, etc.)
// T͟h͟i͟s͟ ͟i͟s͟ ͟u͟n͟d͟e͟r͟l͟i͟n͟e͟d͟ ͟b͟l͟u͟e͟ ͟t͟e͟x͟t͟
func Sectionf(format string, a ...interface{}) {
	if !IsJSON() {
		blue := color.New(color.FgBlue).Add(color.Underline).SprintFunc()
		if runtime.GOOS == "windows" {
			fmt.Fprintf(GetStdout(), "\n- %s\n", blue(fmt.Sprintf(format, a...)))
		} else {
			fmt.Fprintf(GetStdout(), "\n↪ %s\n", blue(fmt.Sprintf(format, a...)))
		}
	}
}

// Section outputs a title in BLUE and underlined for separating a section (such as building a container, deploying files, etc.)
// T͟h͟i͟s͟ ͟i͟s͟ ͟u͟n͟d͟e͟r͟l͟i͟n͟e͟d͟ ͟b͟l͟u͟e͟ ͟t͟e͟x͟t͟
func Section(a ...interface{}) {
	if !IsJSON() {
		blue := color.New(color.FgBlue).Add(color.Underline).SprintFunc()
		if runtime.GOOS == "windows" {
			fmt.Fprintf(GetStdout(), "\n- %s", blue(fmt.Sprintln(a...)))
		} else {
			fmt.Fprintf(GetStdout(), "\n↪ %s", blue(fmt.Sprintln(a...)))
		}
	}
}

// Deprecate will output a warning symbol and then "Deprecated" at the end of the output in YELLOW
//
//	⚠ <message all yellow>
func Deprecate(what, nextAction string) {
	if !IsJSON() {
		yellow := color.New(color.FgYellow).SprintFunc()
		msg1 := fmt.Sprintf("%s%s%s%s%s", yellow(getWarningString()), suffixSpacing, yellow(fmt.Sprintf("%s Deprecated", what)), suffixSpacing, nextAction)
		fmt.Fprintf(GetStderr(), " %s\n", msg1)
	}
}

// Errorf will output in an appropriate "progress" manner
// ✗ <message>
func Errorf(format string, a ...interface{}) {
	if !IsJSON() {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(GetStderr(), " %s%s%s\n", red(getErrString()), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Ferrorf will output in an appropriate "progress" manner
// ✗ <message>
func Ferrorf(w io.Writer, format string, a ...interface{}) {
	if !IsJSON() {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(w, " %s%s%s\n", red(getErrString()), suffixSpacing, fmt.Sprintf(format, a...))
	}
}

// Error will output in an appropriate "progress" manner
// ✗ <message>
func Error(a ...interface{}) {
	if !IsJSON() {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(GetStderr(), "%s%s%s%s", prefixSpacing, red(getErrString()), suffixSpacing, fmt.Sprintln(a...))
	}
}

// Frror will output in an appropriate "progress" manner
// ✗ <message>
func Ferror(w io.Writer, a ...interface{}) {
	if !IsJSON() {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(w, "%s%s%s%s", prefixSpacing, red(getErrString()), suffixSpacing, fmt.Sprintln(a...))
	}
}

// Info will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
// **Line in bold**
func Info(a ...interface{}) {
	if !IsJSON() {
		bold := color.New(color.Bold).SprintFunc()
		fmt.Fprintf(GetStdout(), "%s", bold(fmt.Sprintln(a...)))
	}
}

// Infof will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
// **Line in bold**
func Infof(format string, a ...interface{}) {
	if !IsJSON() {
		bold := color.New(color.Bold).SprintFunc()
		fmt.Fprintf(GetStdout(), "%s\n", bold(fmt.Sprintf(format, a...)))
	}
}

// Finfof will simply print out information on a new (bolded) line
// this is intended as information *after* something has been deployed
// This will also use a WRITER input
// We will have to manually check to see if it's Windows platform or not to
// determine if we are allowed to bold the output or not.
// **Line in bold**
func Finfof(w io.Writer, format string, a ...interface{}) {
	if !IsJSON() {
		bold := color.New(color.Bold).SprintFunc()

		if runtime.GOOS == "windows" {
			fmt.Fprintf(w, "%s\n", fmt.Sprintf(format, a...))
		} else {
			fmt.Fprintf(w, "%s\n", bold(fmt.Sprintf(format, a...)))
		}

	}
}

// Sbold will return a bold string
func Sbold(s string) string {
	bold := color.New(color.Bold).SprintFunc()
	return bold(fmt.Sprint(s))
}

// Bold will print out a bolded string
func Bold(s string) {
	if !IsJSON() {
		bold := color.New(color.Bold).SprintFunc()
		fmt.Fprintf(GetStdout(), "%s\n", bold(fmt.Sprintln(s)))
	}
}

// BoldColor will print out a bolded string with a color (that's passed in)
func SboldColor(c color.Attribute, s string) string {
	chosenColor := color.New(c).SprintFunc()
	return chosenColor(fmt.Sprintln(Sbold(s)))
}

// Describef will print out the first variable as BOLD and then the second not..
// this is intended to be used with `odo describe` and other outputs that list
// a lot of information
func Describef(title string, format string, a ...interface{}) {
	if !IsJSON() {
		bold := color.New(color.Bold).SprintFunc()
		fmt.Fprintf(GetStdout(), "%s%s\n", bold(title), fmt.Sprintf(format, a...))
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

// Fspinnerf creates a spinner, sets the prefix then returns it.
// Remember to use .End(bool) to stop the spin / when you're done.
// For example: defer s.End(false)
// for situations where spinning isn't viable (debug)
func Fspinnerf(w io.Writer, format string, a ...interface{}) *Status {
	s := NewStatus(w)
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

// IsAppleSilicon returns true if we are on a Mac M1 / Apple Silicon natively
func IsAppleSilicon() bool {
	return runtime.GOOS == "darwin" && (strings.HasPrefix(runtime.GOARCH, "arm") || strings.HasPrefix(runtime.GOARCH, "arm64"))
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

// ColorPicker picks a color from colors slice defined at the starting of this file
// It increments the colorCounter variable so that next iteration returns a different color
func ColorPicker() color.Attribute {
	colorCounter++
	return colors[(colorCounter)%len(colors)]
}
