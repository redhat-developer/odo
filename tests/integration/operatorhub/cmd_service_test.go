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

	Context("When creating and deleting an operator backed service", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should be able to create and then delete EtcdCluster from its alm example", func() {
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

			// now test the deletion of the service using odo
			helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/example", "-f")

			// now try deleting the same service again. It should fail with error message
			stdOut := helper.CmdShouldFail("odo", "service", "delete", "EtcdCluster/example", "-f")
			Expect(stdOut).To(ContainSubstring("Couldn't find service named"))
		})
	})

	Context("When deleting an invalid operator backed service", func() {
		It("should correctly detect invalid service names", func() {
			names := []string{"EtcdCluster", "EtcdCluster/", "/example"}

			for _, name := range names {
				stdOut := helper.CmdShouldFail("odo", "service", "delete", name, "-f")
				Expect(stdOut).To(ContainSubstring("Invalid service name"))
			}
		})

		It("should be able to create service with name passed on CLI", func() {
			name := helper.RandString(6)
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster", name)

			// now verify if the pods for the operator have started
			pods := helper.CmdShouldPass("oc", "get", "pods", "-n", project)
			// Look for pod with custom name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(name).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			// Delete the pods created. This should idealy be done by `odo
			// service delete` but that's not implemented for operator backed
			// services yet.
			helper.CmdShouldPass("oc", "delete", "EtcdCluster", name)
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

	Context("Should be able to search from catalog", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should only output the definition of the CR that will be used to start service", func() {
			stdOut := helper.CmdShouldPass("odo", "catalog", "search", "service", "etcd")
			helper.MatchAllInOutput(stdOut, []string{"etcdoperator", "EtcdCluster"})

			stdOut = helper.CmdShouldPass("odo", "catalog", "search", "service", "EtcdCluster")
			helper.MatchAllInOutput(stdOut, []string{"etcdoperator", "EtcdCluster"})

			stdOut = helper.CmdShouldFail("odo", "catalog", "search", "service", "dummy")
			Expect(stdOut).To(ContainSubstring("no service matched the query: dummy"))
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

			helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/example", "-f")
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

			helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/example", "-f")

			// Now let's check the output again to ensure expected behaviour
			stdOut = helper.CmdShouldFail("odo", "service", "list")
			jsonOut = helper.CmdShouldFail("odo", "service", "list", "-o", "json")

			msg := fmt.Sprintf("No operator backed services found in namespace: %s", project)
			msgWithQuote := fmt.Sprintf("\"message\": \"No operator backed services found in namespace: %s\"", project)
			Expect(stdOut).To(ContainSubstring(msg))
			helper.MatchAllInOutput(jsonOut, []string{msg, msgWithQuote})
		})
	})

	Context("When linking devfile component with Operator backed service", func() {
		var context, currentWorkingDirectory, devfilePath string
		const devfile = "devfile.yaml"

		JustBeforeEach(func() {
			preSetup()
			context = helper.CreateNewContext()
			devfilePath = filepath.Join(context, devfile)
			currentWorkingDirectory = helper.Getwd()
			helper.Chdir(context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", devfile), devfilePath)
		})

		JustAfterEach(func() {
			cleanPreSetup()
			helper.Chdir(currentWorkingDirectory)
			helper.DeleteDir(context)
		})

		It("should fail if service name doesn't adhere to <service-type>/<service-name> format", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", componentName)

			stdOut := helper.CmdShouldFail("odo", "link", "EtcdCluster")
			Expect(stdOut).To(ContainSubstring("Invalid service name"))

			stdOut = helper.CmdShouldFail("odo", "link", "EtcdCluster/")
			Expect(stdOut).To(ContainSubstring("Invalid service name"))

			stdOut = helper.CmdShouldFail("odo", "link", "/example")
			Expect(stdOut).To(ContainSubstring("Invalid service name"))
		})

		It("should fail if the provided service doesn't exist in the namespace", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", componentName)
			helper.CmdShouldPass("odo", "push")

			stdOut := helper.CmdShouldFail("odo", "link", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("Couldn't find service named %q", "EtcdCluster/example"))
		})

		It("should successfully connect a component with an existing service", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", componentName)
			helper.CmdShouldPass("odo", "push")

			// start the Operator backed service first
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

			stdOut := helper.CmdShouldPass("odo", "link", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("Successfully created link between component"))
		})
	})
})
