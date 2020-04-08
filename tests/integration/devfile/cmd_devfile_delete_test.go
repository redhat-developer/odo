package devfile

import (
	"github.com/openshift/odo/tests/helper"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile delete command tests", func() {
	var namespace string
	var context string
	var currentWorkingDirectory string

	// TODO: all oc commands in all devfile related test should get replaced by kubectl
	// TODO: to goal is not to use "oc"
	oc := helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		namespace = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
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
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--project", namespace, componentName)

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "--namespace", namespace, "-f")

			oc.WaitAndCheckForExistence("deployments", namespace, 1)
			oc.WaitAndCheckForExistence("pods", namespace, 1)
			oc.WaitAndCheckForExistence("services", namespace, 1)

			// once https://github.com/openshift/odo/issues/2808 is resolved
			// create a url also in this scenario
			//oc.WaitAndCheckForExistence("ingress", namespace, 1)
		})
	})

	Context("when devfile delete command is executed with all flag", func() {

		It("should delete the component created from the devfile and also the env folder", func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--project", namespace, componentName)

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)

			helper.CmdShouldPass("odo", "url", "create", "example", "--host", "1.2.3.4.nip.io", "--context", context)

			helper.CmdShouldPass("odo", "delete", "--devfile", "devfile.yaml", "--namespace", namespace, "-f", "--all")

			oc.WaitAndCheckForExistence("deployments", namespace, 1)

			files := helper.ListFilesInDir(filepath.Join(context, ".odo"))
			Expect(files).To(Not(ContainElement("env")))
		})
	})
})
