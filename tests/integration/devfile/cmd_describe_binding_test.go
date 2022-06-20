package devfile

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo describe binding command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		if helper.IsKubernetesCluster() {
			Skip("Operators have not been setup on Kubernetes cluster yet. Remove this once the issue has been fixed.")
		}
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("creating a component with a binding", func() {
		cmpName := "my-nodejs-app"
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml")).ShouldPass()
		})

		It("should describe the binding without running odo dev", func() {
			By("JSON output", func() {
				res := helper.Cmd("odo", "describe", "binding", "-o", "json").ShouldPass()
				stdout, stderr := res.Out(), res.Err()
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())
				helper.JsonPathContentIs(stdout, "0.name", "my-nodejs-app-cluster-sample")
				helper.JsonPathContentIs(stdout, "0.spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
				helper.JsonPathContentIs(stdout, "0.spec.services.0.kind", "Cluster")
				helper.JsonPathContentIs(stdout, "0.spec.services.0.name", "cluster-sample")
				helper.JsonPathContentIs(stdout, "0.spec.detectBindingResources", "true")
				helper.JsonPathContentIs(stdout, "0.spec.bindAsFiles", "true")
				helper.JsonPathContentIs(stdout, "0.status", "")
			})
			By("human readable output", func() {
				res := helper.Cmd("odo", "describe", "binding").ShouldPass()
				stdout, _ := res.Out(), res.Err()
				Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
				Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample"))
				Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
				Expect(stdout).To(ContainSubstring("Bind as files: true"))
				Expect(stdout).To(ContainSubstring("Detect binding resources: true"))
				Expect(stdout).To(ContainSubstring("Available binding information: unknown"))
				Expect(stdout).To(ContainSubstring("Binding information for one or more ServiceBinding is not available"))
			})
		})
	})

	for _, ctx := range []struct {
		title                     string
		devfile                   string
		assertJsonOutput          func(list bool, stdout, stderr string)
		assertHumanReadableOutput func(list bool, stdout, stderr string)
	}{
		{
			title:   "creating a component with a binding as files",
			devfile: helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml"),
			assertJsonOutput: func(list bool, stdout, stderr string) {
				prefix := ""
				if list {
					prefix = "0."
				}
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())
				helper.JsonPathContentIs(stdout, prefix+"name", "my-nodejs-app-cluster-sample")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.kind", "Cluster")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.name", "cluster-sample")
				helper.JsonPathContentIs(stdout, prefix+"spec.detectBindingResources", "true")
				helper.JsonPathContentIs(stdout, prefix+"spec.bindAsFiles", "true")
				helper.JsonPathContentContain(stdout, prefix+"status.bindingsFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/password")
				helper.JsonPathContentIs(stdout, prefix+"status.bindingEnvVars", "")
			},
			assertHumanReadableOutput: func(list bool, stdout, stderr string) {
				if list {
					Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
				}
				Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample"))
				Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
				Expect(stdout).To(ContainSubstring("Bind as files: true"))
				Expect(stdout).To(ContainSubstring("Detect binding resources: true"))
				Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/password"))
			},
		},
		{
			title:   "creating a component with a binding as environment variables",
			devfile: helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-envvars.yaml"),
			assertJsonOutput: func(list bool, stdout, stderr string) {
				prefix := ""
				if list {
					prefix = "0."
				}
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())
				helper.JsonPathContentIs(stdout, prefix+"name", "my-nodejs-app-cluster-sample")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.kind", "Cluster")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.name", "cluster-sample")
				helper.JsonPathContentIs(stdout, prefix+"spec.detectBindingResources", "true")
				helper.JsonPathContentIs(stdout, prefix+"spec.bindAsFiles", "false")
				helper.JsonPathContentIs(stdout, prefix+"status.bindingsFiles", "")
				helper.JsonPathContentContain(stdout, prefix+"status.bindingEnvVars", "PASSWORD")
			},
			assertHumanReadableOutput: func(list bool, stdout, stderr string) {
				if list {
					Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
				}
				Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample"))
				Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
				Expect(stdout).To(ContainSubstring("Bind as files: false"))
				Expect(stdout).To(ContainSubstring("Detect binding resources: true"))
				Expect(stdout).To(ContainSubstring("PASSWORD"))
			},
		},
		{
			title:   "creating a component with a spec binding",
			devfile: helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-spec-service-binding.yaml"),
			assertJsonOutput: func(list bool, stdout, stderr string) {
				prefix := ""
				if list {
					prefix = "0."
				}
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())
				helper.JsonPathContentIs(stdout, prefix+"name", "my-nodejs-app-cluster-sample")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.kind", "Cluster")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.name", "cluster-sample")
				helper.JsonPathContentIs(stdout, prefix+"spec.detectBindingResources", "false")
				helper.JsonPathContentIs(stdout, prefix+"spec.bindAsFiles", "true")
				helper.JsonPathContentContain(stdout, prefix+"status.bindingsFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/password")
				helper.JsonPathContentIs(stdout, prefix+"status.bindingEnvVars", "")
			},
			assertHumanReadableOutput: func(list bool, stdout, stderr string) {
				if list {
					Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
				}
				Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample"))
				Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
				Expect(stdout).To(ContainSubstring("Bind as files: true"))
				Expect(stdout).To(ContainSubstring("Detect binding resources: false"))
				Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/password"))
			},
		},
		{
			title:   "creating a component with a spec binding and envvars",
			devfile: helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-spec-service-binding-envvars.yaml"),
			assertJsonOutput: func(list bool, stdout, stderr string) {
				prefix := ""
				if list {
					prefix = "0."
				}
				Expect(stderr).To(BeEmpty())
				Expect(helper.IsJSON(stdout)).To(BeTrue())
				helper.JsonPathContentIs(stdout, prefix+"name", "my-nodejs-app-cluster-sample")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.kind", "Cluster")
				helper.JsonPathContentIs(stdout, prefix+"spec.services.0.name", "cluster-sample")
				helper.JsonPathContentIs(stdout, prefix+"spec.detectBindingResources", "false")
				helper.JsonPathContentIs(stdout, prefix+"spec.bindAsFiles", "true")
				helper.JsonPathContentContain(stdout, prefix+"status.bindingsFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/password")
				helper.JsonPathContentContain(stdout, prefix+"status.bindingEnvVars", "PASSWD")
			},
			assertHumanReadableOutput: func(list bool, stdout, stderr string) {
				if list {
					Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
				}
				Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample"))
				Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
				Expect(stdout).To(ContainSubstring("Bind as files: true"))
				Expect(stdout).To(ContainSubstring("Detect binding resources: false"))
				Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/password"))
				Expect(stdout).To(ContainSubstring("PASSWD"))
			},
		},
	} {
		// this is a workaround to ensure that for loop works well with `It` blocks
		ctx := ctx
		When(ctx.title, func() {
			cmpName := "my-nodejs-app"
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", ctx.devfile).ShouldPass()
			})

			When("Starting a Pg service", func() {
				BeforeEach(func() {
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

				When("running dev session", func() {
					var session helper.DevSession
					BeforeEach(func() {
						var err error
						session, _, _, _, err = helper.StartDevMode()
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						session.Kill()
						session.WaitEnd()
					})

					It("should describe the binding", func() {
						By("JSON output", func() {
							res := helper.Cmd("odo", "describe", "binding", "-o", "json").ShouldPass()
							stdout, stderr := res.Out(), res.Err()
							ctx.assertJsonOutput(true, stdout, stderr)
						})
						By("human readable output", func() {
							res := helper.Cmd("odo", "describe", "binding").ShouldPass()
							stdout, stderr := res.Out(), res.Err()
							ctx.assertHumanReadableOutput(true, stdout, stderr)
						})

						By("JSON output from another directory with name flag", func() {
							err := os.Chdir("/")
							Expect(err).ToNot(HaveOccurred())
							res := helper.Cmd("odo", "describe", "binding", "--name", "my-nodejs-app-cluster-sample", "-o", "json").ShouldPass()
							stdout, stderr := res.Out(), res.Err()
							ctx.assertJsonOutput(false, stdout, stderr)
						})
						By("human readable output from another directory with name flag", func() {
							err := os.Chdir("/")
							Expect(err).ToNot(HaveOccurred())
							res := helper.Cmd("odo", "describe", "binding", "--name", "my-nodejs-app-cluster-sample").ShouldPass()
							stdout, stderr := res.Out(), res.Err()
							ctx.assertHumanReadableOutput(false, stdout, stderr)
						})

					})
				})
			})
		})
	}
})
