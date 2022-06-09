package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo remove binding command tests", func() {
	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
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
			helper.Cmd("odo", "init", "--name", "mynode", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml"), "--starter", "nodejs-starter").ShouldPass()
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
	})
})
