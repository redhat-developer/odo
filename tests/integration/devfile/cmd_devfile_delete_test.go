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
	var namespace, context, currentWorkingDirectory, componentName, projectDirPath string
	var projectDir = "/projectDir"

	// TODO: all oc commands in all devfile related test should get replaced by kubectl
	// TODO: to goal is not to use "oc"
	oc := helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		namespace = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		projectDirPath = context + projectDir
		componentName = helper.RandString(6)

		helper.Chdir(context)

		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile commands require experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when devfile delete command is executed", func() {

		It("should delete the component created from the devfile and also the owned resources", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), projectDirPath)

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io")

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "--namespace", namespace, "-f")

			oc.WaitAndCheckForExistence("deployments", namespace, 1)
			oc.WaitAndCheckForExistence("pods", namespace, 1)
			oc.WaitAndCheckForExistence("services", namespace, 1)
			oc.WaitAndCheckForExistence("ingress", namespace, 1)
		})
	})

	Context("when devfile delete command is executed with all flag", func() {

		It("should delete the component created from the devfile and also the env and odo folders and the odo-index-file.json file", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/che-samples/web-nodejs-sample.git", projectDirPath)
			helper.Chdir(projectDirPath)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), projectDirPath)

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--context", projectDirPath)

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "--namespace", namespace, "-f", "--all")

			oc.WaitAndCheckForExistence("deployments", namespace, 1)

			files := helper.ListFilesInDir(projectDirPath)
			Expect(files).To(Not(ContainElement(".odo")))
		})
	})
})
