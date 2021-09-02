package devfile

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/tests/helper"

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

	When("two components in different applications", func() {
		It("should correctly output component information", func() {

			// component created in "app" application
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			By("checking that component correctly shows as 'NotPushed'")
			output := helper.Cmd("odo", "list").ShouldPass().Out()
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sNotPushed", "app", cmpName, commonVar.Project, "nodejs"))))

			output = helper.Cmd("odo", "push").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// component created in different application
			context2 := helper.CreateNewContext()
			cmpName2 := helper.RandString(6)
			appName := helper.RandString(6)

			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, "--app", appName, "--context", context2, cmpName2).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context2)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context2, "devfile.yaml"))

			By("checking that second component correctly shows as 'NotPushed'")
			output = helper.Cmd("odo", "list", "--context", context2).ShouldPass().Out()
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sNotPushed", appName, cmpName2, commonVar.Project, "nodejs"))))

			output2 := helper.Cmd("odo", "push", "--context", context2).ShouldPass().Out()
			Expect(output2).To(ContainSubstring("Changes successfully pushed to component"))

			By("checking listing components in the current application in 'Pushed' state")
			output = helper.Cmd("odo", "list", "--project", commonVar.Project).ShouldPass().Out()
			// this test makes sure that a devfile component doesn't show up as an s2i component as well
			Expect(helper.Suffocate(output)).To(Equal(helper.Suffocate(fmt.Sprintf(`
			APP        NAME       PROJECT        TYPE       STATE        MANAGED BY ODO
			app        %[1]s     %[2]s           nodejs     Pushed		 Yes
			`, cmpName, commonVar.Project))))

			By("checking that it shows components in all applications")
			output = helper.Cmd("odo", "list", "--all-apps", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring(cmpName))
			Expect(output).To(ContainSubstring(cmpName2))

			helper.DeleteDir(context2)
		})

	})

})
