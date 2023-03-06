package integration

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"

	"github.com/redhat-developer/odo/tests/helper"
)

const promptMessageSubString = "Help odo improve by allowing it to collect usage data."

var _ = Describe("odo preference and config command tests", func() {

	// TODO: A neater way to provide odo path. Currently we assume odo and oc in $PATH already.
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	for _, label := range []string{
		helper.LabelNoCluster, helper.LabelUnauth,
	} {
		label := label
		var _ = Context("label "+label, Label(label), func() {

			Context("check that help works", func() {
				It("should display help info", func() {
					helpArgs := []string{"-h", "help", "--help"}
					for _, helpArg := range helpArgs {
						appHelp := helper.Cmd("odo", helpArg).ShouldPass().Out()
						Expect(appHelp).To(ContainSubstring(`Use "odo [command] --help" for more information about a command.`))
					}
				})
			})

			Context("when running help for preference command", func() {
				It("should display the help", func() {
					appHelp := helper.Cmd("odo", "preference", "-h").ShouldPass().Out()
					Expect(appHelp).To(ContainSubstring("Modifies odo specific configuration settings"))
				})
			})

			Context("When viewing global config", func() {
				var newContext string
				// ConsentTelemetry is set to false in helper.CommonBeforeEach so that it does not prompt to set a value
				// during the tests, but we want to check preference values as they would be in real time and hence
				// we set the GLOBALODOCONFIG variable to a value in new context
				var _ = JustBeforeEach(func() {
					newContext = helper.CreateNewContext()
					os.Setenv("GLOBALODOCONFIG", filepath.Join(newContext, "preference.yaml"))
				})
				var _ = JustAfterEach(func() {
					helper.DeleteDir(newContext)
				})
				It("should get the default global config keys", func() {
					configOutput := helper.Cmd("odo", "preference", "view").ShouldPass().Out()
					preferences := []string{"UpdateNotification", "Timeout", "PushTimeout", "RegistryCacheTime", "Ephemeral", "ConsentTelemetry"}
					helper.MatchAllInOutput(configOutput, preferences)
					for _, key := range preferences {
						value := helper.GetPreferenceValue(key)
						Expect(value).To(BeEmpty())
					}
				})
				It("should get the default global config keys in JSON output", func() {
					res := helper.Cmd("odo", "preference", "view", "-o", "json").ShouldPass()
					stdout, stderr := res.Out(), res.Err()
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					preferences := []string{"UpdateNotification", "Timeout", "PushTimeout", "RegistryCacheTime", "ConsentTelemetry", "Ephemeral"}
					for i, pref := range preferences {
						helper.JsonPathContentIs(stdout, fmt.Sprintf("preferences.%d.name", i), pref)
					}
					helper.JsonPathContentIs(stdout, "registries.#", "1")
					helper.JsonPathContentIs(stdout, "registries.0.name", "DefaultDevfileRegistry")
				})
			})

			Context("When configuring global config values", func() {
				preferences := []struct {
					name              string
					value             string
					updateValue       string
					invalidValue      string
					firstSetWithForce bool
				}{
					{"UpdateNotification", "false", "true", "foo", false},
					{"Timeout", "5s", "6s", "foo", false},
					// !! Do not test ConsentTelemetry with true because it sends out the telemetry data and messes up the statistics !!
					{"ConsentTelemetry", "false", "false", "foo", false},
					{"PushTimeout", "4s", "6s", "foo", false},
					{"RegistryCacheTime", "4m", "6m", "foo", false},
					{"Ephemeral", "false", "true", "foo", true},
				}

				It("should successfully updated", func() {
					for _, pref := range preferences {
						// construct arguments for the first command
						firstCmdArgs := []string{"preference", "set"}
						if pref.firstSetWithForce {
							firstCmdArgs = append(firstCmdArgs, "-f")
						}
						firstCmdArgs = append(firstCmdArgs, pref.name, pref.value)

						helper.Cmd("odo", firstCmdArgs...).ShouldPass()
						value := helper.GetPreferenceValue(pref.name)
						Expect(value).To(ContainSubstring(pref.value))

						helper.Cmd("odo", "preference", "set", "-f", pref.name, pref.updateValue).ShouldPass()
						value = helper.GetPreferenceValue(pref.name)
						Expect(value).To(ContainSubstring(pref.updateValue))

						helper.Cmd("odo", "preference", "unset", "-f", pref.name).ShouldPass()
						value = helper.GetPreferenceValue(pref.name)
						Expect(value).To(BeEmpty())
					}
					globalConfPath := os.Getenv("HOME")
					os.RemoveAll(filepath.Join(globalConfPath, ".odo"))
				})
			})
			When("when preference.yaml contains an int value for Timeout", func() {
				BeforeEach(func() {
					preference := `
kind: Preference
apiversion: odo.dev/v1alpha1
OdoSettings:
  UpdateNotification: true
  RegistryList:
  - Name: DefaultDevfileRegistry
    URL: https://registry.devfile.io
    secure: false
  ConsentTelemetry: true
  Timeout: 10
`
					preferencePath := filepath.Join(commonVar.Context, "preference.yaml")
					err := helper.CreateFileWithContent(preferencePath, preference)
					Expect(err).To(BeNil())
					os.Setenv("GLOBALODOCONFIG", preferencePath)
				})
				It("should show warning about incompatible Timeout value when viewing preferences", func() {
					errOut := helper.Cmd("odo", "preference", "view").ShouldPass().Err()
					Expect(helper.GetPreferenceValue("Timeout")).To(ContainSubstring("10ns"))
					Expect(errOut).To(ContainSubstring("Please change the preference value for Timeout"))
				})
			})

			It("should fail to set an incompatible format for a preference that accepts duration", func() {
				errOut := helper.Cmd("odo", "preference", "set", "RegistryCacheTime", "1d").ShouldFail().Err()
				Expect(errOut).To(ContainSubstring("unable to set \"registrycachetime\" to \"1d\""))
			})

			Context("When no ConsentTelemetry preference value is set", func() {
				var _ = JustBeforeEach(func() {
					// unset the preference in case it is already set
					helper.Cmd("odo", "preference", "unset", "ConsentTelemetry", "-f").ShouldPass()
				})

				It("should not prompt when user calls for help", func() {
					output := helper.Cmd("odo", "init", "--help").ShouldPass().Out()
					Expect(output).ToNot(ContainSubstring(promptMessageSubString))
				})

				It("should not prompt when preference command is run", func() {
					output := helper.Cmd("odo", "preference", "view").ShouldPass().Out()
					Expect(output).ToNot(ContainSubstring(promptMessageSubString))

					output = helper.Cmd("odo", "preference", "set", "timeout", "5s", "-f").ShouldPass().Out()
					Expect(output).ToNot(ContainSubstring(promptMessageSubString))

					output = helper.Cmd("odo", "preference", "unset", "timeout", "-f").ShouldPass().Out()
					Expect(output).ToNot(ContainSubstring(promptMessageSubString))
				})
			})

			Context("When ConsentTelemetry preference value is set", func() {
				// !! Do not test with true because it sends out the telemetry data and messes up the statistics !!
				var workingDir string
				BeforeEach(func() {
					workingDir = helper.Getwd()
					helper.Chdir(commonVar.Context)
				})
				AfterEach(func() {
					helper.Chdir(workingDir)
				})
				It("should not prompt the user", func() {
					helper.Cmd("odo", "preference", "set", "ConsentTelemetry", "false", "-f").ShouldPass()
					output := helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass().Out()
					Expect(output).ToNot(ContainSubstring(promptMessageSubString))
				})
			})
		})
	}
	When("DevfileRegistriesList CRD is installed on cluster", func() {
		BeforeEach(func() {
			// install CRDs for devfile registry
			//TODO: install clusterRegestryList from scripts
			// clusterRegistryList := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "clusterdevfileregistrieslists.yaml"))
			// Expect(clusterRegistryList.ExitCode()).To(BeEquivalentTo(0))
			devfileRegistriesLists := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "devfileregistrieslists.yaml"))
			Expect(devfileRegistriesLists.ExitCode()).To(BeEquivalentTo(0))
		})

		When("CR for devfileregistrieslists is installed in namespace", func() {
			var registryURL string

			BeforeEach(func() {
				manifestFilePath := filepath.Join(commonVar.ConfigDir, "devfileRegistryListCR.yaml")
				registryURL = helper.GetDevfileRegistryURL()
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
      skipTLSVerify: true
    - name: ns-devfile-prod
      url: 'https://registry.devfile.io'
`, registryURL))
				Expect(err).ToNot(HaveOccurred())
				command := commonVar.CliRunner.Run("-n", commonVar.Project, "apply", "-f", manifestFilePath)
				Expect(command.ExitCode()).To(BeEquivalentTo(0))
			})

			It("registry list should return registry listed in CR", func() {
				stdout, stderr := helper.Cmd("odo", "preference", "view", "-o", "json").ShouldPass().OutAndErr()
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())

				By("ensuring we have the minimum number of registries returned", func() {
					nbRegistries := gjson.Get(stdout, "registries.#").Int()
					// 3 is the minimum we want here (2 from the cluster namespace, and 1 from the local preferences).
					// But we might have more since we might be on a cluster potentially containing cluster-wide registries, with would be returned here as well.
					// So we'll check that the first registries are the one from the current namespace, and the last one comes from the preference file.
					Expect(nbRegistries).Should(BeNumerically(">=", 3))
				})

				By("ensuring the first registries returned are those in the namespace", func() {
					helper.JsonPathContentIs(stdout, "registries.0.name", "ns-devfile-reg")
					helper.JsonPathContentIs(stdout, "registries.0.url", registryURL)
					helper.JsonPathContentIs(stdout, "registries.0.secure", "false")

					helper.JsonPathContentIs(stdout, "registries.1.name", "ns-devfile-prod")
					helper.JsonPathContentIs(stdout, "registries.1.url", "https://registry.devfile.io")
					helper.JsonPathContentIs(stdout, "registries.1.secure", "true")
				})

				By("ensuring the last registries returned are those coming from the odo preferences", func() {
					helper.JsonPathContentIs(stdout, "registries.@reverse.0.name", "DefaultDevfileRegistry")
					helper.JsonPathContentIs(stdout, "registries.@reverse.0.url", registryURL)
					helper.JsonPathContentIs(stdout, "registries.@reverse.0.secure", "false")
				})
			})

			It("should fail to delete the in-cluster registry", func() {
				regName := "ns-devfile-prod"
				stderr := helper.Cmd("odo", "preference", "remove", "registry", regName).ShouldFail().Err()
				Expect(stderr).Should(ContainSubstring("failed to remove registry: registry %q doesn't exist or it is not managed by odo", regName))
			})
		})
	})
})
