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
var testNamespacedImage = "https://raw.githubusercontent.com/bucharest-gold/centos7-s2i-nodejs/master/imagestreams/nodejs-centos7.json"
var testPHPGitURL = "https://github.com/appuio/example-php-sti-helloworld"
var oc helper.OcRunner
var project string
var context string
var originalDir string

var _ = Describe("odo preference and config command tests", func() {
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
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
			Expect(appHelp).To(ContainSubstring("Modifies odo specific configuration settings within the config file"))
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
			Expect(configOutput).To(ContainSubstring("UpdateNotification"))
			Expect(configOutput).To(ContainSubstring("NamePrefix"))
			Expect(configOutput).To(ContainSubstring("Timeout"))
			updateNotificationValue := helper.GetPreferenceValue("UpdateNotification")
			Expect(updateNotificationValue).To(BeEmpty())
			namePrefixValue := helper.GetPreferenceValue("NamePrefix")
			Expect(namePrefixValue).To(BeEmpty())
			timeoutValue := helper.GetPreferenceValue("Timeout")
			Expect(timeoutValue).To(BeEmpty())
		})
	})

	Context("When configuring global config values", func() {
		It("should successfully updated", func() {
			helper.CmdShouldPass("odo", "preference", "set", "updatenotification", "false")
			helper.CmdShouldPass("odo", "preference", "set", "timeout", "5")
			UpdateNotificationValue := helper.GetPreferenceValue("UpdateNotification")
			Expect(UpdateNotificationValue).To(ContainSubstring("false"))
			TimeoutValue := helper.GetPreferenceValue("Timeout")
			Expect(TimeoutValue).To(ContainSubstring("5"))
			helper.CmdShouldPass("odo", "preference", "unset", "-f", "timeout")
			timeoutValue := helper.GetPreferenceValue("Timeout")
			Expect(timeoutValue).To(BeEmpty())
			globalConfPath := os.Getenv("HOME")
			os.RemoveAll(filepath.Join(globalConfPath, ".odo"))
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
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", project)
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
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", project, "--context", context)
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
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "PORT=4000", "--env", "PORT=1234", "--context", context)
			configPort := helper.GetConfigValueWithContext("PORT", context)
			Expect(configPort).To(ContainSubstring("1234"))
			helper.CmdShouldPass("odo", "config", "set", "--env", "SECRET_KEY=R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q=", "--context", context)
			configSecret := helper.GetConfigValueWithContext("SECRET_KEY", context)
			Expect(configSecret).To(ContainSubstring("R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q"))
			helper.CmdShouldPass("odo", "config", "unset", "--env", "PORT", "--context", context)
			helper.CmdShouldPass("odo", "config", "unset", "--env", "SECRET_KEY", "--context", context)
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", context)
			Expect(configValue).To(Not(ContainSubstring(("PORT"))))
			Expect(configValue).To(Not(ContainSubstring(("SECRET_KEY"))))
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
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "config", "set", "--env", "hello=world", "--context", context)
			kubeconfigOld := os.Getenv("KUBECONFIG")
			os.Setenv("KUBECONFIG", "/no/such/path")
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", context)
			Expect(configValue).To(ContainSubstring("hello"))
			Expect(configValue).To(ContainSubstring("world"))
			os.Setenv("KUBECONFIG", kubeconfigOld)
		})

		It("should set config veriable without logging in", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", project, "--context", context)
			kubeconfigOld := os.Getenv("KUBECONFIG")
			os.Setenv("KUBECONFIG", "/no/such/path")
			helper.CmdShouldPass("odo", "config", "set", "--force", "--context", context, "Name", "foobar")
			configValue := helper.CmdShouldPass("odo", "config", "view", "--context", context)
			Expect(configValue).To(ContainSubstring("foobar"))
			helper.CmdShouldPass("odo", "config", "unset", "--context", context, "Name")
			os.Setenv("KUBECONFIG", kubeconfigOld)
		})
	})
})
