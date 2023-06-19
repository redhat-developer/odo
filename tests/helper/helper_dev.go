package helper

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ActiveState/termtest/expect"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// DevSession represents a session running `odo dev`
/*
	It can be used in different ways:

	# Starting a session for a series of tests and stopping the session after the tests:

	This format can be used when you want to run several independent tests
	when the `odo dev` command is running in the background

	```
	When("running dev session", func() {
		var devSession DevSession
		var outContents []byte
		var errContents []byte
		BeforeEach(func() {
			devSession, outContents, errContents = helper.StartDevMode(nil)
		})
		AfterEach(func() {
			devSession.Stop()
		})

		It("...", func() {
			// Test with `dev odo` running in the background
			// outContents and errContents are contents of std/err output when dev mode is started
		})
		It("...", func() {
			// Test with `dev odo` running in the background
		})
	})

	# Starting a session and stopping it cleanly

	This format can be used to test the behaviour of `odo dev` when it is stopped cleanly

	When("running dev session and stopping it with cleanup", func() {
		var devSession DevSession
		var outContents []byte
		var errContents []byte
		BeforeEach(func() {
			devSession, outContents, errContents = helper.StartDevMode(nil)
			defer devSession.Stop()
			[...]
		})

		It("...", func() {
			// Test after `odo dev` has been stopped cleanly
			// outContents and errContents are contents of std/err output when dev mode is started
		})
		It("...", func() {
			// Test after `odo dev` has been stopped cleanly
		})
	})

	# Starting a session and stopping it immediately without cleanup

	This format can be used to test the behaviour of `odo dev` when it is stopped with a KILL signal

	When("running dev session and stopping it without cleanup", func() {
		var devSession DevSession
		var outContents []byte
		var errContents []byte
		BeforeEach(func() {
			devSession, outContents, errContents = helper.StartDevMode(nil)
			defer devSession.Kill()
			[...]
		})

		It("...", func() {
			// Test after `odo dev` has been killed
			// outContents and errContents are contents of std/err output when dev mode is started
		})
		It("...", func() {
			// Test after `odo dev` has been killed
		})
	})


	# Running a dev session and executing some tests inside this session

	This format can be used to run a series of related tests in dev mode
	All tests will be ran in the same session (ideal for e2e tests)
	To run independent tests, previous formats should be used instead.

	It("should do ... in dev mode", func() {
		helper.RunDevMode(func(session *gexec.Session, outContents []byte, errContents []byte, ports map[string]string) {
			// test on dev mode
			// outContents and errContents are contents of std/err output when dev mode is started
			// ports contains a map where keys are container ports and associated values are local IP:port redirecting to these local ports
		})
	})

	# Waiting for file synchronisation to finish

	The method session.WaitSync() can be used to wait for the synchronization of files to finish.
	The method returns the contents of std/err output since the end of the dev mode started or previous sync, and until the end of the synchronization.
*/

type DevSession struct {
	session           *gexec.Session
	stopped           bool
	console           *expect.Console
	address           string
	StdOut            string
	ErrOut            string
	Endpoints         map[string]string
	APIServerEndpoint string
}

type DevSessionOpts struct {
	EnvVars          []string
	CmdlineArgs      []string
	RunOnPodman      bool
	TimeoutInSeconds int
	NoRandomPorts    bool
	NoWatch          bool
	NoCommands       bool
	CustomAddress    string
	StartAPIServer   bool
	APIServerPort    int
}

// StartDevMode starts a dev session with `odo dev`
// It returns a session structure, the contents of the standard and error outputs
// and the redirections endpoints to access ports opened by component
// when the dev mode is completely started
func StartDevMode(options DevSessionOpts) (devSession DevSession, err error) {
	if options.RunOnPodman {
		options.CmdlineArgs = append(options.CmdlineArgs, "--platform", "podman")
	}
	c, err := expect.NewConsole(expect.WithStdout(os.Stdout))
	if err != nil {
		return DevSession{}, err
	}

	args := []string{"dev"}
	if options.NoCommands {
		args = append(args, "--no-commands")
	}
	if !options.NoRandomPorts {
		args = append(args, "--random-ports")
	}
	if options.NoWatch {
		args = append(args, "--no-watch")
	}
	if options.CustomAddress != "" {
		args = append(args, "--address", options.CustomAddress)
	}
	if options.StartAPIServer {
		args = append(args, "--api-server")
		if options.APIServerPort != 0 {
			args = append(args, "--api-server-port", fmt.Sprintf("%d", options.APIServerPort))
		}
	}
	args = append(args, options.CmdlineArgs...)
	cmd := Cmd("odo", args...)
	cmd.Cmd.Stdin = c.Tty()
	cmd.Cmd.Stdout = c.Tty()
	cmd.Cmd.Stderr = c.Tty()

	session := cmd.AddEnv(options.EnvVars...).Runner().session
	timeoutInSeconds := 420
	if options.TimeoutInSeconds != 0 {
		timeoutInSeconds = options.TimeoutInSeconds
	}
	WaitForOutputToContain("[Ctrl+c] - Exit", timeoutInSeconds, 10, session)
	result := DevSession{
		session: session,
		console: c,
		address: options.CustomAddress,
	}
	outContents := session.Out.Contents()
	errContents := session.Err.Contents()
	err = session.Out.Clear()
	if err != nil {
		return DevSession{}, err
	}
	err = session.Err.Clear()
	if err != nil {
		return DevSession{}, err
	}
	result.StdOut = string(outContents)
	result.ErrOut = string(errContents)
	result.Endpoints = getPorts(string(outContents), options.CustomAddress)
	if options.StartAPIServer {
		// errContents because the server message is still printed as a log/warning
		result.APIServerEndpoint = getAPIServerPort(string(errContents))
	}
	return result, nil

}

// Kill a Dev session abruptly, without handling any cleanup
func (o DevSession) Kill() {
	if o.console != nil {
		err := o.console.Close()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
	o.session.Kill()
}

// Stop a Dev session cleanly (equivalent as hitting Ctrl-c)
func (o *DevSession) Stop() {
	if o.session == nil {
		return
	}
	if o.console != nil {
		err := o.console.Close()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
	if o.stopped {
		return
	}

	if o.session.ExitCode() == -1 {
		err := terminateProc(o.session)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
	o.stopped = true
}

func (o *DevSession) PressKey(p byte) {
	if o.console == nil || o.session == nil {
		return
	}
	_, err := o.console.Write([]byte{p})
	Expect(err).ToNot(HaveOccurred())
}

func (o DevSession) WaitEnd() {
	if o.session == nil {
		return
	}
	o.session.Wait(3 * time.Minute)
}

// WaitSync waits for the synchronization of files to be finished
// It returns the contents of the standard and error outputs
// and the list of forwarded ports
// since the end of the dev mode or the last time WaitSync/UpdateInfo has been called
func (o *DevSession) WaitSync() error {
	WaitForOutputToContainOne([]string{"Pushing files...", "Updating Component..."}, 180, 10, o.session)
	WaitForOutputToContain("Dev mode", 240, 10, o.session)
	return o.UpdateInfo()
}

func (o *DevSession) WaitRestartPortforward() error {
	WaitForOutputToContain("Forwarding from", 240, 10, o.session)
	return o.UpdateInfo()
}

// UpdateInfo returns the contents of the standard and error outputs
// and the list of forwarded ports
// since the end of the dev mode or the last time WaitSync/UpdateInfo has been called
func (o *DevSession) UpdateInfo() error {
	outContents := o.session.Out.Contents()
	errContents := o.session.Err.Contents()
	err := o.session.Out.Clear()
	if err != nil {
		return err
	}
	err = o.session.Err.Clear()
	if err != nil {
		return err
	}
	o.StdOut = string(outContents)
	o.ErrOut = string(errContents)
	endpoints := getPorts(o.StdOut, o.address)
	if len(endpoints) != 0 {
		// when pod was restarted and port forwarding is done again
		o.Endpoints = endpoints
	}
	return nil
}

func (o DevSession) CheckNotSynced(timeout time.Duration) {
	Consistently(func() string {
		return string(o.session.Out.Contents())
	}, timeout).ShouldNot(ContainSubstring("Pushing files..."))
}

// RunDevMode runs a dev session and executes the `inside` code when the dev mode is completely started
// The inside handler is passed the internal session pointer, the contents of the standard and error outputs,
// and a slice of strings - ports - giving the redirections in the form localhost:<port_number> to access ports opened by component
func RunDevMode(options DevSessionOpts, inside func(session *gexec.Session, outContents string, errContents string, ports map[string]string)) error {

	session, err := StartDevMode(options)
	if err != nil {
		return err
	}
	defer func() {
		session.Stop()
		session.WaitEnd()
	}()
	inside(session.session, session.StdOut, session.ErrOut, session.Endpoints)
	return nil
}

// WaitForDevModeToContain runs `odo dev` until it contains a given substring in output or errOut(depending on checkErrOut arg).
// `odo dev` runs in an infinite reconciliation loop, and hence running it with Cmd will not work for a lot of failing cases,
// this function is helpful in such cases.
// If stopSessionAfter is false, it is up to the caller to stop the DevSession returned.
// TODO(pvala): Modify StartDevMode to take substring arg into account, and replace this method with it.
func WaitForDevModeToContain(options DevSessionOpts, substring string, stopSessionAfter bool, checkErrOut bool) (DevSession, error) {
	args := []string{"dev", "--random-ports"}
	args = append(args, options.CmdlineArgs...)
	if options.RunOnPodman {
		args = append(args, "--platform", "podman")
	}
	if options.CustomAddress != "" {
		args = append(args, "--address", options.CustomAddress)
	}
	session := Cmd("odo", args...).AddEnv(options.EnvVars...).Runner().session
	if checkErrOut {
		WaitForErroutToContain(substring, 360, 10, session)
	} else {
		WaitForOutputToContain(substring, 360, 10, session)
	}
	result := DevSession{
		session: session,
		address: options.CustomAddress,
	}
	if stopSessionAfter {
		defer func() {
			result.Stop()
			result.WaitEnd()
		}()
	}

	outContents := session.Out.Contents()
	errContents := session.Err.Contents()
	err := session.Out.Clear()
	if err != nil {
		return DevSession{}, err
	}
	err = session.Err.Clear()
	if err != nil {
		return DevSession{}, err
	}
	result.StdOut = string(outContents)
	result.ErrOut = string(errContents)
	result.Endpoints = getPorts(result.StdOut, options.CustomAddress)
	return result, nil
}

// getPorts returns a map of ports redirected depending on the information in s
//
//	`- Forwarding from 127.0.0.1:20001 -> 3000` will return { "3000": "127.0.0.1:20001" }
func getPorts(s, address string) map[string]string {
	if address == "" {
		address = "127.0.0.1"
	}
	result := map[string]string{}
	re := regexp.MustCompile(fmt.Sprintf("(%s:[0-9]+) -> ([0-9]+)", address))
	matches := re.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		result[match[2]] = match[1]
	}
	return result
}

// getAPIServerPort returns the address at which api server is running
//
// `I0617 11:40:44.124391   49578 starterserver.go:36] API Server started at localhost:20000/api/v1`
func getAPIServerPort(s string) string {
	re := regexp.MustCompile(`(API Server started at localhost:[0-9]+\/api\/v1)`)
	matches := re.FindString(s)
	return strings.Split(matches, "at ")[1]
}
