package devfile

import (
	"fmt"
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
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		imageTag = fmt.Sprintf("image-registry.openshift-image-registry.svc:5000/%s/my-nodejs:1.0", namespace)
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
			output := helper.CmdShouldPass("odo", "deploy", "--tag", imageTag)
			cliRunner.WaitAndCheckForExistence("buildconfig", namespace, 1)
			Expect(output).NotTo(ContainSubstring("does not point to a valid Dockerfile"))
			Expect(output).To(ContainSubstring("Successfully built container image"))
			Expect(output).To(ContainSubstring("Successfully deployed component"))
		})
	})

	Context("Verify error when dockerfile specified in devfile field doesn't point to a valid Dockerfile", func() {
		It("Should error out with 'URL does not point to a valid Dockerfile'", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			err := helper.ReplaceDevfileField("devfile.yaml", "alpha.build-dockerfile", "https://google.com")
			Expect(err).To(BeNil())

			cmdOutput := helper.CmdShouldFail("odo", "deploy", "--tag", imageTag)
			Expect(cmdOutput).To(ContainSubstring("does not reference a valid Dockerfile"))
		})
	})

	// This test depends on the nodejs stack to no have a alpha.build-dockerfile field.
	// This may not be the case in the future when the stack gets updated.
	Context("Verify error when no Dockerfile exists in project and no 'dockerfile' specified in devfile", func() {
		It("Should error out with 'dockerfile required for build.'", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")
			cmdOutput := helper.CmdShouldFail("odo", "deploy", "--tag", imageTag)
			Expect(cmdOutput).To(ContainSubstring("dockerfile required for build. No 'alpha.build-dockerfile' field found in devfile, or Dockerfile found in project directory"))
		})
	})

	Context("Verify error when no manifest definition exists in devfile", func() {
		It("Should error out with 'Unable to deploy as alpha.deployment-manifest is not defined in devfile.yaml'", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile-no-manifest.yaml"), filepath.Join(context, "devfile.yaml"))

			cmdOutput := helper.CmdShouldFail("odo", "deploy", "--tag", imageTag)
			Expect(cmdOutput).To(ContainSubstring("Unable to deploy as alpha.deployment-manifest is not defined in devfile.yaml"))
		})
	})

	Context("Verify error when invalid manifest definition exists in devfile", func() {
		It("Should error out with 'Invalid manifest url'", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			err := helper.ReplaceDevfileField("devfile.yaml", "alpha.deployment-manifest", "google.com")
			Expect(err).To(BeNil())

			cmdOutput := helper.CmdShouldFail("odo", "deploy", "--tag", imageTag)
			Expect(cmdOutput).To(ContainSubstring("invalid url"))
		})
	})

	Context("Verify error when manifest file doesnt exist on web", func() {
		It("Should error out with 'Unable to download manifest'", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			err := helper.ReplaceDevfileField("devfile.yaml", "alpha.deployment-manifest", "http://github.com/myfile.yaml")
			Expect(err).To(BeNil())

			cmdOutput := helper.CmdShouldFail("odo", "deploy", "--tag", imageTag)
			Expect(cmdOutput).To(ContainSubstring("unable to download url"))
		})
	})

	Context("Verify deploy completes when using manifest with deployment/service/route", func() {
		It("Should successfully deploy the application and return a URL", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "3000")
			helper.CopyExampleDevFile(filepath.Join("source", "devfilesV2", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			err := helper.ReplaceDevfileField("devfile.yaml", "alpha.deployment-manifest",
				fmt.Sprintf("file://%s/../../examples/source/manifests/deploy_deployment_clusterip.yaml", currentWorkingDirectory))
			Expect(err).To(BeNil())

			cmdOutput := helper.CmdShouldPass("odo", "deploy", "--tag", imageTag)
			Expect(cmdOutput).To(ContainSubstring(fmt.Sprintf("Successfully deployed component: http://%s-deploy-%s", cmpName, namespace)))
		})
	})
})
