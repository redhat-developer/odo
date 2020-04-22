package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
			// service delete` but that's not implemented for operator backed
			// services yet.
			// Wait for the pods of EtcdCluster service to get cleared off
			// cluster. This is to avoid CI flakes because in next test, we're
			// again doing similar regex check which picks up terminating pod
			// from this run instead of initiating pod from that run
			helper.CmdShouldRunWithTimeout(time.Duration(15)*time.Second, "oc", "delete", "--wait", "EtcdCluster", "example")
		})
	})

	Context("When using dry-run option to create operator backed service", func() {
		It("should only output the definition of the CR that will be used to start service", func() {
			// First let's grab the etcd operator's name from "odo catalog list services" output
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]`).FindString(operators)

			stdOut := helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster", "--dry-run")
			Expect(stdOut).To(ContainSubstring("apiVersion"))
			Expect(stdOut).To(ContainSubstring("kind"))
		})
	})

	Context("When using from-file option", func() {
		It("should be able to create a service", func() {
			// First let's grab the etcd operator's name from "odo catalog list services" output
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]`).FindString(operators)

			stdOut := helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster", "--dry-run")
			// stdOut contains the yaml specification. Store it to a file
			randomFileName := helper.RandString(6) + ".yaml"
			fileName := filepath.Join("/tmp", randomFileName)
			if err := ioutil.WriteFile(fileName, []byte(stdOut), 0644); err != nil {
				fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
			}

			// now create operator backed service
			helper.CmdShouldPass("odo", "service", "create", "--from-file", fileName)

			// now verify if the pods for the operator have started
			pods := helper.CmdShouldPass("oc", "get", "pods", "-n", CI_OPERATOR_HUB_PROJECT)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", CI_OPERATOR_HUB_PROJECT}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			// Delete the pods created. This should idealy be done by `odo
			// service delete` but that's not implemented for operator backed
			// services yet.
			// Wait for the pods of EtcdCluster service to get cleared off
			// cluster. This is to avoid CI flakes because in next test, we're
			// again doing similar regex check which picks up terminating pod
			// from this run instead of initiating pod from that run
			helper.CmdShouldRunWithTimeout(time.Duration(15)*time.Second, "oc", "delete", "--wait", "EtcdCluster", "example")

		})

		It("should fail to create service if metadata doesn't exist or is invalid", func() {
			noMetadata := `
apiVersion: etcd.database.coreos.com/v1beta2
kind: EtcdCluster
spec:
  size: 3
  version: 3.2.13
`

			invalidMetadata := `
apiVersion: etcd.database.coreos.com/v1beta2
kind: EtcdCluster
metadata:
  noname: noname
spec:
  size: 3
  version: 3.2.13
`

			noMetaFile := helper.RandString(6) + ".yaml"
			fileName := filepath.Join("/tmp", noMetaFile)
			if err := ioutil.WriteFile(fileName, []byte(noMetadata), 0644); err != nil {
				fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
			}

			// now create operator backed service
			stdOut := helper.CmdShouldFail("odo", "service", "create", "--from-file", fileName)
			Expect(stdOut).To(ContainSubstring("Couldn't find \"metadata\" in the yaml"))

			invalidMetaFile := helper.RandString(6) + ".yaml"
			fileName = filepath.Join("/tmp", invalidMetaFile)
			if err := ioutil.WriteFile(fileName, []byte(invalidMetadata), 0644); err != nil {
				fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
			}

			// now create operator backed service
			stdOut = helper.CmdShouldFail("odo", "service", "create", "--from-file", fileName)
			Expect(stdOut).To(ContainSubstring("Couldn't find metadata.name in the yaml"))

		})
	})
})
