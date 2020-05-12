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

	// TODO: all oc commands in all devfile related test should get replaced by kubectl
	// TODO: to goal is not to use "oc"
	oc := helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		namespace = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
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
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io")

			helper.CmdShouldPass("odo", "push", "--project", namespace)

			helper.CmdShouldPass("odo", "delete", "--project", namespace, "-f")

			oc.WaitAndCheckForExistence("deployments", namespace, 1)
			oc.WaitAndCheckForExistence("pods", namespace, 1)
			oc.WaitAndCheckForExistence("services", namespace, 1)
			oc.WaitAndCheckForExistence("ingress", namespace, 1)
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

			oc.WaitAndCheckForExistence("deployments", namespace, 1)

			files := helper.ListFilesInDir(context)
			Expect(files).To(Not(ContainElement(".odo")))
		})
	})
})
