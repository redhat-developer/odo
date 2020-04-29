package devfile

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile registry command tests", func() {
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
			project = helper.CreateRandNamespace(context)
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
		} else {
			helper.DeleteProject(project)
		}
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
	})

	Context("When executing registry list", func() {
		It("Should list all default registries", func() {
			output := helper.CmdShouldPass("odo", "registry", "list")
			helper.MatchAllInOutput(output, []string{"CheDevfileRegistry", "DefaultDevfileRegistry"})
		})
	})

	Context("When executing registry commands with the registry is not present", func() {
		It("Should successfully add the registry", func() {
			helper.CmdShouldPass("odo", "registry", "add", "TestRegistryName", "TestRegistryURL")
			output := helper.CmdShouldPass("odo", "registry", "list")
			helper.MatchAllInOutput(output, []string{"TestRegistryName", "TestRegistryURL"})
			helper.CmdShouldPass("odo", "registry", "delete", "TestRegistryName", "-f")
		})

		It("Should fail to update the registry", func() {
			helper.CmdShouldFail("odo", "registry", "update", "TestRegistryName", "TestRegistryURL", "-f")
		})

		It("Should fail to delete the registry", func() {
			helper.CmdShouldFail("odo", "registry", "delete", "TestRegistryName", "-f")
		})
	})

	Context("When executing registry commands with the registry is present", func() {
		It("Should fail to add the registry", func() {
			helper.CmdShouldPass("odo", "registry", "add", "TestRegistryName", "TestRegistryURL")
			helper.CmdShouldFail("odo", "registry", "add", "TestRegistryName", "NewTestRegistryURL")
			helper.CmdShouldPass("odo", "registry", "delete", "TestRegistryName", "-f")
		})

		It("Should successfully update the registry", func() {
			helper.CmdShouldPass("odo", "registry", "add", "TestRegistryName", "TestRegistryURL")
			helper.CmdShouldPass("odo", "registry", "update", "TestRegistryName", "NewTestRegistryURL", "-f")
			output := helper.CmdShouldPass("odo", "registry", "list")
			helper.MatchAllInOutput(output, []string{"TestRegistryName", "NewTestRegistryURL"})
			helper.CmdShouldPass("odo", "registry", "delete", "TestRegistryName", "-f")
		})

		It("Should successfully delete the registry", func() {
			helper.CmdShouldPass("odo", "registry", "add", "TestRegistryName", "TestRegistryURL")
			helper.CmdShouldPass("odo", "registry", "delete", "TestRegistryName", "-f")
		})
	})
})
