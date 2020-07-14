package devfile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link command tests", func() {
	const devfile = "devfile.yaml"
	const envFile = ".odo/env/env.yaml"
	var namespace, context, currentWorkingDirectory string

	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		os.Unsetenv("GLOBALODOCONFIG")
		helper.DeleteDir(context)
	})

	Context("When linking devfile component with Operator backed service", func() {
		It("should fail if service name doesn't adhere to <service-type>/<service-name> format", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", componentName)

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
			helper.CmdShouldPass("odo", "create", "nodejs", componentName)

			stdOut := helper.CmdShouldFail("odo", "link", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("Couldn't find the requested service"))
		})

		It("should successfully connect a component with an existing service", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}

			componentName := helper.RandString(6)
			helper.CmdShouldPass("odo", "create", "nodejs", componentName)

			// start the Operator backed service first
			operators := helper.CmdShouldPass("odo", "catalog", "list", "services")
			etcdOperator := regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(operators)
			helper.CmdShouldPass("odo", "service", "create", etcdOperator, "--crd", "EtcdCluster")

			// now verify if the pods for the operator have started
			pods := helper.CmdShouldPass("oc", "get", "pods", "-n", namespace)
			// Look for pod with example name because that's the name etcd will give to the pods.
			etcdPod := regexp.MustCompile(`example-.[a-z0-9]*`).FindString(pods)

			ocArgs := []string{"get", "pods", etcdPod, "-o", "template=\"{{.status.phase}}\"", "-n", namespace}
			helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
				return strings.Contains(output, "Running")
			})

			// now link the component and service
			stdOut := helper.CmdShouldPass("odo", "link", "EtcdCluster/example")
			Expect(stdOut).To(ContainSubstring("Successfully linked"))
		})
	})
})
