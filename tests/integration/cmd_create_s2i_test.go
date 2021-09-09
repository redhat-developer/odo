package integration

import (
	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
	//. "github.com/onsi/gomega"
)

var _ = Describe("odo create --s2i command tests", func() {
	//var oc helper.OcRunner
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		//	oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("creating a component from s2i wildfly", func() {

		BeforeEach(func() {
			helper.Cmd("odo", "component", "create", "--s2i", "wildfly", "--project", commonVar.Project).ShouldPass()
			// Workaround for https://github.com/openshift/odo/issues/5060
			helper.ReplaceString("devfile.yaml", "/usr/local/s2i", "/usr/libexec/s2i")
		})

		It("should run odo push successfully", func() {
			helper.Cmd("odo", "push").ShouldPass()
		})
	})
})
