package devfile

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo devfile describe command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		if os.Getenv("KUBERNETES") != "true" {
			Skip("Plain Kubernetes scenario only, skipping")
		}

		commonVar = helper.CommonBeforeEach()
		helper.SetDefaultDevfileRegistryAsStaging()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When executing odo describe", func() {
		JustAfterEach(func() {
			// odo delete requires changing directory because it does not work as intended with --context
			// TODO: Remove helper.Chdir after these issues are closed - https://github.com/openshift/odo/issues/4451
			// TODO: and https://github.com/openshift/odo/issues/4135
			helper.Chdir(commonVar.Context)
			helper.Cmd("odo", "delete", "-f", "--all").ShouldPass()
		})

		It("should describe the component when it is not pushed", func() {
			helper.Cmd("odo", "create", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing").ShouldPass()
			helper.Cmd("odo", "url", "create", "url-1", "--port", "3000", "--host", "example.com", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "url", "create", "url-2", "--port", "4000", "--host", "example.com", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context).ShouldPass()
			cmpDescribe := helper.Cmd("odo", "describe", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(cmpDescribe, []string{
				"cmp-git",
				"nodejs",
				"url-1",
				"url-2",
				"storage-1",
			})

			cmpDescribeJSON, err := helper.Unindented(helper.Cmd("odo", "describe", "-o", "json", "--context", commonVar.Context).ShouldPass().Out())
			Expect(err).Should(BeNil())
			valuesDes := gjson.GetMany(cmpDescribeJSON, "kind", "spec.urls.items.0.metadata.name", "spec.urls.items.0.spec.host", "spec.urls.items.1.metadata.name", "spec.urls.items.1.spec.host", "spec.storages.items.0.metadata.name", "spec.storages.items.0.spec.containerName")
			expectedDes := []string{"Component", "url-1", "url-1.example.com", "url-2", "url-2.example.com", "storage-1", "runtime"}
			Expect(helper.GjsonMatcher(valuesDes, expectedDes)).To(Equal(true))

			// odo should describe not pushed component if component name is given.
			helper.Cmd("odo", "describe", "cmp-git", "--context", commonVar.Context).ShouldPass()
			Expect(cmpDescribe).To(ContainSubstring("cmp-git"))
		})

		It("should describe the component when it is pushed", func() {
			helper.Cmd("odo", "create", "nodejs", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing").ShouldPass()
			helper.Cmd("odo", "url", "create", "url-1", "--port", "3000", "--host", "example.com", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "url", "create", "url-2", "--port", "4000", "--host", "example.com", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			cmpDescribe := helper.Cmd("odo", "describe", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(cmpDescribe, []string{
				"cmp-git",
				"nodejs",
				"url-1",
				"url-2",
				"storage-1",
			})

			cmpDescribeJSON, err := helper.Unindented(helper.Cmd("odo", "describe", "-o", "json", "--context", commonVar.Context).ShouldPass().Out())
			Expect(err).Should(BeNil())
			values := gjson.GetMany(cmpDescribeJSON, "kind", "spec.urls.items.0.metadata.name", "spec.urls.items.0.spec.host", "spec.urls.items.1.metadata.name", "spec.urls.items.1.spec.host", "spec.storages.items.0.metadata.name", "spec.storages.items.0.spec.containerName")
			expected := []string{"Component", "url-1", "url-1.example.com", "url-2", "url-2.example.com", "storage-1", "runtime"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))

			// odo should describe not pushed component if component name is given.
			helper.Cmd("odo", "describe", "cmp-git", "--context", commonVar.Context).ShouldPass()
			Expect(cmpDescribe).To(ContainSubstring("cmp-git"))
		})
	})

	Context("when running odo describe for machine readable output", func() {
		It("should show json output for working cluster", func() {
			helper.Cmd("odo", "create", "nodejs", "--context", commonVar.Context).ShouldPass()
			output := helper.Cmd("odo", "describe", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "metadata.name", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Not Pushed"})).To(Equal(true))
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			output = helper.Cmd("odo", "describe", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
			values = gjson.GetMany(output, "kind", "metadata.name", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Pushed"})).To(Equal(true))
		})

		It("should show json output for non connected cluster", func() {
			helper.Cmd("odo", "create", "nodejs", "--context", commonVar.Context).ShouldPass()
			output := helper.Cmd("odo", "describe", "--context", commonVar.Context, "-o", "json").WithEnv("KUBECONFIG=/no/path", "GLOBALODOCONFIG="+os.Getenv("GLOBALODOCONFIG")).ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "metadata.name", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", "nodejs", "Unknown"})).To(Equal(true))
		})
	})

})
