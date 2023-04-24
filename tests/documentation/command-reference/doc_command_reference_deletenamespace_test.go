package docautomation

import (
	"fmt"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("doc command reference odo delete namespace", func() {
	var commonVar helper.CommonVar
	var commonPath = filepath.Join("command-reference", "docs-mdx", "delete-namespace")
	var outputStringFormat = "```console\n$ odo %s\n%s```\n"

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("To delete a namespace resource", func() {

		BeforeEach(func() {
			helper.Cmd("odo", "create", "namespace", "odo-dev").ShouldRun()

		})

		AfterEach(func() {
			commonVar.CliRunner.DeleteNamespaceProject("odo-dev", false)
		})

		It("Deletes a namespace resource for a kubernetes cluster", func() {
			args := []string{"odo", "delete", "namespace", "odo-dev"}
			out, err := helper.RunInteractive(args, []string{"ODO_LOG_LEVEL=0"}, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "? Are you sure you want to delete namespace \"odo-dev\"?")
				helper.SendLine(ctx, "Yes")

			})
			Expect(err).To(BeNil())
			got := helper.StripAnsi(out)
			got = helper.StripInteractiveQuestion(got)
			got = fmt.Sprintf(outputStringFormat, args[1], helper.StripSpinner(got))
			file := "delete_namespace.mdx"
			want := helper.GetMDXContent(filepath.Join(commonPath, file))
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty(), file)
		})

		It("Deletes a project resource for a openshift cluster", func() {
			args := []string{"odo", "delete", "project", "odo-dev"}
			out, err := helper.RunInteractive(args, []string{"ODO_LOG_LEVEL=0"}, func(ctx helper.InteractiveContext) {
				helper.ExpectString(ctx, "? Are you sure you want to delete project \"odo-dev\"?")
				helper.SendLine(ctx, "Yes")

			})
			Expect(err).To(BeNil())
			got := helper.StripAnsi(out)
			got = helper.StripInteractiveQuestion(got)
			got = fmt.Sprintf(outputStringFormat, args[1], helper.StripSpinner(got))
			file := "delete_project.mdx"
			want := helper.GetMDXContent(filepath.Join(commonPath, file))
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty(), file)
		})

	})
})
