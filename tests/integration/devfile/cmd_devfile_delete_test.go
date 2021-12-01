package devfile

import (
	"fmt"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/integration/devfile/utils"

	"os"
	"path/filepath"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile delete command tests", func() {
	const devfile = "devfile.yaml"
	const appName = "app"
	var devfilePath string
	var componentName, invalidNamespace string

	var commonVar helper.CommonVar

	var createNewContext = func() string {
		newContext := path.Join(commonVar.Context, "newContext")
		helper.MakeDir(newContext)
		return newContext
	}
	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	When("a component is created", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", componentName, "--project", commonVar.Project, "--app", appName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})
		It("should delete the component", func() {
			helper.Cmd("odo", "delete", "-f").ShouldPass()
		})

		When("the component is pushed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
			})
			It("should delete the component, env, odo folders and odo-index-file.json with --all flag", func() {
				helper.Cmd("odo", "delete", "-f", "--all").ShouldPass()

				files := helper.ListFilesInDir(commonVar.Context)
				Expect(files).To(Not(ContainElement(".odo")))
				Expect(files).To(Not(ContainElement("devfile.yaml")))
			})

			Describe("deleting a component from other component directory", func() {
				var firstComp, firstDir, secondComp, secondDir string

				BeforeEach(func() {
					// for the sake of verbosity
					firstComp = componentName
					firstDir = commonVar.Context
					secondComp = helper.RandString(6)
					secondDir = createNewContext()
					helper.Chdir(secondDir)
				})
				AfterEach(func() {
					helper.Chdir(commonVar.Context)
				})
				When("the second component is created", func() {
					BeforeEach(func() {
						helper.Cmd("odo", "create", secondComp, "--project", commonVar.Project, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass()
						helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					})
					When("the second component is pushed", func() {
						BeforeEach(func() {
							helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
						})
						It("should delete the context directory's component", func() {
							output := helper.Cmd("odo", "delete", "--context", firstDir, "-f").ShouldPass().Out()
							Expect(output).To(ContainSubstring(firstComp))
							Expect(output).ToNot(ContainSubstring(secondComp))
						})

						It("should delete all the config files and component of the context directory with --all flag", func() {
							output := helper.Cmd("odo", "delete", "--all", "--context", firstDir, "-f").ShouldPass().Out()
							Expect(output).ToNot(ContainSubstring(secondComp))
							Expect(output).To(ContainSubstring(firstComp))

							files := helper.ListFilesInDir(firstDir)
							Expect(files).To(Not(ContainElement(".odo")))
							Expect(files).To(Not(ContainElement("devfile.yaml")))

							files = helper.ListFilesInDir(secondDir)
							Expect(files).To(ContainElement(".odo"))
							Expect(files).To(ContainElement("devfile.yaml"))
						})
						// Marked as pending because it does not work at the moment. It takes the component in current directory into account while deleting.
						XIt("should delete with the component name", func() {
							output := helper.Cmd("odo", "delete", firstComp, "--app", "app", "--project", commonVar.Project, "-f").ShouldPass().Out()
							Expect(output).ToNot(ContainSubstring(secondComp))
							Expect(output).To(ContainSubstring(firstComp))
						})
					})
				})
			})
		})
		It("should throw an error on an invalid delete command", func() {
			By("--project flag")
			helper.Cmd("odo", "delete", "--project", commonVar.Project, "-f").ShouldFail()

			By("component name, --app, --project and --all flag")
			helper.Cmd("odo", "delete", componentName, "--app", appName, "--project", commonVar.Project, "--all", "-f").ShouldFail()

			By("component name and --all flag")
			helper.Cmd("odo", "delete", componentName, "--all", "-f").ShouldFail()

			By("component name and --context flag")
			helper.Cmd("odo", "delete", componentName, "--context", commonVar.Context, "-f").ShouldFail()

			By("--project and --context flag")
			helper.Cmd("odo", "delete", "--project", commonVar.Project, "--context", commonVar.Context, "-f").ShouldFail()
		})

		When("the component has resources attached to it", func() {
			resourceTypes := []string{helper.ResourceTypeDeployment, helper.ResourceTypePod, helper.ResourceTypeService, helper.ResourceTypeIngress, helper.ResourceTypePVC}

			BeforeEach(func() {
				helper.Cmd("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--port", "3000", "--ingress").ShouldPass()

				if os.Getenv("KUBERNETES") != "true" {
					helper.Cmd("odo", "url", "create", "example-1", "--port", "3000").ShouldPass()
					resourceTypes = append(resourceTypes, helper.ResourceTypeRoute)
				}
				helper.Cmd("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context).ShouldPass()
			})
			When("the component is pushed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
				})
				It("should delete the component and its owned resources", func() {
					helper.Cmd("odo", "delete", "-f").ShouldPass()

					for _, resourceType := range resourceTypes {
						commonVar.CliRunner.WaitAndCheckForExistence(resourceType, commonVar.Project, 1)
					}
				})
				It("should delete the component and its owned resources with --wait flag", func() {
					// Pod should exist
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(componentName, commonVar.Project)
					Expect(podName).NotTo(BeEmpty())

					// delete with --wait flag
					helper.Cmd("odo", "delete", "-f", "-w", "--context", commonVar.Context).ShouldPass()

					// Deployment and Pod should be deleted
					helper.VerifyResourcesDeleted(commonVar.CliRunner, []helper.ResourceInfo{
						{

							ResourceType: helper.ResourceTypeDeployment,
							ResourceName: componentName,
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
							ResourceName: componentName,
							Namespace:    commonVar.Project,
						},
						{
							ResourceType: helper.ResourceTypePVC,
							ResourceName: "storage-1",
							Namespace:    commonVar.Project,
						},
						{
							ResourceType: helper.ResourceTypeRoute,
							ResourceName: "example-1",
							Namespace:    commonVar.Project,
						},
					})
				})
			})
		})

		Context("devfile has preStop events", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			})
			When("component is pushed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
				})
				It("should execute the preStop events", func() {
					output := helper.Cmd("odo", "delete", "-f").ShouldPass().Out()
					helper.MatchAllInOutput(output, []string{
						fmt.Sprintf("Executing preStop event commands for component %s", componentName),
						"Executing myprestop command",
						"Executing secondprestop command",
						"Executing thirdprestop command",
					})
				})
			})
		})
	})

	When("the component is created in a non-existent project", func() {
		invalidNamespace = "garbage"
		BeforeEach(func() {
			helper.Cmd("odo", "create", componentName, "--project", invalidNamespace, "--app", appName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass()
		})
		It("should let the user delete the local config files with -a flag", func() {
			// DeleteLocalConfig appends -a flag
			utils.DeleteLocalConfig("delete")
		})
		When("deleting outside a component directory", func() {
			BeforeEach(func() {
				helper.Chdir(createNewContext())

			})
			AfterEach(func() {
				helper.Chdir(commonVar.Context)
			})
			It("should let the user delete the local config files with --context flag", func() {
				// DeleteLocalConfig appends -a flag
				utils.DeleteLocalConfig("delete", "--context", commonVar.Context)
			})
		})
	})
	When("component is created with --devfile flag", func() {
		BeforeEach(func() {
			newContext := createNewContext()
			devfilePath = filepath.Join(newContext, devfile)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), devfilePath)
			helper.Cmd("odo", "create", "nodejs", "--devfile", devfilePath).ShouldPass()
		})
		It("should successfully delete devfile", func() {
			// devfile was copied to top level
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
			helper.Cmd("odo", "delete", "--all", "-f").ShouldPass()
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeFalse())
		})
	})
	When("component is created from an existing devfile present in its directory", func() {
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), path.Join(commonVar.Context, devfile))
			helper.Cmd("odo", "create", "nodejs").ShouldPass()
		})
		It("should not delete the devfile", func() {
			// devfile was copied to top level
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
			helper.Cmd("odo", "delete", "--all", "-f").ShouldPass()
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
		})
	})
})
