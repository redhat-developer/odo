package integration

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
		project = helper.CreateRandProject()
		context = helper.CreateNewDevfileContext()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
	})

	Context("When executing catalog list components", func() {
		It("should list all supported devfile components", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			output := helper.CmdShouldPass("odo", "catalog", "list", "components")
			helper.MatchAllInOutput(output, []string{"Odo Devfile Components", "java-spring-boot", "openLiberty"})
		})
	})

	Context("When executing catalog list components with -a flag", func() {
		It("should list all supported and unsupported devfile components", func() {
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-a")
			helper.MatchAllInOutput(output, []string{"Odo Devfile Components", "java-spring-boot", "java-maven", "php-mysql"})
		})
	})
})
