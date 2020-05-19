package docker

import (
	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo docker devfile catalog command tests", func() {
	var globals helper.Globals

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEachDocker()
		helper.Chdir(globals.Context)

		// Devfile commands require experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeachDocker(globals)

	})

	Context("When executing catalog list components on Docker", func() {
		It("should list all supported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components")
			helper.MatchAllInOutput(output, []string{"Odo Devfile Components", "java-spring-boot", "openLiberty"})
		})
	})

	Context("When executing catalog list components with -a flag on Docker", func() {
		It("should list all supported and unsupported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-a")
			helper.MatchAllInOutput(output, []string{"Odo Devfile Components", "java-spring-boot", "java-maven", "php-mysql"})
		})
	})
})
