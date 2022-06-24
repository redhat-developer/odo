package devfile

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo add binding command tests", func() {
	var commonVar helper.CommonVar
	var devSession helper.DevSession
	var err error

	var _ = BeforeEach(func() {
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

	It("should fail creating a binding without workload parameter", func() {
		stderr := helper.Cmd("odo", "add", "binding", "--name", "aname", "--service", "cluster-sample").ShouldFail().Err()
		Expect(stderr).To(ContainSubstring("missing --workload parameter"))
	})

	It("should create a binding using the workload parameter", func() {
		helper.Cmd("odo", "add", "binding", "--name", "aname", "--service", "cluster-sample", "--workload", "app/Deployment.apps").ShouldPass()
	})

	When("the component is bootstrapped", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", "mynode", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml"), "--starter", "nodejs-starter").ShouldPass()
		})

		It("should fail using the --workload parameter", func() {
			stderr := helper.Cmd("odo", "add", "binding", "--name", "aname", "--service", "cluster-sample", "--workload", "app/Deployment.apps").ShouldFail().Err()
			Expect(stderr).To(ContainSubstring("--workload cannot be used from a directory containing a Devfile"))
		})

		When("adding a binding", func() {
			var bindingName string
			BeforeEach(func() {
				bindingName = fmt.Sprintf("binding-%s", helper.RandString(4))
				helper.Cmd("odo", "add", "binding", "--name", bindingName, "--service", "cluster-sample").ShouldPass()
			})
			It("should successfully add binding between component and service in the devfile", func() {
				components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
				Expect(components).ToNot(BeNil())
			})
			When("odo dev is run", func() {
				BeforeEach(func() {
					devSession, _, _, _, err = helper.StartDevMode()
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})
				It("should successfully bind component and service", func() {
					stdout := commonVar.CliRunner.Run("get", "servicebinding", bindingName).Out.Contents()
					Expect(stdout).To(ContainSubstring("ApplicationsBound"))
				})
				When("odo dev command is stopped", func() {
					BeforeEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})

					It("should have successfully delete the binding", func() {
						_, errOut := commonVar.CliRunner.GetServiceBinding(bindingName, commonVar.Project)
						Expect(errOut).To(ContainSubstring("not found"))
					})
				})
			})
		})

		When("no bindable instance is present on the cluster", func() {
			BeforeEach(func() {
				deleteBindableKind := commonVar.CliRunner.Run("delete", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
				Expect(deleteBindableKind.ExitCode()).To(BeEquivalentTo(0))
			})
			It("should fail to add binding with no bindable instance found error message", func() {
				errOut := helper.Cmd("odo", "add", "binding", "--name", "my-binding", "--service", "cluster-sample").ShouldFail().Err()
				Expect(errOut).To(ContainSubstring("No bindable service instances found"))
			})
		})
	})
})
