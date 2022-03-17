package devfile

import (
	"fmt"
	"path/filepath"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/tidwall/gjson"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo list with devfile", func() {
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("a component created in 'app' application", func() {

		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CreateLocalEnv(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should show the component as 'Not Pushed'", func() {
			output := helper.Cmd("odo", "list").ShouldPass().Out()
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sNotPushed", "app", cmpName, commonVar.Project, "nodejs"))))
		})

		It("should show the component as 'Not Pushed' in JSON output", func() {
			output := helper.Cmd("odo", "list", "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "devfileComponents.0.kind", "devfileComponents.0.metadata.name", "devfileComponents.0.status.state")
			expected := []string{"List", "Component", cmpName, "Not Pushed"}
			Expect(helper.GjsonExactMatcher(values, expected)).To(Equal(true))
		})
	})
	Context("devfile has missing metadata", func() {
		// Note: We will be using SpringBoot example here because it helps to distinguish between language and projectType.
		// In terms of SpringBoot, spring is the projectType and java is the language; see https://github.com/redhat-developer/odo/issues/4815
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
		})
		var metadata devfilepkg.DevfileMetadata

		// checkList checks the list output (both normal and json) to see if it contains the expected componentType
		var checkList = func(componentType string) {
			By("checking the normal output", func() {
				stdOut := helper.Cmd("odo", "list", "--context", commonVar.Context).ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring(componentType))
			})
			By("checking the json output", func() {
				stdOut := helper.Cmd("odo", "list", "--context", commonVar.Context, "-o", "json").ShouldPass().Out()
				Expect(gjson.Get(stdOut, "devfileComponents.0.spec.type").String()).To(Equal(componentType))
			})
		}

		When("projectType is missing", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-with-missing-projectType-metadata.yaml")).ShouldPass()
				helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				metadata = helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
			})

			It("should show the language for 'Type' in odo list", func() {
				checkList(metadata.Language)
			})
			When("the component is pushed in dev mode", func() {
				var devSession helper.DevSession
				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode()
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					devSession.Stop()
				})

				It("should show the language for 'Type' in odo list", func() {
					checkList(metadata.Language)
				})
			})
		})

		When("projectType and language is missing", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", "aname", "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-with-missing-projectType-and-language-metadata.yaml")).ShouldPass()
				helper.CreateLocalEnv(commonVar.Context, "aname", commonVar.Project)
				metadata = helper.GetMetadataFromDevfile(filepath.Join(commonVar.Context, "devfile.yaml"))
			})
			It("should show 'Not available' for 'Type' in odo list", func() {
				checkList(component.NotAvailable)
			})
			When("the component is pushed", func() {
				var devSession helper.DevSession
				BeforeEach(func() {
					var err error
					devSession, _, _, err = helper.StartDevMode()
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					devSession.Stop()
				})
				It("should show 'Not available' for 'Type' in odo list", func() {
					checkList(component.NotAvailable)
				})
			})
		})
	})
})
