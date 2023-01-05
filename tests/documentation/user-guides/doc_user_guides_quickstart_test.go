package docautomation

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("User guides: Quickstart test", func() {
	var commonVar helper.CommonVar
	var commonPath = filepath.Join("user-guides", "quickstart", "docs-mdx")
	var outputStringFormat = "```console\n$ odo %s\n%s```\n"
	const namespace = "odo-dev"

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})
	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Create namespace/project", func() {
		AfterEach(func() {
			helper.DeleteProject(namespace)
		})
		It("should show correct output for namespace/project creation", func() {
			args := []string{"create", "namespace", namespace}
			out := helper.Cmd("odo", args...).ShouldPass().Out()
			got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
			By("checking the output for namespace", func() {
				want := helper.GetMDXContent(filepath.Join(commonPath, "create_namespace_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
			By("checking the output for project", func() {
				got = strings.ReplaceAll(got, "namespace", "project")
				got = strings.ReplaceAll(got, "Namespace", "Project")
				want := helper.GetMDXContent(filepath.Join(commonPath, "create_project_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
		})
	})

	Context("nodejs", func() {
		commonNodeJSPath := filepath.Join(commonPath, "nodejs")
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})
		It("should test the complete nodejs quickstart output in order", func() {
			By("running odo init", func() {
				args := []string{"odo", "init"}
				out, err := helper.RunInteractive(args, []string{"ODO_LOG_LEVEL=0"}, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Is this correct?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "✓  Downloading devfile \"nodejs:2.1.1\" from registry \"DefaultDevfileRegistry\"")

					helper.ExpectString(ctx, "Select container for which you want to change configuration?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-nodejs-app")

					helper.ExpectString(ctx, "Changes will be directly reflected on the cluster.")
				})
				Expect(err).To(BeNil())
				got := helper.StripAnsi(out)
				got = helper.StripInteractiveQuestion(got)
				got = fmt.Sprintf(outputStringFormat, "init", helper.StripSpinner(got))
				want := helper.GetMDXContent(filepath.Join(commonNodeJSPath, "nodejs_odo_init_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
			By("running odo dev", func() {
				session, out, _, cmdEndpointsMap, err := helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).To(BeNil())
				session.Stop()
				session.WaitEnd()
				args := []string{"dev"}
				got := strings.ReplaceAll(string(out), commonVar.Context, "/home/user/quickstart-demo/nodejs-demo")
				got = helper.ReplaceAllForwardedPorts(got, cmdEndpointsMap, map[string]string{"3000": "127.0.0.1:40001", "5858": "127.0.0.1:40002"})
				got = strings.ReplaceAll(got, commonVar.Project, namespace)
				got = fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(got))
				want := helper.GetMDXContent(filepath.Join(commonNodeJSPath, "nodejs_odo_dev_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
		})
	})
	Context("Go quickstart guide", func() {
		commonGoPath := filepath.Join(commonPath, "go")
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "go"), commonVar.Context)
		})
		It("should test the complete go quickstart output in order", func() {
			By("running odo init", func() {
				args := []string{"odo", "init"}
				out, err := helper.RunInteractive(args, []string{"ODO_LOG_LEVEL=0"}, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Is this correct?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "✓  Downloading devfile \"go:1.0.2\" from registry \"DefaultDevfileRegistry\"")

					helper.ExpectString(ctx, "Select container for which you want to change configuration?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-go-app")

					helper.ExpectString(ctx, "Changes will be directly reflected on the cluster.")
				})
				Expect(err).To(BeNil())
				got := helper.StripAnsi(out)
				got = helper.StripInteractiveQuestion(got)
				got = fmt.Sprintf(outputStringFormat, "init", helper.StripSpinner(got))
				want := helper.GetMDXContent(filepath.Join(commonGoPath, "go_odo_init_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
			By("running odo dev", func() {
				session, out, _, cmdEndpointsMap, err := helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).To(BeNil())
				session.Stop()
				session.WaitEnd()
				args := []string{"dev"}
				got := strings.ReplaceAll(string(out), commonVar.Context, "/home/user/quickstart-demo/go-demo")
				got = helper.ReplaceAllForwardedPorts(got, cmdEndpointsMap, map[string]string{"8080": "127.0.0.1:40001"})
				got = strings.ReplaceAll(got, commonVar.Project, namespace)
				got = fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(got))
				want := helper.GetMDXContent(filepath.Join(commonGoPath, "go_odo_dev_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
		})
	})
	Context(".NET quickstart guide", func() {
		commondotnetPath := filepath.Join(commonPath, "dotnet")
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "dotnet"), commonVar.Context)
		})
		It("should test the complete dotnet quickstart output in order", func() {
			By("running odo init", func() {
				// this test is flaky when comparing envvar in the container configuration
				args := []string{"odo", "init"}
				out, err := helper.RunInteractive(args, []string{"ODO_LOG_LEVEL=0"}, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Is this correct?")
					helper.SendLine(ctx, "No")

					helper.ExpectString(ctx, "Select language")
					helper.SendLine(ctx, ".")

					helper.ExpectString(ctx, "Select project type")
					helper.SendLine(ctx, "6")

					helper.ExpectString(ctx, "✓  Downloading devfile \"dotnet60\" from registry \"DefaultDevfileRegistry\"")

					helper.ExpectString(ctx, "Select container for which you want to change configuration?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-dotnet-app")

					helper.ExpectString(ctx, "Changes will be directly reflected on the cluster.")
				})
				Expect(err).To(BeNil())
				got := helper.StripAnsi(out)
				got = helper.StripInteractiveQuestion(got)
				got = strings.ReplaceAll(got, commonVar.Project, namespace)
				got = fmt.Sprintf(outputStringFormat, "init", helper.StripSpinner(got))
				want := helper.GetMDXContent(filepath.Join(commondotnetPath, "dotnet_odo_init_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
			By("running odo dev", func() {
				session, out, _, cmdEndpointsMap, err := helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).To(BeNil())
				session.Stop()
				session.WaitEnd()
				args := []string{"dev"}
				got := strings.ReplaceAll(string(out), commonVar.Context, "/home/user/quickstart-demo/dotnet-demo")
				got = helper.ReplaceAllForwardedPorts(got, cmdEndpointsMap, map[string]string{"8080": "127.0.0.1:40001"})
				got = strings.ReplaceAll(got, commonVar.Project, namespace)
				got = fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(got))
				want := helper.GetMDXContent(filepath.Join(commondotnetPath, "dotnet_odo_dev_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
		})
	})
	Context("Java quickstart guide", func() {
		commonGoPath := filepath.Join(commonPath, "java")
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
		})
		It("should test the complete go quickstart output in order", func() {
			By("running odo init", func() {
				args := []string{"odo", "init"}
				out, err := helper.RunInteractive(args, []string{"ODO_LOG_LEVEL=0"}, func(ctx helper.InteractiveContext) {
					helper.ExpectString(ctx, "Is this correct?")
					helper.SendLine(ctx, "No")

					helper.ExpectString(ctx, "Select language")
					helper.SendLine(ctx, "Java")

					helper.ExpectString(ctx, "Select project type")
					helper.SendLine(ctx, "Spring")

					helper.ExpectString(ctx, "Select version")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "✓  Downloading devfile \"java-springboot:1.2.0\" from registry \"DefaultDevfileRegistry\"")

					helper.ExpectString(ctx, "Select container for which you want to change configuration?")
					helper.SendLine(ctx, "")

					helper.ExpectString(ctx, "Enter component name")
					helper.SendLine(ctx, "my-java-app")

					helper.ExpectString(ctx, "Changes will be directly reflected on the cluster.")
				})
				Expect(err).To(BeNil())
				got := helper.StripAnsi(out)
				got = helper.StripInteractiveQuestion(got)
				got = fmt.Sprintf(outputStringFormat, "init", helper.StripSpinner(got))
				want := helper.GetMDXContent(filepath.Join(commonGoPath, "java_odo_init_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
			By("running odo dev", func() {
				session, out, _, cmdEndpointsMap, err := helper.StartDevMode(helper.DevSessionOpts{TimeoutInSeconds: 420})
				Expect(err).To(BeNil())
				session.Stop()
				session.WaitEnd()
				args := []string{"dev"}
				got := strings.ReplaceAll(string(out), commonVar.Context, "/home/user/quickstart-demo/java-demo")
				got = helper.ReplaceAllForwardedPorts(got, cmdEndpointsMap, map[string]string{"8080": "127.0.0.1:40001", "5858": "127.0.0.1:40002"})
				got = strings.ReplaceAll(got, commonVar.Project, namespace)
				got = fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(got))
				want := helper.GetMDXContent(filepath.Join(commonGoPath, "java_odo_dev_output.mdx"))
				diff := cmp.Diff(want, got)
				Expect(diff).To(BeEmpty())
			})
		})
	})
})
