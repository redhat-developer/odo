package project

import (
	. "github.com/onsi/ginkgo"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo project command tests", func() {

	var commonVar helper.CommonVar
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	ProjectTestScenario()

})
