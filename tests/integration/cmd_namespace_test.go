package integration

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
	for _, command := range []string{"namespace", "project"} {
		When(fmt.Sprintf("using the alias %[1]s to create a %[1]s", command), func() {
			var namespace string
			BeforeEach(func() {
				namespace = fmt.Sprintf("%s-%s", helper.RandString(4), command)
				helper.Cmd("odo", "create", command, namespace, "--wait").ShouldPass()
			})
			AfterEach(func() {
				commonVar.CliRunner.DeleteNamespaceProject(namespace)
			})
			It(fmt.Sprintf("should successfully create the %s", command), func() {
				Expect(commonVar.CliRunner.CheckNamespaceProjectExists(namespace)).To(BeTrue())
			})
		})

	}

	It("should fail to create an existent namespace", func() {
		helper.Cmd("odo", "create", "namespace", commonVar.Project).ShouldFail()
	})
})
