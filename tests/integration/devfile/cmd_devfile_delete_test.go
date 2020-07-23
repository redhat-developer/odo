package devfile

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// deleteLocalConfig helps user to delete the local config with flags
func deleteLocalConfig(args ...string) {
	helper.CmdShouldFail("odo", append(args)...)
	output := helper.CmdShouldPass("odo", append(args, "-af")...)
	expectedOutput := []string{
		"Successfully deleted env file",
		"Successfully deleted devfile.yaml file",
	}
	helper.MatchAllInOutput(output, expectedOutput)
}

var _ = Describe("odo devfile delete command tests", func() {
	var namespace, context, currentWorkingDirectory, componentName, originalKubeconfig, invalidNamespace string

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
		JustBeforeEach(func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
		})

		It("should not throw an error with an existing namespace when no component exists", func() {
			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f")
		})

		It("should delete the component created from the devfile and also the owned resources", func() {
			resourceTypes := []string{"deployments", "pods", "services", "ingress"}

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--ingress", "--context", context)

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example-1", "--context", context)
				resourceTypes = append(resourceTypes, "routes")
			}

			helper.CmdShouldPass("odo", "push", "--project", namespace)

			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f")

			for _, resourceType := range resourceTypes {
				cliRunner.WaitAndCheckForExistence(resourceType, namespace, 1)
			}
		})

		It("should delete the component created from the devfile and also the env and odo folders and the odo-index-file.json file with all flag", func() {

			helper.CmdShouldPass("odo", "push", "--project", namespace)

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--ingress", "--context", context)

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example-1")
			}

			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f", "--all")

			cliRunner.WaitAndCheckForExistence("deployments", namespace, 1)

			files := helper.ListFilesInDir(context)
			Expect(files).To(Not(ContainElement(".odo")))
			Expect(files).To(Not(ContainElement("devfile.yaml")))
		})
	})

	Context("when the project doesn't exist", func() {
		JustBeforeEach(func() {
			invalidNamespace = "garbage"
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", invalidNamespace, componentName)
		})

		It("should let the user delete the local config files with -a flag", func() {
			deleteLocalConfig("delete")
		})

		It("should let the user delete the local config files with -a and -project flags", func() {
			deleteLocalConfig("delete", "--project", invalidNamespace)
		})
	})
})
