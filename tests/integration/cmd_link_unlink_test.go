package integration

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link and unlink command tests", func() {

	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		// wait till odo can see that all operators installed by setup script in the namespace
		odoArgs := []string{"catalog", "list", "services"}
		operators := []string{"service-binding-operator"}
		for _, operator := range operators {
			helper.WaitForCmdOut("odo", odoArgs, 5, true, func(output string) bool {
				return strings.Contains(output, operator)
			})
		}
	})

	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Running the help command", func() {
		By("for link command", func() {
			appHelp := helper.Cmd("odo", "link", "-h").ShouldPass().Out()
			helper.MatchAllInOutput(appHelp, []string{"Link component to a service ", "backed by an Operator or Service Catalog", "or component", "works only with s2i components"})
		})
		By("for unlink command", func() {
			appHelp := helper.Cmd("odo", "unlink", "-h").ShouldPass().Out()
			Expect(appHelp).To(ContainSubstring("Unlink component or service from a component"))
		})
	})

	When("two components are deployed", func() {
		var frontendContext, backendContext, frontendURL, frontendComp, backendComp string
		var oc helper.OcRunner

		// checkDescribe: checks that the linked component and related variables are present in the output of odo describe
		var checkDescribe = func(contextDir string, compName string, pushed bool, bindAsFiles bool) {
			stdOut := helper.Cmd("odo", "describe", "--context", contextDir).ShouldPass().Out()
			Expect(stdOut).To(ContainSubstring("Linked Services"))
			Expect(stdOut).To(ContainSubstring(compName))
			if pushed {
				if bindAsFiles {
					Expect(stdOut).To(ContainSubstring("/bindings"))
				} else {
					Expect(stdOut).To(ContainSubstring("SERVICE_BACKEND_IP"))
					Expect(stdOut).To(ContainSubstring("SERVICE_BACKEND_PORT"))
				}
			}
		}

		// createAndPush: creates component, a URL for it and deploys it
		var createAndPush = func(compType string, compName string, contextDir string) {
			helper.CopyExample(filepath.Join("source", compType), contextDir)
			helper.Cmd("odo", "create", compType, compName, "--context", contextDir, "--project", commonVar.Project).ShouldPass()
			helper.Cmd("odo", "url", "create", "--port", "8080", "--context", contextDir).ShouldPass()
			helper.Cmd("odo", "push", "--context", contextDir).ShouldPass()
		}

		JustBeforeEach(func() {
			oc = helper.NewOcRunner("oc")

			frontendComp = fmt.Sprintf("frontend-%v", helper.RandString(3))
			frontendContext = helper.CreateNewContext()
			createAndPush("nodejs", frontendComp, frontendContext)
			frontendURL = helper.DetermineRouteURL(frontendContext)

			backendComp = fmt.Sprintf("backend-%v", helper.RandString(3))
			backendContext = helper.CreateNewContext()
			createAndPush("python", backendComp, backendContext)
		})

		JustAfterEach(func() {
			helper.DeleteDir(frontendContext)
			helper.DeleteDir(backendContext)
		})

		When("a link is created between the two components", func() {
			JustBeforeEach(func() {
				// we link
				helper.Cmd("odo", "link", backendComp, "--context", frontendContext).ShouldPass()
			})

			It("should find the link in odo describe", func() {
				checkDescribe(frontendContext, backendComp, false, false)
			})

			When("the link is pushed", func() {
				JustBeforeEach(func() {
					helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()
				})

				It("should successfully link", func() {
					By("ensuring that the proper envFrom entry was created", func() {
						envFromOutput := oc.GetEnvFromEntry(frontendComp, "app", commonVar.Project, "deployment")
						Expect(envFromOutput).To(ContainSubstring(backendComp))
						helper.HttpWaitFor(frontendURL, "Hello world from node.js!", 20, 1)
					})
					By("finding the link and environment variables in odo describe", func() {
						checkDescribe(frontendContext, backendComp, true, false)
					})
					By("finding the linked environment variable", func() {
						stdOut := helper.Cmd("odo", "exec", "--context", frontendContext, "--", "sh", "-c", "echo $SERVICE_BACKEND_IP").ShouldPass().Out()
						Expect(stdOut).To(Not(BeEmpty()))
					})
					By("not allowing re-linking", func() {
						outputErr := helper.Cmd("odo", "link", backendComp, "--context", frontendContext).ShouldFail().Err()
						Expect(outputErr).To(ContainSubstring("already linked"))
					})
				})

				It("should successfully delete component after linked component is deleted", func() {
					// Testing: https://github.com/openshift/odo/issues/2355
					helper.Cmd("odo", "delete", "-f", "--context", backendContext).ShouldPass()
					helper.Cmd("odo", "delete", "-f", "--context", frontendContext).ShouldPass()
				})

				When("unlinking the two components", func() {
					JustBeforeEach(func() {
						helper.Cmd("odo", "unlink", backendComp, "--context", frontendContext).ShouldPass()
					})
					It("should find the link in odo describe", func() {
						checkDescribe(frontendContext, backendComp, true, false)
					})
					It("should not allow unlinking again", func() {
						stdOut := helper.Cmd("odo", "unlink", backendComp, "--context", frontendContext).ShouldFail().Err()
						Expect(stdOut).To(ContainSubstring("failed to unlink the component \"backend\" since no link was found in the configuration referring this component"))
					})

					When("odo push is executed", func() {
						JustBeforeEach(func() {
							helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()
						})
						It("should no longer find the link in odo describe", func() {
							checkDescribe(frontendContext, backendComp, true, false)
						})
					})
				})
			})
		})
		When("a link is created between the two components with --bind-as-files", func() {
			JustBeforeEach(func() {
				helper.Cmd("odo", "link", backendComp, "--bind-as-files").ShouldPass()
			})

			When("the component is pushed", func() {
				JustBeforeEach(func() {
					helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()
				})

				It("should successfully link", func() {
					By("finding the link in odo describe", func() {
						checkDescribe(frontendContext, backendComp, true, true)
					})
					By("finding the binding filepath", func() {
						stdOut := helper.Cmd("odo", "exec", "--", "ls", "/bindings").ShouldPass().Out()
						Expect(stdOut).To(Not(ContainSubstring(backendComp)))
					})
					By("not allowing re-linking", func() {
						outputErr := helper.Cmd("odo", "link", backendComp).ShouldFail().Err()
						Expect(outputErr).To(ContainSubstring("already linked"))
					})
				})

				When("unlinking the two components", func() {
					JustBeforeEach(func() {
						helper.Cmd("odo", "unlink", backendComp, "--context", frontendContext).ShouldPass()
					})

					It("should find the link in odo describe", func() {
						checkDescribe(frontendContext, backendComp, true, true)
					})

					It("should not allow unlinking again", func() {
						stdOut := helper.Cmd("odo", "unlink", backendComp, "--context", frontendContext).ShouldFail().Err()
						Expect(stdOut).To(ContainSubstring(fmt.Sprintf("failed to unlink the component %q since no link was found in the configuration referring this component", backendComp)))
					})

					When("odo push is executed", func() {
						JustBeforeEach(func() {
							helper.Cmd("odo", "push", "--context", frontendContext).ShouldPass()
						})

						It("should successfully unlink", func() {
							By("no longer finding the link in odo describe", func() {
								stdOut := helper.Cmd("odo", "describe", "--context", frontendContext).ShouldPass().Out()
								Expect(stdOut).ToNot(ContainSubstring("Linked Services"))
								Expect(stdOut).ToNot(ContainSubstring(backendComp))
							})
							By("not allowing unlinking again", func() {
								stdOut := helper.Cmd("odo", "unlink", backendComp, "--context", frontendContext).ShouldFail().Err()
								Expect(stdOut).To(ContainSubstring(fmt.Sprintf("failed to unlink the component %q since no link was found in the configuration referring this component", backendComp)))
							})
						})
					})
				})

			})

		})
	})
})
