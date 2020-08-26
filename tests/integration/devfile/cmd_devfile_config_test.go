package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile config command tests", func() {
	const (
		testName      = "testname"
		testMemory    = "500Mi"
		testDebugPort = "8888"
		fakeParameter = "fakeParameter"
	)

	var project, context, currentWorkingDirectory, originalKubeconfig string

	// Using program command according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		project = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(project)
		helper.Chdir(currentWorkingDirectory)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
	})

	Context("When executing config view", func() {
		It("Should view all default parameters", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			output := helper.CmdShouldPass("odo", "config", "view")
			wantOutput := []string{
				"nodejs",
				"Ports",
				"Memory",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing config set and unset", func() {
		It("Should successfully set and unset the parameters", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			helper.CmdShouldPass("odo", "config", "set", "Name", testName, "-f")
			helper.CmdShouldPass("odo", "config", "set", "Ports", testDebugPort, "-f")
			helper.CmdShouldPass("odo", "config", "set", "Memory", testMemory, "-f")
			output := helper.CmdShouldPass("odo", "config", "view")
			wantOutput := []string{
				testName,
				testMemory,
				testDebugPort,
			}
			helper.MatchAllInOutput(output, wantOutput)

			helper.CmdShouldPass("odo", "config", "unset", "Ports", "-f")
			output = helper.CmdShouldPass("odo", "config", "view")
			dontWantOutput := []string{
				testDebugPort,
			}
			helper.DontMatchAllInOutput(output, dontWantOutput)
			helper.CmdShouldPass("odo", "push", "--project", project)
		})

		It("Should fail to set and unset an invalid parameter", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			helper.CmdShouldFail("odo", "config", "set", fakeParameter, fakeParameter, "-f")
			helper.CmdShouldFail("odo", "config", "unset", fakeParameter, "-f")
		})
	})
})
