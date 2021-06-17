package devfile

import (
	"fmt"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	"os"
	"path/filepath"

	"github.com/openshift/odo/tests/helper"
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
		JustBeforeEach(func() {
			helper.CmdShouldPass("odo", "create", "nodejs", componentName, "--project", commonVar.Project, "--app", appName)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})
		JustAfterEach(func() {})
		It("should delete the component", func() {
			helper.CmdShouldPass("odo", "delete", "-f")
		})

		When("the component is pushed", func() {
			JustBeforeEach(func() {
				helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			})
			JustAfterEach(func() {})
			It("should delete the component, env, odo folders and odo-index-file.json with --all flag", func() {
				helper.CmdShouldPass("odo", "delete", "-f", "--all")

				files := helper.ListFilesInDir(commonVar.Context)
				Expect(files).To(Not(ContainElement(".odo")))
				Expect(files).To(Not(ContainElement("devfile.yaml")))
			})

			Describe("deleting a component from other component directory", func() {
				var firstComp, firstDir, secondComp, secondDir string

				JustBeforeEach(func() {
					// for the sake of verbosity
					firstComp = componentName
					firstDir = commonVar.Context
					secondComp = helper.RandString(6)
					secondDir = createNewContext()
					helper.Chdir(secondDir)
				})
				JustAfterEach(func() {
					helper.Chdir(commonVar.Context)
				})
				When("the second component is created", func() {
					JustBeforeEach(func() {
						helper.CmdShouldPass("odo", "create", "nodejs", secondComp, "--project", commonVar.Project)
						helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
						helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					})
					JustAfterEach(func() {})
					When("the second component is pushed", func() {
						JustBeforeEach(func() {
							helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
						})
						JustAfterEach(func() {})
						It("should delete the context directory's component", func() {
							output := helper.CmdShouldPass("odo", "delete", "--context", firstDir, "-f")
							Expect(output).To(ContainSubstring(firstComp))
							Expect(output).ToNot(ContainSubstring(secondComp))
						})

						It("should delete all the config files and component of the context directory with --all flag", func() {
							output := helper.CmdShouldPass("odo", "delete", "--all", "--context", firstDir, "-f")
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
							output := helper.CmdShouldPass("odo", "delete", firstComp, "--app", "app", "--project", commonVar.Project, "-f")
							Expect(output).ToNot(ContainSubstring(secondComp))
							Expect(output).To(ContainSubstring(firstComp))
						})
					})
				})
			})
		})
		It("should throw an error on an invalid delete command", func() {
			By("--project flag")
			helper.CmdShouldFail("odo", "delete", "--project", commonVar.Project, "-f")

			By("component name, --app, --project and --all flag")
			helper.CmdShouldFail("odo", "delete", componentName, "--app", appName, "--project", commonVar.Project, "--all", "-f")

			By("component name and --all flag")
			helper.CmdShouldFail("odo", "delete", componentName, "--all", "-f")

			By("component name and --context flag")
			helper.CmdShouldFail("odo", "delete", componentName, "--context", commonVar.Context, "-f")

			By("--project and --context flag")
			helper.CmdShouldFail("odo", "delete", "--project", commonVar.Project, "--context", commonVar.Context, "-f")
		})

		When("the component has resources attached to it", func() {
			resourceTypes := []string{helper.ResourceTypeDeployment, helper.ResourceTypePod, helper.ResourceTypeService, helper.ResourceTypeIngress, helper.ResourceTypePVC}

			JustBeforeEach(func() {
				helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--port", "3000", "--ingress")

				if os.Getenv("KUBERNETES") != "true" {
					helper.CmdShouldPass("odo", "url", "create", "example-1", "--port", "3000")
					resourceTypes = append(resourceTypes, helper.ResourceTypeRoute)
				}
				helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context)
			})
			JustAfterEach(func() {})
			When("the component is pushed", func() {
				JustBeforeEach(func() {
					helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
				})
				It("should delete the component and its owned resources", func() {
					helper.CmdShouldPass("odo", "delete", "-f")

					for _, resourceType := range resourceTypes {
						commonVar.CliRunner.WaitAndCheckForExistence(resourceType, commonVar.Project, 1)
					}
				})
				It("should delete the component and its owned resources with --wait flag", func() {
					// Pod should exist
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(componentName, commonVar.Project)
					Expect(podName).NotTo(BeEmpty())

					// delete with --wait flag
					helper.CmdShouldPass("odo", "delete", "-f", "-w", "--context", commonVar.Context)

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

					// Dependent resources should be marked to be deleted (see https://github.com/openshift/odo/issues/4593)
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
			JustBeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			})
			JustAfterEach(func() {})
			When("component is pushed", func() {
				JustBeforeEach(func() {
					helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
				})
				JustAfterEach(func() {})
				It("should execute the preStop events", func() {
					output := helper.CmdShouldPass("odo", "delete", "-f")
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
		JustBeforeEach(func() {
			helper.CmdShouldPass("odo", "create", "nodejs", componentName, "--project", invalidNamespace, "--app", appName)
		})
		JustAfterEach(func() {})
		It("should let the user delete the local config files with -a flag", func() {
			// DeleteLocalConfig appends -a flag
			utils.DeleteLocalConfig("delete")
		})
		When("deleting outside a component directory", func() {
			JustBeforeEach(func() {
				helper.Chdir(createNewContext())

			})
			JustAfterEach(func() {
				helper.Chdir(commonVar.Context)
			})
			It("should let the user delete the local config files with --context flag", func() {
				// DeleteLocalConfig appends -a flag
				utils.DeleteLocalConfig("delete", "--context", commonVar.Context)
			})
		})
	})
	When("component is created with --devfile flag", func() {
		JustBeforeEach(func() {
			newContext := createNewContext()
			devfilePath = filepath.Join(newContext, devfile)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), devfilePath)
			helper.CmdShouldPass("odo", "create", "nodejs", "--devfile", devfilePath)
		})
		It("should successfully delete devfile", func() {
			// devfile was copied to top level
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
			helper.CmdShouldPass("odo", "delete", "--all", "-f")
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeFalse())
		})
	})
	When("component is created from an existing devfile present in its directory", func() {
		JustBeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), path.Join(commonVar.Context, devfile))
			helper.CmdShouldPass("odo", "create", "nodejs")
		})
		JustAfterEach(func() {})
		It("should not delete the devfile", func() {
			// devfile was copied to top level
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
			helper.CmdShouldPass("odo", "delete", "--all", "-f")
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
		})
	})
})
