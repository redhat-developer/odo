package devfile

import (
	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile env command tests", func() {
	const (
		testName      = "testname"
		testProject   = "testproject"
		testDebugPort = "8888"
		fakeParameter = "fakeParameter"
	)

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		helper.SetDefaultDevfileRegistryAsStaging()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When executing env view", func() {
		It("Should view all default parameters", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project).ShouldPass()
			output := helper.Cmd("odo", "env", "view").ShouldPass().Out()
			wantOutput := []string{
				"PARAMETER NAME",
				"PARAMETER VALUE",
				"NAME",
				"nodejs",
				"Project",
				commonVar.Project,
				"DebugPort",
				"Application",
				"app",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing env set and unset", func() {
		It("Should successfully set and unset the parameters", func() {
			helper.Cmd("odo", "create", "nodejs").ShouldPass()
			helper.Cmd("odo", "env", "set", "Name", testName, "-f").ShouldPass()
			helper.Cmd("odo", "env", "set", "Project", testProject, "-f").ShouldPass()
			helper.Cmd("odo", "env", "set", "DebugPort", testDebugPort, "-f").ShouldPass()
			output := helper.Cmd("odo", "env", "view").ShouldPass().Out()
			wantOutput := []string{
				"PARAMETER NAME",
				"PARAMETER VALUE",
				"NAME",
				testName,
				"Project",
				testProject,
				"DebugPort",
				testDebugPort,
			}
			helper.MatchAllInOutput(output, wantOutput)

			helper.Cmd("odo", "env", "unset", "DebugPort", "-f").ShouldPass()
			output = helper.Cmd("odo", "env", "view").ShouldPass().Out()
			dontWantOutput := []string{
				testDebugPort,
			}
			helper.DontMatchAllInOutput(output, dontWantOutput)
			helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
		})

		It("Should fail to set and unset an invalid parameter", func() {
			helper.Cmd("odo", "create", "nodejs").ShouldPass()
			helper.Cmd("odo", "env", "set", fakeParameter, fakeParameter, "-f").ShouldFail()
			helper.Cmd("odo", "env", "unset", fakeParameter, "-f").ShouldFail()
		})
	})

})
