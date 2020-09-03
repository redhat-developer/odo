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
	var namespace, context, cmpName, currentWorkingDirectory, originalKubeconfig, projectDirPath string
	var projectDir = "/projectDir"
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
		projectDirPath = context + projectDir
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
			helper.MakeDir(projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
			helper.CmdShouldPass("odo", "push", "--project", namespace)
			output := helper.CmdShouldPass("odo", "log")
			Expect(output).To(ContainSubstring("ODO_COMMAND_RUN"))

			/*
				Flaky Test odo log -f, see issue https://github.com/openshift/odo/issues/3809
				match, err := helper.RunCmdWithMatchOutputFromBuffer(30*time.Second, "program=devrun", "odo", "log", "-f")
				Expect(err).To(BeNil())
				Expect(match).To(BeTrue())
			*/

		})

		It("should error out if component does not exist", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldFail("odo", "log")
		})

		It("should log debug command output", func() {

			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), projectDirPath)
			helper.RenameFile("devfile-with-debugrun.yaml", "devfile.yaml")
			helper.CmdShouldPass("odo", "push", "--debug")

			output := helper.CmdShouldPass("odo", "log", "--debug")
			Expect(output).To(ContainSubstring("ODO_COMMAND_DEBUG"))

			/*
				Flaky Test odo log -f, see issue https://github.com/openshift/odo/issues/3809
				match, err := helper.RunCmdWithMatchOutputFromBuffer(30*time.Second, "program=debugrun", "odo", "log", "-f")
				Expect(err).To(BeNil())
				Expect(match).To(BeTrue())
			*/

		})

		// we do not need test for run command as odo push fails
		// if there is no run command in devfile.
		It("should give error if no debug command in devfile", func() {

			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace)
			helper.CmdShouldPass("odo", "push", "--project", namespace)
			helper.CmdShouldFail("odo", "log", "--debug")
		})

	})

})
