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

var _ = Describe("odo service command tests for OperatorHub", func() {

	var project string

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		// TODO: remove this when OperatorHub integration is fully baked into odo
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
	})

	preSetup := func() {
		project = helper.CreateRandProject()
		helper.CmdShouldPass("odo", "project", "set", project)

		// wait till oc can see the all operators installed by setup script in the namespace
		ocArgs := []string{"get", "csv"}
		operators := []string{"etcd", "mongodb"}
		for _, operator := range operators {
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, operator)
			})
		}
	}

	cleanPreSetup := func() {
		helper.DeleteProject(project)
	}

	Context("When experimental mode is enabled", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should list operators installed in the namespace", func() {
			stdOut := helper.CmdShouldPass("odo", "catalog", "list", "services")
			helper.MatchAllInOutput(stdOut, []string{"Operators available in the cluster", "mongodb-enterprise", "etcdoperator"})
		})
	})

	Context("When creating an operator backed service", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should be able to create EtcdCluster from its alm example", func() {
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster")

			// now verify if the pods for the operator have started
			pods := helper.CmdShouldPass("oc", "get", "pods", "-n", project)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			// Delete the pods created. This should idealy be done by `odo
			// service delete` but that's not implemented for operator backed
			// services yet.
			helper.CmdShouldPass("oc", "delete", "EtcdCluster", "example")
		})
	})

	Context("When using dry-run option to create operator backed service", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should only output the definition of the CR that will be used to start service", func() {
			// First let's grab the etcd operator's name from "odo catalog list services" output
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)

			stdOut := helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster", "--dry-run")
			helper.MatchAllInOutput(stdOut, []string{"apiVersion", "kind"})
		})
	})

	Context("When using from-file option", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should be able to create a service", func() {
			// First let's grab the etcd operator's name from "odo catalog list services" output
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)

			stdOut := helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster", "--dry-run")

			// stdOut contains the yaml specification. Store it to a file
			randomFileName := helper.RandString(6) + ".yaml"
			fileName := filepath.Join(os.TempDir(), randomFileName)
			if err := ioutil.WriteFile(fileName, []byte(stdOut), 0644); err != nil {
				fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
			}

			// now create operator backed service
			helper.CmdShouldPass("odo", "service", "create", "--from-file", fileName)

			// now verify if the pods for the operator have started
			pods := helper.CmdShouldPass("oc", "get", "pods", "-n", project)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			// Delete the pods created. This should idealy be done by `odo
			// service delete` but that's not implemented for operator backed
			// services yet.
			helper.CmdShouldPass("oc", "delete", "EtcdCluster", "example")
		})
	})

	Context("When using from-file option", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
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

	Context("JSON output", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("listing catalog of services", func() {
			jsonOut := helper.CmdShouldPass("odo", "catalog", "list", "services", "-o", "json")
			helper.MatchAllInOutput(jsonOut, []string{"mongodb-enterprise", "etcdoperator"})
		})
	})

	Context("When operator backed services are created", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should list the services if they exist", func() {
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster")

			// now verify if the pods for the operator have started
			pods := helper.CmdShouldPass("oc", "get", "pods", "-n", project)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			stdOut := helper.CmdShouldPass("odo", "service", "list")
			helper.MatchAllInOutput(stdOut, []string{"example", "EtcdCluster"})

			// now check for json output
			jsonOut := helper.CmdShouldPass("odo", "service", "list", "-o", "json")
			helper.MatchAllInOutput(jsonOut, []string{"\"apiVersion\": \"etcd.database.coreos.com/v1beta2\"", "\"kind\": \"EtcdCluster\"", "\"name\": \"example\""})

			// Delete the pods created. This should idealy be done by `odo
			// service delete` but that's not implemented for operator backed
			// services yet.
			helper.CmdShouldPass("oc", "delete", "EtcdCluster", "example")

			// Now let's check the output again to ensure expected behaviour
			stdOut = helper.CmdShouldFail("odo", "service", "list")
			jsonOut = helper.CmdShouldFail("odo", "service", "list", "-o", "json")
			Expect(stdOut).To(ContainSubstring("No operator backed services found in the namesapce"))
			helper.MatchAllInOutput(jsonOut, []string{"No operator backed services found in the namesapce", "\"message\": \"No operator backed services found in the namesapce\""})
		})
	})
})
