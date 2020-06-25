package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile deploy command tests", func() {
	var namespace, context, cmpName, currentWorkingDirectory, originalKubeconfig, imageTag string
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
		imageTag = "image-registry.openshift-image-registry.svc:5000/default/my-nodejs:1.0"
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
	})

	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Verify deploy completes when passing a valid Dockerfile URL from the devfile", func() {
		It("Should succesfully download the dockerfile and build the project", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")

			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			output := helper.CmdShouldPass("odo", "deploy", "--tag", imageTag, "--devfile", "devfile.yaml")
			Expect(output).NotTo(ContainSubstring("does not point to a valid Dockerfile"))
		})
	})

	Context("Verify error when dockerfile specified in devfile field doesn't point to a valid Dockerfile", func() {
		It("Should error out with 'URL does not point to a valid Dockerfile'", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			err := helper.ReplaceDevfileField("devfile.yaml", "dockerfile", "https://google.com")
			Expect(err).To(BeNil())

			cmdOutput := helper.CmdShouldFail("odo", "deploy", "--tag", imageTag, "--devfile", "devfile.yaml")
			Expect(cmdOutput).To(ContainSubstring("does not point to a valid Dockerfile"))
		})
	})

	Context("Verify error when no Dockerfile exists in project and no 'dockerfile' specified in devfile", func() {
		It("Should error out with 'dockerfile required for build.'", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")
			cmdOutput := helper.CmdShouldFail("odo", "deploy", "--tag", imageTag, "--devfile", "devfile.yaml")
			Expect(cmdOutput).To(ContainSubstring("dockerfile required for build. No 'dockerfile' field found in devfile, or Dockerfile found in project directory"))
		})
	})
})
