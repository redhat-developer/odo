package devfile

import (
	"fmt"

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

	It("Should list all default registries with json", func() {
		output := helper.Cmd("odo", "preference", "registry", "list", "-o", "json").ShouldPass().Out()
		helper.MatchAllInOutput(output, []string{"DefaultDevfileRegistry"})
	})

	It("Should fail with an error with no registries", func() {
		helper.Cmd("odo", "preference", "registry", "delete", "DefaultDevfileRegistry", "-f").ShouldPass()
		output := helper.Cmd("odo", "preference", "registry", "list").ShouldFail().Err()
		helper.MatchAllInOutput(output, []string{"No devfile registries added to the configuration. Refer `odo registry add -h` to add one"})
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

		It("should pass, when doing odo create with --registry flag", func() {
			helper.Cmd("odo", "create", "nodejs", "--registry", registryName).ShouldPass()
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
			helper.Cmd("odo", "create", "java-maven", "--registry", registryName).ShouldFail()
		})
	})

	When("using a git based registries", func() {
		var deprecated, docLink, out, err, co string
		BeforeEach(func() {
			deprecated = "Deprecated"
			docLink = "https://github.com/redhat-developer/odo/tree/main/docs/public/git-registry-deprecation.adoc"
		})

		It("should not show deprication warning, if git registry is not used", func() {
			out, err = helper.Cmd("odo", "preference", "registry", "list").ShouldPass().OutAndErr()
			helper.DontMatchAllInOutput(fmt.Sprintln(out, err), []string{deprecated, docLink})
		})

		When("adding git based registries", func() {
			BeforeEach(func() {
				out, err = helper.Cmd("odo", "preference", "registry", "add", "RegistryFromGitHub", "https://github.com/odo-devfiles/registry").ShouldPass().OutAndErr()

			})
			It("should show deprication warning", func() {
				co = fmt.Sprintln(out, err)
				helper.MatchAllInOutput(co, []string{deprecated, docLink})

				By("odo resgistry list is executed, should show the warning", func() {
					out, err = helper.Cmd("odo", "preference", "registry", "list").ShouldPass().OutAndErr()
					co = fmt.Sprintln(out, err)
					helper.MatchAllInOutput(co, []string{deprecated, docLink})
				})
				By("should successfully delete registry", func() {
					out, err = helper.Cmd("odo", "create", "nodejs", "--registry", "RegistryFromGitHub").ShouldPass().OutAndErr()
					co = fmt.Sprintln(out, err)
					helper.MatchAllInOutput(co, []string{deprecated, docLink})
				})
			})
			It("should not show deprication warning if git registry is not used for component creation", func() {
				out, err = helper.Cmd("odo", "create", "nodejs", "--registry", "DefaultDevfileRegistry").ShouldPass().OutAndErr()
				helper.DontMatchAllInOutput(fmt.Sprintln(out, err), []string{deprecated, docLink})
			})
		})
	})
})
