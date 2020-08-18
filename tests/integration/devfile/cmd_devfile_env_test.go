package devfile

import (
	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile env command tests", func() {
	const (
		testName      = "testname"
		testNamepace  = "testNamepace"
		testDebugPort = "8888"
		fakeParameter = "fakeParameter"
	)

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()

		helper.Chdir(commonVar.Context)
		// Devfile requires experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When executing env view", func() {
		It("Should view all default parameters", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			output := helper.CmdShouldPass("odo", "env", "view")
			wantOutput := []string{
				"PARAMETER NAME",
				"PARAMETER VALUE",
				"NAME",
				"nodejs",
				"Namespace",
				commonVar.Project,
				"DebugPort",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing env set and unset", func() {
		It("Should successfully set and unset the parameters", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			helper.CmdShouldPass("odo", "env", "set", "Name", testName, "-f")
			helper.CmdShouldPass("odo", "env", "set", "Namespace", testNamepace, "-f")
			helper.CmdShouldPass("odo", "env", "set", "DebugPort", testDebugPort, "-f")
			output := helper.CmdShouldPass("odo", "env", "view")
			wantOutput := []string{
				"PARAMETER NAME",
				"PARAMETER VALUE",
				"NAME",
				testName,
				"Namespace",
				testNamepace,
				"DebugPort",
				testDebugPort,
			}
			helper.MatchAllInOutput(output, wantOutput)

			helper.CmdShouldPass("odo", "env", "unset", "DebugPort", "-f")
			output = helper.CmdShouldPass("odo", "env", "view")
			dontWantOutput := []string{
				testDebugPort,
			}
			helper.DontMatchAllInOutput(output, dontWantOutput)
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
		})

		It("Should fail to set and unset an invalid parameter", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			helper.CmdShouldFail("odo", "env", "set", fakeParameter, fakeParameter, "-f")
			helper.CmdShouldFail("odo", "env", "unset", fakeParameter, "-f")
		})
	})

})
