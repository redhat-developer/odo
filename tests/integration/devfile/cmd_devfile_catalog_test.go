package devfile

import (
	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile catalog command tests", func() {
	var globals helper.Globals

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEach()
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		helper.Chdir(globals.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeach(globals)
	})

	Context("When executing catalog list components", func() {
		It("should list all supported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components")
			wantOutput := []string{
				"Odo Devfile Components",
				"NAME",
				"java-spring-boot",
				"openLiberty",
				"quarkus",
				"DESCRIPTION",
				"REGISTRY",
				"SUPPORTED",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing catalog list components with -a flag", func() {
		It("should list all supported and unsupported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-a")
			wantOutput := []string{
				"Odo Devfile Components",
				"NAME",
				"java-spring-boot",
				"java-maven",
				"quarkus",
				"php-mysql",
				"DESCRIPTION",
				"REGISTRY",
				"SUPPORTED",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})
})
