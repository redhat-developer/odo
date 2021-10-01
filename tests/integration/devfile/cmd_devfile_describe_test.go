package devfile

import (
	"os"
	"path/filepath"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/pkg/component"
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
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	When("a component is created", func() {
		BeforeEach(func() {
			// Using Django example here because it helps to distinguish between language and projectType.
			// With nodejs, both projectType and language is nodejs, but with python-django, django is the projectType and python is the language
			helper.CopyExample(filepath.Join("source", "python"), commonVar.Context)
			helper.Cmd("odo", "create", "python-django", "cmp-git", "--project", commonVar.Project, "--context", commonVar.Context, "--app", "testing").ShouldPass()
			helper.Cmd("odo", "url", "create", "url-1", "--port", "3000", "--host", "example.com", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "url", "create", "url-2", "--port", "4000", "--host", "example.com", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "storage", "create", "storage-1", "--size", "1Gi", "--path", "/data1", "--context", commonVar.Context).ShouldPass()
		})
		AfterEach(func() {
			// odo delete requires changing directory because it does not work as intended with --context
			// TODO: Remove helper.Chdir after these issues are closed - https://github.com/openshift/odo/issues/4451
			// TODO: and https://github.com/openshift/odo/issues/4135
			helper.Chdir(commonVar.Context)
			helper.Cmd("odo", "delete", "-f", "--all").ShouldPass()
		})
		var checkDescribe = func(status string) {
			cmpDescribe := helper.Cmd("odo", "describe", "--context", commonVar.Context).ShouldPass().Out()
			helper.MatchAllInOutput(cmpDescribe, []string{
				"cmp-git",
				"django",
				"url-1",
				"url-2",
				"storage-1",
			})
			By("checking describe works with json output", func() {
				cmpDescribeJSON, err := helper.Unindented(helper.Cmd("odo", "describe", "-o", "json", "--context", commonVar.Context).ShouldPass().Out())
				Expect(err).Should(BeNil())
				valuesDes := gjson.GetMany(cmpDescribeJSON, "kind", "metadata.name", "status.state", "spec.urls.items.0.metadata.name", "spec.urls.items.0.spec.host", "spec.urls.items.1.metadata.name", "spec.urls.items.1.spec.host", "spec.storages.items.0.metadata.name", "spec.storages.items.0.spec.containerName")
				expectedDes := []string{"Component", "cmp-git", status, "url-1", "url-1.example.com", "url-2", "url-2.example.com", "storage-1", "py-web"}
				Expect(helper.GjsonMatcher(valuesDes, expectedDes)).To(Equal(true))
			})

			By("checking describe with component name works", func() {
				// odo should describe not-pushed component if component name is given.
				helper.Cmd("odo", "describe", "cmp-git", "--context", commonVar.Context).ShouldPass()
				Expect(cmpDescribe).To(ContainSubstring("cmp-git"))
			})
		}

		It("should describe the component correctly", func() {
			checkDescribe("Not Pushed")
		})
		When("the component is pushed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			})
			It("should describe the component correctly", func() {
				checkDescribe("Pushed")
			})
		})
	})

	When("executing odo describe from a disconnected cluster", func() {
		var compName = "mynode"
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "nodejs", compName, "--context", commonVar.Context).ShouldPass()
		})
		It("should describe successfully", func() {
			output := helper.Cmd("odo", "describe", "--context", commonVar.Context).WithEnv("KUBECONFIG=/no/path", "GLOBALODOCONFIG="+os.Getenv("GLOBALODOCONFIG")).ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{compName, "nodejs"})
		})

		It("should show json output", func() {
			output := helper.Cmd("odo", "describe", "--context", commonVar.Context, "-o", "json").WithEnv("KUBECONFIG=/no/path", "GLOBALODOCONFIG="+os.Getenv("GLOBALODOCONFIG")).ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "metadata.name", "spec.type", "status.state")
			Expect(helper.GjsonMatcher(values, []string{"Component", compName, "nodejs", "Unknown"})).To(Equal(true))
		})
	})

	Context("devfile has missing metadata", func() {
		// Note: We will be using SpringBoot example here because it helps to distinguish between language and projectType.
		// In terms of SpringBoot, spring is the projectType and java is the language; see https://github.com/openshift/odo/issues/4815

		var metadata devfilepkg.DevfileMetadata

		// checkDescribe checks the describe output (both normal and json) to see if it contains the expected componentType
		var checkDescribe = func(componentType string) {
			By("checking the normal output", func() {
				stdOut := helper.Cmd("odo", "describe", "--context", commonVar.Context).ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring(componentType))
			})
			By("checking the json output", func() {
				stdOut := helper.Cmd("odo", "describe", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
				Expect(gjson.Get(stdOut, "spec.type").String()).To(Equal(componentType))
			})
		}

		When("projectType is missing", func() {
			BeforeEach(func() {
				helper.CopyAndCreate(filepath.Join("source", "devfiles", "springboot", "project"), filepath.Join("source", "devfiles", "springboot", "devfile-with-missing-projectType-metadata.yaml"), commonVar.Context)
				metadata = helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
			})

			It("should show the language for 'Type' in odo describe", func() {
				checkDescribe(metadata.Language)
			})
			When("the component is pushed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
				})
				It("should show the language for 'Type' in odo describe", func() {
					checkDescribe(metadata.Language)
				})
			})

		})
		When("projectType and language is missing", func() {
			BeforeEach(func() {
				helper.CopyAndCreate(filepath.Join("source", "devfiles", "springboot", "project"), filepath.Join("source", "devfiles", "springboot", "devfile-with-missing-projectType-and-language-metadata.yaml"), commonVar.Context)
				metadata = helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
			})
			It("should show 'Not available' for 'Type' in odo describe", func() {
				checkDescribe(component.NotAvailable)
			})
			When("the component is pushed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
				})
				It("should show 'Not available' for 'Type' in odo describe", func() {
					checkDescribe(component.NotAvailable)
				})
			})
		})
	})
})
