package integration

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
	"github.com/tidwall/gjson"
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
	})

	Context("When configuring global config values", func() {
		preferences := []struct {
			name, value, updateValue, invalidValue string
		}{
			{"UpdateNotification", "false", "true", "foo"},
			{"Timeout", "5", "6", "foo"},
			// !! Do not test ConsentTelemetry with true because it sends out the telemetry data and messes up the statistics !!
			{"ConsentTelemetry", "false", "false", "foo"},
			{"PushTimeout", "4", "6", "foo"},
			{"RegistryCacheTime", "4", "6", "foo"},
			{"Ephemeral", "false", "true", "foo"},
		}

		It("should successfully updated", func() {
			for _, pref := range preferences {
				helper.Cmd("odo", "preference", "set", pref.name, pref.value).ShouldPass()
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

		It("should show json output", func() {
			prefJSONOutput, err := helper.Unindented(helper.Cmd("odo", "preference", "view", "-o", "json").ShouldPass().Out())
			Expect(err).Should(BeNil())
			values := gjson.GetMany(prefJSONOutput, "kind", "items.0.Description")
			expected := []string{"PreferenceList", "Flag to control if an update notification is shown or not (Default: true)"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})
	})

	Context("When no ConsentTelemetry preference value is set", func() {
		var _ = JustBeforeEach(func() {
			// unset the preference in case it is already set
			helper.Cmd("odo", "preference", "unset", "ConsentTelemetry", "-f").ShouldPass()
		})

		It("should not prompt when user calls for help", func() {
			output := helper.Cmd("odo", "create", "--help").ShouldPass().Out()
			Expect(output).ToNot(ContainSubstring(promptMessageSubString))
		})

		It("should not prompt when preference command is run", func() {
			output := helper.Cmd("odo", "preference", "view").ShouldPass().Out()
			Expect(output).ToNot(ContainSubstring(promptMessageSubString))

			output = helper.Cmd("odo", "preference", "set", "timeout", "5", "-f").ShouldPass().Out()
			Expect(output).ToNot(ContainSubstring(promptMessageSubString))

			output = helper.Cmd("odo", "preference", "unset", "timeout", "-f").ShouldPass().Out()
			Expect(output).ToNot(ContainSubstring(promptMessageSubString))
		})
	})

	Context("When ConsentTelemetry preference value is set", func() {
		// !! Do not test with true because it sends out the telemetry data and messes up the statistics !!
		It("should not prompt the user", func() {
			helper.Cmd("odo", "preference", "set", "ConsentTelemetry", "false", "-f").ShouldPass()
			output := helper.Cmd("odo", "create", "--context", commonVar.Context, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass().Out()
			Expect(output).ToNot(ContainSubstring(promptMessageSubString))
		})
	})

})
