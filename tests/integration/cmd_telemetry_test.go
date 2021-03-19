package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

const promtMessageSubString = "Help odo improve by allowing it to collect usage data."

var _ = FDescribe("odo telemetry", func() {
	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	Context("When no ConsentTelemetry preference value is set", func() {
		var _ = JustBeforeEach(func() {
			// unset the preference in case it is already set
			helper.CmdShouldPass("odo", "preference", "unset", "ConsentTelemetry", "-f")
		})
		It("prompt should not appear when preference command is run", func() {
			output := helper.CmdShouldPass("odo", "preference", "view")
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))

			output = helper.CmdShouldPass("odo", "preference", "set", "buildtimeout", "5", "-f")
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))

			output = helper.CmdShouldPass("odo", "preference", "unset", "buildtimeout", "-f")
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))
		})
		It("prompt should appear when non-preference command is run", func() {
			output := helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring(promtMessageSubString))
		})
	})

	Context("Prompt should not appear when", func() {
		It("ConsentTelemetry is set to true", func() {
			helper.CmdShouldPass("odo", "preference", "set", "ConsentTelemetry", "true", "-f")
			output := helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context)
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))
		})
		It("ConsentTelemetry is set to false", func() {
			helper.CmdShouldPass("odo", "preference", "set", "ConsentTelemetry", "false", "-f")
			output := helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context)
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))
		})
	})
})
