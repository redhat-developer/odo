package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile delete command tests", func() {
	var namespace, context, currentWorkingDirectory, componentName, originalKubeconfig string

	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile commands require experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

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

		It("should delete the component created from the devfile and also the owned resources", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io")

			helper.CmdShouldPass("odo", "push", "--project", namespace)

			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f")

			resourceTypes := []string{"deployments", "pods", "services", "ingress"}
			for _, resourceType := range resourceTypes {
				cliRunner.WaitAndCheckForExistence(resourceType, namespace, 1)
			}
		})
	})

	Context("when no component exists", func() {

		It("should not throw an error with an existing namespace", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f")
		})
	})

	Context("when devfile delete command is executed with all flag", func() {

		It("should delete the component created from the devfile and also the env and odo folders and the odo-index-file.json file", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--project", namespace)

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--context", context)

			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f", "--all")

			cliRunner.WaitAndCheckForExistence("deployments", namespace, 1)

			files := helper.ListFilesInDir(context)
			Expect(files).To(Not(ContainElement(".odo")))
			Expect(files).To(Not(ContainElement("devfile.yaml")))
		})
	})

	Context("when the project doesn't exist", func() {

		It("should let the user delete the local config files with -a flag", func() {
			invalidNamespace := "garbage"
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)

			output := helper.CmdShouldFail("odo", "delete")
			helper.MatchAllInOutput(output, []string{
				fmt.Sprint("the namespace doesn't exist"),
			})

			output = helper.CmdShouldPass("odo", "delete", "-af")
			expectedOutput := []string{
				"Successfully deleted env file",
				"Successfully deleted devfile.yaml file",
			}
			helper.MatchAllInOutput(output, expectedOutput)
		})

		It("should let the user delete the local config files with -a and -project flags", func() {
			invalidNamespace := "garbage"
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)

			helper.CmdShouldFail("odo", "delete", "--project", invalidNamespace)

			output := helper.CmdShouldPass("odo", "delete", "--project", invalidNamespace, "-af")
			expectedOutput := []string{
				"Successfully deleted env file",
				"Successfully deleted devfile.yaml file",
			}
			helper.MatchAllInOutput(output, expectedOutput)
		})
	})
})
