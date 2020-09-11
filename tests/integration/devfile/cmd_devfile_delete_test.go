package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile delete command tests", func() {
	var namespace, context, currentWorkingDirectory, componentName, originalKubeconfig, invalidNamespace string

	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		componentName = helper.RandString(6)
		helper.Chdir(context)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when devfile delete command is executed", func() {
		It("should not throw an error with an existing namespace when no component exists", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f")
		})

		It("should delete the component created from the devfile and also the owned resources", func() {
			resourceTypes := []string{"deployments", "pods", "services", "ingress"}

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--port", "3000", "--ingress", "--context", context)

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example-1", "--port", "3000", "--context", context)
				resourceTypes = append(resourceTypes, "routes")
			}

			helper.CmdShouldPass("odo", "push", "--project", namespace)

			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f")

			for _, resourceType := range resourceTypes {
				cliRunner.WaitAndCheckForExistence(resourceType, namespace, 1)
			}
		})

		It("should delete the component created from the devfile and also the env and odo folders and the odo-index-file.json file with all flag", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--project", namespace)

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--port", "3000", "--ingress", "--context", context)

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example-1", "--port", "3000")
			}

			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f", "--all")

			cliRunner.WaitAndCheckForExistence("deployments", namespace, 1)

			files := helper.ListFilesInDir(context)
			Expect(files).To(Not(ContainElement(".odo")))
			Expect(files).To(Not(ContainElement("devfile.yaml")))
		})

		It("should execute preStop events if present", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--project", namespace)

			output := helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f")
			helper.MatchAllInOutput(output, []string{
				fmt.Sprintf("Executing preStop event commands for component %s", componentName),
				"Executing myprestop command",
				"Executing secondprestop command",
				"Executing thirdprestop command",
			})

		})

		It("should error out on devfile flag", func() {
			helper.CmdShouldFail("odo", "delete", "--devfile", "invalid.yaml")
		})
	})

	Context("when the project doesn't exist", func() {
		JustBeforeEach(func() {
			invalidNamespace = "garbage"
		})

		It("should let the user delete the local config files with -a flag", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)
			utils.DeleteLocalConfig("delete")
		})

		It("should let the user delete the local config files with -a and -project flags", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)
			utils.DeleteLocalConfig("delete", "--project", invalidNamespace)
		})
	})
})
