package integration

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

// TODO: A neater way to provide odo path. Currently we assume \
// odo and oc in $PATH already.
var oc helper.OcRunner
var project string
var context string
var originalDir string

var _ = Describe("odo preference and config command tests", func() {
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
	})

	Context("check that help works", func() {
		It("should display help when using odo -h", func() {
			appHelp := helper.CmdShouldPass("odo", "-h")
			Expect(appHelp).To(ContainSubstring(`Use "odo [command] --help" for more information about a command.`))
		})

		It("should display help when using odo help", func() {
			appHelp := helper.CmdShouldPass("odo", "help")
			Expect(appHelp).To(ContainSubstring(`Use "odo [command] --help" for more information about a command.`))
		})

		It("should display help when using odo --help", func() {
			appHelp := helper.CmdShouldPass("odo", "--help")
			Expect(appHelp).To(ContainSubstring(`Use "odo [command] --help" for more information about a command.`))
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
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "preference.yaml"))
		})
		JustAfterEach(func() {
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should get the default global config keys", func() {
			configOutput := helper.CmdShouldPass("odo", "preference", "view")
			helper.MatchAllInOutput(configOutput, []string{"UpdateNotification", "NamePrefix", "Timeout", "PushTarget"})
			updateNotificationValue := helper.GetPreferenceValue("UpdateNotification")
			Expect(updateNotificationValue).To(BeEmpty())
			namePrefixValue := helper.GetPreferenceValue("NamePrefix")
			Expect(namePrefixValue).To(BeEmpty())
			timeoutValue := helper.GetPreferenceValue("Timeout")
			Expect(timeoutValue).To(BeEmpty())
		})
	})

	Context("When configuring global config values", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "preference.yaml"))
		})
		JustAfterEach(func() {
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		})
		It("should successfully updated", func() {
			helper.CmdShouldPass("odo", "preference", "set", "updatenotification", "false")
			helper.CmdShouldPass("odo", "preference", "set", "timeout", "5")
			helper.CmdShouldPass("odo", "preference", "set", "pushtarget", "docker")
			UpdateNotificationValue := helper.GetPreferenceValue("UpdateNotification")
			Expect(UpdateNotificationValue).To(ContainSubstring("false"))
			TimeoutValue := helper.GetPreferenceValue("Timeout")
			Expect(TimeoutValue).To(ContainSubstring("5"))
			PushTargetValue := helper.GetPreferenceValue("PushTarget")
			Expect(PushTargetValue).To(ContainSubstring("docker"))
			helper.CmdShouldPass("odo", "preference", "set", "-f", "pushtarget", "kube")
			PushTargetValue = helper.GetPreferenceValue("PushTarget")
			Expect(PushTargetValue).To(ContainSubstring("kube"))
			helper.CmdShouldPass("odo", "preference", "unset", "-f", "timeout")
			timeoutValue := helper.GetPreferenceValue("Timeout")
			Expect(timeoutValue).To(BeEmpty())
			helper.CmdShouldPass("odo", "preference", "unset", "-f", "pushtarget")
			PushTargetValue = helper.GetPreferenceValue("PushTarget")
			Expect(PushTargetValue).To(BeEmpty())
			globalConfPath := os.Getenv("HOME")
			os.RemoveAll(filepath.Join(globalConfPath, ".odo"))
		})
		It("should unsuccessfully update", func() {
			helper.CmdShouldFail("odo", "preference", "set", "-f", "pushtarget", "invalid-value")
			helper.CmdShouldFail("odo", "preference", "set", "-f", "experimental", "invalid-value")
			helper.CmdShouldFail("odo", "preference", "set", "-f", "updatenotification", "invalid-value")
		})

		It("should show json output", func() {
			prefJSONOutput, err := helper.Unindented(helper.CmdShouldPass("odo", "preference", "view", "-o", "json"))
			Expect(err).Should(BeNil())
			expected, err := helper.Unindented(`{"kind":"PreferenceList","apiVersion":"odo.dev/v1alpha1","items":[{"Name":"UpdateNotification","Value":null,"Default":true,"Type":"bool","Description":"Flag to control if an update notification is shown or not (Default: true)"},{"Name":"NamePrefix","Value":null,"Default":"","Type":"string","Description":"Use this value to set a default name prefix (Default: current directory name)"},{"Name":"Timeout","Value":null,"Default":1,"Type":"int","Description":"Timeout (in seconds) for OpenShift server connection check (Default: 1)"},{"Name":"BuildTimeout","Value":null,"Default":300,"Type":"int","Description":"BuildTimeout (in seconds) for waiting for a build of the git component to complete (Default: 300)"},{"Name":"PushTimeout","Value":null,"Default":240,"Type":"int","Description":"PushTimeout (in seconds) for waiting for a Pod to come up (Default: 240)"},{"Name":"Experimental","Value":null,"Default":false,"Type":"bool","Description":"Set this value to true to expose features in development/experimental mode"},{"Name":"PushTarget","Value":null,"Default":"kube","Type":"string","Description":"Set this value to 'kube' or 'docker' to tell odo where to push applications to. (Default: kube)"}]}`)
			Expect(err).Should(BeNil())
			Expect(prefJSONOutput).Should(MatchJSON(expected))
		})

	})

	Context("when creating odo local config in the same config dir", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "preference.yaml"))
			project = helper.CreateRandProject()
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.Unsetenv("GLOBALODOCONFIG")
			helper.Chdir(originalDir)
			helper.DeleteDir(context)
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
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", project)
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
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "preference.yaml"))
			project = helper.CreateRandProject()
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.Unsetenv("GLOBALODOCONFIG")
			helper.DeleteDir(context)
		})
		It("should allow setting and unsetting a config locally with context", func() {
			context := helper.CreateNewContext()
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
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", project, "--context", context)
			for _, testCase := range cases {

				helper.CmdShouldPass("odo", "config", "set", "-f", "--context", context, testCase.paramName, testCase.paramValue)
				configOutput := helper.CmdShouldPass("odo", "config", "unset", "-f", "--context", context, testCase.paramName)
				Expect(configOutput).To(ContainSubstring("Local config was successfully updated."))
				Value := helper.GetConfigValueWithContext(testCase.paramName, context)
				Expect(Value).To(BeEmpty())
			}
		})
	})

	Context("when creating odo local config with env variables", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "preference.yaml"))
			project = helper.CreateRandProject()
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.Unsetenv("GLOBALODOCONFIG")
			helper.DeleteDir(context)
		})
		It("should set and unset env variables", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "PORT=4000", "--env", "PORT=1234", "--context", context)
			configPort := helper.GetConfigValueWithContext("PORT", context)
			Expect(configPort).To(ContainSubstring("1234"))
			helper.CmdShouldPass("odo", "config", "set", "--env", "SECRET_KEY=R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q=", "--context", context)
			configSecret := helper.GetConfigValueWithContext("SECRET_KEY", context)
			Expect(configSecret).To(ContainSubstring("R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q"))
			helper.CmdShouldPass("odo", "config", "unset", "--env", "PORT", "--context", context)
			helper.CmdShouldPass("odo", "config", "unset", "--env", "SECRET_KEY", "--context", context)
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", context)
			helper.DontMatchAllInOutput(configValue, []string{"PORT", "SECRET_KEY"})
		})
		It("should check for existence of environment variable in config before unsetting it", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "PORT=4000", "--env", "PORT=1234", "--context", context)

			// unset a valid env var
			helper.CmdShouldPass("odo", "config", "unset", "--env", "PORT", "--context", context)

			// try to unset an env var that doesn't exist
			stdOut := helper.CmdShouldFail("odo", "config", "unset", "--env", "nosuchenv", "--context", context)
			Expect(stdOut).To(ContainSubstring("unable to find environment variable nosuchenv in the component"))
		})
	})

	Context("when viewing local config without logging into the OpenShift cluster", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "preference.yaml"))
			project = helper.CreateRandProject()
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.Unsetenv("GLOBALODOCONFIG")
			helper.DeleteDir(context)
		})
		It("should list config successfully", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "hello=world", "--context", context)
			kubeconfigOld := os.Getenv("KUBECONFIG")
			os.Setenv("KUBECONFIG", "/no/such/path")
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", context)
			helper.MatchAllInOutput(configValue, []string{"hello", "world"})
			os.Setenv("KUBECONFIG", kubeconfigOld)
		})

		It("should set config variable without logging in", func() {
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "--project", project, "--context", context)
			kubeconfigOld := os.Getenv("KUBECONFIG")
			os.Setenv("KUBECONFIG", "/no/such/path")
			helper.CmdShouldPass("odo", "config", "set", "--force", "--context", context, "Name", "foobar")
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", context)
			Expect(configValue).To(ContainSubstring("foobar"))
			helper.CmdShouldPass("odo", "config", "unset", "--force", "--context", context, "Name")
			os.Setenv("KUBECONFIG", kubeconfigOld)
		})
	})

	Context("when using --now with config command", func() {
		JustBeforeEach(func() {
			context = helper.CreateNewContext()
			os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "preference.yaml"))
			project = helper.CreateRandProject()
		})
		JustAfterEach(func() {
			helper.DeleteDir(context)
			helper.DeleteProject(project)
			os.Unsetenv("GLOBALODOCONFIG")
		})

		It("should successfully set and unset variables", func() {
			//set env var
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "nodejs", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "config", "set", "--now", "--env", "hello=world", "--context", context)
			//*Check config
			configValue1 := helper.CmdShouldPass("odo", "config", "view", "--context", context)
			helper.MatchAllInOutput(configValue1, []string{"hello", "world"})
			//*Check dc
			envs := oc.GetEnvs("nodejs", "app", project)
			val, ok := envs["hello"]
			Expect(ok).To(BeTrue())
			Expect(val).To(ContainSubstring("world"))
			// unset a valid env var
			helper.CmdShouldPass("odo", "config", "unset", "--now", "--env", "hello", "--context", context)
			configValue2 := helper.CmdShouldPass("odo", "config", "view", "--context", context)
			helper.DontMatchAllInOutput(configValue2, []string{"hello", "world"})
			envs = oc.GetEnvs("nodejs", "app", project)
			_, ok = envs["hello"]
			Expect(ok).To(BeFalse())
		})
	})
})
