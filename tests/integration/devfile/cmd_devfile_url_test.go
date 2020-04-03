package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile url command tests", func() {
	var namespace string
	var context string
	var currentWorkingDirectory string
	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		namespace = helper.CreateRandProject()
		context = helper.CreateNewDevfileContext()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
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
			componentName := helper.RandString(6)
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--namespace", namespace, componentName)
			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))

			stdout = helper.CmdShouldFail("odo", "url", "create", url1, "--port", "8080")
			Expect(stdout).To(ContainSubstring("is not exposed"))

			stdout = helper.CmdShouldFail("odo", "url", "create", url1, "--port", "9090")
			Expect(stdout).To(ContainSubstring("host must be provided"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host)

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)
			helper.WaitTillUrlexist(namespace, url1, 1)
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{url1, url1 + "." + host})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)
			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))
		})

		It("should be able to list url in machine readable json format", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"
			componentName := helper.RandString(6)

			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--namespace", namespace, componentName)
			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host)

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)
			helper.WaitTillUrlexist(namespace, url1, 1)
			// odo url list -o json
			actualURLListJSON := helper.CmdShouldPass("odo", "url", "list", "-o", "json")
			desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"udo.udo.io/v1alpha1","metadata":{},"items":[{"kind":"Ingress","apiVersion":"extensions/v1beta1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"rules":[{"host":"%s","http":{"paths":[{"path":"/","backend":{"serviceName":"%s","servicePort":9090}}]}}]},"status":{"loadBalancer":{}}}]}`, url1, url1+"."+host, componentName)
			Expect(desiredURLListJSON).Should(MatchJSON(actualURLListJSON))
		})
	})

	Context("Creating urls", func() {
		It("should create a secure URL", func() {
			var stdout string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--namespace", namespace, componentName)

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host, "--secure")
			stdout = helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)
			helper.MatchAllInOutput(stdout, []string{"https:", url1 + "." + host})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)
			stdout = helper.CmdShouldFail("odo", "url", "list")
			Expect(stdout).To(ContainSubstring("no URLs found"))
		})

		It("create with now flag should pass", func() {
			var stdout string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--namespace", namespace, componentName)
			stdout = helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host, "--now")
			helper.MatchAllInOutput(stdout, []string{"URL created for component", "http:", url1 + "." + host})
		})
	})

	Context("Describing urls", func() {
		It("should describe appropriate URL and error messages", func() {
			var stdout string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"
			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--namespace", namespace, componentName)

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host)
			stdout = helper.CmdShouldFail("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "exists in local", "odo push"})
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", namespace)
			helper.WaitTillUrlexist(namespace, url1, 1)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, url1 + "." + host})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
		})
	})

})
