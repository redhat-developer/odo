//go:build linux || darwin || dragonfly || solaris || openbsd || netbsd || freebsd
// +build linux darwin dragonfly solaris openbsd netbsd freebsd

package interactive

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo init interactive command tests", func() {

	var commonVar helper.CommonVar
	var serviceName string

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)

		// We make EXPLICITLY sure that we are outputting with NO COLOR
		// this is because in some cases we are comparing the output with a colorized one
		os.Setenv("NO_COLOR", "true")

		// Ensure that the operators are installed
		commonVar.CliRunner.EnsureOperatorIsInstalled("service-binding-operator")
		commonVar.CliRunner.EnsureOperatorIsInstalled("cloud-native-postgresql")
		Eventually(func() string {
			out, _ := commonVar.CliRunner.GetBindableKinds()
			return out
		}, 120, 3).Should(ContainSubstring("Cluster"))
		addBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
		Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
		serviceName = "cluster-sample" // Hard coded from bindablekind-instance.yaml
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("the component is bootstrapped", func() {
		var componentName = "mynode"
		var bindingName string
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", componentName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml"), "--starter", "nodejs-starter").ShouldPass()
			bindingName = fmt.Sprintf("%s-%s", componentName, serviceName)
		})

		It("should successsfully add binding to the devfile (Bind as Environment Variables)", func() {
			command := []string{"odo", "add", "binding"}

			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "Select service instance you want to bind to:")
				helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

				helper.ExpectString(ctx, "Enter the Binding's name")
				helper.SendLine(ctx, "\n")

				helper.ExpectString(ctx, "How do you want to bind the service?")
				helper.SendLine(ctx, "Bind as Environment Variables")

				helper.ExpectString(ctx, "Successfully added the binding to the devfile.")

				helper.ExpectString(ctx, fmt.Sprintf("odo add binding --service cluster-sample.Cluster.postgresql.k8s.enterprisedb.io --name %s --bind-as-files=false", bindingName))
			})

			Expect(err).To(BeNil())
			components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
			Expect(components).ToNot(BeNil())
		})

		It("should successsfully add binding to the devfile (Bind as Files)", func() {
			command := []string{"odo", "add", "binding"}

			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "Select service instance you want to bind to:")
				helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

				helper.ExpectString(ctx, "Enter the Binding's name")
				helper.SendLine(ctx, "\n")

				helper.ExpectString(ctx, "How do you want to bind the service?")
				helper.SendLine(ctx, "Bind as Files")

				helper.ExpectString(ctx, "Successfully added the binding to the devfile.")
			})

			Expect(err).To(BeNil())
			components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
			Expect(components).ToNot(BeNil())
		})
	})
})
