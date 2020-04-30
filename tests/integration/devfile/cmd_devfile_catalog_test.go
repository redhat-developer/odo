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
			helper.MatchAllInOutput(output, []string{"Odo Devfile Components", "java-spring-boot", "openLiberty"})
		})
	})

	Context("When executing catalog list components with -a flag", func() {
		It("should list all supported and unsupported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-a")
			helper.MatchAllInOutput(output, []string{"Odo Devfile Components", "java-spring-boot", "java-maven", "php-mysql"})
		})
	})
	Context("When executing catalog list components on a devfile component with projects", func() {
		It("should list all the default project information for the openLiberty component", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "openLiberty")
			helper.MatchAllInOutput(output, []string{"Devfile Starter Project(s):", "TYPE", "git", "LOCATION", "https://github.com/rajivnathan/openLiberty.git"})
		})
	})
	Context("When executing catalog list components on a devfile component with no projects", func() {
		It("should list the maven devfile information, but show there is no starter projects available", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "maven")
			helper.MatchAllInOutput(output, []string{"This devfile component does not have any available starter projects."})
		})
	})
	Context("When executing catalog list components with -p flag", func() {
		It("should list all the project information for the java-spring-boot component in yaml format", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "java-spring-boot", "-p")
			helper.MatchAllInOutput(output, []string{"Devfile Starter Project(s):", "source", "type", "git", "location", "https://github.com/maysunfaisal/springboot.git"})
		})
	})
	Context("When executing catalog list components with an invalid component arg", func() {
		It("should give an error if invalid component given as argument", func() {
			output := helper.CmdShouldFail("odo", "catalog", "list", "components", "invalidComponent")
			helper.MatchAllInOutput(output, []string{"The component \"invalidComponent\" is not a valid Odo component."})
		})
	})
	Context("When executing catalog list components with more than one argument", func() {
		It("should give an error if more than one argument is given", func() {
			output := helper.CmdShouldFail("odo", "catalog", "list", "components", "too", "many", "args")
			helper.MatchAllInOutput(output, []string{"accepts between 0 and 1 arg(s)"})
		})
	})
})
