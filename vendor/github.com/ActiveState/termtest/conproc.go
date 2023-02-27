// Copyright 2020 ActiveState Software. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file

package termtest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ActiveState/termtest/expect"
	"github.com/ActiveState/termtest/internal/osutils"
)

var (
	// ErrNoProcess is returned when a process was expected to be running
	ErrNoProcess = errors.New("no command process seems to be running")
)

type errWaitTimeout struct {
	error
}

func (errWaitTimeout) Timeout() bool { return true }

// ErrWaitTimeout is returned when we time out waiting for the console process to exit
var ErrWaitTimeout = errWaitTimeout{fmt.Errorf("timeout waiting for exit code")}

// ConsoleProcess bonds a command with a pseudo-terminal for automation
type ConsoleProcess struct {
	opts    Options
	errs    chan error
	console *expect.Console
	cmd     *exec.Cmd
	cmdName string
	ctx     context.Context
	cancel  func()
}

// NewTest bonds a command process with a console pty and sets it up for testing
func NewTest(t *testing.T, opts Options) (*ConsoleProcess, error) {
	opts.ObserveExpect = TestExpectObserveFn(t)
	opts.ObserveSend = TestSendObserveFn(t)
	return New(opts)
}

// New bonds a command process with a console pty.
func New(opts Options) (*ConsoleProcess, error) {
	if err := opts.Normalize(); err != nil {
		return nil, err
	}

	cmd := exec.Command(opts.CmdName, opts.Args...)
	cmd.Dir = opts.WorkDirectory
	cmd.Env = opts.Environment

	// Create the process in a new process group.
	// This makes the behavior more consistent, as it isolates the signal handling from
	// the parent processes, which are dependent on the test environment.
	cmd.SysProcAttr = osutils.SysProcAttrForNewProcessGroup()
	cmdString := osutils.CmdString(cmd)
	if opts.HideCmdLine {
		cmdString = "*****"
	}
	fmt.Printf("Spawning '%s' from %s\n", cmdString, opts.WorkDirectory)

	conOpts := []expect.ConsoleOpt{
		expect.WithDefaultTimeout(opts.DefaultTimeout),
		expect.WithSendObserver(expect.SendObserver(opts.ObserveSend)),
		expect.WithExpectObserver(opts.ObserveExpect),
	}
	conOpts = append(conOpts, opts.ExtraOpts...)

	console, err := expect.NewConsole(conOpts...)

	if err != nil {
		return nil, err
	}

	if err = console.Pty.StartProcessInTerminal(cmd); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	cp := ConsoleProcess{
		opts:    opts,
		errs:    make(chan error),
		console: console,
		cmd:     cmd,
		cmdName: opts.CmdName,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Asynchronously wait for the underlying process to finish and communicate
	// results to `cp.errs` channel
	// Once the error has been received (by the `wait` function, the TTY is closed)
	go func() {
		defer close(cp.errs)

		err := cmd.Wait()

		select {
		case cp.errs <- err:
		case <-cp.ctx.Done():
			log.Println("ConsoleProcess cancelled!  You may have forgotten to call ExpectExitCode()")
			_ = console.Close()
			return
		}

		// wait till passthrough-pipe has caught up
		cp.console.Pty.WaitTillDrained()
		_ = console.Pty.CloseTTY()
	}()

	return &cp, nil
}

// Close cleans up all the resources allocated by the ConsoleProcess
// If the underlying process is still running, it is terminated with a SIGTERM signal.
func (cp *ConsoleProcess) Close() error {
	cp.cancel()

	_ = cp.opts.CleanUp()

	if cp.cmd == nil || cp.cmd.Process == nil {
		return nil
	}

	if cp.cmd.ProcessState != nil && cp.cmd.ProcessState.Exited() {
		return nil
	}

	if err := cp.cmd.Process.Kill(); err == nil {
		return nil
	}

	return cp.cmd.Process.Signal(syscall.SIGTERM)
}

// Executable returns the command name to be executed
func (cp *ConsoleProcess) Executable() string {
	return cp.cmdName
}

// Cmd returns the underlying command
func (cp *ConsoleProcess) Cmd() *exec.Cmd {
	return cp.cmd
}

// WorkDirectory returns the directory in which the command shall be run
func (cp *ConsoleProcess) WorkDirectory() string {
	return cp.opts.WorkDirectory
}

// Snapshot returns a string containing a terminal snap-shot as a user would see it in a "real" terminal
func (cp *ConsoleProcess) Snapshot() string {
	return cp.console.Pty.State.String()
}

// TrimmedSnapshot displays the terminal output a user would see
// however the goroutine that creates this output is separate from this
// function so any output is not synced
func (cp *ConsoleProcess) TrimmedSnapshot() string {
	// When the PTY reaches 80 characters it continues output on a new line.
	// On Windows this means both a carriage return and a new line. Windows
	// also picks up any spaces at the end of the console output, hence all
	// the cleaning we must do here.
	newlineRe := regexp.MustCompile(`\r?\n`)
	return newlineRe.ReplaceAllString(strings.TrimSpace(cp.Snapshot()), "")
}

// ExpectRe listens to the terminal output and returns once the expected regular expression is matched or
// a timeout occurs
// Default timeout is 10 seconds
func (cp *ConsoleProcess) ExpectRe(value string, timeout ...time.Duration) (string, error) {
	opts := []expect.ExpectOpt{expect.RegexpPattern(value)}
	if len(timeout) > 0 {
		opts = append(opts, expect.WithTimeout(timeout[0]))
	}

	return cp.console.Expect(opts...)
}

// ExpectLongString listens to the terminal output and returns once the expected value is found or
// a timeout occurs
// This function ignores mismatches caused by newline and space characters to account
// for wrappings at the maximum terminal width.
// Default timeout is 10 seconds
func (cp *ConsoleProcess) ExpectLongString(value string, timeout ...time.Duration) (string, error) {
	opts := []expect.ExpectOpt{expect.LongString(value)}
	if len(timeout) > 0 {
		opts = append(opts, expect.WithTimeout(timeout[0]))
	}

	return cp.console.Expect(opts...)
}

// Expect listens to the terminal output and returns once the expected value is found or
// a timeout occurs
// Default timeout is 10 seconds
func (cp *ConsoleProcess) Expect(value string, timeout ...time.Duration) (string, error) {
	opts := []expect.ExpectOpt{expect.String(value)}
	if len(timeout) > 0 {
		opts = append(opts, expect.WithTimeout(timeout[0]))
	}

	return cp.console.Expect(opts...)
}

// ExpectCustom listens to the terminal output and returns once the supplied condition is satisfied or
// a timeout occurs
// Default timeout is 10 seconds
func (cp *ConsoleProcess) ExpectCustom(opt expect.ExpectOpt, timeout ...time.Duration) (string, error) {
	opts := []expect.ExpectOpt{opt}
	if len(timeout) > 0 {
		opts = append(opts, expect.WithTimeout(timeout[0]))
	}

	return cp.console.Expect(opts...)
}

// WaitForInput returns once a shell prompt is active on the terminal
// Default timeout is 10 seconds
func (cp *ConsoleProcess) WaitForInput(timeout ...time.Duration) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	msg := "echo wait_ready_$HOME"
	if runtime.GOOS == "windows" {
		msg = "echo wait_ready_%USERPROFILE%"
	}

	cp.SendLine(msg)
	return cp.Expect("wait_ready_"+homeDir, timeout...)
}

// Send sends a new line to the terminal, as if a user typed it
func (cp *ConsoleProcess) Send(value string) {
	_, _ = cp.console.SendLine(value)
}

// SendLine sends a new line to the terminal, as if a user typed it, the newline sequence is OS aware
func (cp *ConsoleProcess) SendLine(value string) {
	_, _ = cp.console.SendOSLine(value)
}

// SendUnterminated sends a string to the terminal as if a user typed it
func (cp *ConsoleProcess) SendUnterminated(value string) {
	_, _ = cp.console.Send(value)
}

// Signal sends an arbitrary signal to the running process
func (cp *ConsoleProcess) Signal(sig os.Signal) error {
	return cp.cmd.Process.Signal(sig)
}

// SendCtrlC tries to emulate what would happen in an interactive shell, when the user presses Ctrl-C
// Note: On Windows the Ctrl-C event is only reliable caught when the receiving process is
// listening for os.Interrupt signals.
func (cp *ConsoleProcess) SendCtrlC() {
	cp.SendUnterminated(string([]byte{0x03})) // 0x03 is ASCII character for ^C
}

// Stop sends an interrupt signal for the tested process and fails if no process has been started yet.
// Note: This is not supported on Windows
func (cp *ConsoleProcess) Stop() error {
	if cp.cmd == nil || cp.cmd.Process == nil {
		return ErrNoProcess
	}
	return cp.cmd.Process.Signal(os.Interrupt)
}

// MatchState returns the current state of the expect-matcher
func (cp *ConsoleProcess) MatchState() *expect.MatchState {
	return cp.console.MatchState
}

func (cp *ConsoleProcess) rawString() string {
	if cp.console.MatchState.Buf == nil {
		return ""
	}
	return cp.console.MatchState.Buf.String()
}

type exitCodeMatcher struct {
	exitCode int
	expected bool
}

func (em *exitCodeMatcher) Match(_ interface{}) bool {
	return true
}

func (em *exitCodeMatcher) Criteria() interface{} {
	comparator := "=="
	if !em.expected {
		comparator = "!="
	}

	return fmt.Sprintf("exit code %s %d", comparator, em.exitCode)
}

// ExpectExitCode waits for the program under test to terminate, and checks that the returned exit code meets expectations
func (cp *ConsoleProcess) ExpectExitCode(exitCode int, timeout ...time.Duration) (string, error) {
	_, err := cp.wait(timeout...)
	if err == nil && exitCode == 0 {
		return cp.rawString(), nil
	}
	matchers := []expect.Matcher{&exitCodeMatcher{exitCode, true}}
	eexit, ok := err.(*exec.ExitError)
	if !ok {
		e := fmt.Errorf("process failed with error: %w", err)
		cp.opts.ObserveExpect(matchers, cp.MatchState(), e)
		return cp.rawString(), e
	}
	if eexit.ExitCode() != exitCode {
		e := fmt.Errorf("exit code wrong: was %d (expected %d)", eexit.ExitCode(), exitCode)
		cp.opts.ObserveExpect(matchers, cp.MatchState(), e)
		return cp.rawString(), e
	}
	return cp.rawString(), nil
}

// ExpectNotExitCode waits for the program under test to terminate, and checks that the returned exit code is not the value provide
func (cp *ConsoleProcess) ExpectNotExitCode(exitCode int, timeout ...time.Duration) (string, error) {
	_, err := cp.wait(timeout...)
	matchers := []expect.Matcher{&exitCodeMatcher{exitCode, false}}
	if err == nil {
		if exitCode == 0 {
			e := fmt.Errorf("exit code wrong: should not have been 0")
			cp.opts.ObserveExpect(matchers, cp.MatchState(), e)
			return cp.rawString(), e
		}
		return cp.rawString(), nil
	}
	eexit, ok := err.(*exec.ExitError)
	if !ok {
		e := fmt.Errorf("process failed with error: %w", err)
		cp.opts.ObserveExpect(matchers, cp.MatchState(), e)
		return cp.rawString(), e
	}
	if eexit.ExitCode() == exitCode {
		e := fmt.Errorf("exit code wrong: should not have been %d", exitCode)
		cp.opts.ObserveExpect(matchers, cp.MatchState(), e)
		return cp.rawString(), e
	}
	return cp.rawString(), nil
}

// Wait waits for the program under test to terminate, not caring about the exit code at all
func (cp *ConsoleProcess) Wait(timeout ...time.Duration) {
	_, err := cp.wait(timeout...)
	if err != nil {
		fmt.Printf("Process exited with error: %v (This is not fatal when using Wait())", err)
	}
}

// forceKill kills the underlying process and waits until it return the exit error
func (cp *ConsoleProcess) forceKill() {
	if err := cp.cmd.Process.Kill(); err != nil {
		panic(err)
	}
	<-cp.errs
}

// wait waits for a console to finish and cleans up all resources
// First it consistently flushes/drains the pipe until the underlying process finishes.
// Note, that without draining the output pipe, the process might hang.
// As soon as the process actually finishes, it waits for the underlying console to be closed
// and gives all readers a chance to read remaining bytes.
func (cp *ConsoleProcess) wait(timeout ...time.Duration) (*os.ProcessState, error) {
	if cp.cmd == nil || cp.cmd.Process == nil {
		panic(ErrNoProcess.Error())
	}

	t := cp.opts.DefaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}

	finalErrCh := make(chan error)
	defer close(finalErrCh)
	go func() {
		_, err := cp.console.Expect(
			expect.Any(expect.PTSClosed, expect.StdinClosed, expect.EOF),
			expect.WithTimeout(t),
		)
		finalErrCh <- err
	}()

	select {
	case perr := <-cp.errs:
		// wait for the expect call to find EOF in stream
		expErr := <-finalErrCh
		// close the readers after all bytes from the terminal have been consumed
		err := cp.console.CloseReaders()
		if err != nil {
			log.Printf("Failed to close the console readers: %v", err)
		}
		// we only expect timeout or EOF errors here, otherwise something went wrong
		if expErr != nil && !(os.IsTimeout(expErr) || expErr == io.EOF) {
			return nil, fmt.Errorf("unexpected error while waiting for exit code: %v", expErr)
		}
		return cp.cmd.ProcessState, perr
	case <-time.After(t):
		// we can ignore the error from the expect (this will also time out)
		<-finalErrCh
		log.Println("killing process after timeout")
		cp.forceKill()
		return nil, ErrWaitTimeout
	case <-cp.ctx.Done():
		// wait until expect returns (will be forced by closed console)
		<-finalErrCh
		return nil, fmt.Errorf("ConsoleProcess context canceled")
	}
}
