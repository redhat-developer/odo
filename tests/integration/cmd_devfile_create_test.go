package integration

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile create command tests", func() {
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

	Context("When executing odo create with devfile component type argument", func() {
		It("should successfully create the devfile component", func() {
			helper.CmdShouldPass("odo", "create", "openLiberty")
		})
	})

	Context("When executing odo create with devfile component type and component name arguments", func() {
		It("should successfully create the devfile component", func() {
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "openLiberty", componentName)
		})
	})

	Context("When executing odo create with devfile component type argument and --project flag", func() {
		It("should successfully create the devfile component", func() {
			componentNamespace := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "openLiberty", "--project", componentNamespace)
		})
	})

	Context("When executing odo create with devfile component name that contains unsupported character", func() {
		It("", func() {
			componentName := "BAD@123"
			helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
		})
	})

	Context("When executing odo create with devfile component name that contains all numeric values", func() {
		It("", func() {
			componentName := "123456"
			helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
		})
	})

	Context("When executing odo create with devfile component name that contains more than 63 characters ", func() {
		It("", func() {
			componentName := helper.RandString(64)
			helper.CmdShouldFail("odo", "create", "openLiberty", componentName)
		})
	})
})
