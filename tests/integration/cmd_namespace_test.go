package integration

import (
	"fmt"
	"path/filepath"
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

	for _, commandName := range []string{"namespace", "project"} {
		When(fmt.Sprintf("using the alias %[1]s to set the current active %[1]s", commandName), func() {

			It(fmt.Sprintf("should succeed to set the current active %s", commandName), func() {
				By("using a namespace already set as current", func() {
					Expect(commonVar.CliRunner.GetActiveNamespace()).Should(Equal(commonVar.Project))
					helper.Cmd("odo", "set", commandName, commonVar.Project).ShouldPass()
					Expect(commonVar.CliRunner.GetActiveNamespace()).Should(Equal(commonVar.Project))
				})

				By("using a namespace that does not exist in the cluster", func() {
					fakeNamespace := "my-fake-ns-" + helper.RandString(3)
					Expect(commonVar.CliRunner.GetAllNamespaceProjects()).ShouldNot(ContainElement(fakeNamespace))
					helper.Cmd("odo", "set", commandName, fakeNamespace).ShouldPass()
					Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(fakeNamespace))
				})
			})

			It(fmt.Sprintf("should not succeed to set the %s", commandName), func() {
				invalidNs := "234567"
				helper.Cmd("odo", "set", commandName, invalidNs).ShouldFail()
				Expect(commonVar.CliRunner.GetActiveNamespace()).ShouldNot(Equal(invalidNs))
			})

			When("running inside a component directory", func() {
				activeNs := "my-current-ns"

				BeforeEach(func() {
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"))
					helper.Chdir(commonVar.Context)

					// Bootstrap the component with a .odo/env/env.yaml file
					odoDir := filepath.Join(commonVar.Context, ".odo", "env")
					helper.MakeDir(odoDir)
					err := helper.CreateFileWithContent(filepath.Join(odoDir, "env.yaml"), fmt.Sprintf(`
ComponentSettings:
  Name: my-component
  Project: %s
  AppName: app
`, commonVar.Project))
					Expect(err).ShouldNot(HaveOccurred())
				})

				It(fmt.Sprintf("should succeed to set the %s", commandName), func() {
					var stdout, stderr string
					By("setting the current active " + commandName, func() {
						Expect(commonVar.CliRunner.GetActiveNamespace()).ToNot(Equal(activeNs))
						cmd := helper.Cmd("odo", "set", commandName, activeNs).ShouldPass()
						Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(activeNs))
						stdout, stderr = cmd.OutAndErr()
					})

					By("displaying warning message", func() {
						Expect(stdout).To(
							ContainSubstring(fmt.Sprintf("Current active %s set to %q", commandName, activeNs)))
						Expect(stderr).To(
							ContainSubstring(fmt.Sprintf("This is being executed inside a component directory. " +
								"This will not update the %s of the existing component", commandName)))
					})

					By("not changing the namespace of the existing component", func() {
						helper.FileShouldContainSubstring(".odo/env/env.yaml", "Project: "+commonVar.Project)
					})
				})
			})

		})
	}
})
