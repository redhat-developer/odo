package integration

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

const promtMessageSubString = "Help odo improve by allowing it to collect usage data."

var _ = Describe("odo preference and config command tests", func() {
	// TODO: A neater way to provide odo path. Currently we assume \
	// odo and oc in $PATH already.
	var oc helper.OcRunner
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
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
				appHelp := helper.CmdShouldPass("odo", helpArg)
				Expect(appHelp).To(ContainSubstring(`Use "odo [command] --help" for more information about a command.`))
			}
		})
	})

	Context("when running help for preference command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "preference", "-h")
			Expect(appHelp).To(ContainSubstring("Modifies odo specific configuration settings"))
		})
	})

	Context("when running help for config command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "config", "-h")
			Expect(appHelp).To(ContainSubstring("Modifies odo specific configuration settings within the devfile or config file"))
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
			configOutput := helper.CmdShouldPass("odo", "preference", "view")
			preferences := []string{"UpdateNotification", "NamePrefix", "Timeout", "PushTarget", "BuildTimeout", "PushTimeout", "Experimental", "Ephemeral", "ConsentTelemetry"}
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
			{"PushTarget", "docker", "kube", "smh"},
			{"NamePrefix", "foo", "bar", ""},
			{"BuildTimeout", "5", "7", "foo"},
			{"Experimental", "false", "true", "foo"},
			// !! Do not test ConsentTelemetry with true because it sends out the telemetry data and messes up the statistics !!
			{"ConsentTelemetry", "false", "false", "foo"},
			{"PushTimeout", "4", "6", "f00"},
		}

		It("should successfully updated", func() {
			for _, pref := range preferences {
				helper.CmdShouldPass("odo", "preference", "set", pref.name, pref.value)
				value := helper.GetPreferenceValue(pref.name)
				Expect(value).To(ContainSubstring(pref.value))

				helper.CmdShouldPass("odo", "preference", "set", "-f", pref.name, pref.updateValue)
				value = helper.GetPreferenceValue(pref.name)
				Expect(value).To(ContainSubstring(pref.updateValue))

				helper.CmdShouldPass("odo", "preference", "unset", "-f", pref.name)
				value = helper.GetPreferenceValue(pref.name)
				Expect(value).To(BeEmpty())
			}
			globalConfPath := os.Getenv("HOME")
			os.RemoveAll(filepath.Join(globalConfPath, ".odo"))
		})

		It("should unsuccessfully update", func() {
			for _, pref := range preferences {
				// TODO: Remove this once we decide something on checking NamePrefix
				if pref.name != "NamePrefix" {
					helper.CmdShouldFail("odo", "preference", "set", "-f", pref.name, pref.invalidValue)
				}
			}
		})

		It("should show json output", func() {
			prefJSONOutput, err := helper.Unindented(helper.CmdShouldPass("odo", "preference", "view", "-o", "json"))
			Expect(err).Should(BeNil())
			values := gjson.GetMany(prefJSONOutput, "kind", "items.0.Description")
			expected := []string{"PreferenceList", "Flag to control if an update notification is shown or not (Default: true)"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})
	})

	Context("when creating odo local config in the same config dir", func() {
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})
		It("should set, unset local config successfully", func() {
			cases := []struct {
				paramName  string
				paramValue string
			}{
				{
					paramName:  "Type",
					paramValue: "java",
				},
				{
					paramName:  "Name",
					paramValue: "odo-java",
				},
				{
					paramName:  "MinCPU",
					paramValue: "0.2",
				},
				{
					paramName:  "MaxCPU",
					paramValue: "2",
				},
				{
					paramName:  "MinMemory",
					paramValue: "100M",
				},
				{
					paramName:  "MaxMemory",
					paramValue: "500M",
				},
				{
					paramName:  "Ports",
					paramValue: "8080/TCP,45/UDP",
				},
				{
					paramName:  "Application",
					paramValue: "odotestapp",
				},
				{
					paramName:  "Project",
					paramValue: "odotestproject",
				},
				{
					paramName:  "SourceType",
					paramValue: "git",
				},
				{
					paramName:  "Ref",
					paramValue: "develop",
				},
				{
					paramName:  "SourceLocation",
					paramValue: "https://github.com/sclorg/nodejs-ex",
				},
			}
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", commonVar.Project, "--git", "https://github.com/sclorg/nodejs-ex")
			for _, testCase := range cases {
				helper.CmdShouldPass("odo", "config", "set", testCase.paramName, testCase.paramValue, "-f")
				setValue := helper.GetConfigValue(testCase.paramName)
				Expect(setValue).To(ContainSubstring(testCase.paramValue))
				// cleanup
				helper.CmdShouldPass("odo", "config", "unset", testCase.paramName, "-f")
				UnsetValue := helper.GetConfigValue(testCase.paramName)
				Expect(UnsetValue).To(BeEmpty())
			}
		})
	})

	Context("when creating odo local config with context flag", func() {
		It("should allow setting and unsetting a config locally with context", func() {
			cases := []struct {
				paramName  string
				paramValue string
			}{
				{
					paramName:  "Type",
					paramValue: "java",
				},
				{
					paramName:  "Name",
					paramValue: "odo-java",
				},
				{
					paramName:  "MinCPU",
					paramValue: "0.2",
				},
				{
					paramName:  "MaxCPU",
					paramValue: "2",
				},
				{
					paramName:  "MinMemory",
					paramValue: "100M",
				},
				{
					paramName:  "MaxMemory",
					paramValue: "500M",
				},
				{
					paramName:  "Ports",
					paramValue: "8080/TCP,45/UDP",
				},
				{
					paramName:  "Application",
					paramValue: "odotestapp",
				},
				{
					paramName:  "Project",
					paramValue: "odotestproject",
				},
				{
					paramName:  "SourceType",
					paramValue: "git",
				},
				{
					paramName:  "Ref",
					paramValue: "develop",
				},
				{
					paramName:  "SourceLocation",
					paramValue: "https://github.com/sclorg/nodejs-ex",
				},
			}
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context, "--git", "https://github.com/odo-devfiles/nodejs-ex.git")
			for _, testCase := range cases {

				helper.CmdShouldPass("odo", "config", "set", "-f", "--context", commonVar.Context, testCase.paramName, testCase.paramValue)
				configOutput := helper.CmdShouldPass("odo", "config", "unset", "-f", "--context", commonVar.Context, testCase.paramName)
				Expect(configOutput).To(ContainSubstring("Local config was successfully updated."))
				Value := helper.GetConfigValueWithContext(testCase.paramName, commonVar.Context)
				Expect(Value).To(BeEmpty())
			}
		})
	})

	Context("when creating odo local config with env variables", func() {
		It("should set and unset env variables", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "--git", "https://github.com/openshift/nodejs-ex", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "PORT=4000", "--env", "PORT=1234", "--context", commonVar.Context)
			configPort := helper.GetConfigValueWithContext("PORT", commonVar.Context)
			Expect(configPort).To(ContainSubstring("1234"))
			helper.CmdShouldPass("odo", "config", "set", "--env", "SECRET_KEY=R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q=", "--context", commonVar.Context)
			configSecret := helper.GetConfigValueWithContext("SECRET_KEY", commonVar.Context)
			Expect(configSecret).To(ContainSubstring("R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q"))
			helper.CmdShouldPass("odo", "config", "unset", "--env", "PORT", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "config", "unset", "--env", "SECRET_KEY", "--context", commonVar.Context)
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", commonVar.Context)
			helper.DontMatchAllInOutput(configValue, []string{"PORT", "SECRET_KEY"})
		})
		It("should check for existence of environment variable in config before unsetting it", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--git", "https://github.com/openshift/nodejs-ex", "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "PORT=4000", "--env", "PORT=1234", "--context", commonVar.Context)

			// unset a valid env var
			helper.CmdShouldPass("odo", "config", "unset", "--env", "PORT", "--context", commonVar.Context)

			// try to unset an env var that doesn't exist
			stdOut := helper.CmdShouldFail("odo", "config", "unset", "--env", "nosuchenv", "--context", commonVar.Context)
			Expect(stdOut).To(ContainSubstring("unable to find environment variable nosuchenv in the component"))
		})
	})

	Context("when viewing local config without logging into the OpenShift cluster", func() {
		It("should list config successfully", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--git", "https://github.com/openshift/nodejs-ex", "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "hello=world", "--context", commonVar.Context)
			kubeconfigOld := os.Getenv("KUBECONFIG")
			os.Setenv("KUBECONFIG", "/no/such/path")
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", commonVar.Context)
			helper.MatchAllInOutput(configValue, []string{"hello", "world"})
			os.Setenv("KUBECONFIG", kubeconfigOld)
		})

		It("should set config variable without logging in", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context)
			kubeconfigOld := os.Getenv("KUBECONFIG")
			os.Setenv("KUBECONFIG", "/no/such/path")
			helper.CmdShouldPass("odo", "config", "set", "--force", "--context", commonVar.Context, "Name", "foobar")
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", commonVar.Context)
			Expect(configValue).To(ContainSubstring("foobar"))
			helper.CmdShouldPass("odo", "config", "unset", "--force", "--context", commonVar.Context, "Name")
			os.Setenv("KUBECONFIG", kubeconfigOld)
		})
	})

	// issue https://github.com/openshift/odo/issues/4594
	// Context("when using --now with config command", func() {
	// 	It("should successfully set and unset variables", func() {
	// 		//set env var
	// 		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
	// 		helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context)
	// 		helper.CmdShouldPass("odo", "config", "set", "--now", "--env", "hello=world", "--context", commonVar.Context)
	// 		//*Check config
	// 		configValue1 := helper.CmdShouldPass("odo", "config", "view", "--context", commonVar.Context)
	// 		helper.MatchAllInOutput(configValue1, []string{"hello", "world"})
	// 		//*Check dc
	// 		envs := oc.GetEnvsDevFileDeployment("nodejs", commonVar.Project)
	// 		val, ok := envs["hello"]
	// 		Expect(ok).To(BeTrue())
	// 		Expect(val).To(ContainSubstring("world"))
	// 		// unset a valid env var
	// 		helper.CmdShouldPass("odo", "config", "unset", "--now", "--env", "hello", "--context", commonVar.Context)
	// 		configValue2 := helper.CmdShouldPass("odo", "config", "view", "--context", commonVar.Context)
	// 		helper.DontMatchAllInOutput(configValue2, []string{"hello", "world"})
	// 		envs = oc.GetEnvsDevFileDeployment("nodejs", commonVar.Project)
	// 		_, ok = envs["hello"]
	// 		Expect(ok).To(BeFalse())
	// 	})
	// })

	Context("When no ConsentTelemetry preference value is set", func() {
		var _ = JustBeforeEach(func() {
			// unset the preference in case it is already set
			helper.CmdShouldPass("odo", "preference", "unset", "ConsentTelemetry", "-f")
		})
		It("prompt should not appear when user calls for help", func() {
			output := helper.CmdShouldPass("odo", "create", "--help")
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))
		})

		It("prompt should not appear when preference command is run", func() {
			output := helper.CmdShouldPass("odo", "preference", "view")
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))

			output = helper.CmdShouldPass("odo", "preference", "set", "buildtimeout", "5", "-f")
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))

			output = helper.CmdShouldPass("odo", "preference", "unset", "buildtimeout", "-f")
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))
		})
	})

	Context("Prompt should not appear when", func() {

		// !! Do not test with true because it sends out the telemetry data and messes up the statistics !!
		It("ConsentTelemetry is set", func() {
			helper.CmdShouldPass("odo", "preference", "set", "ConsentTelemetry", "false", "-f")
			output := helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context)
			Expect(output).ToNot(ContainSubstring(promtMessageSubString))
		})
	})

})
