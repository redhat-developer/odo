package devfile

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"

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
		helper.SetDefaultDevfileRegistryAsStaging()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("a component created in 'app' application", func() {

		BeforeEach(func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

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

		When("the first component is pushed and a second component is created in different application", func() {
			var context2 string
			var cmpName2 string
			var appName string

			BeforeEach(func() {
				output := helper.Cmd("odo", "push").ShouldPass().Out()
				Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

				context2 = helper.CreateNewContext()
				cmpName2 = helper.RandString(6)
				appName = helper.RandString(6)

				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, "--app", appName, "--context", context2, cmpName2).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context2)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context2, "devfile.yaml"))
			})

			AfterEach(func() {
				helper.DeleteDir(context2)
			})

			It("should show the second component as 'NotPushed'", func() {
				output := helper.Cmd("odo", "list", "--context", context2).ShouldPass().Out()
				Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sNotPushed", appName, cmpName2, commonVar.Project, "nodejs"))))
			})

			It("should show the second component as 'Not Pushed' in JSON output", func() {
				output := helper.Cmd("odo", "list", "-o", "json", "--context", context2).ShouldPass().Out()
				values := gjson.GetMany(output, "kind", "devfileComponents.0.kind", "devfileComponents.0.metadata.name", "devfileComponents.0.status.state")
				expected := []string{"List", "Component", cmpName2, "Not Pushed"}
				Expect(helper.GjsonExactMatcher(values, expected)).To(Equal(true))
			})

			When("second component is pushed", func() {
				BeforeEach(func() {
					output2 := helper.Cmd("odo", "push", "--context", context2).ShouldPass().Out()
					Expect(output2).To(ContainSubstring("Changes successfully pushed to component"))
				})

				It("should show components in the current application in 'Pushed' state", func() {
					output := helper.Cmd("odo", "list", "--project", commonVar.Project).ShouldPass().Out()
					// this test makes sure that a devfile component doesn't show up as an s2i component as well
					Expect(helper.Suffocate(output)).To(Equal(helper.Suffocate(fmt.Sprintf(`
					APP        NAME       PROJECT        TYPE       STATE        MANAGED BY ODO
					app        %[1]s     %[2]s           nodejs     Pushed                Yes
					`, cmpName, commonVar.Project))))
				})

				It("should show components in the current application in 'Pushed' state in JSON output", func() {
					output := helper.Cmd("odo", "list", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
					values := gjson.GetMany(output, "kind", "devfileComponents.0.kind", "devfileComponents.0.metadata.name", "devfileComponents.0.status.state")
					expected := []string{"List", "Component", cmpName, "Pushed"}
					Expect(helper.GjsonExactMatcher(values, expected)).To(Equal(true))
				})

				It("should show components in all applications", func() {
					output := helper.Cmd("odo", "list", "--all-apps", "--project", commonVar.Project).ShouldPass().Out()
					Expect(output).To(ContainSubstring(cmpName))
					Expect(output).To(ContainSubstring(cmpName2))
				})
			})
		})
	})
})
