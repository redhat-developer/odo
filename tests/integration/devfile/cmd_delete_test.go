package devfile

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
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

	When("a component is bootstrapped", func() {
		BeforeEach(func() {
			// Hardcoded names from devfile-deploy-with-multiple-resources.yaml
			cmpName = "mynodejs"
			deploymentName = "my-component"
			serviceName = "my-cs"
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-deploy-with-multiple-resources.yaml")).ShouldPass()
			// Note:	component will be automatically bootstrapped when `odo dev` or `odo deploy` is run
		})
		It("should fail when the directory does not contain a .odo/env.yaml file", func() {
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).ToNot(ContainElement(".odo"))
			errOut := helper.Cmd("odo", "delete", "component", "-f").ShouldFail().Err()
			Expect(errOut).To(ContainSubstring("The current directory does not represent an odo component"))
		})
		When("the components are not deployed", func() {
			var stdOut string
			BeforeEach(func() {
				// Bootstrap the component with a .odo/env/env.yaml file
				odoDir := filepath.Join(commonVar.Context, ".odo", "env")
				helper.MakeDir(odoDir)
				err := helper.CreateFileWithContent(filepath.Join(odoDir, "env.yaml"), fmt.Sprintf(`
ComponentSettings:
  Name: %s
  Project: %s
  AppName: app
`, cmpName, commonVar.Project))
				Expect(err).To(BeNil())
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
				devSession, _, _, _, err = helper.StartDevMode()
				Expect(err).ToNot(HaveOccurred())
				defer devSession.Kill()
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
					By("ensuring that devfile.yaml and .odo still exists", func() {
						files := helper.ListFilesInDir(commonVar.Context)
						Expect(files).To(ContainElement(".odo"))
						Expect(files).To(ContainElement("devfile.yaml"))
					})
				})
			})

		})
		When("component is deployed to the cluster in the namespace set in env.yaml which is not the same as the current active namespace", func() {
			var projectName string
			BeforeEach(func() {
				// deploy the component to the cluster
				session := helper.CmdRunner("odo", "dev")
				defer session.Kill()
				helper.WaitForOutputToContain("Press Ctrl+c to exit", 180, 10, session)
				Expect(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents())).To(ContainSubstring(cmpName))

				helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
				Expect(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents())).To(ContainSubstring(deploymentName))
				Expect(string(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents())).To(ContainSubstring(serviceName))

				// create and set a new namespace
				projectName = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
			})
			AfterEach(func() {
				commonVar.CliRunner.DeleteNamespaceProject(projectName)
			})
			When("the component is deleted", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "delete", "component", "-f").ShouldPass().Out()
				})
				It("should have deleted the component", func() {
					By("deleting the component", func() {
						Eventually(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(cmpName))
					})
					By("deleting the deployment", func() {
						Eventually(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(deploymentName))
					})
					By("deleting the service", func() {
						Eventually(string(commonVar.CliRunner.Run(getSVCArgs...).Out.Contents()), 60, 3).ShouldNot(ContainSubstring(serviceName))
					})
				})
			})
		})
	})
	When("deleting a component containing preStop event that is deployed with DEV", func() {
		var out string
		BeforeEach(func() {
			// Hardcoded names from devfile-with-valid-events.yaml
			cmpName = "nodejs"
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml")).ShouldPass()
			session := helper.CmdRunner("odo", "dev")
			defer session.Kill()
			helper.WaitForOutputToContain("Press Ctrl+c to exit", 180, 10, session)
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

	//Test reused and adapted from the now-removed `cmd_devfile_delete_test.go`.
	// cf. https://github.com/redhat-developer/odo/blob/24fd02673d25eb4c7bb166ec3369554a8e64b59c/tests/integration/devfile/cmd_devfile_delete_test.go#L172-L238
	When("deleting a component that has resources attached to it after a Dev session is started and abruptly killed", func() {
		resourceTypes := []string{
			helper.ResourceTypeDeployment,
			helper.ResourceTypePod,
			helper.ResourceTypeService,
			helper.ResourceTypeIngress,
			helper.ResourceTypePVC,
		}

		BeforeEach(func() {
			// Component name comes from devfile-with-endpoints.{k8s,ocp}.yaml
			cmpName = "nodejs-with-endpoints"
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

			var devfileName string
			isKubernetesCluster := helper.IsKubernetesCluster()
			if isKubernetesCluster {
				devfileName = "devfile-with-endpoints.k8s.yaml"
			} else {
				devfileName = "devfile-with-endpoints.ocp.yaml"
			}

			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path",
				helper.GetExamplePath("source", "devfiles", "nodejs", devfileName)).ShouldPass()

			// Mimic the behavior of `odo url create` by bootstrapping the component with a .odo/env/env.yaml file
			odoEnvDir := filepath.Join(commonVar.Context, ".odo", "env")
			helper.MakeDir(odoEnvDir)
			if isKubernetesCluster {
				err := helper.CreateFileWithContent(filepath.Join(odoEnvDir, "env.yaml"), fmt.Sprintf(`
ComponentSettings:
  AppName: app
  Name: %s
  Project: %s
  Url:
  - Name: example
    Host: 1.2.3.4.nip.io
    Kind: ingress
`, cmpName, commonVar.Project))
				Expect(err).To(BeNil())
			} else {
				err := helper.CreateFileWithContent(filepath.Join(odoEnvDir, "env.yaml"), fmt.Sprintf(`
ComponentSettings:
  AppName: app
  Name: %s
  Project: %s
  Url:
  - Name: example
    Host: 1.2.3.4.nip.io
    Kind: ingress
  - Name: example-1
    Kind: route
`, cmpName, commonVar.Project))
				Expect(err).To(BeNil())
				resourceTypes = append(resourceTypes, helper.ResourceTypeRoute)
			}

			helper.StartDevMode().Kill()
		})

		It("should delete the component and its owned resources", func() {
			// Pod should exist
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			Expect(podName).NotTo(BeEmpty())
			services := commonVar.CliRunner.GetServices(commonVar.Project)
			Expect(services).To(SatisfyAll(
				Not(BeEmpty()),
				ContainSubstring(fmt.Sprintf("%s-app", cmpName)),
			))

			isKubernetesCluster := helper.IsKubernetesCluster()

			ingressesOut := commonVar.CliRunner.Run("get", "ingress",
				"-n", commonVar.Project,
				"-o", "custom-columns=NAME:.metadata.name",
				"--no-headers").Out.Contents()
			ingresses, err := helper.ExtractLines(string(ingressesOut))
			Expect(err).To(BeNil())
			Expect(ingresses).To(HaveLen(1))
			Expect(ingresses[0]).To(HavePrefix("example-"))

			if !isKubernetesCluster {
				routesOut := commonVar.CliRunner.Run("get", "routes",
					"-n", commonVar.Project,
					"-o", "custom-columns=NAME:.metadata.name",
					"--no-headers").Out.Contents()
				routes, err := helper.ExtractLines(string(routesOut))
				Expect(err).To(BeNil())
				Expect(routes).To(HaveLen(3))
				Expect(routesOut).To(SatisfyAll(
					ContainSubstring("example"),
					ContainSubstring("example-1"),
				))
			}

			helper.Cmd("odo", "delete", "component", "-f").ShouldPass()
			for _, resourceType := range resourceTypes {
				commonVar.CliRunner.WaitAndCheckForExistence(resourceType, commonVar.Project, 1)
			}
			// Deployment and Pod should be deleted
			helper.VerifyResourcesDeleted(commonVar.CliRunner, []helper.ResourceInfo{
				{
					ResourceType: helper.ResourceTypeDeployment,
					ResourceName: cmpName,
					Namespace:    commonVar.Project,
				},
				{
					ResourceType: helper.ResourceTypePod,
					ResourceName: podName,
					Namespace:    commonVar.Project,
				},
			})
			// Dependent resources should be marked to be deleted (see https://github.com/redhat-developer/odo/issues/4593)
			helper.VerifyResourcesToBeDeleted(commonVar.CliRunner, []helper.ResourceInfo{
				{
					ResourceType: helper.ResourceTypeIngress,
					ResourceName: "example",
					Namespace:    commonVar.Project,
				},
				{
					ResourceType: helper.ResourceTypeService,
					ResourceName: cmpName,
					Namespace:    commonVar.Project,
				},
			})
			if !isKubernetesCluster {
				helper.VerifyResourcesToBeDeleted(commonVar.CliRunner, []helper.ResourceInfo{
					{
						ResourceType: helper.ResourceTypeRoute,
						ResourceName: "example-1",
						Namespace:    commonVar.Project,
					},
				})
			}
		})
	})

})
