package devfile

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile exec command tests", func() {
	var namespace, context, cmpName, currentWorkingDirectory, originalKubeconfig string

	// Using program command according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)

		helper.Chdir(context)

		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
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

	Context("When devfile exec command is executed", func() {

		It("should execute the given command successfully in the container", func() {
			utils.ExecCommand(context, cmpName)
			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)
			listDir := cliRunner.ExecListDir(podName, namespace, "/projects")
			Expect(listDir).To(ContainSubstring("blah.js"))
		})

		It("should error out when no command is given by the user", func() {
			utils.ExecWithoutCommand(context, cmpName)
		})

		It("should error out when a invalid command is given by the user", func() {
			utils.ExecWithInvalidCommand(context, cmpName, "kube")
		})

		It("should error out when a component is not present or when a devfile flag is used", func() {
			utils.ExecCommandWithoutComponentAndDevfileFlag(context, cmpName)
		})
	})
})
