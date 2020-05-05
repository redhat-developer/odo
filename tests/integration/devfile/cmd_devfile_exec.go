package devfile

import (
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile exec command tests", func() {
	var namespace, context, cmpName, currentWorkingDirectory string

	// TODO: all oc commands in all devfile related test should get replaced by kubectl
	// TODO: to goal is not to use "oc"
	oc := helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		namespace = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)

		helper.Chdir(context)

		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile push requires experimental mode to be set
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

	Context("When devfile exec command is executed", func() {

		It("should execute the given command successfully in the container", func() {
			utils.ExecCommand(context, cmpName)
			podName := oc.GetRunningPodNameByComponent(cmpName, namespace)
			listDir := oc.ExecListDir(podName, namespace, "/projects")
			Expect(listDir).To(ContainSubstring("blah.js"))
		})

		It("should error out when no command is given by the user", func() {
			utils.ExecWithoutCommand(context, cmpName)
		})

		It("should error out when a invalid command is given by the user", func() {
			utils.ExecWithInvalidCommand(context, cmpName, "kube")
		})
	})
})
