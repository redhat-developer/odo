package devfile

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo remove binding command tests", func() {
	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		if helper.IsKubernetesCluster() {
			Skip("Operators have not been setup on Kubernetes cluster yet. Remove this once the issue has been fixed.")
		}
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		// Ensure that the operators are installed
		commonVar.CliRunner.EnsureOperatorIsInstalled("service-binding-operator")
		commonVar.CliRunner.EnsureOperatorIsInstalled("cloud-native-postgresql")
		Eventually(func() string {
			out, _ := commonVar.CliRunner.GetBindableKinds()
			return out
		}, 120, 3).Should(ContainSubstring("Cluster"))
		addBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
		Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})
	When("the component with binding is bootstrapped", func() {
		var bindingName string
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", "mynode", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml"), "--starter", "nodejs-starter").ShouldPass()
			bindingName = fmt.Sprintf("binding-%s", helper.RandString(4))
			helper.Cmd("odo", "add", "binding", "--name", bindingName, "--service", "cluster-sample").ShouldPass()
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
