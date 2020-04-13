package integration

import (
	"os"
	"regexp"
	"strings"

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

	Context("When creating an operator backed service", func() {
		It("should be able to create EtcdCluster from its alm example", func() {
			// First let's grab the etcd operator's name from "odo catalog list services" output
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]`).FindString(operators)

			helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster")

			pods := helper.CmdShouldPass("oc", "get", "pods", "-n", CI_OPERATOR_HUB_PROJECT)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", CI_OPERATOR_HUB_PROJECT}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			// Delete the pods created. This should idealy be done by `odo
			// service delete` but that's implemented for operator backed
			// services yet.
			helper.CmdShouldPass("oc", "delete", "EtcdCluster", "example")
		})
	})

})
