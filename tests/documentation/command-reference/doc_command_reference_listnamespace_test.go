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

var _ = Describe("doc command reference odo list namespace", func() {
	var commonVar helper.CommonVar
	var commonPath = filepath.Join("command-reference", "docs-mdx", "list-namespace")
	var outputStringFormat = "```console\n$ odo %s\n%s```\n"
	var namespacelist = `
*          default
	kube-node-lease
	kube-public
	kube-system
	mynamespace
	myproject
	olm
	operators`

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("To list all available namespaces", func() {

		It("Lists all namespace resources available in a kubernetes cluster", func() {
			args := []string{"list", "namespace"}
			out := helper.Cmd("odo", args...).ShouldPass().Out()
			out = strings.SplitAfter(out, "NAME")[0] + namespacelist
			got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
			file := "list_namespace.mdx"
			want := helper.GetMDXContent(filepath.Join(commonPath, file))
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty(), file)
		})

		It("Lists all project resources available in a openshift cluster", func() {
			args := []string{"list", "project"}
			out := helper.Cmd("odo", args...).ShouldPass().Out()
			out = strings.SplitAfter(out, "NAME")[0] + namespacelist
			got := fmt.Sprintf(outputStringFormat, strings.Join(args, " "), helper.StripSpinner(out))
			file := "list_project.mdx"
			want := helper.GetMDXContent(filepath.Join(commonPath, file))
			diff := cmp.Diff(want, got)
			Expect(diff).To(BeEmpty(), file)
		})
	})

})
