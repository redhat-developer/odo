package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo service command tests for OperatorHub", func() {

	var commonVar helper.CommonVar
	var oc helper.OcRunner

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		oc = helper.NewOcRunner("oc")
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	preSetup := func() {
		// wait till oc can see the all operators installed by setup script in the namespace
		odoArgs := []string{"catalog", "list", "services"}
		operators := []string{"etcdoperator", "service-binding-operator"}
		for _, operator := range operators {
			helper.WaitForCmdOut("odo", odoArgs, 5, true, func(output string) bool {
				return strings.Contains(output, operator)
			})
		}
	}

	cleanPreSetup := func() {
		helper.DeleteProject(commonVar.Project)
	}

	Context("When Operators are installed in the cluster", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should list operators installed in the namespace", func() {
			stdOut := helper.CmdShouldPass("odo", "catalog", "list", "services")
			helper.MatchAllInOutput(stdOut, []string{"Services available through Operators", "etcdoperator"})
		})

		It("should describe an installed operator with json output", func() {
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			etcdCluster := fmt.Sprintf("%s/EtcdCluster", etcdOperator)

			output := helper.CmdShouldPass("odo", "catalog", "describe", "service", etcdCluster)
			Expect(output).To(ContainSubstring("Kind: EtcdCluster"))
			outputJSON := helper.CmdShouldPass("odo", "catalog", "describe", "service", etcdCluster, "-o", "json")
			values := gjson.GetMany(outputJSON, "spec.kind", "spec.displayName")
			expected := []string{"EtcdCluster", "etcd Cluster"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})

		It("should not allow creating service without valid context, and fail for interactive mode", func() {
			stdOut := helper.CmdShouldFail("odo", "service", "create")
			Expect(stdOut).To(ContainSubstring("service can be created/deleted from a valid component directory only"))

			helper.CmdShouldPass("odo", "create", "nodejs")
			stdOut = helper.CmdShouldFail("odo", "service", "create")
			Expect(stdOut).To(ContainSubstring("odo doesn't support interactive mode for creating Operator backed service"))
		})

		It("should successfully push a second service after a first service is deployed", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)

			// create a first service
			stdOut := helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "myetcd1", "--project", commonVar.Project)
			Expect(stdOut).To(ContainSubstring("Successfully added service to the configuration"))

			// deploy it
			helper.CmdShouldPass("odo", "push")
			stdOut = helper.CmdShouldPass("odo", "service", "list")
			Expect(stdOut).To(ContainSubstring("EtcdCluster/myetcd1"))

			// create a second service
			stdOut = helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "myetcd2", "--project", commonVar.Project)
			Expect(stdOut).To(ContainSubstring("Successfully added service to the configuration"))

			// deploy it
			helper.CmdShouldPass("odo", "push")
			stdOut = helper.CmdShouldPass("odo", "service", "list")
			// first service still here
			Expect(stdOut).To(ContainSubstring("EtcdCluster/myetcd1"))
			// second service created
			Expect(stdOut).To(ContainSubstring("EtcdCluster/myetcd2"))
		})
	})

	Context("When creating and deleting an operator backed service", func() {

		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should be able to create, list and then delete EtcdCluster from its alm example", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			stdOut := helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "--project", commonVar.Project)
			Expect(stdOut).To(ContainSubstring("Successfully added service to the configuration"))

			// read the devfile.yaml to check if service definition was rightly inserted
			devfilePath := filepath.Join(commonVar.Context, "devfile.yaml")
			content, err := ioutil.ReadFile(devfilePath)
			Expect(err).To(BeNil())
			matchInOutput := []string{"kubernetes", "inlined", "EtcdCluster", "example"}
			helper.MatchAllInOutput(string(content), matchInOutput)

			// now create the service on cluster and verify if the pods for the operator have started
			helper.CmdShouldPass("odo", "push")
			pods := oc.GetAllPodsInNs(commonVar.Project)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			// now test listing of the service using odo
			stdOut = helper.CmdShouldPass("odo", "service", "list")
			Expect(stdOut).To(ContainSubstring("EtcdCluster/example"))

			// now test the deletion of the service using odo
			helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/example", "-f")

			// read the devfile.yaml to check if service definition was deleted
			content, err = ioutil.ReadFile(devfilePath)
			Expect(err).To(BeNil())
			helper.DontMatchAllInOutput(string(content), matchInOutput)

			// now try deleting the same service again. It should fail with error message
			stdOut = helper.CmdShouldFail("odo", "service", "delete", "EtcdCluster/example", "-f")
			Expect(stdOut).To(ContainSubstring("couldn't find service named"))
		})

		It("should be able to create service with name passed on CLI", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
			name := helper.RandString(6)
			svcFullName := strings.Join([]string{"EtcdCluster", name}, "/")
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), name, "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "push")

			// now verify if the pods for the operator have started
			pods := oc.GetAllPodsInNs(commonVar.Project)
			// Look for pod with custom name because that's the name etcd will give to the pods.
			compileString := name + `-.[a-z0-9]*`
			etcdPod := regexp.MustCompile(compileString).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			// now try creating service with same name again. it should fail
			stdOut := helper.CmdShouldFail("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), name, "--project", commonVar.Project)
			Expect(stdOut).To(ContainSubstring(fmt.Sprintf("service %q already exists", svcFullName)))

			helper.CmdShouldPass("odo", "service", "delete", svcFullName, "-f")
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
			helper.CmdShouldPass("odo", "create", "nodejs")
			// First let's grab the etcd operator's name from "odo catalog list services" output
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)

			stdOut := helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "--dry-run", "--project", commonVar.Project)
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
			helper.CmdShouldPass("odo", "create", "nodejs")

			// First let's grab the etcd operator's name from "odo catalog list services" output
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)

			stdOut := helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "--dry-run", "--project", commonVar.Project)

			// stdOut contains the yaml specification. Store it to a file
			randomFileName := helper.RandString(6) + ".yaml"
			fileName := filepath.Join(os.TempDir(), randomFileName)
			if err := ioutil.WriteFile(fileName, []byte(stdOut), 0644); err != nil {
				fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
			}

			// now create operator backed service
			helper.CmdShouldPass("odo", "service", "create", "--from-file", fileName, "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "push")

			// now verify if the pods for the operator have started
			pods := oc.GetAllPodsInNs(commonVar.Project)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/example", "-f")
		})

		It("should be able to create a service with name passed on CLI", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")

			name := helper.RandString(6)
			// First let's grab the etcd operator's name from "odo catalog list services" output
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)

			stdOut := helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "--dry-run", "--project", commonVar.Project)

			// stdOut contains the yaml specification. Store it to a file
			randomFileName := helper.RandString(6) + ".yaml"
			fileName := filepath.Join(os.TempDir(), randomFileName)
			if err := ioutil.WriteFile(fileName, []byte(stdOut), 0644); err != nil {
				fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
			}

			// now create operator backed service
			helper.CmdShouldPass("odo", "service", "create", "--from-file", fileName, name, "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "push")

			// Attempting to create service with same name should fail
			stdOut = helper.CmdShouldFail("odo", "service", "create", "--from-file", fileName, name, "--project", commonVar.Project)
			Expect(stdOut).To(ContainSubstring("please provide a different name or delete the existing service first"))
		})
	})

	Context("When using from-file option", func() {

		var tmpContext string

		JustBeforeEach(func() {
			tmpContext = helper.CreateNewContext()
			preSetup()
		})

		JustAfterEach(func() {
			helper.DeleteDir(tmpContext)
			cleanPreSetup()
		})

		It("should fail to create service if metadata doesn't exist or is invalid", func() {
			helper.CmdShouldPass("odo", "create", "nodejs")
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
			fileName := filepath.Join(tmpContext, noMetaFile)
			if err := ioutil.WriteFile(fileName, []byte(noMetadata), 0644); err != nil {
				fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
			}

			// now create operator backed service
			stdOut := helper.CmdShouldFail("odo", "service", "create", "--from-file", fileName, "--project", commonVar.Project)
			Expect(stdOut).To(ContainSubstring("couldn't find \"metadata\" in the yaml"))

			invalidMetaFile := helper.RandString(6) + ".yaml"
			fileName = filepath.Join(tmpContext, invalidMetaFile)
			if err := ioutil.WriteFile(fileName, []byte(invalidMetadata), 0644); err != nil {
				fmt.Printf("Could not write yaml spec to file %s because of the error %v", fileName, err.Error())
			}

			// now create operator backed service
			stdOut = helper.CmdShouldFail("odo", "service", "create", "--from-file", fileName, "--project", commonVar.Project)
			Expect(stdOut).To(ContainSubstring("couldn't find metadata.name in the yaml"))

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
			helper.MatchAllInOutput(jsonOut, []string{"etcdoperator"})
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
			helper.CmdShouldPass("odo", "create", "nodejs")

			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "push")

			// now verify if the pods for the operator have started
			pods := oc.GetAllPodsInNs(commonVar.Project)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", commonVar.Project}
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

			msg := fmt.Sprintf("no operator backed services found in namespace: %s", commonVar.Project)
			msgWithQuote := fmt.Sprintf("\"message\": \"no operator backed services found in namespace: %s\"", commonVar.Project)
			Expect(stdOut).To(ContainSubstring(msg))
			helper.MatchAllInOutput(jsonOut, []string{msg, msgWithQuote})
		})
	})

	Context("When linking devfile component with Operator backed service", func() {
		JustBeforeEach(func() {
			preSetup()
		})

		JustAfterEach(func() {
			cleanPreSetup()
		})

		It("should fail if service name doesn't adhere to <service-type>/<service-name> format", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", componentName)

			stdOut := helper.CmdShouldFail("odo", "link", "EtcdCluster")
			Expect(stdOut).To(ContainSubstring("invalid service name"))

			stdOut = helper.CmdShouldFail("odo", "link", "EtcdCluster/")
			Expect(stdOut).To(ContainSubstring("invalid service name"))

			stdOut = helper.CmdShouldFail("odo", "link", "/example")
			Expect(stdOut).To(ContainSubstring("invalid service name"))
		})

		It("should fail if the provided service doesn't exist in the namespace", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", componentName)
			helper.CmdShouldPass("odo", "push")

			stdOut := helper.CmdShouldFail("odo", "link", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("Couldn't find service named %q", "EtcdCluster/example"))
		})

		It("should successfully connect and disconnect a component with an existing service", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", componentName)
			helper.CmdShouldPass("odo", "push")

			// start the Operator backed service first
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "--project", commonVar.Project)
			helper.CmdShouldPass("odo", "push")

			// now verify if the pods for the operator have started
			pods := oc.GetAllPodsInNs(commonVar.Project)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			stdOut := helper.CmdShouldPass("odo", "link", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("Successfully created link between component"))
			helper.CmdShouldPass("odo", "push")
			stdOut = helper.CmdShouldFail("odo", "link", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("already linked with the service"))

			// Before running "odo unlink" checks, wait for the pod to come up from "odo push" done after "odo link"
			pods = oc.GetAllPodsInNs(commonVar.Project)
			componentPod := regexp.MustCompile(fmt.Sprintf(`%s-.[a-z0-9]*-.[a-z0-9\-]*`, componentName)).FindString(pods)
			ocArgs = []string{"get", "pods", componentPod, "-o", "template=\"{{.status.phase}}\"", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			stdOut = helper.CmdShouldPass("odo", "unlink", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("Successfully unlinked component"))
			helper.CmdShouldPass("odo", "push")

			// verify that sbr is deleted
			stdOut = helper.CmdShouldFail("odo", "unlink", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("failed to unlink the service"))
		})

		It("should successfully link the component to the etcd service with a specific name", func() {

			// create a component and deploy it
			helper.CmdShouldPass("odo", "create", "nodejs")

			// create an etcd service
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "etcd")

			// deploy the component and service
			helper.CmdShouldPass("odo", "push")

			// create the link with a specific name
			helper.CmdShouldPass("odo", "link", "EtcdCluster/etcd", "--name", "etcd-config")

			// for the moment, odo push is not necessary

			// check the link exists with the specific name
			ocArgs := []string{"get", "servicebinding", "etcd-config", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "etcd-config")
			})

			// delete the link
			helper.CmdShouldPass("odo", "unlink", "EtcdCluster/etcd")

			// delete the etcd service
			helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/etcd", "-f")
		})

		It("should successfully link the component to the etcd service with a specific name and activating bindAsFiles", func() {

			// create a component and deploy it
			helper.CmdShouldPass("odo", "create", "nodejs")

			// create an etcd service
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", fmt.Sprintf("%s/EtcdCluster", etcdOperator), "etcd")

			// deploy the component and service
			helper.CmdShouldPass("odo", "push")

			// create the link with a specific name and with bind-as-files
			helper.CmdShouldPass("odo", "link", "EtcdCluster/etcd", "--name", "etcd-config", "--bind-as-files")

			// for the moment, odo push is not necessary

			// check the link exists with the specific name
			ocArgs := []string{"get", "servicebinding", "etcd-config", "-o", "jsonpath='{.spec.bindAsFiles}'", "-n", commonVar.Project}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "true")
			})

			// delete the link
			helper.CmdShouldPass("odo", "unlink", "EtcdCluster/etcd")

			// delete the etcd service
			helper.CmdShouldPass("odo", "service", "delete", "EtcdCluster/etcd", "-f")
		})
	})
})
