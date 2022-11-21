package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo add binding interactive command tests", func() {

	var commonVar helper.CommonVar
	var serviceName string

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)

		// We make EXPLICITLY sure that we are outputting with NO COLOR
		// this is because in some cases we are comparing the output with a colorized one
		os.Setenv("NO_COLOR", "true")

		// Ensure that the operators are installed
		commonVar.CliRunner.EnsureOperatorIsInstalled("service-binding-operator")
		commonVar.CliRunner.EnsureOperatorIsInstalled("cloud-native-postgresql")
		Eventually(func() string {
			out, _ := commonVar.CliRunner.GetBindableKinds()
			return out
		}, 120, 3).Should(ContainSubstring("Cluster"))
		addBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
		Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
		commonVar.CliRunner.EnsurePodIsUp(commonVar.Project, "cluster-sample-1")
		serviceName = "cluster-sample" // Hard coded from bindablekind-instance.yaml
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("the component is bootstrapped", func() {
		var componentName = "mynode"
		var bindingName string
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", componentName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml"), "--starter", "nodejs-starter").ShouldPass()
			bindingName = fmt.Sprintf("%s-%s", componentName, serviceName)
		})

		checkBindingInDevfile := func(bindingName string, bindAsFiles bool, namingStrategy string) {
			components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
			Expect(components).ToNot(BeNil())
			Expect(components).To(HaveLen(1))
			cmp := components[0]
			Expect(cmp.Kubernetes).ToNot(BeNil())

			Expect(cmp.Kubernetes.Inlined).To(ContainSubstring("bindAsFiles: " + strconv.FormatBool(bindAsFiles)))

			if namingStrategy != "" {
				Expect(cmp.Kubernetes.Inlined).To(ContainSubstring("namingStrategy: " + namingStrategy))
			} else {
				Expect(cmp.Kubernetes.Inlined).ToNot(ContainSubstring("namingStrategy: "))
			}
		}

		It("should successfully add binding to the devfile (Bind as Environment Variables)", func() {
			command := []string{"odo", "add", "binding"}

			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "Do you want to list services from:")
				helper.SendLine(ctx, "current namespace")

				helper.ExpectString(ctx, "Select service instance you want to bind to:")
				helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

				helper.ExpectString(ctx, "Enter the Binding's name")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "How do you want to bind the service?")
				helper.SendLine(ctx, "Bind as Environment Variables")

				helper.ExpectString(ctx, "Select naming strategy for binding names")
				helper.SendLine(ctx, "DEFAULT")

				helper.ExpectString(ctx, "Successfully added the binding to the devfile.")

				helper.ExpectString(ctx, fmt.Sprintf("odo add binding --service cluster-sample.Cluster.postgresql.k8s.enterprisedb.io --name %s --bind-as-files=false", bindingName))
			})

			Expect(err).To(BeNil())
			checkBindingInDevfile(bindingName, false, "")
		})

		It("should successfully add binding to the devfile (Bind as Files)", func() {
			command := []string{"odo", "add", "binding"}

			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "Do you want to list services from:")
				helper.SendLine(ctx, "current namespace")

				helper.ExpectString(ctx, "Select service instance you want to bind to:")
				helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

				helper.ExpectString(ctx, "Enter the Binding's name")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "How do you want to bind the service?")
				helper.SendLine(ctx, "Bind as Files")

				helper.ExpectString(ctx, "Select naming strategy for binding names")
				helper.SendLine(ctx, "DEFAULT")

				helper.ExpectString(ctx, "Successfully added the binding to the devfile.")

				helper.ExpectString(ctx, fmt.Sprintf("odo add binding --service cluster-sample.Cluster.postgresql.k8s.enterprisedb.io --name %s", bindingName))
			})

			Expect(err).To(BeNil())
			checkBindingInDevfile(bindingName, true, "")
		})

		for _, predefinedNamingStrategy := range []string{"none", "lowercase", "uppercase"} {
			predefinedNamingStrategy := predefinedNamingStrategy
			It(fmt.Sprintf("should successfully add binding to the devfile (%q as naming strategy)", predefinedNamingStrategy), func() {
				command := []string{"odo", "add", "binding"}

				_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Do you want to list services from:")
					helper.SendLine(ctx, "current namespace")

					helper.ExpectString(ctx, "Select service instance you want to bind to:")
					helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

					helper.ExpectString(ctx, "Enter the Binding's name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "How do you want to bind the service?")
					helper.SendLine(ctx, "Bind as Environment Variables")

					helper.ExpectString(ctx, "Select naming strategy for binding names")
					helper.SendLine(ctx, predefinedNamingStrategy)

					helper.ExpectString(ctx, "Successfully added the binding to the devfile.")

					helper.ExpectString(ctx,
						fmt.Sprintf("odo add binding --service cluster-sample.Cluster.postgresql.k8s.enterprisedb.io --name %s --bind-as-files=false --naming-strategy='%s'",
							bindingName, predefinedNamingStrategy))
				})

				Expect(err).To(BeNil())
				checkBindingInDevfile(bindingName, false, predefinedNamingStrategy)
			})
		}

		for _, tt := range []struct {
			namingStrategy string
			wantInDevfile  string
		}{
			{
				namingStrategy: "",
				wantInDevfile:  "",
			},
			{
				namingStrategy: "any string",
				wantInDevfile:  "any string",
			},
			{
				namingStrategy: "{ .name | upper }",
				wantInDevfile:  "'{ .name | upper }'",
			},
		} {
			tt := tt
			It(fmt.Sprintf("should successfully add binding to the devfile (custom naming strategy: %q)", tt.namingStrategy), func() {
				command := []string{"odo", "add", "binding"}

				_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Do you want to list services from:")
					helper.SendLine(ctx, "current namespace")

					helper.ExpectString(ctx, "Select service instance you want to bind to:")
					helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

					helper.ExpectString(ctx, "Enter the Binding's name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "How do you want to bind the service?")
					helper.SendLine(ctx, "Bind as Files")

					helper.ExpectString(ctx, "Select naming strategy for binding names")
					helper.SendLine(ctx, "CUSTOM")

					helper.ExpectString(ctx, "Enter the naming strategy")
					inputNamingStrategy := tt.namingStrategy
					if inputNamingStrategy == "" {
						inputNamingStrategy = "\n"
					}
					helper.SendLine(ctx, inputNamingStrategy)

					helper.ExpectString(ctx, "Successfully added the binding to the devfile.")

					automationMsg := fmt.Sprintf("odo add binding --service cluster-sample.Cluster.postgresql.k8s.enterprisedb.io --name %s", bindingName)
					if tt.namingStrategy != "" {
						automationMsg += fmt.Sprintf(" --naming-strategy='%s'", tt.namingStrategy)
					}
					helper.ExpectString(ctx, automationMsg)
				})

				Expect(err).To(BeNil())
				checkBindingInDevfile(bindingName, true, tt.wantInDevfile)
			})
		}

		When("binding to a service in a different namespace", func() {
			var otherNS string
			var nsWithNoService string

			BeforeEach(func() {
				otherNS = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
				addBindableKindInOtherNs := commonVar.CliRunner.Run("-n", otherNS, "apply", "-f",
					helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
				Expect(addBindableKindInOtherNs.ExitCode()).To(BeEquivalentTo(0))
				commonVar.CliRunner.EnsurePodIsUp(otherNS, "cluster-sample-1")
				nsWithNoService = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
				commonVar.CliRunner.ListNamespaceProject(nsWithNoService)
				commonVar.CliRunner.ListNamespaceProject(otherNS)

				commonVar.CliRunner.SetProject(commonVar.Project)
			})

			AfterEach(func() {
				commonVar.CliRunner.DeleteNamespaceProject(nsWithNoService, false)
				commonVar.CliRunner.DeleteNamespaceProject(otherNS, false)
			})

			It("should error out if service is not found in the namespace selected", func() {
				_, err := helper.RunInteractive([]string{"odo", "add", "binding"}, nil, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Do you want to list services from:")
					helper.SendLine(ctx, "all accessible namespaces")

					helper.ExpectString(ctx, "Select the namespace containing the service instances:")
					helper.SendLine(ctx, nsWithNoService)

					helper.ExpectString(ctx, fmt.Sprintf("No bindable service instances found in namespace %q", nsWithNoService))
				})
				Expect(err).To(HaveOccurred())
				components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
				Expect(components).To(HaveLen(0))
			})

			It("should successfully add binding to the devfile", func() {
				command := []string{"odo", "add", "binding"}

				_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Do you want to list services from:")
					helper.SendLine(ctx, "all accessible namespaces")

					helper.ExpectString(ctx, "Select the namespace containing the service instances:")
					helper.SendLine(ctx, otherNS)

					helper.ExpectString(ctx, "Select service instance you want to bind to:")
					helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

					helper.ExpectString(ctx, "Enter the Binding's name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "How do you want to bind the service?")
					helper.SendLine(ctx, "Bind as Files")

					helper.ExpectString(ctx, "Select naming strategy for binding names")
					helper.SendLine(ctx, "DEFAULT")

					helper.ExpectString(ctx, "Successfully added the binding to the devfile.")
				})

				Expect(err).To(BeNil())
				components := helper.GetDevfileComponents(filepath.Join(commonVar.Context, "devfile.yaml"), bindingName)
				Expect(components).ToNot(BeNil())
				Expect(components).To(HaveLen(1))
				cmp := components[0]
				Expect(cmp.Kubernetes).ToNot(BeNil())
				Expect(cmp.Kubernetes.Inlined).To(ContainSubstring(fmt.Sprintf(`
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: mynode-cluster-sample
    kind: Cluster
    name: cluster-sample
    namespace: %s
    resource: clusters
    version: v1`, otherNS)))
			})

		})
	})

	When("running a deployment", func() {
		BeforeEach(func() {
			commonVar.CliRunner.Run("create", "deployment", "nginx", "--image=nginx")
		})

		AfterEach(func() {
			commonVar.CliRunner.Run("delete", "deployment", "nginx")
		})

		It("should successfully add binding without devfile (default naming strategy)", func() {
			command := []string{"odo", "add", "binding"}

			_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
				outputFile := "binding.yaml"
				expected := `spec:
  application:
    group: apps
    kind: Deployment
    name: nginx
    version: v1
  bindAsFiles: true
  detectBindingResources: true
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: nginx-cluster-sample
    kind: Cluster
    name: cluster-sample
    resource: clusters
    version: v1`

				helper.ExpectString(ctx, "Do you want to list services from:")
				helper.SendLine(ctx, "current namespace")

				helper.ExpectString(ctx, "Select service instance you want to bind to:")
				helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

				helper.ExpectString(ctx, "Select workload resource you want to bind:")
				helper.SendLine(ctx, "Deployment")

				helper.ExpectString(ctx, "Select workload resource name you want to bind:")
				helper.SendLine(ctx, "nginx")

				helper.ExpectString(ctx, "Enter the Binding's name")
				helper.SendLine(ctx, "")

				helper.ExpectString(ctx, "How do you want to bind the service?")
				helper.SendLine(ctx, "Bind as Files")

				helper.ExpectString(ctx, "Select naming strategy for binding names")
				helper.SendLine(ctx, "DEFAULT")

				helper.ExpectString(ctx, "Check(with Space Bar) one or more operations to perform with the ServiceBinding")
				helper.SendLine(ctx, " \x1B[B \x1B[B ")

				helper.ExpectString(ctx, "Save the ServiceBinding to file:")
				helper.SendLine(ctx, outputFile)

				for _, line := range strings.Split(expected, "\n") {
					helper.ExpectString(ctx, line)
				}
				helper.ExpectString(ctx, "The ServiceBinding has been created in the cluster")
				inCluster := commonVar.CliRunner.Run("get", "servicebinding", "nginx-cluster-sample", "-o", "yaml").Out.Contents()
				Expect(string(inCluster)).To(ContainSubstring(expected))

				helper.ExpectString(ctx, fmt.Sprintf("The ServiceBinding has been written to the file %q", outputFile))
				helper.VerifyFileExists("binding.yaml")
				fileContent, err := os.ReadFile(filepath.Join(commonVar.Context, outputFile))
				Expect(err).Should(Succeed())
				Expect(string(fileContent)).To(ContainSubstring(expected))
			})

			Expect(err).To(BeNil())
		})

		for _, predefinedNamingStrategy := range []string{"none", "lowercase", "uppercase"} {
			predefinedNamingStrategy := predefinedNamingStrategy
			It(fmt.Sprintf("should successfully add binding without devfile (naming strategy: %q)", predefinedNamingStrategy), func() {
				command := []string{"odo", "add", "binding"}

				_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
					outputFile := "binding.yaml"
					expected := fmt.Sprintf(`spec:
  application:
    group: apps
    kind: Deployment
    name: nginx
    version: v1
  bindAsFiles: true
  detectBindingResources: true
  namingStrategy: %s
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: nginx-cluster-sample
    kind: Cluster
    name: cluster-sample
    resource: clusters
    version: v1`, predefinedNamingStrategy)

					helper.ExpectString(ctx, "Do you want to list services from:")
					helper.SendLine(ctx, "current namespace")

					helper.ExpectString(ctx, "Select service instance you want to bind to:")
					helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

					helper.ExpectString(ctx, "Select workload resource you want to bind:")
					helper.SendLine(ctx, "Deployment")

					helper.ExpectString(ctx, "Select workload resource name you want to bind:")
					helper.SendLine(ctx, "nginx")

					helper.ExpectString(ctx, "Enter the Binding's name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "How do you want to bind the service?")
					helper.SendLine(ctx, "Bind as Files")

					helper.ExpectString(ctx, "Select naming strategy for binding names")
					helper.SendLine(ctx, predefinedNamingStrategy)

					helper.ExpectString(ctx, "Check(with Space Bar) one or more operations to perform with the ServiceBinding")
					helper.SendLine(ctx, " \x1B[B \x1B[B ")

					helper.ExpectString(ctx, "Save the ServiceBinding to file:")
					helper.SendLine(ctx, outputFile)

					for _, line := range strings.Split(expected, "\n") {
						helper.ExpectString(ctx, line)
					}
					helper.ExpectString(ctx, "The ServiceBinding has been created in the cluster")
					inCluster := commonVar.CliRunner.Run("get", "servicebinding", "nginx-cluster-sample", "-o", "yaml").Out.Contents()
					Expect(string(inCluster)).To(ContainSubstring(expected))

					helper.ExpectString(ctx, fmt.Sprintf("The ServiceBinding has been written to the file %q", outputFile))
					helper.VerifyFileExists("binding.yaml")
					fileContent, err := os.ReadFile(filepath.Join(commonVar.Context, outputFile))
					Expect(err).Should(Succeed())
					Expect(string(fileContent)).To(ContainSubstring(expected))
				})

				Expect(err).To(BeNil())
			})
		}

		for _, tt := range []struct {
			namingStrategy string
			wantInYaml     string
		}{
			{
				namingStrategy: "",
				wantInYaml:     "",
			},
			{
				namingStrategy: "any string",
				wantInYaml:     "any string",
			},
			{
				namingStrategy: "{ .name | upper }",
				wantInYaml:     "'{ .name | upper }'",
			},
		} {
			tt := tt
			It(fmt.Sprintf("should successfully add binding without devfile (custom naming strategy: %q)", tt.namingStrategy), func() {
				command := []string{"odo", "add", "binding"}

				_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
					outputFile := "binding.yaml"
					expected := fmt.Sprintf(`spec:
  application:
    group: apps
    kind: Deployment
    name: nginx
    version: v1
  bindAsFiles: true
  detectBindingResources: true
  namingStrategy: %s
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: nginx-cluster-sample
    kind: Cluster
    name: cluster-sample
    resource: clusters
    version: v1`, tt.wantInYaml)
					if tt.namingStrategy == "" {
						expected = `spec:
  application:
    group: apps
    kind: Deployment
    name: nginx
    version: v1
  bindAsFiles: true
  detectBindingResources: true
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: nginx-cluster-sample
    kind: Cluster
    name: cluster-sample
    resource: clusters
    version: v1`
					}

					helper.ExpectString(ctx, "Do you want to list services from:")
					helper.SendLine(ctx, "current namespace")

					helper.ExpectString(ctx, "Select service instance you want to bind to:")
					helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

					helper.ExpectString(ctx, "Select workload resource you want to bind:")
					helper.SendLine(ctx, "Deployment")

					helper.ExpectString(ctx, "Select workload resource name you want to bind:")
					helper.SendLine(ctx, "nginx")

					helper.ExpectString(ctx, "Enter the Binding's name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "How do you want to bind the service?")
					helper.SendLine(ctx, "Bind as Files")

					helper.ExpectString(ctx, "Select naming strategy for binding names")
					helper.SendLine(ctx, "CUSTOM")

					helper.ExpectString(ctx, "Enter the naming strategy")
					inputNamingStrategy := tt.namingStrategy
					if inputNamingStrategy == "" {
						inputNamingStrategy = "\n"
					}
					helper.SendLine(ctx, inputNamingStrategy)
					helper.ExpectString(ctx, "Check(with Space Bar) one or more operations to perform with the ServiceBinding")
					helper.SendLine(ctx, " \x1B[B \x1B[B ")

					helper.ExpectString(ctx, "Save the ServiceBinding to file:")
					helper.SendLine(ctx, outputFile)

					for _, line := range strings.Split(expected, "\n") {
						helper.ExpectString(ctx, line)
					}
					helper.ExpectString(ctx, "The ServiceBinding has been created in the cluster")
					inCluster := commonVar.CliRunner.Run("get", "servicebinding", "nginx-cluster-sample", "-o", "yaml").Out.Contents()
					Expect(string(inCluster)).To(ContainSubstring(expected))

					helper.ExpectString(ctx, fmt.Sprintf("The ServiceBinding has been written to the file %q", outputFile))
					helper.VerifyFileExists("binding.yaml")
					fileContent, err := os.ReadFile(filepath.Join(commonVar.Context, outputFile))
					Expect(err).Should(Succeed())
					Expect(string(fileContent)).To(ContainSubstring(expected))
				})

				Expect(err).To(BeNil())
			})
		}

		When("binding to a service in a different namespace", func() {
			var otherNS string
			var nsWithNoService string

			BeforeEach(func() {
				otherNS = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
				addBindableKindInOtherNs := commonVar.CliRunner.Run("-n", otherNS, "apply", "-f",
					helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
				Expect(addBindableKindInOtherNs.ExitCode()).To(BeEquivalentTo(0))
				commonVar.CliRunner.EnsurePodIsUp(otherNS, "cluster-sample-1")
				nsWithNoService = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
				commonVar.CliRunner.ListNamespaceProject(nsWithNoService)
				commonVar.CliRunner.ListNamespaceProject(otherNS)

				commonVar.CliRunner.SetProject(commonVar.Project)
			})

			AfterEach(func() {
				commonVar.CliRunner.DeleteNamespaceProject(nsWithNoService, false)
				commonVar.CliRunner.DeleteNamespaceProject(otherNS, false)

				commonVar.CliRunner.SetProject(commonVar.Project)
			})

			It("should error out if service is not found in the namespace selected", func() {
				_, err := helper.RunInteractive([]string{"odo", "add", "binding"}, nil, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Do you want to list services from:")
					helper.SendLine(ctx, "all accessible namespaces")

					helper.ExpectString(ctx, "Select the namespace containing the service instances:")
					helper.SendLine(ctx, nsWithNoService)

					helper.ExpectString(ctx, fmt.Sprintf("No bindable service instances found in namespace %q", nsWithNoService))
				})
				Expect(err).To(HaveOccurred())
				Expect(helper.VerifyFileExists("binding.yaml")).To(BeFalse())
			})

			It("should successfully add binding without devfile", func() {
				command := []string{"odo", "add", "binding"}

				_, err := helper.RunInteractive(command, nil, func(ctx helper.InteractiveContext) {
					outputFile := "binding.yaml"
					expected := fmt.Sprintf(`spec:
  application:
    group: apps
    kind: Deployment
    name: nginx
    version: v1
  bindAsFiles: true
  detectBindingResources: true
  services:
  - group: postgresql.k8s.enterprisedb.io
    id: nginx-cluster-sample
    kind: Cluster
    name: cluster-sample
    namespace: %s
    resource: clusters
    version: v1`, otherNS)

					helper.ExpectString(ctx, "Do you want to list services from:")
					helper.SendLine(ctx, "all accessible namespaces")

					helper.ExpectString(ctx, "Select the namespace containing the service instances:")
					helper.SendLine(ctx, otherNS)

					helper.ExpectString(ctx, "Select service instance you want to bind to:")
					helper.SendLine(ctx, "cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)")

					helper.ExpectString(ctx, "Select workload resource you want to bind:")
					helper.SendLine(ctx, "Deployment")

					helper.ExpectString(ctx, "Select workload resource name you want to bind:")
					helper.SendLine(ctx, "nginx")

					helper.ExpectString(ctx, "Enter the Binding's name")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "How do you want to bind the service?")
					helper.SendLine(ctx, "Bind as Files")

					helper.ExpectString(ctx, "Select naming strategy for binding names")
					helper.SendLine(ctx, "DEFAULT")

					helper.ExpectString(ctx, "Check(with Space Bar) one or more operations to perform with the ServiceBinding")
					helper.SendLine(ctx, " \x1B[B \x1B[B ")

					helper.ExpectString(ctx, "Save the ServiceBinding to file:")
					helper.SendLine(ctx, outputFile)

					for _, line := range strings.Split(expected, "\n") {
						helper.ExpectString(ctx, line)
					}
					helper.ExpectString(ctx, "The ServiceBinding has been created in the cluster")
					inCluster := commonVar.CliRunner.Run("get", "servicebinding", "nginx-cluster-sample", "-o", "yaml").Out.Contents()
					Expect(string(inCluster)).To(ContainSubstring(expected))

					helper.ExpectString(ctx, fmt.Sprintf("The ServiceBinding has been written to the file %q", outputFile))
					helper.VerifyFileExists("binding.yaml")
					fileContent, err := os.ReadFile(filepath.Join(commonVar.Context, outputFile))
					Expect(err).Should(Succeed())
					Expect(string(fileContent)).To(ContainSubstring(expected))
				})

				Expect(err).To(BeNil())
			})
		})
	})
})
