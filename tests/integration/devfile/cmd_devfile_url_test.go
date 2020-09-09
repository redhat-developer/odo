package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile url command tests", func() {
	var componentName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()

		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		// Devfile requires experimental mode to be set
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Listing urls", func() {
		It("should list url after push", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			stdout := helper.CmdShouldFail("odo", "url", "create", url1, "--port", "3000", "--ingress")
			Expect(stdout).To(ContainSubstring("host must be provided"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--ingress")
			stdout = helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			Expect(stdout).Should(ContainSubstring(url1 + "." + host))

			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "false", "ingress"})
		})

		It("should list ingress url with appropriate state", func() {
			url1 := helper.RandString(5)
			url2 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--context", context, "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host, "--secure", "--ingress", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			stdout := helper.CmdShouldPass("odo", "url", "list", "--context", context)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "true", "ingress"})

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f", "--context", context)
			helper.CmdShouldPass("odo", "url", "create", url2, "--port", "8080", "--host", host, "--ingress", "--context", context)
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", context)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", "true", "ingress"})
			helper.MatchAllInOutput(stdout, []string{url2, "Not Pushed", "false", "ingress"})
		})

		It("should be able to list ingress url in machine readable json format", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			// remove the endpoint came with the devfile
			// need to create an ingress to be more general for openshift/non-openshift cluster to run
			helper.CmdShouldPass("odo", "url", "delete", "3000/tcp", "-f")
			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--ingress")
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)

			// odo url list -o json
			helper.WaitForCmdOut("odo", []string{"url", "list", "-o", "json"}, 1, true, func(output string) bool {
				desiredURLListJSON := fmt.Sprintf(`{"kind":"List","apiVersion":"odo.dev/v1alpha1","metadata":{},"items":[{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"host":"%s","port":3000,"secure": false,"path": "/", "kind":"ingress"},"status":{"state":"Pushed"}}]}`, url1, url1+"."+host)
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
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host, "--secure", "--ingress")

			stdout := helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.MatchAllInOutput(stdout, []string{"https:", url1 + "." + host})
			stdout = helper.CmdShouldPass("odo", "url", "list")
			helper.MatchAllInOutput(stdout, []string{"https:", url1 + "." + host, "true"})
		})

		It("create and delete with now flag should pass", func() {
			var stdout string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			stdout = helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--now", "--ingress", "--context", commonVar.Context)

			// check the env for the runMode
			envOutput, err := helper.ReadFile(filepath.Join(commonVar.Context, ".odo/env/env.yaml"))
			Expect(err).To(BeNil())
			Expect(envOutput).To(ContainSubstring(" RunMode: run"))

			helper.MatchAllInOutput(stdout, []string{"URL " + url1 + " created for component", "http:", url1 + "." + host})
			stdout = helper.CmdShouldPass("odo", "url", "delete", url1, "--now", "-f", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{"URL " + url1 + " successfully deleted", "Applying URL changes"})
		})

		It("should be able to push again twice after creating and deleting a url", func() {
			var stdOut string
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--ingress")

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			stdOut = helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.DontMatchAllInOutput(stdOut, []string{"successfully deleted", "created"})
			Expect(stdOut).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			stdOut = helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.DontMatchAllInOutput(stdOut, []string{"successfully deleted", "created"})
			Expect(stdOut).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))
		})

		It("should not allow creating an invalid host", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project)
			stdOut := helper.CmdShouldFail("odo", "url", "create", "--host", "https://127.0.0.1:60104", "--port", "3000", "--ingress")
			Expect(stdOut).To(ContainSubstring("is not a valid host name"))
		})

		It("should not allow using tls secret if url is not secure", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project)
			stdOut := helper.CmdShouldFail("odo", "url", "create", "--tls-secret", "foo", "--port", "3000", "--ingress")
			Expect(stdOut).To(ContainSubstring("TLS secret is only available for secure URLs of Ingress kind"))
		})

		It("should report multiple issues when it's the case", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project)
			stdOut := helper.CmdShouldFail("odo", "url", "create", "--host", "https://127.0.0.1:60104", "--tls-secret", "foo", "--port", "3000", "--ingress")
			Expect(stdOut).To(And(ContainSubstring("is not a valid host name"), ContainSubstring("TLS secret is only available for secure URLs of Ingress kind")))
		})

		It("should not allow creating under an invalid container", func() {
			containerName := helper.RandString(5)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace)
			stdOut := helper.CmdShouldFail("odo", "url", "create", "--host", "com", "--port", "3000", "--container", containerName, "--ingress")
			Expect(stdOut).To(ContainSubstring(fmt.Sprintf("the container specified: %s does not exist in devfile", containerName)))
		})

		It("should not allow creating an endpoint with same name", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			stdOut := helper.CmdShouldFail("odo", "url", "create", "3000/tcp", "--host", "com", "--port", "3000", "--ingress")
			Expect(stdOut).To(ContainSubstring("url 3000/tcp already exist in devfile endpoint entry"))
		})

		It("should create URL with path defined in Endpoint", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8090", "--host", host, "--path", "testpath", "--ingress")

			stdout := helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.MatchAllInOutput(stdout, []string{url1, "/testpath", "created"})
		})

		It("should create URLs under different container names", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"
			url2 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--host", host, "--container", "runtime", "--ingress")
			helper.CmdShouldPass("odo", "url", "create", url2, "--port", "9090", "--host", host, "--container", "tools", "--ingress")
			stdout := helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.MatchAllInOutput(stdout, []string{url1, url2, "created"})
		})

		It("should not create URLs under different container names with same port number", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "java-springboot", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			stdout := helper.CmdShouldFail("odo", "url", "create", url1, "--port", "8080", "--host", host, "--container", "tools", "--ingress")
			helper.MatchAllInOutput(stdout, []string{fmt.Sprintf("cannot set URL %s under container tools", url1), "TargetPort 8080 is being used under container runtime"})
		})

		It("should error out on devfile flag", func() {
			helper.CmdShouldFail("odo", "url", "create", "mynodejs", "--devfile", "invalid.yaml")
			helper.CmdShouldFail("odo", "url", "delete", "mynodejs", "--devfile", "invalid.yaml")
		})

	})

	Context("Describing urls", func() {
		It("should describe appropriate Ingress URLs", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--host", host, "--ingress")

			stdout := helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1 + "." + host, "Not Pushed", "false", "ingress", "odo push"})

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1 + "." + host, "Pushed", "false", "ingress"})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1 + "." + host, "Locally Deleted", "false", "ingress"})

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--host", host, "--secure", "--ingress")
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1 + "." + host, "Pushed", "true", "ingress"})
		})

		It("should describe Ingress URL in json format", func() {
			url1 := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))
			// remove the endpoint came with the devfile
			// need to create an ingress to be more general for openshift/non-openshift cluster to run
			helper.CmdShouldPass("odo", "url", "delete", "3000/tcp", "-f")
			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000", "--host", host, "--ingress")
			helper.CmdShouldPass("odo", "push", "--project", namespace)

			// odo url describe url1 -o json
			stdout := helper.CmdShouldPass("odo", "url", "describe", url1, "-o", "json")
			desiredURLListJSON := fmt.Sprintf(`{"kind":"url","apiVersion":"odo.dev/v1alpha1","metadata":{"name":"%s","creationTimestamp":null},"spec":{"host":"%s","port":3000,"secure": false,"path": "/", "kind":"ingress"},"status":{"state":"Pushed"}}`, url1, url1+"."+host)
			Expect(desiredURLListJSON).Should(MatchJSON(stdout))
		})
	})

	Context("Testing URLs for OpenShift specific scenarios", func() {
		JustBeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}
		})

		It("should error out when a host is provided with a route on a openShift cluster", func() {
			url1 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.CmdShouldFail("odo", "url", "create", url1, "--host", "com", "--port", "3000")
			Expect(output).To(ContainSubstring("host is not supported"))
		})

		It("should list route and ingress urls with appropriate state", func() {
			url1 := helper.RandString(5)
			url2 := helper.RandString(5)
			ingressurl := helper.RandString(5)
			host := helper.RandString(5) + ".com"

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--secure")
			helper.CmdShouldPass("odo", "url", "create", ingressurl, "--port", "8080", "--host", host, "--ingress")
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "url", "create", url2, "--port", "8080")
			stdout := helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "true", "route"})
			helper.MatchAllInOutput(stdout, []string{url2, "Not Pushed", "false", "route"})
			helper.MatchAllInOutput(stdout, []string{ingressurl, "Pushed", "false", "ingress"})

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			stdout = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", "true", "route"})
			helper.MatchAllInOutput(stdout, []string{url2, "Not Pushed", "false", "route"})
			helper.MatchAllInOutput(stdout, []string{ingressurl, "Pushed", "false", "ingress"})
		})

		It("should create a automatically route on a openShift cluster", func() {
			url1 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "3000")

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			pushStdOut := helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.DontMatchAllInOutput(pushStdOut, []string{"successfully deleted", "created"})
			Expect(pushStdOut).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))

			output := helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			Expect(output).Should(ContainSubstring(url1))

			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			pushStdOut = helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.DontMatchAllInOutput(pushStdOut, []string{"successfully deleted", "created"})
			Expect(pushStdOut).To(ContainSubstring("URLs are synced with the cluster, no changes are required"))

			output = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			Expect(output).ShouldNot(ContainSubstring(url1))
		})

		It("should create a route on a openShift cluster without calling url create", func() {

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			output := helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			helper.MatchAllInOutput(output, []string{"URL 3000-tcp", "created"})

			output = helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			Expect(output).Should(ContainSubstring("3000-tcp"))
		})

		It("should describe appropriate Route URLs", func() {
			url1 := helper.RandString(5)

			helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, componentName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080")

			stdout := helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "Not Pushed", "false", "route", "odo push"})

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "false", "route"})
			helper.CmdShouldPass("odo", "url", "delete", url1, "-f")
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", "false", "route"})

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "9090", "--secure")
			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)
			stdout = helper.CmdShouldPass("odo", "url", "describe", url1)
			helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "true", "route"})
		})

		It("should create a url for a unsupported devfile component", func() {
			url1 := helper.RandString(5)

			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)

			helper.CmdShouldPass("odo", "create", "python", "--project", commonVar.Project, componentName)

			helper.CmdShouldPass("odo", "url", "create", url1)

			helper.CmdShouldPass("odo", "push", "--project", commonVar.Project)

			output := helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			Expect(output).Should(ContainSubstring(url1))
		})

		// remove once https://github.com/openshift/odo/issues/3550 is resolved
		It("should list URLs for s2i components", func() {
			url1 := helper.RandString(5)
			url2 := helper.RandString(5)

			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "nodejs", "--context", commonVar.Context, "--project", commonVar.Project, componentName, "--s2i")

			helper.CmdShouldPass("odo", "url", "create", url1, "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "url", "create", url2, "--port", "8080", "--context", commonVar.Context, "--ingress", "--host", "com")

			stdout := helper.CmdShouldPass("odo", "url", "list", "--context", commonVar.Context)
			helper.MatchAllInOutput(stdout, []string{url1, url2})
		})
	})
})
