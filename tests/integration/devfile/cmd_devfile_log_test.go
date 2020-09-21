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

		It("should log run command output and fail for debug command", func() {

			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", namespace, cmpName, "--context", context)
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--context", context)
			output := helper.CmdShouldPass("odo", "log", "--context", context)
			Expect(output).To(ContainSubstring("ODO_COMMAND_RUN"))

			// It should fail for debug command as no debug command in devfile
			helper.CmdShouldFail("odo", "log", "--debug")

			/*
				Flaky Test odo log -f, see issue https://github.com/openshift/odo/issues/3809
				match, err := helper.RunCmdWithMatchOutputFromBuffer(30*time.Second, "program=devrun", "odo", "log", "-f")
				Expect(err).To(BeNil())
				Expect(match).To(BeTrue())
			*/

		})

		It("should error out if component does not exist", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName, "--context", context)
			helper.CmdShouldFail("odo", "log")
		})

		It("should log debug command output", func() {

			projectDir := filepath.Join(context, "projectDir")

			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDir)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName, "--context", projectDir)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(projectDir, "devfile.yaml"))
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDir)

			output := helper.CmdShouldPass("odo", "log", "--debug", "--context", projectDir)
			Expect(output).To(ContainSubstring("ODO_COMMAND_DEBUG"))

			/*
				Flaky Test odo log -f, see issue https://github.com/openshift/odo/issues/3809
				match, err := helper.RunCmdWithMatchOutputFromBuffer(30*time.Second, "program=debugrun", "odo", "log", "-f")
				Expect(err).To(BeNil())
				Expect(match).To(BeTrue())
			*/

		})

	})

})
