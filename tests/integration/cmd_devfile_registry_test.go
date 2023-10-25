package integration

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile registry command tests", func() {

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		helper.CreateInvalidDevfile(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	for _, label := range []string{
		helper.LabelNoCluster, helper.LabelUnauth,
	} {
		label := label
		var _ = Context("label "+label, Label(label), func() {

			const registryName string = "RegistryName"

			// Use staging OCI-based registry for tests to avoid overload
			addRegistryURL := helper.GetDevfileRegistryURL()

			It("Should list all default registries", func() {
				output := helper.Cmd("odo", "preference", "view").ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{"DefaultDevfileRegistry"})
			})

			It("Should list at least one nodejs component from the default registry", func() {
				output := helper.Cmd("odo", "registry").ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{"nodejs"})
			})

			It("Should list detailed information regarding nodejs", func() {
				args := []string{"registry", "--details", "--devfile", "nodejs", "--devfile-registry", "DefaultDevfileRegistry"}

				By("using human readable output", func() {
					output := helper.Cmd("odo", args...).ShouldPass().Out()
					By("checking headers", func() {
						for _, h := range []string{
							"Name",
							"Display Name",
							"Registry",
							"Registry URL",
							"Version",
							"Description",
							"Tags",
							"Project Type",
							"Language",
							"Starter Projects",
							"Supported odo Features",
							"Versions",
						} {
							Expect(output).Should(ContainSubstring(h + ":"))
						}
					})
					By("not displaying Architectures", func() {
						Expect(output).ShouldNot(ContainSubstring("Architectures:"))
					})
					By("checking values", func() {
						helper.MatchAllInOutput(output, []string{"nodejs-starter", "JavaScript", "Node.js Runtime", "Dev: Y"})
					})
				})
				By("using JSON output", func() {
					args = append(args, "-o", "json")
					res := helper.Cmd("odo", args...).ShouldPass()
					stdout, stderr := res.Out(), res.Err()
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "0.name", "nodejs")
					helper.JsonPathContentContain(stdout, "0.displayName", "Node")
					helper.JsonPathContentContain(stdout, "0.description", "Node")
					helper.JsonPathContentContain(stdout, "0.language", "JavaScript")
					helper.JsonPathContentContain(stdout, "0.projectType", "Node.js")
					helper.JsonPathDoesNotExist(stdout, "0.architectures")
					helper.JsonPathContentContain(stdout, "0.devfileData.devfile.metadata.name", "nodejs")
					helper.JsonPathContentContain(stdout, "0.devfileData.supportedOdoFeatures.dev", "true")
					helper.JsonPathContentContain(stdout, "0.versions.0.commandGroups.run", "true")

					defaultVersion := gjson.Get(stdout, "0.version").String()
					By("returning backward-compatible information linked to the default stack version", func() {
						helper.JsonPathContentContain(stdout, "0.starterProjects.0", "nodejs-starter")
						Expect(defaultVersion).ShouldNot(BeEmpty())
					})
					By("returning a non-empty list of versions", func() {
						versions := gjson.Get(stdout, "0.versions").Array()
						Expect(versions).ShouldNot(BeEmpty())
					})
					By("listing the default version as such in the versions list", func() {
						Expect(gjson.Get(stdout, "0.versions.#(isDefault==true).version").String()).Should(Equal(defaultVersion))
					})
				})
			})

			It("Should list java-openliberty specifically", func() {
				args := []string{"registry", "--devfile", "java-openliberty", "--devfile-registry", "DefaultDevfileRegistry"}
				By("using human readable output", func() {
					output := helper.Cmd("odo", args...).ShouldPass().Out()
					By("checking table header", func() {
						helper.MatchAllInOutput(output, []string{"NAME", "REGISTRY", "DESCRIPTION", "ARCHITECTURES", "VERSIONS"})
					})
					By("checking table row", func() {
						helper.MatchAllInOutput(output, []string{"java-openliberty"})
					})
				})
				By("using JSON output", func() {
					args = append(args, "-o", "json")
					res := helper.Cmd("odo", args...).ShouldPass()
					stdout, stderr := res.Out(), res.Err()
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "0.name", "java-openliberty")
					helper.JsonPathContentContain(stdout, "0.displayName", "Open Liberty Maven")
					helper.JsonPathContentContain(stdout, "0.description", "using the Open Liberty runtime")
					helper.JsonPathContentContain(stdout, "0.language", "Java")
					helper.JsonPathContentContain(stdout, "0.projectType", "Open Liberty")
					helper.JsonPathContentContain(stdout, "0.devfileData", "")

					By("checking architectures", func() {
						architectures := gjson.Get(stdout, "0.architectures").Array()
						Expect(architectures).ShouldNot(BeEmpty())
					})

					defaultVersion := gjson.Get(stdout, "0.version").String()
					By("returning backward-compatible information linked to the default stack version", func() {
						helper.JsonPathContentContain(stdout, "0.starterProjects.0", "rest")
						Expect(defaultVersion).ShouldNot(BeEmpty())
					})
					By("returning a non-empty list of versions", func() {
						versions := gjson.Get(stdout, "0.versions").Array()
						Expect(versions).ShouldNot(BeEmpty())
					})
					By("listing the default version as such in the versions list", func() {
						Expect(gjson.Get(stdout, "0.versions.#(isDefault==true).version").String()).Should(Equal(defaultVersion))
					})
				})
			})

			It("Should fail with an error with no registries", func() {
				helper.Cmd("odo", "preference", "remove", "registry", "DefaultDevfileRegistry", "-f").ShouldPass()
				output := helper.Cmd("odo", "preference", "view").ShouldRun().Err()
				helper.MatchAllInOutput(output, []string{"No devfile registries added to the configuration. Refer to `odo preference add registry -h` to add one"})
			})

			It("Should fail to delete the registry, when registry is not present", func() {
				helper.Cmd("odo", "preference", "remove", "registry", registryName, "-f").ShouldFail()
			})

			When("adding a registry", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "preference", "add", "registry", registryName, addRegistryURL).ShouldPass()
				})

				It("should list newly added registry", func() {
					output := helper.Cmd("odo", "preference", "view").ShouldPass().Out()
					helper.MatchAllInOutput(output, []string{registryName, addRegistryURL})
				})

				It("should pass, when doing odo init with --devfile-registry flag", func() {
					helper.DeleteInvalidDevfile(commonVar.Context)
					helper.Cmd("odo", "init", "--name", "aname", "--devfile", "nodejs", "--devfile-registry", registryName).ShouldPass()
				})

				It("should fail, when adding same registry", func() {
					helper.Cmd("odo", "preference", "add", "registry", registryName, addRegistryURL).ShouldFail()
				})

				It("should successfully delete registry", func() {
					helper.Cmd("odo", "preference", "remove", "registry", registryName, "-f").ShouldPass()
				})

				It("deleting registry and creating component with registry flag ", func() {
					helper.DeleteInvalidDevfile(commonVar.Context)
					helper.Cmd("odo", "preference", "remove", "registry", registryName, "-f").ShouldPass()
					helper.Cmd("odo", "init", "--name", "aname", "--devfile", "java-maven", "--devfile-registry", registryName).ShouldFail()
				})

				It("should list registry with recently added registry on top", func() {
					By("for json output", func() {
						output := helper.Cmd("odo", "preference", "view", "-o", "json").ShouldPass().Out()
						Expect(helper.IsJSON(output)).To(BeTrue())
						helper.JsonPathContentIs(output, "registries.0.name", registryName)
						helper.JsonPathContentIs(output, "registries.0.url", addRegistryURL)
						helper.JsonPathContentIs(output, "registries.1.name", "DefaultDevfileRegistry")
						helper.JsonPathContentIs(output, "registries.1.url", addRegistryURL) // as we are using its updated in case of Proxy
					})

				})
			})

			It("should fail when adding a git based registry", func() {
				err := helper.Cmd("odo", "preference", "add", "registry", "RegistryFromGitHub", "https://github.com/devfile/registry").ShouldFail().Err()
				helper.MatchAllInOutput(err, []string{"github", "no", "supported", "https://github.com/devfile/registry-support"})
			})
		})
	}

	When("DevfileRegistriesList CRD is installed on cluster", func() {
		BeforeEach(func() {
			if !helper.IsKubernetesCluster() {
				Skip("skipped on non Kubernetes clusters")
			}
			// install CRDs for devfile registry
			devfileRegistriesLists := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "devfileregistrieslists.yaml"))
			Expect(devfileRegistriesLists.ExitCode()).To(BeEquivalentTo(0))
		})

		When("CR for devfileregistrieslists is installed in namespace", func() {
			BeforeEach(func() {
				manifestFilePath := filepath.Join(commonVar.ConfigDir, "devfileRegistryListCR.yaml")
				// NOTE: Use reachable URLs as we might be on a cluster with the registry operator installed, which would perform validations.
				err := helper.CreateFileWithContent(manifestFilePath, fmt.Sprintf(`
apiVersion: registry.devfile.io/v1alpha1
kind: DevfileRegistriesList
metadata:
  name: namespace-list
spec:
  devfileRegistries:
    - name: ns-devfile-reg
      url: %q
`, helper.GetDevfileRegistryURL()))
				Expect(err).ToNot(HaveOccurred())
				command := commonVar.CliRunner.Run("-n", commonVar.Project, "apply", "-f", manifestFilePath)
				Expect(command.ExitCode()).To(BeEquivalentTo(0))
			})

			It("should list detailed information regarding nodejs when using an in-cluster registry", func() {
				args := []string{"registry", "--details", "--devfile", "nodejs", "--devfile-registry", "ns-devfile-reg"}

				By("using human readable output", func() {
					output := helper.Cmd("odo", args...).ShouldPass().Out()
					By("checking headers", func() {
						for _, h := range []string{
							"Name",
							"Display Name",
							"Registry",
							"Registry URL",
							"Version",
							"Description",
							"Tags",
							"Project Type",
							"Language",
							"Starter Projects",
							"Supported odo Features",
							"Versions",
						} {
							Expect(output).Should(ContainSubstring(h + ":"))
						}
					})
					By("checking values", func() {
						helper.MatchAllInOutput(output, []string{"nodejs-starter", "JavaScript", "Node.js Runtime", "Dev: Y"})
					})
				})
				By("using JSON output", func() {
					args = append(args, "-o", "json")
					res := helper.Cmd("odo", args...).ShouldPass()
					stdout, stderr := res.Out(), res.Err()
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "0.name", "nodejs")
					helper.JsonPathContentContain(stdout, "0.displayName", "Node")
					helper.JsonPathContentContain(stdout, "0.description", "Node")
					helper.JsonPathContentContain(stdout, "0.language", "JavaScript")
					helper.JsonPathContentContain(stdout, "0.projectType", "Node.js")
					helper.JsonPathContentContain(stdout, "0.devfileData.devfile.metadata.name", "nodejs")
					helper.JsonPathContentContain(stdout, "0.devfileData.supportedOdoFeatures.dev", "true")
					helper.JsonPathContentContain(stdout, "0.versions.0.commandGroups.run", "true")

					defaultVersion := gjson.Get(stdout, "0.version").String()
					By("returning backward-compatible information linked to the default stack version", func() {
						helper.JsonPathContentContain(stdout, "0.starterProjects.0", "nodejs-starter")
						Expect(defaultVersion).ShouldNot(BeEmpty())
					})
					By("returning a non-empty list of versions", func() {
						versions := gjson.Get(stdout, "0.versions").Array()
						Expect(versions).ShouldNot(BeEmpty())
					})
					By("listing the default version as such in the versions list", func() {
						Expect(gjson.Get(stdout, "0.versions.#(isDefault==true).version").String()).Should(Equal(defaultVersion))
					})
				})
			})

		})
	})
})
