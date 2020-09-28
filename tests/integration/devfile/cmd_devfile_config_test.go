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
			helper.CmdShouldPass("odo", "create", "nodejs")
			output := helper.CmdShouldPass("odo", "config", "view")
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
			helper.CmdShouldPass("odo", "create", "nodejs")
			helper.CmdShouldPass("odo", "config", "set", "Name", testName, "-f")
			helper.CmdShouldPass("odo", "config", "set", "Ports", testDebugPort, "-f")
			helper.CmdShouldPass("odo", "config", "set", "Memory", testMemory, "-f")
			output := helper.CmdShouldPass("odo", "config", "view")
			wantOutput := []string{
				testName,
				testMemory,
				testDebugPort,
			}
			helper.MatchAllInOutput(output, wantOutput)

			helper.CmdShouldPass("odo", "config", "unset", "Ports", "-f")
			output = helper.CmdShouldPass("odo", "config", "view")
			dontWantOutput := []string{
				testDebugPort,
			}
			helper.DontMatchAllInOutput(output, dontWantOutput)
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
		})

		It("Should fail to set and unset an invalid parameter", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			helper.CmdShouldFail("odo", "config", "set", fakeParameter, fakeParameter, "-f")
			helper.CmdShouldFail("odo", "config", "unset", fakeParameter, "-f")
		})
	})
})
