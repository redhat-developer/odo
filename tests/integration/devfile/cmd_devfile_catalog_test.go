package devfile

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo devfile catalog command tests", func() {
	const registryName string = "RegistryName"
	// Use staging OCI-based registry for tests to avoid overload
	const addRegistryURL string = "https://registry.stage.devfile.io"

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)

		// For some reason on TravisCI, there are flakes with regards to registrycachetime and doing
		// odo catalog list components.
		// TODO: Investigate this more.
		helper.Cmd("odo", "preference", "set", "registrycachetime", "0").ShouldPass()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("When executing catalog list components", func() {
		It("should list all supported devfile components", func() {
			output := helper.Cmd("odo", "catalog", "list", "components").ShouldPass().Out()
			wantOutput := []string{
				"Odo Devfile Components",
				"NAME",
				"java-springboot",
				"java-openliberty",
				"java-quarkus",
				"DESCRIPTION",
				"REGISTRY",
				"DefaultDevfileRegistry",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
		It("should list components successfully even with an invalid kubeconfig path or path points to existing directory", func() {
			originalKC := os.Getenv("KUBECONFIG")
			err := os.Setenv("KUBECONFIG", "/idonotexist")
			Expect(err).ToNot(HaveOccurred())
			helper.Cmd("odo", "catalog", "list", "components").ShouldPass()
			err = os.Setenv("KUBECONFIG", commonVar.Context)
			Expect(err).ToNot(HaveOccurred())
			helper.Cmd("odo", "catalog", "list", "components").ShouldPass()
			err = os.Setenv("KUBECONFIG", originalKC)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When executing catalog list components with -o json flag", func() {
		It("should list devfile components in json format", func() {
			output := helper.Cmd("odo", "catalog", "list", "components", "-o", "json").ShouldPass().Out()

			var outputData interface{}
			unmarshalErr := json.Unmarshal([]byte(output), &outputData)
			Expect(unmarshalErr).NotTo(HaveOccurred(), "Output is not a valid JSON")

			wantOutput := []string{
				"odo.dev/v1alpha1",
				"devfileItems",
				"java-openliberty",
				"java-springboot",
				"nodejs",
				"java-quarkus",
				"java-maven",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})
	})

	Context("When executing catalog describe component with -o json", func() {
		It("should display a valid JSON", func() {
			output := helper.Cmd("odo", "catalog", "describe", "component", "nodejs", "-o", "json").ShouldPass().Out()
			var outputData interface{}
			unmarshalErr := json.Unmarshal([]byte(output), &outputData)
			Expect(unmarshalErr).NotTo(HaveOccurred(), "Output is not a valid JSON")
		})
	})

	Context("When executing catalog list components with registry that is not set up properly", func() {
		It("should list components from valid registry", func() {
			helper.Cmd("odo", "registry", "add", "fake", "http://fake").ShouldPass()
			output := helper.Cmd("odo", "catalog", "list", "components").ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{
				"Odo Devfile Components",
				"java-springboot",
				"java-quarkus",
			})
			helper.Cmd("odo", "registry", "delete", "fake", "-f").ShouldPass()
		})
	})

	Context("When executing catalog describe component with a component name with multiple components", func() {
		It("should print multiple devfiles from different registries", func() {
			helper.Cmd("odo", "registry", "add", registryName, addRegistryURL).ShouldPass()
			output := helper.Cmd("odo", "catalog", "describe", "component", "nodejs").ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{"name: nodejs-starter", "Registry: " + registryName})
		})
	})

	Context("When checking catalog for installed services", func() {
		It("should succeed", func() {
			helper.Cmd("odo", "catalog", "list", "services").ShouldPass()
		})
	})
})
