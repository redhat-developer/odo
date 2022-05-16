package integration

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo create/delete/list/set namespace/project tests", func() {
	var commonVar helper.CommonVar

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	for _, commandName := range []string{"namespace", "project"} {

		Describe("create "+commandName, func() {
			namespace := fmt.Sprintf("%s-%s", helper.RandString(4), commandName)

			It(fmt.Sprintf("should successfully create the %s", commandName), func() {
				helper.Cmd("odo", "create", commandName, namespace, "--wait").ShouldPass()
				defer func(ns string) {
					commonVar.CliRunner.DeleteNamespaceProject(ns)
				}(namespace)
				Expect(commonVar.CliRunner.CheckNamespaceProjectExists(namespace)).To(BeTrue())
				Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(namespace))
			})

			It(fmt.Sprintf("should fail to create %s", commandName), func() {
				By("using an existent name", func() {
					helper.Cmd("odo", "create", commandName, commonVar.Project).ShouldFail()
				})
				By("using an invalid name", func() {
					helper.Cmd("odo", "create", commandName, "12345").ShouldFail()
					Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(commonVar.Project))
				})
			})
		})

		Describe("delete "+commandName, func() {

			When("force-deleting a valid "+commandName, func() {
				var namespace string

				BeforeEach(func() {
					namespace = helper.CreateRandProject()
					Expect(commonVar.CliRunner.CheckNamespaceProjectExists(namespace)).To(BeTrue())
				})

				checkNsDeletionFunc := func(additionalArgs []string, nsCheckerFunc func()) {
					args := []string{"delete", commandName, namespace, "--force"}
					if additionalArgs != nil {
						args = append(args, additionalArgs...)
					}
					out := helper.Cmd("odo", args...).ShouldPass().Out()
					if nsCheckerFunc != nil {
						nsCheckerFunc()
					}
					Expect(out).To(
						ContainSubstring(fmt.Sprintf("%s %q deleted", strings.Title(commandName), namespace)))
				}

				It(fmt.Sprintf("should successfully delete the %s asynchronously", commandName), func() {
					checkNsDeletionFunc(nil, func() {
						Eventually(func() []string {
							return commonVar.CliRunner.GetAllNamespaceProjects()
						}, 60*time.Second).ShouldNot(ContainElement(namespace))
					})
				})

				It(fmt.Sprintf("should successfully delete the %s synchronously with --wait", commandName), func() {
					checkNsDeletionFunc([]string{"--wait"}, func() {
						Expect(commonVar.CliRunner.GetAllNamespaceProjects()).ShouldNot(ContainElement(namespace))
					})
				})
			})

			It("should not succeed to delete a non-existent "+commandName, func() {
				fakeNamespace := "my-fake-ns-" + helper.RandString(3)
				By("using the force flag and asynchronously", func() {
					helper.Cmd("odo", "delete", commandName, fakeNamespace, "--force").ShouldFail()
				})

				By("using the force flag and waiting", func() {
					helper.Cmd("odo", "delete", commandName, fakeNamespace, "--force", "--wait").ShouldFail()
				})
			})

		})

		Describe("set "+commandName, func() {

			BeforeEach(func() {
				Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(commonVar.Project))
			})

			AfterEach(func() {
				if commonVar.CliRunner.GetActiveNamespace() != commonVar.Project {
					commonVar.CliRunner.SetProject(commonVar.Project)
				}
			})

			It("should set again the " + commandName, func() {
				helper.Cmd("odo", "set", commandName, commonVar.Project).ShouldPass()
				Expect(commonVar.CliRunner.GetActiveNamespace()).Should(Equal(commonVar.Project))
			})

			It(fmt.Sprintf("should set the %s even if it does not exist in the cluster", commandName), func() {
				fakeNamespace := "my-fake-ns-" + helper.RandString(3)
				Expect(commonVar.CliRunner.GetAllNamespaceProjects()).ShouldNot(ContainElement(fakeNamespace))
				helper.Cmd("odo", "set", commandName, fakeNamespace).ShouldPass()
				Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(fakeNamespace))
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

				It(fmt.Sprintf("should set the %s", commandName), func() {
					var stdout, stderr string
					By("setting the current active "+commandName, func() {
						Expect(commonVar.CliRunner.GetActiveNamespace()).ToNot(Equal(activeNs))
						cmd := helper.Cmd("odo", "set", commandName, activeNs).ShouldPass()
						Expect(commonVar.CliRunner.GetActiveNamespace()).To(Equal(activeNs))
						stdout, stderr = cmd.OutAndErr()
					})

					By("displaying warning message", func() {
						Expect(stdout).To(
							ContainSubstring(fmt.Sprintf("Current active %s set to %q", commandName, activeNs)))
						Expect(stderr).To(
							ContainSubstring(fmt.Sprintf("This is being executed inside a component directory. "+
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
