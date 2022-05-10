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

	When("using the alias namespace to create a namespace", func() {
		namespace := fmt.Sprintf("%s-namespace", helper.RandString(4))
		AfterEach(func() {
			commonVar.CliRunner.DeleteNamespaceProject(namespace)
		})
		It("should successfully create the namespace", func() {
			helper.Cmd("odo", "create", "namespace", namespace).ShouldPass()
			Expect(commonVar.CliRunner.GetNamespaceProject()).To(ContainSubstring(namespace))
		})
	})

	It("should fail to create an existent namespace", func() {
		helper.Cmd("odo", "create", "namespace", commonVar.Project).ShouldFail()
	})

	When("using the alias project to create a project", func() {
		project := fmt.Sprintf("%s-project", helper.RandString(4))
		AfterEach(func() {
			commonVar.CliRunner.DeleteNamespaceProject(project)
		})
		It("should successfully create the project", func() {
			helper.Cmd("odo", "create", "project", project).ShouldPass()
			Expect(commonVar.CliRunner.GetNamespaceProject()).To(ContainSubstring(project))
		})
	})
})
