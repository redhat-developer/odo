package integration

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile catalog command tests", func() {
	var project string
	var context string
	var originalDir string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
		originalDir = helper.Getwd()
		helper.Chdir(context)
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.Chdir(originalDir)
		helper.DeleteDir(context)
	})

	Context("When executing catalog list components", func() {
		It("should list all supported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components")
			Expect(output).To(ContainSubstring("Odo Devfile Components"))
			Expect(output).To(ContainSubstring("java-spring-boot"))
			Expect(output).To(ContainSubstring("openLiberty"))
		})
	})

	Context("When executing catalog list components with -a flag", func() {
		It("should list all supported and unsupported devfile components", func() {
			output := helper.CmdShouldPass("odo", "catalog", "list", "components", "-a")
			Expect(output).To(ContainSubstring("Odo Devfile Components"))
			Expect(output).To(ContainSubstring("java-spring-boot"))
			Expect(output).To(ContainSubstring("java-maven"))
			Expect(output).To(ContainSubstring("php-mysql"))
		})
	})
})
