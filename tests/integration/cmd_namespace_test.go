package integration

import (
	. "github.com/onsi/ginkgo"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("create/delete/list/get/set namespace tests", func() {
	var commonVar helper.CommonVar

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})
	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	It("should successfully create the namespace", func() {
		helper.Cmd("odo", "create", "namespace", "my-namespace").ShouldPass()
	})
	It("should successfully create the project", func() {
		helper.Cmd("odo", "create", "project", "my-project").ShouldPass()
	})
	It("should fail when an existent namespace is created again", func() {
		helper.Cmd("odo", "create", "project", commonVar.Project).ShouldFail()
	})
})
