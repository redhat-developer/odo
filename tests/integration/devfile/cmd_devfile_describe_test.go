package devfile

import (
	"github.com/openshift/odo/pkg/component"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile describe command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		if os.Getenv("KUBERNETES") != "true" {
			Skip("Plain Kubernetes scenario only, skipping")
		}
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When executing odo describe", func() {
		It("should describe the component when it is not pushed", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")
			helper.CmdShouldPass("odo", "url", "create", "url-1", "--port", "3000", "--host", "example.com", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "url", "create", "url-2", "--port", "4000", "--host", "example.com", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context)
			cmpDescribe := helper.CmdShouldPass("odo", "describe", "--context", commonVar.Context)
			helper.MatchAllInOutput(cmpDescribe, []string{
				"cmp-git",
				"nodejs",
				"url-1",
				"url-2",
				"storage-1",
			})

			cmpDescribeJSON, err := helper.Unindented(helper.CmdShouldPass("odo", "describe", "-o", "json", "--context", commonVar.Context))
			Expect(err).Should(BeNil())
			expected, err := helper.Unindented(`{"kind": "Component","apiVersion": "odo.dev/v1alpha1","metadata": {"name": "cmp-git","namespace": "` + commonVar.Project + `","creationTimestamp": null},"spec":{"app": "testing","type":"nodejs","urls": {"kind": "List", "apiVersion": "odo.dev/v1alpha1", "metadata": {}, "items": [{"kind": "url", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "url-1", "creationTimestamp": null}, "spec": {"host": "url-1.example.com",
            "kind": "ingress", "port": 3000, "secure": false}, "status": {"state": "` + string(component.StateTypeNotPushed) + `"}}, {"kind": "url", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "url-2", "creationTimestamp": null}, "spec": {"host": "url-2.example.com", "port": 4000, "secure": false, "kind": "ingress"}, "status": {"state": "` + string(component.StateTypeNotPushed) + `"}}]},"storages": {"kind": "List", "apiVersion": "odo.dev/v1alpha1", "metadata": {}, "items": [{"kind": "storage", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "storage-1", "creationTimestamp": null}, "spec": {"containerName": "runtime","size": "1Gi", "path": "/data1"}}]},"ports": ["5858"]},"status": {"state": "` + string(component.StateTypeNotPushed) + `"}}`)
			Expect(err).Should(BeNil())
			Expect(cmpDescribeJSON).Should(MatchJSON(expected))

			// odo should describe not pushed component if component name is given.
			helper.CmdShouldPass("odo", "describe", "cmp-git", "--context", commonVar.Context)
			Expect(cmpDescribe).To(ContainSubstring("cmp-git"))

			helper.CmdShouldPass("odo", "delete", "-f", "--all", "--context", commonVar.Context)
		})

		It("should describe the component when it is pushed", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing")
			helper.CmdShouldPass("odo", "url", "create", "url-1", "--port", "3000", "--host", "example.com", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "url", "create", "url-2", "--port", "4000", "--host", "example.com", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			cmpDescribe := helper.CmdShouldPass("odo", "describe", "--context", commonVar.Context)
			helper.MatchAllInOutput(cmpDescribe, []string{
				"cmp-git",
				"nodejs",
				"url-1",
				"url-2",
				"storage-1",
			})

			cmpDescribeJSON, err := helper.Unindented(helper.CmdShouldPass("odo", "describe", "-o", "json", "--context", commonVar.Context))
			Expect(err).Should(BeNil())
			expected, err := helper.Unindented(`{"kind": "Component","apiVersion": "odo.dev/v1alpha1","metadata": {"name": "cmp-git","namespace": "` + commonVar.Project + `","creationTimestamp": null},"spec":{"app": "testing", "env": [{"name": "PROJECTS_ROOT", "value": "/project"},{"name": "PROJECT_SOURCE", "value": "/project"}, {"name": "DEBUG_PORT","value": "5858"}],"type":"nodejs","urls": {"kind": "List", "apiVersion": "odo.dev/v1alpha1", "metadata": {}, "items": [{"kind": "url", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "url-1", "creationTimestamp": null}, "spec": {"host": "url-1.example.com", "path": "/",
            "kind": "ingress", "port": 3000, "secure": false}, "status": {"state": "` + string(component.StateTypePushed) + `"}}, {"kind": "url", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "url-2", "creationTimestamp": null}, "spec": {"host": "url-2.example.com", "port": 4000, "path": "/", "secure": false, "kind": "ingress"}, "status": {"state": "` + string(component.StateTypePushed) + `"}}]},"storages": {"kind": "List", "apiVersion": "odo.dev/v1alpha1", "metadata": {}, "items": [{"kind": "storage", "apiVersion": "odo.dev/v1alpha1", "metadata": {"name": "storage-1", "creationTimestamp": null}, "spec": {"containerName": "runtime","size": "1Gi", "path": "/data1"}}]},"ports": ["5858"]},"status": {"state": "` + string(component.StateTypePushed) + `"}}`)
			Expect(err).Should(BeNil())
			Expect(cmpDescribeJSON).Should(MatchJSON(expected))

			// odo should describe not pushed component if component name is given.
			helper.CmdShouldPass("odo", "describe", "cmp-git", "--context", commonVar.Context)
			Expect(cmpDescribe).To(ContainSubstring("cmp-git"))

			helper.CmdShouldPass("odo", "delete", "-f", "--all", "--context", commonVar.Context)
		})
	})
})
