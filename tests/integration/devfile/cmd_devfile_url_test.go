package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile url command tests", func() {
	var namespace, context, componentName, currentWorkingDirectory string

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		namespace = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		componentName = helper.RandString(6)

		helper.Chdir(context)

		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		// Devfile push requires experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Listing urls", func() {
		It("should list url after push", func() {
			var stdout string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))

			stdout = helper.CmdShouldFail("odo", "url", "create", url1, "--port", "8080")
			Expect(stdout).To(ContainSubstring("is not exposed"))

			stdout = helper.CmdShouldFail("odo", "url", "create", url1, "--port", "3000", "--ingress")
			Expect(stdout).To(ContainSubstring("host must be provided"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--ingress")
			helper.CmdShouldPass("odo", "push")
			helper.WaitForCmdOut("odo", []string{"url", "list"}, 1, false, func(output string) bool {
				if strings.Contains(output, url1) {
					Expect(output).Should(ContainSubstring(url1 + "." + host))
					return true
				}
				return false
			})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			helper.CmdShouldPass("odo", "push", "--project", namespace)

			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))
		})

		It("should be able to list url in machine readable json format", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--ingress")
			helper.CmdShouldPass("odo", "push", "--project", namespace)

			// odo url list -o json
			helper.WaitForCmdOut("odo", []string{"url", "list", "-o", "json"}, 1, true, func(output string) bool {
				desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"Ingress","apiVersion":"extensions/v1beta1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"rules":[{"host":"%s","http":{"paths":[{"path":"/","backend":{"serviceName":"%s","servicePort":3000}}]}}]},"status":{"loadBalancer":{}}}]}`, url1, url1+"."+host, componentName)
				if strings.Contains(output, url1) {
					Expect(desiredURLListJSON).Should(MatchJSON(output))
					return true
				}
				return false
			})
		})
	})

	Context("Creating urls", func() {
		It("should create a secure URL", func() {
			var stdout string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--secure", "--ingress")

			stdout = helper.CmdShouldPass("odo", "push", "--project", namespace)
			helper.MatchAllInOutput(stdout, []string{"https:", url1 + "." + host})
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{"https:", url1 + "." + host, "true"})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			helper.CmdShouldPass("odo", "push", "--project", namespace)

			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))
		})

		It("create with now flag should pass", func() {
			var stdout string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			stdout = helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--now", "--ingress")
			helper.MatchAllInOutput(stdout, []string{"URL " + url1 + " created for component", "http:", url1 + "." + host})
		})

		It("should create a automatically route on a openShift cluster", func() {

			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			url1 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1)

			helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			pushStdOut := helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			Expect(pushStdOut).NotTo(ContainSubstring("successfully deleted"))
			Expect(pushStdOut).NotTo(ContainSubstring("created"))
			Expect(pushStdOut).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))

			output := helper.CmdShouldPass("oc", "get", "routes", "--namespace", namespace)
			Expect(output).Should(ContainSubstring(url1))

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			pushStdOut = helper.CmdShouldPass("odo", "push", "--namespace", namespace)
			Expect(pushStdOut).NotTo(ContainSubstring("successfully deleted"))
			Expect(pushStdOut).NotTo(ContainSubstring("created"))
			Expect(pushStdOut).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))

			output = helper.CmdShouldPass("oc", "get", "routes", "--namespace", namespace)
			Expect(output).ShouldNot(ContainSubstring(url1))
		})

		It("should create a url for a unsupported devfile component", func() {
			url1 := helper.RandString(5)

			helper.CopyExample(filepath.Join("source", "python"), context)
			helper.Chdir(context)

			helper.CmdShouldPass("odo", "create", "python", "--project", namespace, componentName)

			helper.CmdShouldPass("odo", "url", "create", url1)

			helper.CmdShouldPass("odo", "push", "--namespace", namespace)

			output := helper.CmdShouldPass("oc", "get", "routes", "--namespace", namespace)
			Expect(output).Should(ContainSubstring(url1))
		})

		It("should be able to push again twice after creating and deleting a url", func() {
			var stdOut string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--ingress")

			helper.CmdShouldPass("odo", "push", "--project", namespace)
			stdOut = helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(stdOut).NotTo(ContainSubstring("successfully deleted"))
			Expect(stdOut).NotTo(ContainSubstring("created"))
			Expect(stdOut).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")

			helper.CmdShouldPass("odo", "push", "--project", namespace)
			stdOut = helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(stdOut).NotTo(ContainSubstring("successfully deleted"))
			Expect(stdOut).NotTo(ContainSubstring("created"))
			Expect(stdOut).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))
		})
	})

	Context("Describing urls", func() {
		It("should describe appropriate URL and error messages", func() {
			var stdout string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--ingress")

			stdout = helper.CmdShouldFail("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "exists in local", "odo push"})

			helper.CmdShouldPass("odo", "push", "--project", namespace)
			helper.WaitForCmdOut("odo", []string{"url", "describe", url1}, 1, false, func(output string) bool {
				if strings.Contains(output, url1) {
					Expect(output).Should(ContainSubstring(url1 + "." + host))
					return true
				}
				return false
			})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
		})
	})

})
