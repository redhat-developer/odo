package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo describe/list binding command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	for _, ns := range []string{"", fmt.Sprintf("binding-%s", helper.RandString(3))} {
		ns := ns

		When(fmt.Sprintf("creating a component with a binding (service in namespace %q)", ns), func() {
			cmpName := "my-nodejs-app"
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml")).ShouldPass()
				if ns != "" {
					helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"),
						"name: cluster-sample",
						fmt.Sprintf(`name: cluster-sample
          namespace: %s`, ns))
				}
			})

			for _, label := range []string{
				helper.LabelNoCluster, helper.LabelUnauth,
			} {
				label := label
				It("should describe the binding without running odo dev", Label(label), func() {
					By("JSON output", func() {
						res := helper.Cmd("odo", "describe", "binding", "-o", "json").ShouldPass()
						stdout, stderr := res.Out(), res.Err()
						Expect(stderr).To(BeEmpty())
						Expect(helper.IsJSON(stdout)).To(BeTrue())
						helper.JsonPathContentIs(stdout, "0.name", "my-nodejs-app-cluster-sample-k8s")
						helper.JsonPathContentIs(stdout, "0.spec.application.kind", "Deployment")
						helper.JsonPathContentIs(stdout, "0.spec.application.name", "my-nodejs-app-app")
						helper.JsonPathContentIs(stdout, "0.spec.application.apiVersion", "apps/v1")
						helper.JsonPathContentIs(stdout, "0.spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
						helper.JsonPathContentIs(stdout, "0.spec.services.0.kind", "Cluster")
						helper.JsonPathContentIs(stdout, "0.spec.services.0.name", "cluster-sample")
						if ns != "" {
							helper.JsonPathContentIs(stdout, "0.spec.services.0.namespace", ns)
						} else {
							helper.JsonPathDoesNotExist(stdout, "0.spec.services.0.namespace")
						}

						helper.JsonPathContentIs(stdout, "1.name", "my-nodejs-app-cluster-sample-ocp")
					})

					By("human readable output", func() {
						res := helper.Cmd("odo", "describe", "binding").ShouldPass()
						stdout, _ := res.Out(), res.Err()
						Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
						Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-k8s"))
						Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-ocp"))
						if ns != "" {
							Expect(stdout).To(ContainSubstring(fmt.Sprintf("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: %s)", ns)))
						} else {
							Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
							Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
						}
						Expect(stdout).To(ContainSubstring("Bind as files: true"))
						Expect(stdout).To(ContainSubstring("Detect binding resources: true"))
						Expect(stdout).To(ContainSubstring("Naming strategy: lowercase"))
						Expect(stdout).To(ContainSubstring("Available binding information: unknown"))
						Expect(stdout).To(ContainSubstring("Binding information for one or more ServiceBinding is not available"))
					})
				})
			}
			for _, command := range [][]string{
				{"list", "binding"},
				{"list"},
			} {
				command := command
				It(fmt.Sprintf("should list the binding without running odo dev (%s)", command), func() {
					By("JSON output", func() {
						res := helper.Cmd("odo", append(command, "-o", "json")...).ShouldPass()
						stdout, stderr := res.Out(), res.Err()
						Expect(stderr).To(BeEmpty())
						Expect(helper.IsJSON(stdout)).To(BeTrue())
						helper.JsonPathContentIs(stdout, "bindings.0.name", "my-nodejs-app-cluster-sample-k8s")
						helper.JsonPathContentIs(stdout, "bindings.0.spec.application.kind", "Deployment")
						helper.JsonPathContentIs(stdout, "bindings.0.spec.application.name", "my-nodejs-app-app")
						helper.JsonPathContentIs(stdout, "bindings.0.spec.application.apiVersion", "apps/v1")
						helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
						helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.kind", "Cluster")
						helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.name", "cluster-sample")
						if ns != "" {
							helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.namespace", ns)
						} else {
							helper.JsonPathDoesNotExist(stdout, "bindings.0.spec.services.0.namespace")
						}
						helper.JsonPathContentIs(stdout, "bindings.0.spec.detectBindingResources", "true")
						helper.JsonPathContentIs(stdout, "bindings.0.spec.bindAsFiles", "true")
						helper.JsonPathContentIs(stdout, "bindings.0.spec.namingStrategy", "lowercase")
						helper.JsonPathContentIs(stdout, "bindings.0.status", "")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.#", "2")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.0", "my-nodejs-app-cluster-sample-k8s")

						helper.JsonPathContentIs(stdout, "bindings.1.name", "my-nodejs-app-cluster-sample-ocp")
					})
					By("human readable output", func() {
						res := helper.Cmd("odo", command...).ShouldPass()
						stdout, _ := res.Out(), res.Err()
						lines := strings.Split(stdout, "\n")

						if len(command) == 1 {
							Expect(lines[0]).To(ContainSubstring(fmt.Sprintf("Listing resources from the namespace %q", commonVar.Project)))
							lines = lines[6:]
						} else {
							Expect(lines[0]).To(ContainSubstring("Listing ServiceBindings"))
						}
						Expect(lines[3]).To(ContainSubstring("* "))
						Expect(lines[3]).To(ContainSubstring("my-nodejs-app-cluster-sample-k8s"))
						Expect(lines[3]).To(ContainSubstring("my-nodejs-app-app (Deployment)"))
						if ns != "" {
							Expect(lines[3]).To(ContainSubstring(fmt.Sprintf("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: %s)", ns)))
						} else {
							Expect(lines[3]).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
							Expect(lines[3]).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
						}
						Expect(lines[3]).To(ContainSubstring("None"))

						Expect(lines[4]).To(ContainSubstring("my-nodejs-app-cluster-sample-ocp"))
					})
				})
			}

		})

		for _, ctx := range []struct {
			title                                  string
			devfile                                string
			isServiceNsSupported                   bool
			assertDescribeAllJsonOutput            func(stdout, stderr string)
			assertDescribeAllHumanReadableOutput   func(stdout, stderr string)
			assertDescribeNamedJsonOutput          func(stdout, stderr string, name string)
			assertDescribeNamedHumanReadableOutput func(stdout, stderr string, name string)
			assertListJsonOutput                   func(devfile bool, stdout, stderr string)
			assertListHumanReadableOutput          func(devfile bool, stdout, stderr string, command []string)
		}{
			{
				title:                "creating a component with a binding as files",
				devfile:              helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml"),
				isServiceNsSupported: true,
				assertDescribeAllJsonOutput: func(stdout, stderr string) {
					prefixK8s := "0."
					prefixOcp := "1."
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, prefixK8s+"name", "my-nodejs-app-cluster-sample-k8s")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.name", "cluster-sample")
					if ns != "" {
						helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.namespace", ns)
					} else {
						helper.JsonPathDoesNotExist(stdout, prefixK8s+"spec.services.0.namespace")
					}
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.detectBindingResources", "true")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.bindAsFiles", "true")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.namingStrategy", "lowercase")
					helper.JsonPathContentContain(stdout, prefixK8s+"status.bindingFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password")
					helper.JsonPathContentIs(stdout, prefixK8s+"status.bindingEnvVars", "")

					helper.JsonPathContentIs(stdout, prefixOcp+"name", "my-nodejs-app-cluster-sample-ocp")
				},
				assertDescribeAllHumanReadableOutput: func(stdout, stderr string) {
					Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
					Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-k8s"))
					Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-ocp"))
					if ns != "" {
						Expect(stdout).To(ContainSubstring(fmt.Sprintf("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: %s)", ns)))
					} else {
						Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
						Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					}
					Expect(stdout).To(ContainSubstring("Bind as files: true"))
					Expect(stdout).To(ContainSubstring("Detect binding resources: true"))
					Expect(stdout).To(ContainSubstring("Naming strategy: lowercase"))
					Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password"))
				},

				assertDescribeNamedJsonOutput: func(stdout, stderr string, name string) {
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "name", name)
					helper.JsonPathContentIs(stdout, "spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, "spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, "spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, "spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, "spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, "spec.services.0.name", "cluster-sample")
					if ns != "" {
						helper.JsonPathContentIs(stdout, "spec.services.0.namespace", ns)
					} else {
						helper.JsonPathDoesNotExist(stdout, "spec.services.0.namespace")
					}
					helper.JsonPathContentIs(stdout, "spec.detectBindingResources", "true")
					helper.JsonPathContentIs(stdout, "spec.bindAsFiles", "true")
					helper.JsonPathContentIs(stdout, "spec.namingStrategy", "lowercase")
					helper.JsonPathContentContain(stdout, "status.bindingFiles", "${SERVICE_BINDING_ROOT}/"+name+"/password")
					helper.JsonPathContentIs(stdout, "status.bindingEnvVars", "")
				},
				assertDescribeNamedHumanReadableOutput: func(stdout, stderr string, name string) {
					Expect(stdout).To(ContainSubstring("Service Binding Name: " + name))
					if ns != "" {
						Expect(stdout).To(ContainSubstring(fmt.Sprintf("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: %s)", ns)))
					} else {
						Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
						Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					}
					Expect(stdout).To(ContainSubstring("Bind as files: true"))
					Expect(stdout).To(ContainSubstring("Detect binding resources: true"))
					Expect(stdout).To(ContainSubstring("Naming strategy: lowercase"))
					Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/" + name + "/password"))
				},

				assertListJsonOutput: func(devfile bool, stdout, stderr string) {
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "bindings.0.name", "my-nodejs-app-cluster-sample-k8s")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.name", "cluster-sample")
					if ns != "" {
						helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.namespace", ns)
					} else {
						helper.JsonPathDoesNotExist(stdout, "bindings.0.spec.services.0.namespace")
					}
					helper.JsonPathContentIs(stdout, "bindings.0.spec.detectBindingResources", "true")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.bindAsFiles", "true")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.namingStrategy", "lowercase")
					helper.JsonPathContentContain(stdout, "bindings.0.status.bindingFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password")
					helper.JsonPathContentIs(stdout, "bindings.0.status.bindingEnvVars", "")
					if devfile {
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.#", "2")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.0", "my-nodejs-app-cluster-sample-k8s")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.1", "my-nodejs-app-cluster-sample-ocp")
					} else {
						helper.JsonPathContentIs(stdout, "bindingsInDevfile", "")
					}

					helper.JsonPathContentIs(stdout, "bindings.1.name", "my-nodejs-app-cluster-sample-ocp")
				},
				assertListHumanReadableOutput: func(devfile bool, stdout, stderr string, command []string) {
					lines := strings.Split(stdout, "\n")
					if len(command) == 1 {
						Expect(lines[0]).To(ContainSubstring(fmt.Sprintf("Listing resources from the namespace %q", commonVar.Project)))
						lines = lines[6:]
					} else {
						Expect(lines[0]).To(ContainSubstring("Listing ServiceBindings"))
					}
					if devfile {
						Expect(lines[3]).To(ContainSubstring("* "))
					} else {
						Expect(lines[3]).ToNot(ContainSubstring("* "))
					}
					Expect(lines[3]).To(ContainSubstring("my-nodejs-app-cluster-sample-k8s"))
					Expect(lines[3]).To(ContainSubstring("my-nodejs-app-app (Deployment)"))
					if ns != "" {
						Expect(lines[3]).To(ContainSubstring(fmt.Sprintf("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: %s)", ns)))
					} else {
						Expect(lines[3]).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
						Expect(lines[3]).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					}
					Expect(lines[3]).To(ContainSubstring("Dev"))

					Expect(lines[4]).To(ContainSubstring("my-nodejs-app-cluster-sample-ocp"))

				},
			},
			{
				title:                "creating a component with a binding as environment variables",
				devfile:              helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-service-binding-envvars.yaml"),
				isServiceNsSupported: true,
				assertDescribeAllJsonOutput: func(stdout, stderr string) {
					prefixK8s := "0."
					prefixOcp := "1."
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, prefixK8s+"name", "my-nodejs-app-cluster-sample-k8s")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.name", "cluster-sample")
					if ns != "" {
						helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.namespace", ns)
					} else {
						helper.JsonPathDoesNotExist(stdout, prefixK8s+"spec.services.0.namespace")
					}
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.detectBindingResources", "true")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.bindAsFiles", "false")
					helper.JsonPathContentIs(stdout, prefixK8s+"status.bindingFiles", "")
					helper.JsonPathContentContain(stdout, prefixK8s+"status.bindingEnvVars", "PASSWORD")
					helper.JsonPathDoesNotExist(stdout, prefixK8s+"spec.namingStrategy")

					helper.JsonPathContentIs(stdout, prefixOcp+"name", "my-nodejs-app-cluster-sample-ocp")
				},
				assertDescribeAllHumanReadableOutput: func(stdout, stderr string) {
					Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
					Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-k8s"))
					Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-ocp"))
					if ns != "" {
						Expect(stdout).To(ContainSubstring(fmt.Sprintf("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: %s)", ns)))
					} else {
						Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
						Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					}
					Expect(stdout).To(ContainSubstring("Bind as files: false"))
					Expect(stdout).To(ContainSubstring("Detect binding resources: true"))
					Expect(stdout).To(ContainSubstring("PASSWORD"))
					Expect(stdout).ToNot(ContainSubstring("Naming strategy:"))
				},

				assertDescribeNamedJsonOutput: func(stdout, stderr string, name string) {
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "name", name)
					helper.JsonPathContentIs(stdout, "spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, "spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, "spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, "spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, "spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, "spec.services.0.name", "cluster-sample")
					if ns != "" {
						helper.JsonPathContentIs(stdout, "spec.services.0.namespace", ns)
					} else {
						helper.JsonPathDoesNotExist(stdout, "spec.services.0.namespace")
					}
					helper.JsonPathContentIs(stdout, "spec.detectBindingResources", "true")
					helper.JsonPathContentIs(stdout, "spec.bindAsFiles", "false")
					helper.JsonPathContentIs(stdout, "status.bindingFiles", "")
					helper.JsonPathContentContain(stdout, "status.bindingEnvVars", "PASSWORD")
					helper.JsonPathDoesNotExist(stdout, "spec.namingStrategy")
				},
				assertDescribeNamedHumanReadableOutput: func(stdout, stderr string, name string) {
					Expect(stdout).To(ContainSubstring("Service Binding Name: " + name))
					if ns != "" {
						Expect(stdout).To(ContainSubstring(fmt.Sprintf("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: %s)", ns)))
					} else {
						Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
						Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					}
					Expect(stdout).To(ContainSubstring("Bind as files: false"))
					Expect(stdout).To(ContainSubstring("Detect binding resources: true"))
					Expect(stdout).To(ContainSubstring("PASSWORD"))
					Expect(stdout).ToNot(ContainSubstring("Naming strategy:"))
				},
				assertListJsonOutput: func(devfile bool, stdout, stderr string) {
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "bindings.0.name", "my-nodejs-app-cluster-sample-k8s")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.name", "cluster-sample")
					if ns != "" {
						helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.namespace", ns)
					} else {
						helper.JsonPathDoesNotExist(stdout, "bindings.0.spec.services.0.namespace")
					}
					helper.JsonPathContentIs(stdout, "bindings.0.spec.detectBindingResources", "true")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.bindAsFiles", "false")
					helper.JsonPathContentIs(stdout, "bindings.0.status.bindingFiles", "")
					helper.JsonPathContentContain(stdout, "bindings.0.status.bindingEnvVars", "PASSWORD")
					if devfile {
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.#", "2")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.0", "my-nodejs-app-cluster-sample-k8s")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.1", "my-nodejs-app-cluster-sample-ocp")
					} else {
						helper.JsonPathContentIs(stdout, "bindingsInDevfile", "")
					}
					helper.JsonPathDoesNotExist(stdout, "bindings.0.spec.namingStrategy")

					helper.JsonPathContentIs(stdout, "bindings.1.name", "my-nodejs-app-cluster-sample-ocp")

				},
				assertListHumanReadableOutput: func(devfile bool, stdout, stderr string, command []string) {
					lines := strings.Split(stdout, "\n")
					if len(command) == 1 {
						Expect(lines[0]).To(ContainSubstring(fmt.Sprintf("Listing resources from the namespace %q", commonVar.Project)))
						lines = lines[6:]
					} else {
						Expect(lines[0]).To(ContainSubstring("Listing ServiceBindings"))
					}
					if devfile {
						Expect(lines[3]).To(ContainSubstring("* "))
					} else {
						Expect(lines[3]).ToNot(ContainSubstring("* "))
					}
					Expect(lines[3]).To(ContainSubstring("my-nodejs-app-cluster-sample-k8s"))
					Expect(lines[3]).To(ContainSubstring("my-nodejs-app-app (Deployment)"))
					if ns != "" {
						Expect(lines[3]).To(ContainSubstring(fmt.Sprintf("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: %s)", ns)))
					} else {
						Expect(lines[3]).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
						Expect(lines[3]).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					}
					Expect(lines[3]).To(ContainSubstring("Dev"))

					Expect(lines[4]).To(ContainSubstring("my-nodejs-app-cluster-sample-ocp"))
				},
			},

			{
				title:                "creating a component with a spec binding",
				devfile:              helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-spec-service-binding.yaml"),
				isServiceNsSupported: false,
				assertDescribeAllJsonOutput: func(stdout, stderr string) {
					prefixK8s := "0."
					prefixOcp := "1."
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, prefixK8s+"name", "my-nodejs-app-cluster-sample-k8s")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.name", "cluster-sample")
					helper.JsonPathDoesNotExist(stdout, prefixK8s+"spec.services.0.namespace")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.detectBindingResources", "false")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.bindAsFiles", "true")
					helper.JsonPathContentContain(stdout, prefixK8s+"status.bindingFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password")
					helper.JsonPathContentIs(stdout, prefixK8s+"status.bindingEnvVars", "")
					helper.JsonPathDoesNotExist(stdout, prefixK8s+"spec.namingStrategy")

					helper.JsonPathContentIs(stdout, prefixOcp+"name", "my-nodejs-app-cluster-sample-ocp")
				},
				assertDescribeAllHumanReadableOutput: func(stdout, stderr string) {
					Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
					Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-k8s"))
					Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-ocp"))
					Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
					Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					Expect(stdout).To(ContainSubstring("Bind as files: true"))
					Expect(stdout).To(ContainSubstring("Detect binding resources: false"))
					Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password"))
					Expect(stdout).ToNot(ContainSubstring("Naming strategy:"))
				},

				assertDescribeNamedJsonOutput: func(stdout, stderr string, name string) {
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "name", name)
					helper.JsonPathContentIs(stdout, "spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, "spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, "spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, "spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, "spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, "spec.services.0.name", "cluster-sample")
					helper.JsonPathDoesNotExist(stdout, "spec.services.0.namespace")
					helper.JsonPathContentIs(stdout, "spec.detectBindingResources", "false")
					helper.JsonPathContentIs(stdout, "spec.bindAsFiles", "true")
					helper.JsonPathContentContain(stdout, "status.bindingFiles", "${SERVICE_BINDING_ROOT}/"+name+"/password")
					helper.JsonPathContentIs(stdout, "status.bindingEnvVars", "")
					helper.JsonPathDoesNotExist(stdout, "spec.namingStrategy")
				},
				assertDescribeNamedHumanReadableOutput: func(stdout, stderr string, name string) {
					Expect(stdout).To(ContainSubstring("Service Binding Name: " + name))
					Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
					Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					Expect(stdout).To(ContainSubstring("Bind as files: true"))
					Expect(stdout).To(ContainSubstring("Detect binding resources: false"))
					Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/" + name + "/password"))
					Expect(stdout).ToNot(ContainSubstring("Naming strategy:"))
				},

				assertListJsonOutput: func(devfile bool, stdout, stderr string) {
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "bindings.0.name", "my-nodejs-app-cluster-sample-k8s")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.name", "cluster-sample")
					helper.JsonPathDoesNotExist(stdout, "bindings.0.spec.services.0.namespace")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.detectBindingResources", "false")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.bindAsFiles", "true")
					helper.JsonPathContentContain(stdout, "bindings.0.status.bindingFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password")
					helper.JsonPathContentIs(stdout, "bindings.0.status.bindingEnvVars", "")
					if devfile {
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.#", "2")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.0", "my-nodejs-app-cluster-sample-k8s")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.1", "my-nodejs-app-cluster-sample-ocp")
					} else {
						helper.JsonPathContentIs(stdout, "bindingsInDevfile", "")
					}
					helper.JsonPathDoesNotExist(stdout, "bindings.0.spec.namingStrategy")

					helper.JsonPathContentIs(stdout, "bindings.1.name", "my-nodejs-app-cluster-sample-ocp")
				},
				assertListHumanReadableOutput: func(devfile bool, stdout, stderr string, command []string) {
					lines := strings.Split(stdout, "\n")
					if len(command) == 1 {
						Expect(lines[0]).To(ContainSubstring(fmt.Sprintf("Listing resources from the namespace %q", commonVar.Project)))
						lines = lines[6:]
					} else {
						Expect(lines[0]).To(ContainSubstring("Listing ServiceBindings"))
					}
					if devfile {
						Expect(lines[3]).To(ContainSubstring("* "))
					} else {
						Expect(lines[3]).ToNot(ContainSubstring("* "))
					}
					Expect(lines[3]).To(ContainSubstring("my-nodejs-app-cluster-sample-k8s"))
					Expect(lines[3]).To(ContainSubstring("my-nodejs-app-app (Deployment)"))
					Expect(lines[3]).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
					Expect(lines[3]).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					Expect(lines[3]).To(ContainSubstring("Dev"))

					Expect(lines[4]).To(ContainSubstring("my-nodejs-app-cluster-sample-ocp"))
				},
			},
			{
				title:                "creating a component with a spec binding and envvars",
				devfile:              helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-spec-service-binding-envvars.yaml"),
				isServiceNsSupported: false,
				assertDescribeAllJsonOutput: func(stdout, stderr string) {
					prefixK8s := "0."
					prefixOcp := "1."
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, prefixK8s+"name", "my-nodejs-app-cluster-sample-k8s")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.services.0.name", "cluster-sample")
					helper.JsonPathDoesNotExist(stdout, prefixK8s+"spec.services.0.namespace")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.detectBindingResources", "false")
					helper.JsonPathContentIs(stdout, prefixK8s+"spec.bindAsFiles", "true")
					helper.JsonPathContentContain(stdout, prefixK8s+"status.bindingFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password")
					helper.JsonPathContentContain(stdout, prefixK8s+"status.bindingEnvVars", "PASSWD")
					helper.JsonPathDoesNotExist(stdout, prefixK8s+"spec.namingStrategy")

					helper.JsonPathContentIs(stdout, prefixOcp+"name", "my-nodejs-app-cluster-sample-ocp")
				},
				assertDescribeAllHumanReadableOutput: func(stdout, stderr string) {
					Expect(stdout).To(ContainSubstring("ServiceBinding used by the current component"))
					Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-k8s"))
					Expect(stdout).To(ContainSubstring("Service Binding Name: my-nodejs-app-cluster-sample-ocp"))
					Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
					Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					Expect(stdout).To(ContainSubstring("Bind as files: true"))
					Expect(stdout).To(ContainSubstring("Detect binding resources: false"))
					Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password"))
					Expect(stdout).To(ContainSubstring("PASSWD"))
					Expect(stdout).ToNot(ContainSubstring("Naming strategy:"))
				},

				assertDescribeNamedJsonOutput: func(stdout, stderr string, name string) {
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "name", name)
					helper.JsonPathContentIs(stdout, "spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, "spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, "spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, "spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, "spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, "spec.services.0.name", "cluster-sample")
					helper.JsonPathDoesNotExist(stdout, "spec.services.0.namespace")
					helper.JsonPathContentIs(stdout, "spec.detectBindingResources", "false")
					helper.JsonPathContentIs(stdout, "spec.bindAsFiles", "true")
					helper.JsonPathContentContain(stdout, "status.bindingFiles", "${SERVICE_BINDING_ROOT}/"+name+"/password")
					helper.JsonPathContentContain(stdout, "status.bindingEnvVars", "PASSWD")
					helper.JsonPathDoesNotExist(stdout, "spec.namingStrategy")
				},
				assertDescribeNamedHumanReadableOutput: func(stdout, stderr string, name string) {
					Expect(stdout).To(ContainSubstring("Service Binding Name: " + name))
					Expect(stdout).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
					Expect(stdout).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					Expect(stdout).To(ContainSubstring("Bind as files: true"))
					Expect(stdout).To(ContainSubstring("Detect binding resources: false"))
					Expect(stdout).To(ContainSubstring("${SERVICE_BINDING_ROOT}/" + name + "/password"))
					Expect(stdout).To(ContainSubstring("PASSWD"))
					Expect(stdout).ToNot(ContainSubstring("Naming strategy:"))
				},

				assertListJsonOutput: func(devfile bool, stdout, stderr string) {
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "bindings.0.name", "my-nodejs-app-cluster-sample-k8s")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.kind", "Deployment")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.name", "my-nodejs-app-app")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.application.apiVersion", "apps/v1")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.apiVersion", "postgresql.k8s.enterprisedb.io/v1")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.kind", "Cluster")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.services.0.name", "cluster-sample")
					helper.JsonPathDoesNotExist(stdout, "bindings.0.spec.services.0.namespace")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.detectBindingResources", "false")
					helper.JsonPathContentIs(stdout, "bindings.0.spec.bindAsFiles", "true")
					helper.JsonPathContentContain(stdout, "bindings.0.status.bindingFiles", "${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample-k8s/password")
					helper.JsonPathContentContain(stdout, "bindings.0.status.bindingEnvVars", "PASSWD")
					if devfile {
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.#", "2")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.0", "my-nodejs-app-cluster-sample-k8s")
						helper.JsonPathContentIs(stdout, "bindingsInDevfile.1", "my-nodejs-app-cluster-sample-ocp")
					} else {
						helper.JsonPathContentIs(stdout, "bindingsInDevfile", "")
					}
					helper.JsonPathDoesNotExist(stdout, "bindings.0.spec.namingStrategy")

					helper.JsonPathContentIs(stdout, "bindings.1.name", "my-nodejs-app-cluster-sample-ocp")
				},
				assertListHumanReadableOutput: func(devfile bool, stdout, stderr string, command []string) {
					lines := strings.Split(stdout, "\n")
					if len(command) == 1 {
						Expect(lines[0]).To(ContainSubstring(fmt.Sprintf("Listing resources from the namespace %q", commonVar.Project)))
						lines = lines[6:]
					} else {
						Expect(lines[0]).To(ContainSubstring("Listing ServiceBindings"))
					}
					if devfile {
						Expect(lines[3]).To(ContainSubstring("* "))
					} else {
						Expect(lines[3]).ToNot(ContainSubstring("* "))
					}
					Expect(lines[3]).To(ContainSubstring("my-nodejs-app-cluster-sample-k8s"))
					Expect(lines[3]).To(ContainSubstring("my-nodejs-app-app (Deployment)"))
					Expect(lines[3]).To(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)"))
					Expect(lines[3]).ToNot(ContainSubstring("cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: "))
					Expect(lines[3]).To(ContainSubstring("Dev"))

					Expect(lines[4]).To(ContainSubstring("my-nodejs-app-cluster-sample-ocp"))
				},
			},
		} {
			// this is a workaround to ensure that for loop works well with `It` blocks
			ctx := ctx
			When(fmt.Sprintf("%s (service in namespace %q)", ctx.title, ns), func() {
				cmpName := "my-nodejs-app"
				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
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

						if ctx.isServiceNsSupported && ns != "" {
							ns = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
							// Reset the original project
							commonVar.CliRunner.SetProject(commonVar.Project)

							addBindableKindInOtherNs := commonVar.CliRunner.Run("-n", ns, "apply", "-f",
								helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
							Expect(addBindableKindInOtherNs.ExitCode()).To(BeEquivalentTo(0))

							helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"),
								"name: cluster-sample",
								fmt.Sprintf(`name: cluster-sample
          namespace: %s`, ns))
							commonVar.CliRunner.EnsurePodIsUp(ns, "cluster-sample-1")
						} else {
							addBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
							Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
							commonVar.CliRunner.EnsurePodIsUp(commonVar.Project, "cluster-sample-1")
						}
					})

					AfterEach(func() {
						if ctx.isServiceNsSupported && ns != "" {
							commonVar.CliRunner.DeleteNamespaceProject(ns, false)
						}
					})

					When("running dev session", func() {
						var session helper.DevSession
						BeforeEach(func() {
							var err error
							session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
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
								ctx.assertDescribeAllJsonOutput(stdout, stderr)
							})
							By("human readable output", func() {
								res := helper.Cmd("odo", "describe", "binding").ShouldPass()
								stdout, stderr := res.Out(), res.Err()
								ctx.assertDescribeAllHumanReadableOutput(stdout, stderr)
							})

							By("JSON output from another directory with name flag", func() {
								err := os.Chdir("/")
								Expect(err).ToNot(HaveOccurred())
								res := helper.Cmd("odo", "describe", "binding", "--name", "my-nodejs-app-cluster-sample-k8s", "-o", "json").ShouldPass()
								stdout, stderr := res.Out(), res.Err()
								ctx.assertDescribeNamedJsonOutput(stdout, stderr, "my-nodejs-app-cluster-sample-k8s")

								res = helper.Cmd("odo", "describe", "binding", "--name", "my-nodejs-app-cluster-sample-ocp", "-o", "json").ShouldPass()
								stdout, stderr = res.Out(), res.Err()
								ctx.assertDescribeNamedJsonOutput(stdout, stderr, "my-nodejs-app-cluster-sample-ocp")
							})
							By("human readable output from another directory with name flag", func() {
								err := os.Chdir("/")
								Expect(err).ToNot(HaveOccurred())
								res := helper.Cmd("odo", "describe", "binding", "--name", "my-nodejs-app-cluster-sample-k8s").ShouldPass()
								stdout, stderr := res.Out(), res.Err()
								ctx.assertDescribeNamedHumanReadableOutput(stdout, stderr, "my-nodejs-app-cluster-sample-k8s")

								res = helper.Cmd("odo", "describe", "binding", "--name", "my-nodejs-app-cluster-sample-ocp").ShouldPass()
								stdout, stderr = res.Out(), res.Err()
								ctx.assertDescribeNamedHumanReadableOutput(stdout, stderr, "my-nodejs-app-cluster-sample-ocp")
							})

						})

						for _, command := range [][]string{
							{"list"},
							{"list", "binding"},
						} {
							command := command
							It(fmt.Sprintf("should list the binding - command: %v", command), func() {
								By("JSON output", func() {
									res := helper.Cmd("odo", append(command, "-o", "json")...).ShouldPass()
									stdout, stderr := res.Out(), res.Err()
									if ctx.assertListJsonOutput != nil {
										ctx.assertListJsonOutput(true, stdout, stderr)
									}
								})
								By("human readable output", func() {
									res := helper.Cmd("odo", command...).ShouldPass()
									stdout, stderr := res.Out(), res.Err()
									if ctx.assertListHumanReadableOutput != nil {
										ctx.assertListHumanReadableOutput(true, stdout, stderr, command)
									}
								})

								By("JSON output from another directory", func() {
									err := os.Chdir("/")
									Expect(err).ToNot(HaveOccurred())
									res := helper.Cmd("odo", append(command, "-o", "json")...).ShouldPass()
									stdout, stderr := res.Out(), res.Err()
									if ctx.assertListJsonOutput != nil {
										ctx.assertListJsonOutput(false, stdout, stderr)
									}
								})
								By("human readable output from another directory with name flag", func() {
									err := os.Chdir("/")
									Expect(err).ToNot(HaveOccurred())
									res := helper.Cmd("odo", command...).ShouldPass()
									stdout, stderr := res.Out(), res.Err()
									if ctx.assertListHumanReadableOutput != nil {
										ctx.assertListHumanReadableOutput(false, stdout, stderr, command)
									}
								})
							})

							When("changing the current namespace", func() {
								BeforeEach(func() {
									commonVar.CliRunner.SetProject("default")
								})

								AfterEach(func() {
									commonVar.CliRunner.SetProject(commonVar.Project)
								})

								It(fmt.Sprintf("should list the binding with --namespace flag - command: %s", command), func() {
									By("JSON output from another directory", func() {
										err := os.Chdir("/")
										Expect(err).ToNot(HaveOccurred())
										res := helper.Cmd("odo", append(command, "-o", "json", "--namespace", commonVar.Project)...).ShouldPass()
										stdout, stderr := res.Out(), res.Err()
										if ctx.assertListJsonOutput != nil {
											ctx.assertListJsonOutput(false, stdout, stderr)
										}
									})
									By("human readable output from another directory with name flag", func() {
										err := os.Chdir("/")
										Expect(err).ToNot(HaveOccurred())
										res := helper.Cmd("odo", append(command, "--namespace", commonVar.Project)...).ShouldPass()
										stdout, stderr := res.Out(), res.Err()
										if ctx.assertListHumanReadableOutput != nil {
											ctx.assertListHumanReadableOutput(false, stdout, stderr, command)
										}
									})
								})
							})
						}
					})
				})
			})
		}
	}
})
