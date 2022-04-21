package devfile

import (
	. "github.com/onsi/ginkgo"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile registry command tests", func() {
	const registryName string = "RegistryName"
	// Use staging OCI-based registry for tests to avoid overload
	const addRegistryURL string = "https://registry.stage.devfile.io"

	const updateRegistryURL string = "http://www.example.com/update"
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("Should list all default registries", func() {
		output := helper.Cmd("odo", "preference", "registry", "list").ShouldPass().Out()
		helper.MatchAllInOutput(output, []string{"DefaultDevfileRegistry"})
	})

	It("Should list at least one nodejs component from the default registry", func() {
		output := helper.Cmd("odo", "registry").ShouldPass().Out()
		helper.MatchAllInOutput(output, []string{"nodejs"})
	})

	It("Should list detailed information regarding nodejs", func() {
		output := helper.Cmd("odo", "registry", "--details", "--devfile", "nodejs", "--devfile-registry", "DefaultDevfileRegistry").ShouldPass().Out()
		helper.MatchAllInOutput(output, []string{"nodejs-starter", "javascript", "Node.js Runtime"})
	})

	It("Should list python specifically", func() {
		output := helper.Cmd("odo", "registry", "--devfile", "python", "--devfile-registry", "DefaultDevfileRegistry").ShouldPass().Out()
		helper.MatchAllInOutput(output, []string{"python"})
	})

	It("Should fail with an error with no registries", func() {
		helper.Cmd("odo", "preference", "registry", "delete", "DefaultDevfileRegistry", "-f").ShouldPass()
		output := helper.Cmd("odo", "preference", "registry", "list").ShouldFail().Err()
		helper.MatchAllInOutput(output, []string{"No devfile registries added to the configuration. Refer `odo preference registry add -h` to add one"})
	})

	It("Should fail to update the registry, when registry is not present", func() {
		helper.Cmd("odo", "preference", "registry", "update", registryName, updateRegistryURL, "-f").ShouldFail()
	})

	It("Should fail to delete the registry, when registry is not present", func() {
		helper.Cmd("odo", "preference", "registry", "delete", registryName, "-f").ShouldFail()
	})

	When("adding a registry", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "preference", "registry", "add", registryName, addRegistryURL).ShouldPass()
		})

		It("should list newly added registry", func() {
			output := helper.Cmd("odo", "preference", "registry", "list").ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{registryName, addRegistryURL})
		})

		It("should pass, when doing odo init with --devfile-registry flag", func() {
			helper.Cmd("odo", "init", "--name", "aname", "--devfile", "nodejs", "--devfile-registry", registryName).ShouldPass()
		})

		It("should fail, when adding same registry", func() {
			helper.Cmd("odo", "preference", "registry", "add", registryName, addRegistryURL).ShouldFail()
		})

		It("should successfully delete registry", func() {
			helper.Cmd("odo", "preference", "registry", "delete", registryName, "-f").ShouldPass()
		})

		It("should successfully update the registry", func() {
			helper.Cmd("odo", "preference", "registry", "update", registryName, updateRegistryURL, "-f").ShouldPass()
			output := helper.Cmd("odo", "preference", "registry", "list").ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{registryName, updateRegistryURL})
		})

		It("deleting registry and creating component with registry flag ", func() {
			helper.Cmd("odo", "preference", "registry", "delete", registryName, "-f").ShouldPass()
			helper.Cmd("odo", "init", "--name", "aname", "--devfile", "java-maven", "--devfile-registry", registryName).ShouldFail()
		})
	})

	It("should fail when adding a git based registry", func() {
		err := helper.Cmd("odo", "preference", "registry", "add", "RegistryFromGitHub", "https://github.com/devfile/registry").ShouldFail().Err()
		helper.MatchAllInOutput(err, []string{"github", "no", "supported", "https://github.com/devfile/registry-support"})
	})
})
