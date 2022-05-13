package integration

import (
	"fmt"
	"strings"
	"time"

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
				Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(namespace))
			})
		})

	}

	It("should fail to create namespace", func() {
		By("using an existent namespace name", func() {
			helper.Cmd("odo", "create", "namespace", commonVar.Project).ShouldFail()
		})
		By("using an invalid namespace name", func() {
			helper.Cmd("odo", "create", "namespace", "12345").ShouldFail()
			Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(commonVar.Project))
		})
	})

	for _, commandName := range []string{"namespace", "project"} {
		When(fmt.Sprintf("using the alias %[1]s to delete a %[1]s", commandName), func() {
			var namespace string

			BeforeEach(func() {
				namespace = helper.CreateRandProject()
				Expect(commonVar.CliRunner.CheckNamespaceProjectExists(namespace)).To(BeTrue())
			})

			checkNsDeletionFunc := func(wait bool) {
				args := []string{"delete", commandName, namespace, "--force"}
				if wait {
					args = append(args, "--wait")
				}
				out := helper.Cmd("odo", args...).ShouldPass().Out()
				if wait {
					Expect(commonVar.CliRunner.GetAllNamespaceProjects()).ShouldNot(ContainElement(namespace))
				} else {
					Eventually(func() []string {
						return commonVar.CliRunner.GetAllNamespaceProjects()
					}, 60*time.Second).ShouldNot(ContainElement(namespace))
				}
				Expect(out).To(
					ContainSubstring(fmt.Sprintf("%s %q deleted", strings.Title(commandName), namespace)))
			}

			It(fmt.Sprintf("should successfully delete the %s using the force flag and asynchronously", commandName), func() {
				checkNsDeletionFunc(false)
			})

			It(fmt.Sprintf("should successfully delete the %s using the force flag and waiting", commandName), func() {
				checkNsDeletionFunc(true)
			})

			It(fmt.Sprintf("should not succeed to delete a non-existent %s", commandName), func() {
				fakeNamespace := "my-fake-ns-" + helper.RandString(3)
				By("using the force flag and asynchronously", func() {
					helper.Cmd("odo", "delete", commandName, fakeNamespace, "--force").ShouldFail()
				})

				By("using the force flag and waiting", func() {
					helper.Cmd("odo", "delete", commandName, fakeNamespace, "--force", "--wait").ShouldFail()
				})
			})
		})
	}
})
