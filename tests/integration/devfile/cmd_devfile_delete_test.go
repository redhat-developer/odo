package devfile

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile delete command tests", func() {
	var namespace, context, currentWorkingDirectory, componentName string
	var cliRunner helper.CliRunner

	// Using program commmand according to cluter type in devfile
	if os.Getenv("KUBERNETES") == "true" {
		cliRunner = helper.NewKubectlRunner("kubectl")
	} else {
		cliRunner = helper.NewOcRunner("oc")
	}

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile commands require experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

		if os.Getenv("KUBERNETES") == "true" {
			homeDir := helper.GetUserHomeDir()
			kubeConfigFile := helper.CopyKubeConfigFile(filepath.Join(homeDir, ".kube", "config"), filepath.Join(context, "config"))
			namespace = helper.CreateRandNamespace(kubeConfigFile)
		} else {
			namespace = helper.CreateRandProject()
		}
		currentWorkingDirectory = helper.Getwd()
		componentName = helper.RandString(6)

		helper.Chdir(context)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		if os.Getenv("KUBERNETES") == "true" {
			helper.DeleteNamespace(namespace)
			os.Unsetenv("KUBECONFIG")
		} else {
			helper.DeleteProject(namespace)
		}
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when devfile delete command is executed", func() {

		It("should delete the component created from the devfile and also the owned resources", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io")
			}

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "--project", namespace, "-f")

			cliRunner.WaitAndCheckForExistence("deployments", namespace, 1)
			cliRunner.WaitAndCheckForExistence("pods", namespace, 1)
			cliRunner.WaitAndCheckForExistence("services", namespace, 1)
			cliRunner.WaitAndCheckForExistence("ingress", namespace, 1)
		})
	})

	Context("when devfile delete command is executed with all flag", func() {

		It("should delete the component created from the devfile and also the env and odo folders and the odo-index-file.json file", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--project", namespace)

			if os.Getenv("KUBERNETES") != "true" {
				helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--context", context)
			}

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "--project", namespace, "-f", "--all")

			cliRunner.WaitAndCheckForExistence("deployments", namespace, 1)

			files := helper.ListFilesInDir(context)
			Expect(files).To(Not(ContainElement(".odo")))
		})
	})
})
