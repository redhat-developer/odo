package integration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo list services tests", func() {
	var commonVar helper.CommonVar
	var randomProject string

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()

		// Ensure that the operators are installed
		commonVar.CliRunner.EnsureOperatorIsInstalled("service-binding-operator")
		commonVar.CliRunner.EnsureOperatorIsInstalled("cloud-native-postgresql")
		Eventually(func() string {
			out, _ := commonVar.CliRunner.GetBindableKinds()
			return out
		}, 120, 3).Should(ContainSubstring("Cluster"))
		addBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
		Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
		commonVar.CliRunner.EnsurePodIsUp(commonVar.Project, "cluster-sample-1")
		randomProject = helper.CreateRandProject()
		addBindableKind = commonVar.CliRunner.Run("apply", "-n", randomProject, "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
		commonVar.CliRunner.EnsurePodIsUp(randomProject, "cluster-sample-1")
		Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
		helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
		helper.Cmd("odo", "delete", "project", randomProject, "-f").ShouldPass()
	})

	It("should list bindable services in human readable format", func() {
		// from current namespace
		out := helper.Cmd("odo", "list", "services", "-n", commonVar.Project).ShouldPass().Out()
		helper.MatchAllInOutput(out, []string{"cluster-sample", commonVar.Project, "Listing bindable services from namespace"})

		// from all namespaces
		out = helper.Cmd("odo", "list", "services", "-A").ShouldPass().Out()
		helper.MatchAllInOutput(out, []string{"cluster-sample", commonVar.Project, randomProject, "Listing bindable services from all namespaces"})

		// fail if -A and -n flags are used together
		out = helper.Cmd("odo", "list", "services", "-A", "-n", commonVar.Project).ShouldFail().Err()
		Expect(out).To(ContainSubstring("cannot use --all-namespaces and --namespace flags together"))
	})

	It("should list bindable services in JSON format", func() {
		// from current namespace
		out := helper.Cmd("odo", "list", "services", "-o", "json", "-n", commonVar.Project).ShouldPass().Out()
		Expect(helper.IsJSON(out)).To(BeTrue())
		Expect(gjson.Get(out, "bindableServices.0.name").String()).To(ContainSubstring("cluster-sample"))
		Expect(gjson.Get(out, "bindableServices.0.namespace").String()).To(Equal(commonVar.Project))
		Expect(gjson.Get(out, "bindableServices.0.kind").String()).To(Equal("Cluster"))
		Expect(gjson.Get(out, "bindableServices.0.group").String()).To(Equal("postgresql.k8s.enterprisedb.io"))

		// from all namespaces
		out = helper.Cmd("odo", "list", "services", "-A", "-o", "json").ShouldPass().Out()
		Expect(helper.IsJSON(out)).To(BeTrue())
		helper.MatchAllInOutput(out, []string{"cluster-sample", commonVar.Project, randomProject, "Cluster", "postgresql.k8s.enterprisedb.io"})

		// fail if -A and -n flags are used together
		out = helper.Cmd("odo", "list", "services", "-o", "json", "-A", "-n", commonVar.Project).ShouldFail().Err()
		Expect(gjson.Get(out, "message").String()).To(Equal("cannot use --all-namespaces and --namespace flags together"))
	})
})
