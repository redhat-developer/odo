package integration

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo add binding command tests", func() {
	var commonVar helper.CommonVar
	var devSession helper.DevSession
	var err error

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
		helper.Chdir(commonVar.Context)
		// Ensure that the operators are installed
		commonVar.CliRunner.EnsureOperatorIsInstalled("service-binding-operator")
		commonVar.CliRunner.EnsureOperatorIsInstalled("cloud-native-postgresql")
		Eventually(func() string {
			out, _ := commonVar.CliRunner.GetBindableKinds()
			return out
		}, 120, 3).Should(ContainSubstring("Cluster"))
		addBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
		Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should fail creating a binding without workload parameter", func() {
		stderr := helper.Cmd("odo", "add", "binding", "--name", "aname", "--service", "cluster-sample").ShouldFail().Err()
		Expect(stderr).To(ContainSubstring("missing --workload parameter"))
	})

	It("should create a binding using the workload parameter", func() {
		stdout := helper.Cmd("odo", "add", "binding",
			"--name", "aname",
			"--service", "cluster-sample",
			"--workload", "app/Deployment.apps",
		).ShouldPass().Out()
		Expect(stdout).To(BeEquivalentTo(`apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  creationTimestamp: null
  name: aname
spec:
  application:
    group: apps
    kind: Deployment
    name: app
    version: v1
  bindAsFiles: true
  detectBindingResources: true
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: aname
    kind: Cluster
    name: cluster-sample
    resource: clusters
    version: v1
status:
  secret: ""

`))
	})

	It("should create a binding using the workload parameter and naming strategy", func() {
		stdout := helper.Cmd("odo", "add", "binding",
			"--name", "aname",
			"--service", "cluster-sample",
			"--workload", "app/Deployment.apps",
			"--naming-strategy", "lowercase",
		).ShouldPass().Out()
		Expect(stdout).To(BeEquivalentTo(`apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  creationTimestamp: null
  name: aname
spec:
  application:
    group: apps
    kind: Deployment
    name: app
    version: v1
  bindAsFiles: true
  detectBindingResources: true
  namingStrategy: lowercase
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: aname
    kind: Cluster
    name: cluster-sample
    resource: clusters
    version: v1
status:
  secret: ""

`))
	})

	When("binding to a service in a different namespace", func() {
		var otherNS string
		var nsWithNoService string

		BeforeEach(func() {
			otherNS = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
			addBindableKindInOtherNs := commonVar.CliRunner.Run("-n", otherNS, "apply", "-f",
				helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
			Expect(addBindableKindInOtherNs.ExitCode()).To(BeEquivalentTo(0))

			nsWithNoService = commonVar.CliRunner.CreateAndSetRandNamespaceProject()

			commonVar.CliRunner.SetProject(commonVar.Project)
		})

		AfterEach(func() {
			commonVar.CliRunner.DeleteNamespaceProject(nsWithNoService, false)
			commonVar.CliRunner.DeleteNamespaceProject(otherNS, false)
		})

		It("should error out if service is not found in the namespace selected", func() {
			stderr := helper.Cmd("odo", "add", "binding",
				"--name", "aname",
				"--service-namespace", nsWithNoService, "--service", "cluster-sample",
				"--workload", "app/Deployment.apps").ShouldFail().Err()
			Expect(stderr).To(ContainSubstring(fmt.Sprintf("No bindable service instances found in namespace %q", nsWithNoService)))
		})

		It("should error out if service is not found in list of services of namespace selected", func() {
			unknownService := "cluster-sample-not-found"
			stderr := helper.Cmd("odo", "add", "binding",
				"--name", "aname",
				"--service-namespace", otherNS, "--service", unknownService,
				"--workload", "app/Deployment.apps").ShouldFail().Err()
			Expect(stderr).To(ContainSubstring(fmt.Sprintf("%q service not found", unknownService)))
		})

		It("should create a binding using the workload parameter", func() {
			helper.Cmd("odo", "add", "binding",
				"--name", "aname",
				"--service-namespace", otherNS, "--service", "cluster-sample",
				"--workload", "app/Deployment.apps").ShouldPass()
		})
	})

	When("the component is bootstrapped", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", "mynode", "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml"), "--starter", "nodejs-starter").ShouldPass()
		})

		It("should fail using the --workload parameter", func() {
			stderr := helper.Cmd("odo", "add", "binding", "--name", "aname", "--service", "cluster-sample", "--workload", "app/Deployment.apps").ShouldFail().Err()
			Expect(stderr).To(ContainSubstring("--workload cannot be used from a directory containing a Devfile"))
		})

		for _, ctx := range []struct {
			name           string
			beforeEachFunc func(bindingName string) string
			afterEachFunc  func(ns string)
			checkFunc      func(bindingName string, ns string, k8sComponentInlined string)
		}{
			{
				name: "current namespace",
				beforeEachFunc: func(bindingName string) string {
					helper.Cmd("odo", "add", "binding", "--name", bindingName, "--service", "cluster-sample").ShouldPass()
					return commonVar.Project
				},
				checkFunc: func(bindingName string, _ string, k8sComponentInlined string) {
					Expect(k8sComponentInlined).To(ContainSubstring(fmt.Sprintf(`
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: %s
    kind: Cluster
    name: cluster-sample
    resource: clusters
    version: v1`, bindingName)))
				},
			},
			{
				name: "other namespace",
				beforeEachFunc: func(bindingName string) string {
					otherNS := commonVar.CliRunner.CreateAndSetRandNamespaceProject()
					addBindableKindInOtherNs := commonVar.CliRunner.Run("-n", otherNS, "apply", "-f",
						helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
					Expect(addBindableKindInOtherNs.ExitCode()).To(BeEquivalentTo(0))

					commonVar.CliRunner.SetProject(commonVar.Project)

					helper.Cmd("odo", "add", "binding", "--name", bindingName,
						"--service", "cluster-sample", "--service-namespace", otherNS).ShouldPass()

					return otherNS
				},
				afterEachFunc: func(ns string) {
					commonVar.CliRunner.DeleteNamespaceProject(ns, false)
				},
				checkFunc: func(bindingName string, ns string, k8sComponentInlined string) {
					Expect(k8sComponentInlined).To(ContainSubstring(fmt.Sprintf(`
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: %s
    kind: Cluster
    name: cluster-sample
    namespace: %s
    resource: clusters
    version: v1`, bindingName, ns)))
				},
			},
		} {
			ctx := ctx

			When("adding a binding ("+ctx.name+")", func() {
				var bindingName string
				var ns string

				BeforeEach(func() {
					bindingName = fmt.Sprintf("binding-%s", helper.RandString(4))
					ns = ctx.beforeEachFunc(bindingName)
				})

				AfterEach(func() {
					if ctx.afterEachFunc != nil {
						ctx.afterEachFunc(ns)
					}
				})

				It("should successfully add binding between component and service in the devfile", func() {
					components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
					Expect(components).ToNot(BeNil())
					Expect(components).To(HaveLen(1))
					Expect(components[0].Kubernetes).ToNot(BeNil())
					Expect(components[0].Kubernetes.Inlined).ToNot(BeEmpty())
					ctx.checkFunc(bindingName, ns, components[0].Kubernetes.Inlined)
				})

				When("odo dev is run", func() {
					BeforeEach(func() {
						devSession, _, _, _, err = helper.StartDevMode(nil)
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})

					It("should successfully bind component and service", func() {
						stdout := commonVar.CliRunner.Run("get", "servicebinding", bindingName).Out.Contents()
						Expect(stdout).To(ContainSubstring("ApplicationsBound"))
					})
				})

				When("odo dev is run", func() {
					BeforeEach(func() {
						devSession, _, _, _, err = helper.StartDevMode(nil)
						Expect(err).ToNot(HaveOccurred())
					})

					When("odo dev command is stopped", func() {
						BeforeEach(func() {
							devSession.Stop()
							devSession.WaitEnd()
						})

						It("should have successfully delete the binding", func() {
							_, errOut := commonVar.CliRunner.GetServiceBinding(bindingName, commonVar.Project)
							Expect(errOut).To(ContainSubstring("not found"))
						})
					})
				})
			})
		}

		When("no bindable instance is present on the cluster", func() {
			BeforeEach(func() {
				deleteBindableKind := commonVar.CliRunner.Run("delete", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
				Expect(deleteBindableKind.ExitCode()).To(BeEquivalentTo(0))
			})
			It("should fail to add binding with no bindable instance found error message", func() {
				errOut := helper.Cmd("odo", "add", "binding", "--name", "my-binding", "--service", "cluster-sample").ShouldFail().Err()
				Expect(errOut).To(ContainSubstring("No bindable service instances found"))
			})
		})
	})
})
