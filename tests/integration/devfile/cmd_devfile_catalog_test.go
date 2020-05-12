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
	var project string
	var context string
	var currentWorkingDirectory string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		if os.Getenv("KUBERNETES") == "true" {
			homeDir := helper.GetUserHomeDir()
			kubeConfigFile := helper.CopyKubeConfigFile(filepath.Join(homeDir, ".kube", "config"), filepath.Join(context, "config"))
			project = helper.CreateRandNamespace(kubeConfigFile)
		} else {
			project = helper.CreateRandProject()
		}
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		if os.Getenv("KUBERNETES") == "true" {
			helper.DeleteNamespace(project)
			os.Unsetenv("KUBECONFIG")
		} else {
			helper.DeleteProject(project)
		}
		helper.Chdir(currentWorkingDirectory)
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
				"php-mysql",
				"DESCRIPTION",
				"REGISTRY",
				"SUPPORTED",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})
	Context("When executing catalog describe component with a component name with a single component", func() {
		It("should only give information about one component", func() {
			output := helper.CmdShouldPass("odo", "catalog", "describe", "component", "maven")
			helper.MatchAllInOutput(output, []string{"generateName: maven-", "- command: java -jar target/*.jar"})
		})
	})
	Context("When executing catalog describe component with a component name with multiple components", func() {
		It("should give a warning that there are multiple components with this name in different registries", func() {
			output := helper.CmdShouldPass("odo", "catalog", "describe", "component", "nodejs")
			helper.MatchAllInOutput(output, []string{"There are multiple components named \"nodejs\" in different multiple devfile registries"})
		})
	})
	Context("When executing catalog describe component with a component name that does not have a devfile component", func() {
		It("should return the component name and say there is no Odo devfile component available", func() {
			output := helper.CmdShouldPass("odo", "catalog", "describe", "component", "java")
			helper.MatchAllInOutput(output, []string{"There are no Odo devfile components with the name \"java\"", "S2I Based Components:", "-java"})
		})
	})
	Context("When executing catalog describe component with a component name that is invalid", func() {
		It("should fail and give an error", func() {
			output := helper.CmdShouldFail("odo", "catalog", "describe", "component", "invalidcomponent")
			helper.MatchAllInOutput(output, []string{"No components with the name \"invalidcomponent\" found"})
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
})
