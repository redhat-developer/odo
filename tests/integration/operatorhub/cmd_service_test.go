package integration

import (
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
	})
	Context("When experimental mode is enabled", func() {
		It("should list list operators installed in the namespace", func() {
			stdOut := helper.CmdShouldPass("ODO_EXPERIMENTAL=true", "odo", "catalog", "list", "services")
			Expect(stdOut).To(ContainSubstring("Operators available through Operator Hub"))
			Expect(stdOut).To(ContainSubstring("mongodb-enterprise"))
			Expect(stdOut).To(ContainSubstring("etcdoperator"))
		})
	})
})
