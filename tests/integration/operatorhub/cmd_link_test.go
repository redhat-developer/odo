package integration

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo link command tests for OperatorHub", func() {

	var commonVar helper.CommonVar

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Operators are installed in the cluster", func() {

		var etcdOperator string
		var etcdCluster string

		BeforeEach(func() {
			// wait till odo can see that all operators installed by setup script in the namespace
			odoArgs := []string{"catalog", "list", "services"}
			operators := []string{"etcdoperator", "service-binding-operator"}
			for _, operator := range operators {
				helper.WaitForCmdOut("odo", odoArgs, 5, true, func(output string) bool {
					return strings.Contains(output, operator)
				})
			}

			list := helper.Cmd("odo", "catalog", "list", "services").ShouldPass().Out()
			etcdOperator = regexp.MustCompile(`etcdoperator\.*[a-z][0-9]\.[0-9]\.[0-9]-clusterwide`).FindString(list)
			etcdCluster = fmt.Sprintf("%s/EtcdCluster", etcdOperator)
		})

		AfterEach(func() {
			helper.DeleteProject(commonVar.Project)
		})

		When("a component and a service are deployed", func() {

			var componentName string
			var svcFullName string

			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				componentName = "cmp-" + helper.RandString(6)
				helper.Cmd("odo", "create", "nodejs", componentName).ShouldPass()

				serviceName := "service-" + helper.RandString(6)
				svcFullName = strings.Join([]string{"EtcdCluster", serviceName}, "/")
				helper.Cmd("odo", "service", "create", etcdCluster, serviceName, "--project", commonVar.Project).ShouldPass()

				helper.Cmd("odo", "push").ShouldPass()
				name := commonVar.CliRunner.GetRunningPodNameByComponent(componentName, commonVar.Project)
				Expect(name).To(Not(BeEmpty()))
			})

			It("should find files in component container", func() {
				helper.Cmd("odo", "exec", "--", "ls", "/project/server.js").ShouldPass()
			})

			When("a link between the component and the service is created and deployed", func() {

				BeforeEach(func() {
					helper.Cmd("odo", "link", svcFullName).ShouldPass()
					helper.Cmd("odo", "push").ShouldPass()
					name := commonVar.CliRunner.GetRunningPodNameByComponent(componentName, commonVar.Project)
					Expect(name).To(Not(BeEmpty()))
				})

				It("should find files in component container", func() {
					helper.Cmd("odo", "exec", "--", "ls", "/project/server.js").ShouldPass()
				})

				It("should find the link environment variable", func() {
					stdOut := helper.Cmd("odo", "exec", "--", "sh", "-c", "echo $ETCDCLUSTER_CLUSTERIP").ShouldPass().Out()
					Expect(stdOut).To(Not(BeEmpty()))
				})
			})

			When("a link with between the component and the service is created with --bind-as-files and deployed", func() {

				BeforeEach(func() {
					helper.Cmd("odo", "link", svcFullName, "--bind-as-files").ShouldPass()
					helper.Cmd("odo", "push").ShouldPass()
					name := commonVar.CliRunner.GetRunningPodNameByComponent(componentName, commonVar.Project)
					Expect(name).To(Not(BeEmpty()))
				})

				It("should find files in component container", func() {
					helper.Cmd("odo", "exec", "--", "ls", "/project/server.js").ShouldPass()
				})

				It("should find bindings for service", func() {
					helper.Cmd("odo", "exec", "--", "ls", "/bindings/etcd-link/clusterIP").ShouldPass()
				})
			})
		})

		When("getting sources, a devfile defining a component, a service and a link, and executing odo push", func() {

			BeforeEach(func() {
				componentName := "api" // this is the name of the component in the devfile
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-link.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.Cmd("odo", "create", componentName).ShouldPass()

				helper.Cmd("odo", "push").ShouldPass()
				name := commonVar.CliRunner.GetRunningPodNameByComponent(componentName, commonVar.Project)
				Expect(name).To(Not(BeEmpty()))
			})

			It("should find files in component container", func() {
				helper.Cmd("odo", "exec", "--", "ls", "/project/server.js").ShouldPass()
			})

			It("should find bindings for service", func() {
				helper.Cmd("odo", "exec", "--", "ls", "/bindings/etcd-link/clusterIP").ShouldPass()
			})

			It("should find owner references on link and service", func() {
				ocArgs := []string{"get", "servicebinding", "etcd-link", "-o", "jsonpath='{.metadata.ownerReferences.*.name}'", "-n", commonVar.Project}
				helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
					return strings.Contains(output, "api-app")
				})

				ocArgs = []string{"get", "etcdclusters.etcd.database.coreos.com", "myetcd", "-o", "jsonpath='{.metadata.ownerReferences.*.name}'", "-n", commonVar.Project}
				helper.WaitForCmdOut("oc", ocArgs, 1, true, func(output string) bool {
					return strings.Contains(output, "api-app")
				})
			})
		})
	})
})
