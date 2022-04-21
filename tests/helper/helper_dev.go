package helper

import (
	"regexp"
	"time"

	"github.com/onsi/gomega"
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
			devSession, outContents, errContents = helper.StartDevMode()
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
			devSession, outContents, errContents = helper.StartDevMode()
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
			devSession, outContents, errContents = helper.StartDevMode()
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
		helper.RunDevMode(func(session *gexec.Session, outContents []byte, errContents []byte) {
			// test on dev mode
			// outContents and errContents are contents of std/err output when dev mode is started
		})
	})

	# Waiting for file synchronisation to finish

	The method session.WaitSync() can be used to wait for the synchronization of files to finish.
	The method returns the contents of std/err output since the end of the dev mode started or previous sync, and until the end of the synchronization.
*/

const localhostRegexp = "127.0.0.1:[0-9]*"

type DevSession struct {
	session *gexec.Session
}

// StartDevMode starts a dev session with `odo dev`
// It returns a session structure, the contents of the standard and error outputs
// and the redirections endpoints to access ports opened by component
// when the dev mode is completely started
func StartDevMode(opts ...string) (DevSession, []byte, []byte, []string, error) {
	args := []string{"dev", "--random-ports"}
	args = append(args, opts...)
	session := CmdRunner("odo", args...)
	WaitForOutputToContain("Press Ctrl+c to exit `odo dev` and delete resources from the cluster", 360, 10, session)
	result := DevSession{
		session: session,
	}
	outContents := session.Out.Contents()
	errContents := session.Err.Contents()
	err := session.Out.Clear()
	if err != nil {
		return DevSession{}, nil, nil, nil, err
	}
	err = session.Err.Clear()
	if err != nil {
		return DevSession{}, nil, nil, nil, err
	}
	return result, outContents, errContents, FindAllMatchingStrings(string(outContents), localhostRegexp), nil

}

// Kill a Dev session abruptly, without handling any cleanup
func (o DevSession) Kill() {
	o.session.Kill()
}

// Stop a Dev session cleanly (equivalent as hitting Ctrl-c)
func (o DevSession) Stop() {
	err := terminateProc(o.session)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

func (o DevSession) WaitEnd() {
	o.session.Wait(3 * time.Minute)
}

//  WaitSync waits for the synchronization of files to be finished
// It returns the contents of the standard and error outputs
// since the end of the dev mode started or previous sync, and until the end of the synchronization.
func (o DevSession) WaitSync() ([]byte, []byte, error) {
	WaitForOutputToContain("Pushing files...", 180, 10, o.session)
	WaitForOutputToContain("Watching for changes in the current directory", 240, 10, o.session)
	outContents := o.session.Out.Contents()
	errContents := o.session.Err.Contents()
	err := o.session.Out.Clear()
	if err != nil {
		return nil, nil, err
	}
	err = o.session.Err.Clear()
	if err != nil {
		return nil, nil, err
	}
	return outContents, errContents, nil
}

// RunDevMode runs a dev session and executes the `inside` code when the dev mode is completely started
// The inside handler is passed the internal session pointer, the contents of the standard and error outputs,
// and a slice of strings - urls - giving the redirections in the form localhost:<port_number> to access ports opened by component
func RunDevMode(inside func(session *gexec.Session, outContents []byte, errContents []byte, urls []string)) error {
	session, outContents, errContents, urls, err := StartDevMode()
	if err != nil {
		return err
	}
	defer func() {
		session.Stop()
		session.WaitEnd()
	}()
	inside(session.session, outContents, errContents, urls)
	return nil
}

// FindAllMatchingStrings returns all matches for the provided regExp as a slice of strings
func FindAllMatchingStrings(s, regExp string) []string {
	re := regexp.MustCompile(regExp)
	return re.FindAllString(s, -1)
}
