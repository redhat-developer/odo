package devfile

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile delete command tests", func() {
	const devfile = "devfile.yaml"
	var componentName string

	var commonVar helper.CommonVar

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

	Context("when devfile delete command is executed", func() {
		JustBeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", componentName, "--project", commonVar.Project)
		})

		// Why do we have this test case? `It` part doesn't make sense
		It("should not throw an error with an existing namespace when no component exists", func() {
			helper.CmdShouldPass("odo", "delete", componentName, "--project", commonVar.Project, "-f")
		})

		// TODO: This is bound to fail until https://github.com/openshift/odo/issues/4593 is fixed
		//It("should wait for the pods to terminate while using --wait flag to delete", func() {
		//	helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
		//	helper.CmdShouldPass("odo", "delete", "--project", commonVar.Project, "-f", "--wait")
		//	// This check will fail if the wait is longer, this check should happen immediately after the command is run.
		//	Expect(commonVar.CliRunner.GetRunningPodNameByComponent(componentName, commonVar.Project)).To(BeEmpty())
		//	Expect(true).ToNot(BeFalse())
		//})

		It("should delete the component created from the devfile and also the owned resources", func() {
			resourceTypes := []string{"deployments", "pods", "services", "ingress"}
			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--port", "3000", "--ingress")

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example-1", "--port", "3000")
				resourceTypes = append(resourceTypes, "routes")
			}

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)

			helper.CmdShouldPass("odo", "delete", "--project", commonVar.Project, "-f")

			for _, resourceType := range resourceTypes {
				commonVar.CliRunner.WaitAndCheckForExistence(resourceType, commonVar.Project, 1)
			}
		})

		It("should delete the component created from the devfile and also the env and odo folders and the odo-index-file.json file with all flag", func() {
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--port", "3000", "--ingress")

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example-1", "--port", "3000")
			}

			helper.CmdShouldPass("odo", "delete", "--project", commonVar.Project, "-f", "--all")

			commonVar.CliRunner.WaitAndCheckForExistence("deployments", commonVar.Project, 1)

			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).To(Not(ContainElement(".odo")))
			Expect(files).To(Not(ContainElement("devfile.yaml")))
		})

		It("should execute preStop events if present", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)

			output := helper.CmdShouldPass("odo", "delete", "--project", commonVar.Project, "-f")
			helper.MatchAllInOutput(output, []string{
				fmt.Sprintf("Executing preStop event commands for component %s", componentName),
				"Executing myprestop command",
				"Executing secondprestop command",
				"Executing thirdprestop command",
			})
		})
	})

	Context("when the project doesn't exist", func() {
		var invalidNamespace = "garbage"

		It("should let the user delete the local config files with -a flag", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)
			// utils.DeleteLocalConfig appends -a flag
			utils.DeleteLocalConfig("delete")
		})

		It("should let the user delete the local config files with -a and --project flags", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)
			// utils.DeleteLocalConfig appends -a flag
			utils.DeleteLocalConfig("delete", "--project", invalidNamespace)
		})
	})

	Context("When devfile does not exist in user's working directory and user specifies the devfile path via --devfile", func() {
		It("should successfully delete the devfile as its not present in root on delete", func() {
			newContext := path.Join(commonVar.Context, "newContext")
			devfilePath := filepath.Join(newContext, devfile)
			helper.MakeDir(newContext)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), devfilePath)

			helper.CmdShouldPass("odo", "create", "nodejs", "--devfile", devfilePath)
			// devfile was copied to top level
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
			helper.CmdShouldPass("odo", "delete", "--all", "-f")
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeFalse())
		})

		It("should not delete the devfile if its already present", func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), path.Join(commonVar.Context, devfile))
			helper.CmdShouldPass("odo", "create", "nodejs")
			// devfile was copied to top level
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
			helper.CmdShouldPass("odo", "delete", "--all", "-f")
			Expect(helper.VerifyFileExists(path.Join(commonVar.Context, devfile))).To(BeTrue())
		})
	})

	Context("When deleting component data from other component's directory", func() {
		var firstComponent, secondComponent, firstContext, secondContext, appName string
		appName = "myapp"
		var setup = func(componentName, contextName string) {
			helper.Chdir(contextName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), contextName)
			helper.CmdShouldPass("odo", "create", "nodejs", componentName, "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project, "--app", appName)
		}
		JustBeforeEach(func() {
			// Create the second component in a new context dir
			secondContext = path.Join(commonVar.Context, "newContext")
			secondComponent = helper.RandString(6)
			helper.MakeDir(secondContext)
			setup(secondComponent, secondContext)
			// Create the first component in commonVar.Context
			// redefining the variables for better verbosity
			firstContext = commonVar.Context
			firstComponent = componentName
			setup(firstComponent, firstContext)
			// we are passing secondContext to --context,
			// hence it is required that we be in firstContext when all the commands are run.
			cwd, _ := os.Getwd()
			Expect(cwd).To(BeEquivalentTo(firstContext))
		})
		JustAfterEach(func() {
			// Delete any pushed component and related config files in both the directories
			for _, dir := range []string{secondContext, firstContext} {
				helper.Chdir(dir)
				helper.CmdRunner("odo", "delete", "-f")
			}
		})
		It("should delete the context directory's component with --context flag", func() {
			output := helper.CmdShouldPass("odo", "delete", "-f", "--context", secondContext)
			Expect(output).To(ContainSubstring(secondComponent))
			Expect(output).ToNot(ContainSubstring(firstComponent))
		})
		It("should delete all the config files and component with -a and --context flag of the context directory", func() {
			output := helper.CmdShouldPass("odo", "delete", "-af", "--context", secondContext)
			Expect(output).To(ContainSubstring(secondComponent))
			Expect(output).ToNot(ContainSubstring(firstComponent))

			files := helper.ListFilesInDir(secondContext)
			Expect(files).To(Not(ContainElement(".odo")))
			// TODO: This is bound to fail until https://github.com/openshift/odo/issues/4135 is fixed
			//Expect(files).To(Not(ContainElement("devfile.yaml")))
		})

		// TODO: This is bound to fail until https://github.com/openshift/odo/issues/4451 is fixed
		//It("should delete the component when deleting with component name, --app and --project flags", func() {
		//	output := helper.CmdShouldPass("odo", "delete", secondComponent, "--project", commonVar.Project, "-f", "--app", appName)
		//	Expect(output).To(ContainSubstring(secondComponent))
		//	Expect(output).ToNot(ContainSubstring(firstComponent))
		//})

		Context("odo component delete should clean owned resources", func() {
			appName := helper.RandString(5)
			It("should delete the devfile component and the owned resources with wait flag", func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CmdShouldPass("odo", "create", "nodejs", componentName, "--app", appName, "--project", commonVar.Project, "--context", commonVar.Context)
				helper.CmdShouldPass("odo", "url", "create", "example-1", "--context", commonVar.Context, "--host", "com", "--ingress")

				helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context)
				info := helper.LocalEnvInfo(commonVar.Context)
				Expect(info.GetApplication(), appName)
				Expect(info.GetName(), componentName)
				helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

				helper.CmdShouldPass("odo", "url", "create", "example-2", "--context", commonVar.Context, "--host", "com", "--ingress")
				helper.CmdShouldPass("odo", "storage", "create", "storage-2", "--size", "1Gi", "--path", "/data2", "--context", commonVar.Context)
				helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

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
						ResourceType: helper.ResourceTypePVC,
						ResourceName: "storage-2",
						Namespace:    commonVar.Project,
					},
				})
			})
		})

		It("should throw an error when passing --app and --project flags with --context flag", func() {
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project, "--app", appName)
			output := helper.CmdShouldFail("odo", "delete", "--project", commonVar.Project, "--app", appName)
			Expect(output).To(ContainSubstring("cannot provide --app, --project or --component flag when --context is provided"))
		})
	})
})
