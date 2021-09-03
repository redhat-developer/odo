package integration

import (
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link and unlink command tests without the service binding operator", func() {
	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		// wait until timeout(sec) for odo to see all the operators installed by setup script in the namespace
		odoArgs := []string{"catalog", "list", "services"}
		operator := "service-binding-operator"
		helper.WaitForCmdOut("odo", odoArgs, 5, true, func(output string) bool {
			return strings.Contains(output, operator)
		})
	})
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

})
