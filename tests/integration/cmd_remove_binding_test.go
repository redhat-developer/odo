package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo remove binding command tests", func() {
	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
		helper.Chdir(commonVar.Context)
		// Note: We do not add any operators here because `odo remove binding` is simply about removing the ServiceBinding from devfile.
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	When("the component with binding is bootstrapped", func() {
		var bindingName = "my-nodejs-app-cluster-sample" // Hard coded from the devfile-with-service-binding-files.yaml
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", "mynode", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml")).ShouldPass()
		})

		When("removing the binding", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "remove", "binding", "--name", bindingName).ShouldPass()
			})
			It("should successfully remove binding between component and service in the devfile", func() {
				components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
				Expect(components).To(BeNil())
			})
		})
		It("should fail to remove binding that does not exist", func() {
			helper.Cmd("odo", "remove", "binding", "--name", "my-binding").ShouldFail()
		})
		When("odo dev is running", func() {
			var session helper.DevSession
			BeforeEach(func() {
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})
			When("binding is removed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "remove", "binding", "--name", bindingName).ShouldPass()
					_, _, _, err := session.WaitSync()
					Expect(err).ToNot(HaveOccurred())
				})
				It("should have led odo dev to delete ServiceBinding from the cluster", func() {
					_, errOut := commonVar.CliRunner.GetServiceBinding(bindingName, commonVar.Project)
					Expect(errOut).To(ContainSubstring("not found"))
				})
			})
		})
	})
})
