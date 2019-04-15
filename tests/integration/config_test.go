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

var _ = Describe("odo config test", func() {
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
	})

	Context("Creating odo global config", func() {

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

		It("should be checking to see if global config values are the same as the configured ones", func() {
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

	Context("Creating odo local config", func() {
		JustBeforeEach(func() {
			project = helper.CreateRandProject()
		})
		JustAfterEach(func() {
			helper.DeleteProject(project)
			os.RemoveAll(".odo")
		})
		It("should be checking to see if local config values are the same as the configured ones", func() {
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
			os.RemoveAll(context)

		})

		It("should set and unset env variables", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", project)
			helper.CmdShouldPass("odo", "config", "set", "--env", "PORT=4000", "--env", "PORT=1234")
			configPort := helper.GetConfigValue("PORT")
			Expect(configPort).To(ContainSubstring("1234"))
			helper.CmdShouldPass("odo", "config", "set", "--env", "SECRET_KEY=R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q=")
			configSecret := helper.GetConfigValue("SECRET_KEY")
			Expect(configSecret).To(ContainSubstring("R2lyaXNoIFJhbW5hbmkgaXMgdGhlIGJlc3Q"))
			helper.CmdShouldPass("odo", "config", "unset", "--env", "PORT")
			helper.CmdShouldPass("odo", "config", "unset", "--env", "SECRET_KEY")
			configValue := helper.CmdShouldPass("odo", "config", "view")
			Expect(configValue).To(Not(ContainSubstring(("PORT"))))
			Expect(configValue).To(Not(ContainSubstring(("SECRET_KEY"))))
		})
	})
})
