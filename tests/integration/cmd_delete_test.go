package integration

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo delete command tests", func() {
	var commonVar helper.CommonVar
	var cmpName, deploymentName, serviceName string
	var getDeployArgs, getSVCArgs []string

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		getDeployArgs = []string{"get", "deployment", "-n", commonVar.Project}
		getSVCArgs = []string{"get", "svc", "-n", commonVar.Project}
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("running odo delete from a non-component directory", func() {
		var files []string
		BeforeEach(func() {
			files = helper.ListFilesInDir(commonVar.Context)
			Expect(files).ToNot(ContainElement(".odo"))
		})
		When("the directory is empty", func() {
			BeforeEach(func() {
				Expect(len(files)).To(BeZero())
			})
			It("should fail", func() {
				errOut := helper.Cmd("odo", "delete", "component", "-f").ShouldFail().Err()
				helper.MatchAllInOutput(errOut, []string{"The current directory does not represent an odo component"})
			})
		})
		When("the directory is not empty", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			})
			It("should fail", func() {
				errOut := helper.Cmd("odo", "delete", "component", "-f").ShouldFail().Err()
				helper.MatchAllInOutput(errOut, []string{"The current directory does not represent an odo component"})
			})
		})
	})

	for _, ctx := range []struct {
		title       string
		devfileName string
		setupFunc   func()
	}{
		{
			title:       "a component is bootstrapped",
			devfileName: "devfile-deploy-with-multiple-resources.yaml",
		},
		{
			title:       "a component is bootstrapped using a devfile.yaml with URI-referenced Kubernetes components",
			devfileName: "devfile-deploy-with-multiple-resources-and-k8s-uri.yaml",
			setupFunc: func() {
				helper.CopyExample(
					filepath.Join("source", "devfiles", "nodejs", "kubernetes", "devfile-deploy-with-multiple-resources-and-k8s-uri"),
					filepath.Join(commonVar.Context, "kubernetes", "devfile-deploy-with-multiple-resources-and-k8s-uri"))
			},
		},
	} {
		// this is a workaround to ensure that the for loop works with `It` blocks
		ctx := ctx
		When(ctx.title, func() {
			BeforeEach(func() {
				// Hardcoded names from `ctx.devfileName`
				cmpName = "mynodejs"
				deploymentName = "my-component"
				serviceName = "my-cs"
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path",
					helper.GetExamplePath("source", "devfiles", "nodejs", ctx.devfileName)).ShouldPass()
				// Note:	component will be automatically bootstrapped when `odo dev` or `odo deploy` is run
				if ctx.setupFunc != nil {
					ctx.setupFunc()
				}
			})
			When("the components are not deployed", func() {
				var stdOut string
				BeforeEach(func() {
					stdOut = helper.Cmd("odo", "delete", "component", "-f").ShouldPass().Out()
				})
				It("should output that there are no resources to be deleted", func() {
					Expect(stdOut).To(ContainSubstring("No resource found for component %q in namespace %q", cmpName, commonVar.Project))
				})
			})
			When("the component is deployed in DEV mode and dev mode stopped", func() {
				var devSession helper.DevSession
				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(nil)
					Expect(err).ToNot(HaveOccurred())
					defer func() {
						devSession.Kill()
						devSession.WaitEnd()
					}()
					Expect(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()).To(ContainSubstring(cmpName))
				})

				When("the component is deleted using its name and namespace from another directory", func() {
					var out string
					BeforeEach(func() {
						otherDir := filepath.Join(commonVar.Context, "tmp")
						helper.MakeDir(otherDir)
						helper.Chdir(otherDir)
						out = helper.Cmd("odo", "delete", "component", "--name", cmpName, "--namespace", commonVar.Project, "-f").ShouldPass().Out()
					})

					It("should have deleted the component", func() {
						By("listing the resource to delete", func() {
							Expect(out).To(ContainSubstring("Deployment: " + cmpName))
						})
						By("deleting the deployment", func() {
							// odo delete does not wait for resources to be deleted; hence we wait for .
							Eventually(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents(), 60, 3).ShouldNot(ContainSubstring(cmpName))
						})
					})
					When("odo delete command is run again with nothing deployed on the cluster", func() {
						var stdOut string
						BeforeEach(func() {
							// wait until the resources are deleted from the first delete
							Eventually(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(deploymentName))
							Eventually(string(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(serviceName))
						})
						It("should output that there are no resources to be deleted", func() {
							Eventually(func() string {
								stdOut = helper.Cmd("odo", "delete", "component", "--name", cmpName, "--namespace", commonVar.Project, "-f").ShouldPass().Out()
								return stdOut
							}, 60, 3).Should(ContainSubstring("No resource found for component %q in namespace %q", cmpName, commonVar.Project))
						})
					})
				})
				When("the component is deleted while having access to the devfile.yaml", func() {
					var stdOut string
					BeforeEach(func() {
						stdOut = helper.Cmd("odo", "delete", "component", "-f").ShouldPass().Out()
					})
					It("should have deleted the component", func() {
						By("listing the resource to delete", func() {
							Expect(stdOut).To(ContainSubstring(cmpName))
						})
						By("deleting the deployment", func() {
							Eventually(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents(), 60, 3).ShouldNot(ContainSubstring(cmpName))
						})
						By("ensuring that devfile.yaml and .odo still exists", func() {
							files := helper.ListFilesInDir(commonVar.Context)
							Expect(files).To(ContainElement(".odo"))
							Expect(files).To(ContainElement("devfile.yaml"))
						})
					})
					When("odo delete command is run again with nothing deployed on the cluster", func() {
						var stdOut string
						BeforeEach(func() {
							// wait until the resources are deleted from the first delete
							Eventually(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(deploymentName))
							Eventually(string(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(serviceName))
							stdOut = helper.Cmd("odo", "delete", "component", "-f").ShouldPass().Out()
						})
						It("should output that there are no resources to be deleted", func() {
							Expect(stdOut).To(ContainSubstring("No resource found for component %q in namespace %q", cmpName, commonVar.Project))
						})
					})
				})
			})

			When("the component is deployed in DEPLOY mode", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
					Expect(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()).To(ContainSubstring(deploymentName))
					Expect(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents()).To(ContainSubstring(serviceName))
				})
				When("the component is deleted using its name and namespace from another directory", func() {
					var out string
					BeforeEach(func() {
						otherDir := filepath.Join(commonVar.Context, "tmp")
						helper.MakeDir(otherDir)
						helper.Chdir(otherDir)
						out = helper.Cmd("odo", "delete", "component", "--name", cmpName, "--namespace", commonVar.Project, "-f").ShouldPass().Out()
					})

					It("should have deleted the component", func() {
						By("listing the resource to delete", func() {
							Expect(out).To(ContainSubstring("Deployment: " + deploymentName))
							Expect(out).To(ContainSubstring("Service: " + serviceName))
						})
						By("deleting the deployment", func() {
							Eventually(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents(), 60, 3).ShouldNot(ContainSubstring(deploymentName))
						})
						By("deleting the service", func() {
							Eventually(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents(), 60, 3).ShouldNot(ContainSubstring(serviceName))
						})
					})
				})
				When("a resource is changed in the devfile and the component is deleted while having access to the devfile.yaml", func() {
					var changedServiceName, stdout string
					BeforeEach(func() {
						changedServiceName = "my-changed-cs"
						helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), fmt.Sprintf("name: %s", serviceName), fmt.Sprintf("name: %s", changedServiceName))

						stdout = helper.Cmd("odo", "delete", "component", "-f").ShouldPass().Out()
					})
					It("should show warning about undeleted service belonging to the component", func() {
						Expect(stdout).To(SatisfyAll(
							ContainSubstring("There are still resources left in the cluster that might be belonging to the deleted component"),
							Not(ContainSubstring(changedServiceName)),
							ContainSubstring(serviceName),
							ContainSubstring("odo delete component --name %s --namespace %s", cmpName, commonVar.Project),
						))
					})

				})
				When("the component is deleted while having access to the devfile.yaml", func() {
					var stdOut string
					BeforeEach(func() {
						stdOut = helper.Cmd("odo", "delete", "component", "-f").ShouldPass().Out()
					})
					It("should have deleted the component", func() {
						By("listing the resources to delete", func() {
							Expect(stdOut).To(ContainSubstring(cmpName))
							Expect(stdOut).To(ContainSubstring("Deployment: " + deploymentName))
							Expect(stdOut).To(ContainSubstring("Service: " + serviceName))
						})
						By("deleting the deployment", func() {
							Eventually(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents(), 60, 3).ShouldNot(ContainSubstring(deploymentName))
						})
						By("deleting the service", func() {
							Eventually(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents(), 60, 3).ShouldNot(ContainSubstring(serviceName))
						})
						By("ensuring that devfile.yaml still exists", func() {
							files := helper.ListFilesInDir(commonVar.Context)
							Expect(files).To(ContainElement("devfile.yaml"))
						})
					})
				})

			})
		})
	}

	When("deleting a component containing preStop event that is deployed with DEV", func() {
		var out string
		BeforeEach(func() {
			// Hardcoded names from devfile-with-valid-events.yaml
			cmpName = "nodejs"
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml")).ShouldPass()
			session := helper.CmdRunner("odo", "dev", "--random-ports")
			defer session.Kill()
			helper.WaitForOutputToContain("[Ctrl+c] - Exit", 180, 10, session)
			// Ensure that the pod is in running state
			Eventually(string(commonVar.CliRunner.Run("get", "pods", "-n", commonVar.Project).Out.Contents()), 60, 3).Should(ContainSubstring(cmpName))
			// running in verbosity since the preStop events information is only printed in v4
			out = helper.Cmd("odo", "delete", "component", "-v", "4", "-f").ShouldPass().Out()
		})
		It("should contain preStop events list", func() {
			helper.MatchAllInOutput(out, []string{
				"Executing myprestop command",
				"Executing secondprestop command",
				"Executing thirdprestop command",
			})
		})
	})
})
