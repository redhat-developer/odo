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

var _ = Describe("odo devfile deploy delete command tests", func() {
	var namespace, context, currentWorkingDirectory, componentName, originalKubeconfig, imageTag string

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
		imageTag = fmt.Sprintf("image-registry.openshift-image-registry.svc:5000/%s/my-nodejs:1.0", namespace)
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

	Context("when manifest.yaml isnt present in .odo folder", func() {
		It("should fail and alert the user that there isn't a manifest.yaml present", func() {

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldFail("odo", "deploy", "delete")
			expectedString := "stat .odo/manifest.yaml: no such file or directory"

			helper.MatchAllInOutput(output, []string{expectedString})
		})

	})

	Context("when manifest.yaml is present, but deployment doesn't exist", func() {
		It("should pass, by deleting the manifest.yaml, but warn the user that deployment doesn't exist ", func() {

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "manifest.yaml"), filepath.Join(context, ".odo", "manifest.yaml"))

			helper.CmdShouldPass("odo", "deploy", "delete")
			Expect(helper.VerifyFileExists(filepath.Join(context, ".odo", "manifest.yaml"))).To(Equal(false))
		})

	})

	Context("when manifest.yaml is present, and deployment exists", func() {
		It("should pass, by deleting the manifest.yaml, and deleting deployment in cluster", func() {

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")

			err := helper.ReplaceDevfileField("devfile.yaml", "alpha.deployment-manifest",
				fmt.Sprintf("file://%s/../../examples/source/manifests/deploy_deployment_clusterip.yaml", currentWorkingDirectory))
			Expect(err).To(BeNil())
			helper.CmdShouldPass("odo", "deploy", "--tag", imageTag)

			helper.CmdShouldPass("odo", "deploy", "delete")
			cliRunner.WaitAndCheckForExistence("deployments", namespace, 1)
			cliRunner.WaitAndCheckForExistence("services", namespace, 1)
			cliRunner.WaitAndCheckForExistence("routes", namespace, 1)
			Expect(helper.VerifyFileExists(filepath.Join(context, ".odo", "manifest.yaml"))).To(Equal(false))
		})

	})
})
