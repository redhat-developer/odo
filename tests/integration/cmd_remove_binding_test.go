package integration

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo remove binding command tests", Label(helper.LabelServiceBinding), func() {
	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		skipLogin := os.Getenv("SKIP_SERVICE_BINDING_TESTS")
		if skipLogin == "true" {
			Skip("Skipping service binding tests as SKIP_SERVICE_BINDING_TESTS is true")
		}

		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		// Note: We do not add any operators here because `odo remove binding` is simply about removing the ServiceBinding from devfile.
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	for _, bindingName := range []string{"my-nodejs-app-cluster-sample-k8s", "my-nodejs-app-cluster-sample-ocp"} {
		bindingName := bindingName
		When(fmt.Sprintf("the component with binding is bootstrapped (bindingName=%s)", bindingName), func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.Cmd("odo", "init", "--name", "mynode", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml")).ShouldPass()
			})

			for _, label := range []string{
				helper.LabelNoCluster, helper.LabelUnauth,
			} {
				label := label
				When("removing the binding", Label(label), func() {
					BeforeEach(func() {
						helper.Cmd("odo", "remove", "binding", "--name", bindingName).ShouldPass()
					})
					It("should successfully remove binding between component and service in the devfile", func() {
						components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
						Expect(components).To(BeNil())
					})
				})
			}

			It("should fail to remove binding that does not exist", func() {
				helper.Cmd("odo", "remove", "binding", "--name", "my-binding").ShouldFail()
			})

			When("odo dev is running", func() {
				var devSession helper.DevSession
				BeforeEach(func() {
					var err error
					devSession, err = helper.StartDevMode(helper.DevSessionOpts{})
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})
				When("binding is removed", func() {
					BeforeEach(func() {
						helper.Cmd("odo", "remove", "binding", "--name", bindingName).ShouldPass()
						err := devSession.WaitSync()
						Expect(err).ToNot(HaveOccurred())
					})
					It("should have led odo dev to delete ServiceBinding from the cluster", func() {
						_, errOut := commonVar.CliRunner.GetServiceBinding(bindingName, commonVar.Project)
						Expect(errOut).To(ContainSubstring("not found"))
					})
				})
			})
		})
	}
})
