package integration

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
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

	Context("when running help for config command", func() {
		It("should display the help", func() {
			appHelp := helper.Cmd("odo", "config", "-h").ShouldPass().Out()
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
			configOutput := helper.Cmd("odo", "preference", "view").ShouldPass().Out()
			preferences := []string{"UpdateNotification", "NamePrefix", "Timeout", "BuildTimeout", "PushTimeout", "Experimental", "Ephemeral", "ConsentTelemetry"}
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
			{"NamePrefix", "foo", "bar", ""},
			{"BuildTimeout", "5", "7", "foo"},
			{"Experimental", "false", "true", "foo"},
			// !! Do not test ConsentTelemetry with true because it sends out the telemetry data and messes up the statistics !!
			{"ConsentTelemetry", "false", "false", "foo"},
			{"PushTimeout", "4", "6", "f00"},
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

		It("should unsuccessfully update", func() {
			for _, pref := range preferences {
				// TODO: Remove this once we decide something on checking NamePrefix
				if pref.name != "NamePrefix" {
					helper.Cmd("odo", "preference", "set", "-f", pref.name, pref.invalidValue).ShouldFail()
				}
			}
		})

		It("should show json output", func() {
			prefJSONOutput, err := helper.Unindented(helper.Cmd("odo", "preference", "view", "-o", "json").ShouldPass().Out())
			Expect(err).Should(BeNil())
			values := gjson.GetMany(prefJSONOutput, "kind", "items.0.Description")
			expected := []string{"PreferenceList", "Flag to control if an update notification is shown or not (Default: true)"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})
	})

	Context("when setting or unsetting odo local config", func() {
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--project", commonVar.Project, "--git", "https://github.com/sclorg/nodejs-ex").ShouldPass()
		})
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

		It("should run successfully", func() {
			for _, testCase := range cases {
				helper.Cmd("odo", "config", "set", testCase.paramName, testCase.paramValue, "-f").ShouldPass()
				setValue := helper.GetConfigValue(testCase.paramName)
				Expect(setValue).To(ContainSubstring(testCase.paramValue))
				// cleanup
				helper.Cmd("odo", "config", "unset", testCase.paramName, "-f").ShouldPass()
				UnsetValue := helper.GetConfigValue(testCase.paramName)
				Expect(UnsetValue).To(BeEmpty())
			}
		})
		It("should run successfully with context", func() {
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			for _, testCase := range cases {
				helper.Cmd("odo", "config", "set", "-f", "--context", commonVar.Context, testCase.paramName, testCase.paramValue).ShouldPass()
				configOutput := helper.Cmd("odo", "config", "unset", "-f", "--context", commonVar.Context, testCase.paramName).ShouldPass().Out()
				Expect(configOutput).To(ContainSubstring("Local config was successfully updated."))
				Value := helper.GetConfigValueWithContext(testCase.paramName, commonVar.Context)
				Expect(Value).To(BeEmpty())
			}
		})
	})

	Context("when creating odo local config with env variables", func() {
		It("should set and unset env variables", func() {
			helper.Cmd("odo", "create", "--s2i", "--git", "https://github.com/openshift/nodejs-ex", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "config", "set", "--env", "PORT=4000", "--env", "PORT=1234", "--context", commonVar.Context).ShouldPass()
			configPort := helper.GetConfigValueWithContext("PORT", commonVar.Context)
			Expect(configPort).To(ContainSubstring("1234"))
			helper.Cmd("odo", "config", "set", "--env", "SECRET_KEY=R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q=", "--context", commonVar.Context).ShouldPass()
			configSecret := helper.GetConfigValueWithContext("SECRET_KEY", commonVar.Context)
			Expect(configSecret).To(ContainSubstring("R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q"))
			helper.Cmd("odo", "config", "unset", "--env", "PORT", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "config", "unset", "--env", "SECRET_KEY", "--context", commonVar.Context).ShouldPass()
			configValue := helper.Cmd("odo", "config", "view", "--context", commonVar.Context).ShouldPass().Out()
			helper.DontMatchAllInOutput(configValue, []string{"PORT", "SECRET_KEY"})
		})
		It("should check for existence of environment variable in config before unsetting it", func() {
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--git", "https://github.com/openshift/nodejs-ex", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "config", "set", "--env", "PORT=4000", "--env", "PORT=1234", "--context", commonVar.Context).ShouldPass()

			// unset a valid env var
			helper.Cmd("odo", "config", "unset", "--env", "PORT", "--context", commonVar.Context).ShouldPass()

			// try to unset an env var that doesn't exist
			stdOut := helper.Cmd("odo", "config", "unset", "--env", "nosuchenv", "--context", commonVar.Context).ShouldFail().Err()
			Expect(stdOut).To(ContainSubstring("unable to find environment variable nosuchenv in the component"))
		})
	})

	Context("when viewing local config without logging into the OpenShift cluster", func() {
		var ocRunner helper.OcRunner
		var token string
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			ocRunner = helper.NewOcRunner("oc")
			token = ocRunner.GetToken()
			ocRunner.Logout()
		})
		AfterEach(func() {
			ocRunner.LoginUsingToken(token)
		})
		When("user is working with a devfile component", func() {
			It("should set, list and delete config successfully", func() {
				if helper.IsKubernetesCluster() {
					Skip("skipping for kubernetes until we can figure out how to simulate logged out state there")
				}
				helper.Cmd("odo", "config", "set", "--force", "--context", commonVar.Context, "Name", "foobar").ShouldPass()
				configValue := helper.Cmd("odo", "config", "view", "--context", commonVar.Context).ShouldPass().Out()
				Expect(configValue).To(ContainSubstring("foobar"))
				helper.Cmd("odo", "config", "unset", "--force", "--context", commonVar.Context, "Name").ShouldPass()
			})
			It("should set, list and delete config envs successfully", func() {
				if helper.IsKubernetesCluster() {
					Skip("skipping for kubernetes until we can figure out how to simulate logged out state there")
				}
				helper.Cmd("odo", "config", "set", "--force", "--env", "hello=world", "--context", commonVar.Context).ShouldPass()
				configValue := helper.Cmd("odo", "config", "view", "--context", commonVar.Context).ShouldPass().Out()
				helper.MatchAllInOutput(configValue, []string{"hello", "world"})
			})
		})
	})

	Context("when using --now with config command", func() {
		var oc helper.OcRunner
		BeforeEach(func() {
			oc = helper.NewOcRunner("oc")
		})
		It("should successfully set and unset variables", func() {
			//set env var
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "config", "set", "--now", "--env", "hello=world", "--context", commonVar.Context).ShouldPass()
			//*Check config
			configValue1 := helper.Cmd("odo", "config", "view", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(configValue1, []string{"hello", "world"})
			//*Check dc
			envs := oc.GetEnvsDevFileDeployment("nodejs", "app", commonVar.Project)
			val, ok := envs["hello"]
			Expect(ok).To(BeTrue())
			Expect(val).To(ContainSubstring("world"))
			// unset a valid env var
			helper.Cmd("odo", "config", "unset", "--now", "--env", "hello", "--context", commonVar.Context).ShouldPass()
			configValue2 := helper.Cmd("odo", "config", "view", "--context", commonVar.Context).ShouldPass().Out()
			helper.DontMatchAllInOutput(configValue2, []string{"hello", "world"})
			envs = oc.GetEnvsDevFileDeployment("nodejs", "app", commonVar.Project)
			_, ok = envs["hello"]
			Expect(ok).To(BeFalse())
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

			output = helper.Cmd("odo", "preference", "set", "buildtimeout", "5", "-f").ShouldPass().Out()
			Expect(output).ToNot(ContainSubstring(promptMessageSubString))

			output = helper.Cmd("odo", "preference", "unset", "buildtimeout", "-f").ShouldPass().Out()
			Expect(output).ToNot(ContainSubstring(promptMessageSubString))
		})
	})

	Context("When ConsentTelemetry preference value is set", func() {
		// !! Do not test with true because it sends out the telemetry data and messes up the statistics !!
		It("should not prompt the user", func() {
			helper.Cmd("odo", "preference", "set", "ConsentTelemetry", "false", "-f").ShouldPass()
			output := helper.Cmd("odo", "create", "nodejs", "--context", commonVar.Context).ShouldPass().Out()
			Expect(output).ToNot(ContainSubstring(promptMessageSubString))
		})
	})

})
