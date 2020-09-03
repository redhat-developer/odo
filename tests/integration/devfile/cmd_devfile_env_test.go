package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile env command tests", func() {
	const (
		testName      = "testname"
		testNamepace  = "testNamepace"
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

	Context("When executing env view", func() {
		It("Should view all default parameters", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			output := helper.CmdShouldPass("odo", "env", "view")
			wantOutput := []string{
				"PARAMETER NAME",
				"PARAMETER VALUE",
				"NAME",
				"nodejs",
				"Namespace",
				project,
				"DebugPort",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing env set and unset", func() {
		It("Should successfully set and unset the parameters", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			helper.CmdShouldPass("odo", "env", "set", "Name", testName, "-f")
			helper.CmdShouldPass("odo", "env", "set", "Namespace", testNamepace, "-f")
			helper.CmdShouldPass("odo", "env", "set", "DebugPort", testDebugPort, "-f")
			output := helper.CmdShouldPass("odo", "env", "view")
			wantOutput := []string{
				"PARAMETER NAME",
				"PARAMETER VALUE",
				"NAME",
				testName,
				"Namespace",
				testNamepace,
				"DebugPort",
				testDebugPort,
			}
			helper.MatchAllInOutput(output, wantOutput)

			helper.CmdShouldPass("odo", "env", "unset", "DebugPort", "-f")
			output = helper.CmdShouldPass("odo", "env", "view")
			dontWantOutput := []string{
				testDebugPort,
			}
			helper.DontMatchAllInOutput(output, dontWantOutput)
			helper.CmdShouldPass("odo", "push", "--project", project)
		})

		It("Should fail to set and unset an invalid parameter", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			helper.CmdShouldFail("odo", "env", "set", fakeParameter, fakeParameter, "-f")
			helper.CmdShouldFail("odo", "env", "unset", fakeParameter, "-f")
		})
	})

})
