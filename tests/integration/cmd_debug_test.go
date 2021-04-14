package integration

import (
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo debug command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("odo debug on a nodejs:latest component", func() {

		It("should expect a ws connection when tried to connect on different debug port locally and remotely", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context)
			// commented due to https://github.com/openshift/odo/issues/4615
			// helper.CmdShouldPass("odo", "env", "set", "--force", "DebugPort", "9292", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "DEBUG_PORT=9292", "--context", commonVar.Context)
			dbgPort := helper.GetLocalEnvInfoValueWithContext("DebugPort", commonVar.Context)
			Expect(dbgPort).To(Equal("9292"))
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", "5050", "--context", commonVar.Context)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:5050", "WebSockets request was expected", 12, 5, 400)
			stopChannel <- true
		})

	})

})
