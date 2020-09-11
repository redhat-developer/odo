package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"
)

var _ = Describe("odo devfile catalog command tests", func() {
	var project, context, currentWorkingDirectory, originalKubeconfig string
	const registryName string = "RegistryName"
	const addRegistryURL string = "https://github.com/odo-devfiles/registry"

	// Using program commmand according to cliRunner in devfile
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

		// For some reason on TravisCI, there are flakes with regards to registrycachetime and doing
		// odo catalog list components.
		// TODO: Investigate this more.
		helper.CmdShouldPass("odo", "preference", "set", "registrycachetime", "0")
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
				"java-springboot",
				"java-openliberty",
				"java-quarkus",
				"DESCRIPTION",
				"REGISTRY",
				"DefaultDevfileRegistry",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing catalog list components with -o json flag", func() {
		It("should list devfile components in json format", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-o", "json")
			wantOutput := []string{
				"odo.dev/v1alpha1",
				"devfileItems",
				"java-openliberty",
				"java-springboot",
				"nodejs",
				"java-quarkus",
				"java-maven",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing catalog list components with registry that is not set up properly", func() {
		It("should list components from valid registry", func() {
			helper.CmdShouldPass("odo", "registry", "add", "fake", "http://fake")
			output := helper.CmdShouldPass("odo", "catalog", "list", "components")
			helper.MatchAllInOutput(output, []string{
				"Odo Devfile Components",
				"java-springboot",
				"java-quarkus",
			})
			helper.CmdShouldPass("odo", "registry", "delete", "fake", "-f")
		})
	})

	Context("When executing catalog describe component with a component name with a single project", func() {
		It("should only give information about one project", func() {
			output := helper.CmdShouldPass("odo", "catalog", "describe", "component", "java-openliberty")
			Expect(output).To(MatchRegexp("origin: .+"))
		})
	})
	Context("When executing catalog describe component with a component name with no starter projects", func() {
		It("should print message that the component has no starter projects", func() {
			output := helper.CmdShouldPass("odo", "catalog", "describe", "component", "java-maven")
			helper.MatchAllInOutput(output, []string{"The Odo devfile component \"java-maven\" has no starter projects."})
		})
	})
	Context("When executing catalog describe component with a component name with multiple components", func() {
		It("should print multiple devfiles from different registries", func() {
			helper.CmdShouldPass("odo", "registry", "add", registryName, addRegistryURL)
			output := helper.CmdShouldPass("odo", "catalog", "describe", "component", "nodejs")
			helper.MatchAllInOutput(output, []string{"name: nodejs-starter", "Registry: " + registryName})
		})
	})
	Context("When executing catalog describe component with a component name that does not have a devfile component", func() {
		It("should print message that there is no Odo devfile component available", func() {
			output := helper.CmdShouldPass("odo", "catalog", "describe", "component", "java")
			helper.MatchAllInOutput(output, []string{"There are no Odo devfile components with the name \"java\""})
		})
	})
	Context("When executing catalog describe component with more than one argument", func() {
		It("should give an error saying it received too many arguments", func() {
			output := helper.CmdShouldFail("odo", "catalog", "describe", "component", "too", "many", "args")
			helper.MatchAllInOutput(output, []string{"accepts 1 arg(s), received 3"})
		})
	})
	Context("When executing catalog describe component with no arguments", func() {
		It("should give an error saying it expects exactly one argument", func() {
			output := helper.CmdShouldFail("odo", "catalog", "describe", "component")
			helper.MatchAllInOutput(output, []string{"accepts 1 arg(s), received 0"})
		})
	})

	Context("When executing catalog list components with experimental mode set to true", func() {
		It("should prove that nodejs is present in both S2I Component list and Devfile Component list", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-o", "json")
			cmpName := []string{"nodejs"}
			err := utils.VerifyCatalogListComponent(output, cmpName)
			Expect(err).Should(BeNil())
		})
	})

})
