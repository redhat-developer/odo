package devfile

import (
	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile config command tests", func() {
	const (
		testName      = "testname"
		testMemory    = "500Mi"
		testDebugPort = "8888"
		fakeParameter = "fakeParameter"
	)

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When executing config view", func() {
		It("Should view all default parameters", func() {
			helper.Cmd("odo", "create", "nodejs").ShouldPass()
			output := helper.Cmd("odo", "config", "view").ShouldPass().Out()
			wantOutput := []string{
				"nodejs",
				"Ports",
				"Memory",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing config set and unset", func() {
		It("Should successfully set and unset the parameters", func() {
			helper.Cmd("odo", "create", "nodejs").ShouldPass()
			helper.Cmd("odo", "config", "set", "Name", testName, "-f").ShouldPass()
			helper.Cmd("odo", "config", "set", "Ports", testDebugPort, "-f").ShouldPass()
			helper.Cmd("odo", "config", "set", "Memory", testMemory, "-f").ShouldPass()
			output := helper.Cmd("odo", "config", "view").ShouldPass().Out()
			wantOutput := []string{
				testName,
				testMemory,
				testDebugPort,
			}
			helper.MatchAllInOutput(output, wantOutput)

			helper.Cmd("odo", "config", "unset", "Ports", "-f").ShouldPass()
			output = helper.Cmd("odo", "config", "view").ShouldPass().Out()
			dontWantOutput := []string{
				testDebugPort,
			}
			helper.DontMatchAllInOutput(output, dontWantOutput)
			helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
		})

		It("Should fail to set and unset an invalid parameter", func() {
			helper.Cmd("odo", "create", "nodejs").ShouldPass()
			helper.Cmd("odo", "config", "set", fakeParameter, fakeParameter, "-f").ShouldFail()
			helper.Cmd("odo", "config", "unset", fakeParameter, "-f").ShouldFail()
		})
	})
})
