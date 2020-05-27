package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile catalog command tests", func() {
	var project, context, currentWorkingDirectory, originalKubeconfig string
	var cliRunner helper.CliRunner

	// Using program commmand according to cliRunner in devfile
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
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("When executing catalog list components", func() {
		It("should list all supported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components")
			wantOutput := []string{
				"Odo Devfile Components",
				"NAME",
				"java-spring-boot",
				"openLiberty",
				"quarkus",
				"DESCRIPTION",
				"REGISTRY",
				"SUPPORTED",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing catalog list components with -a flag", func() {
		It("should list all supported and unsupported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-a")
			wantOutput := []string{
				"Odo Devfile Components",
				"NAME",
				"java-spring-boot",
				"java-maven",
				"quarkus",
				"php-mysql",
				"DESCRIPTION",
				"REGISTRY",
				"SUPPORTED",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})
})
