package helper

import (
	"github.com/onsi/gomega/gexec"
)

// DevSession represents a session running `odo dev`
/*
	It can be used in different ways:

	# Starting a session for a series of tests and stopping the session after the tests:

	```
	When("running dev session", func() {
		var devSession DevSession
		BeforeEach(func() {
			devSession = helper.StartDevMode()
		})
		AfterEach(func() {
			session.Stop()
		})

		It([...])
	})

	# Starting a session and stopping it immediately without cleanup

	When("running dev session and stopping it without cleanup", func() {
		BeforeEach(func() {
			devSession := helper.StartDevMode()
			defer devSession.Kill()
			[...]
		})

		It([...])
	})

	# Starting a session and stopping it cleanly

	When("running dev session and stopping it with cleanup", func() {
		var devSession DevSession
		BeforeEach(func() {
			devSession := helper.StartDevMode()
			defer devSession.Stop()
			[...]
		})

		It([...])
	})

	# Running a dev session and executing some tests inside this session
	It("should do ... in dev mode", func() {
		helper.RunDevMode(func(session *gexec.Session) {
			// tests on dev mode
		})
	})
*/
type DevSession struct {
	session *gexec.Session
}

// StartDevMode starts a dev session with `odo dev`
func StartDevMode() DevSession {
	session := CmdRunner("odo", "dev")
	WaitForOutputToContain("Waiting for something to change", 180, 10, session)
	return DevSession{
		session: session,
	}
}

// Kill a Dev session abruptly, without handling any cleanup
func (o DevSession) Kill() {
	o.session.Kill()
}

// Stop a Dev session cleanly (equivalent as hitting Ctrl-c)
func (o DevSession) Stop() {
	o.session.Interrupt()
}

// RunDevMode runs a dev session and executes the `inside` code
func RunDevMode(inside func(session *gexec.Session)) {
	session := StartDevMode()
	defer session.Stop()
	WaitForOutputToContain("Waiting for something to change", 180, 10, session.session)
	inside(session.session)
}
