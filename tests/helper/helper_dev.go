package helper

import (
	"fmt"
	"regexp"

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
		BeforeEach(func() {
			devSession = helper.StartDevMode()
		})
		AfterEach(func() {
			devSession.Stop()
		})

		It("...", func() {
			// Test with `dev odo` running in the background
		})
		It("...", func() {
			// Test with `dev odo` running in the background
		})
	})

	# Starting a session and stopping it cleanly

	This format can be used to test the behaviour of `odo dev` when it is stopped cleanly

	When("running dev session and stopping it with cleanup", func() {
		BeforeEach(func() {
			devSession := helper.StartDevMode()
			defer devSession.Stop()
			[...]
		})

		It("...", func() {
			// Test after `odo dev` has been stopped cleanly
		})
		It("...", func() {
			// Test after `odo dev` has been stopped cleanly
		})
	})

	# Starting a session and stopping it immediately without cleanup

	This format can be used to test the behaviour of `odo dev` when it is stopped with a KILL signal

	When("running dev session and stopping it without cleanup", func() {
		BeforeEach(func() {
			devSession := helper.StartDevMode()
			defer devSession.Kill()
			[...]
		})

		It("...", func() {
			// Test after `odo dev` has been killed
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
		helper.RunDevMode(func(session *gexec.Session) {
			// test on dev mode
		})
	})
*/
type DevSession struct {
	session *gexec.Session
}

// StartDevMode takes a regExp which it compiles against the output of the session structure
// It returns a session structure, the contents of the standard and error outputs, and matches it finds for the regExp
// when the dev mode is completely started
func StartDevMode(regExp string) (DevSession, []byte, []byte, []string, error) {
	session := CmdRunner("odo", "dev")
	WaitForOutputToContain("Watching for changes in the current directory", 180, 10, session)
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
	return result, outContents, errContents, FindAllMatchingStrings(string(outContents), regExp), nil
}

// Kill a Dev session abruptly, without handling any cleanup
func (o DevSession) Kill() {
	o.session.Kill()
}

// Stop a Dev session cleanly (equivalent as hitting Ctrl-c)
func (o DevSession) Stop() {
	o.session.Interrupt()
}

// RunDevMode takes a regExp that it passes to StartDevMode to find matches in the output of its session structure
// It runs a dev session and executes the `inside` code when the dev mode is completely started.
// The inside handler is passed the internal session pointer
// and the contents of the standard and error outputs at the time the handler is called
func RunDevMode(regExp string, inside func(session *gexec.Session, outContents []byte, errContents []byte, matches []string)) error {
	session, outContents, errContents, matches, err := StartDevMode(regExp)
	if err != nil {
		return err
	}
	defer session.Stop()
	inside(session.session, outContents, errContents, matches)
	return nil
}

// FindAllMatchingStrings returns all matches for the provided regExp as a slice of strings
func FindAllMatchingStrings(s, regExp string) []string {
	if regExp == "" {
		return nil
	}
	regExp = fmt.Sprintf("%s.*", regExp)
	re := regexp.MustCompile(regExp)
	return re.FindAllString(s, -1)
}
