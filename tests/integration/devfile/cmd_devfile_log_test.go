package devfile

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile log command tests", func() {
	var namespace, context, cmpName, currentWorkingDirectory, originalKubeconfig string
	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile push requires experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
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

	Context("Verify odo log for devfile works", func() {

		It("should log run command output", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)

			helper.CmdShouldPass("odo", "push", "--project", namespace)
			output := helper.CmdShouldPass("odo", "log")
			Expect(output).To(ContainSubstring("ODO_COMMAND_RUN"))

			// test with follow flag
			output = helper.CmdShouldRunWithTimeout(1*time.Second, "odo", "log", "-f")
			Expect(output).To(ContainSubstring("program=devrun"))

		})

		It("should error out if component does not exist", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldFail("odo", "log")
		})

	})

})
