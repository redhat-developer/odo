package devfile

import (
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/pkg/component"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile describe command tests", func() {
	var namespace, context, currentWorkingDirectory, originalKubeconfig string

	// Using program command according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		if os.Getenv("KUBERNETES") != "true" {
			Skip("Plain Kubernetes scenario only, skipping")
		}
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("When executing odo describe", func() {
		It("should describe the component when it is not pushed", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "cmp-git", "--project", namespace, "--context", context, "--app", "testing")
			helper.CmdShouldPass("odo", "url", "create", "url-1", "--port", "3000", "--host", "example.com", "--context", context)
			helper.CmdShouldPass("odo", "url", "create", "url-2", "--port", "4000", "--host", "example.com", "--context", context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", context)
			cmpDescribe := helper.CmdShouldPass("odo", "describe", "--context", context)
			helper.MatchAllInOutput(cmpDescribe, []string{
				"cmp-git",
				"nodejs",
				"url-1",
				"url-2",
				"storage-1",
			})

			cmpDescribeJSON, err := helper.Unindented(helper.CmdShouldPass("odo", "describe", "-o", "json", "--context", context))
			Expect(err).Should(BeNil())
			expected, err := helper.Unindented(`{"kind": "Component","apiVersion": "odo.dev/v1alpha1","metadata": {"name": "cmp-git","namespace": "` + namespace + `","creationTimestamp": null},"spec":{"app": "testing","type":"nodejs","urls": {"kind": "List", "apiVersion": "odo.dev/v1alpha1", "metadata": {}, "items": [{"kind": "url", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "url-1", "creationTimestamp": null}, "spec": {"host": "url-1.example.com",
            "kind": "ingress", "port": 3000, "secure": false}, "status": {"state": "` + string(component.StateTypeNotPushed) + `"}}, {"kind": "url", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "url-2", "creationTimestamp": null}, "spec": {"host": "url-2.example.com", "port": 4000, "secure": false, "kind": "ingress"}, "status": {"state": "` + string(component.StateTypeNotPushed) + `"}}]},"storages": {"kind": "List", "apiVersion": "odo.dev/v1alpha1", "metadata": {}, "items": [{"kind": "storage", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "storage-1", "creationTimestamp": null}, "spec": {"containerName": "runtime","size": "1Gi", "path": "/data1"}}]},"ports": ["5858"]},"status": {"state": "` + string(component.StateTypeNotPushed) + `"}}`)
			Expect(err).Should(BeNil())
			Expect(cmpDescribeJSON).Should(MatchJSON(expected))

			// odo should describe not pushed component if component name is given.
			helper.CmdShouldPass("odo", "describe", "cmp-git", "--context", context)
			Expect(cmpDescribe).To(ContainSubstring("cmp-git"))

			helper.CmdShouldPass("odo", "delete", "-f", "--all", "--context", context)
		})

		It("should describe the component when it is pushed", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "cmp-git", "--project", namespace, "--context", context, "--app", "testing")
			helper.CmdShouldPass("odo", "url", "create", "url-1", "--port", "3000", "--host", "example.com", "--context", context)
			helper.CmdShouldPass("odo", "url", "create", "url-2", "--port", "4000", "--host", "example.com", "--context", context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)
			cmpDescribe := helper.CmdShouldPass("odo", "describe", "--context", context)
			helper.MatchAllInOutput(cmpDescribe, []string{
				"cmp-git",
				"nodejs",
				"url-1",
				"url-2",
				"storage-1",
			})

			cmpDescribeJSON, err := helper.Unindented(helper.CmdShouldPass("odo", "describe", "-o", "json", "--context", context))
			Expect(err).Should(BeNil())
			expected, err := helper.Unindented(`{"kind": "Component","apiVersion": "odo.dev/v1alpha1","metadata": {"name": "cmp-git","namespace": "` + namespace + `","creationTimestamp": null},"spec":{"app": "testing", "env": [{"name": "PROJECTS_ROOT", "value": "/project"},{"name": "PROJECT_SOURCE", "value": "/project"}, {"name": "DEBUG_PORT","value": "5858"}],"type":"nodejs","urls": {"kind": "List", "apiVersion": "odo.dev/v1alpha1", "metadata": {}, "items": [{"kind": "url", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "url-1", "creationTimestamp": null}, "spec": {"host": "url-1.example.com", "path": "/",
            "kind": "ingress", "port": 3000, "secure": false}, "status": {"state": "` + string(component.StateTypePushed) + `"}}, {"kind": "url", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "url-2", "creationTimestamp": null}, "spec": {"host": "url-2.example.com", "port": 4000, "path": "/", "secure": false, "kind": "ingress"}, "status": {"state": "` + string(component.StateTypePushed) + `"}}]},"storages": {"kind": "List", "apiVersion": "odo.dev/v1alpha1", "metadata": {}, "items": [{"kind": "storage", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "storage-1", "creationTimestamp": null}, "spec": {"containerName": "runtime","size": "1Gi", "path": "/data1"}}]},"ports": ["5858"]},"status": {"state": "` + string(component.StateTypePushed) + `"}}`)
			Expect(err).Should(BeNil())
			Expect(cmpDescribeJSON).Should(MatchJSON(expected))

			// odo should describe not pushed component if component name is given.
			helper.CmdShouldPass("odo", "describe", "cmp-git", "--context", context)
			Expect(cmpDescribe).To(ContainSubstring("cmp-git"))

			helper.CmdShouldPass("odo", "delete", "-f", "--all", "--context", context)
		})
	})
})
