package integration

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

const (
	CI_OPERATOR_HUB_PROJECT = "ci-operator-hub-project"
)

var _ = Describe("odo service command tests for OperatorHub", func() {

	BeforeEach(func() {
		helper.CmdShouldPass("odo", "project", "set", CI_OPERATOR_HUB_PROJECT)
		// TODO: remove this when OperatorHub integration is fully baked into odo
		os.Setenv("ODO_EXPERIMENTAL", "true")
	})
	Context("When experimental mode is enabled", func() {
		It("should list operators installed in the namespace", func() {
			stdOut := helper.CmdShouldPass("odo", "catalog", "list", "services")
			Expect(stdOut).To(ContainSubstring("Operators available in the cluster"))
			Expect(stdOut).To(ContainSubstring("mongodb-enterprise"))
			Expect(stdOut).To(ContainSubstring("etcdoperator"))
		})
	})
})
