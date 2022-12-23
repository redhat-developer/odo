package integration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
	"path/filepath"
)

var _ = Describe("doc command reference init", Label(helper.LabelNoCluster), func() {
	var commonVar helper.CommonVar
	var commonPath = filepath.Join("command-reference", "docs-mdx", "init")
	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should check good versioned devfile output", func() {
		out := helper.Cmd("odo", "init", "--devfile", "go", "--name", "my-go-app", "--devfile-version", "2.0.0").ShouldPass().Out()
		stringsMissingFromCmdOut, stringsMissingFromFile, err := helper.CompareDocOutput(out, filepath.Join(commonPath, "versioned_devfile_output.mdx"))
		Expect(err).To(BeNil())
		Expect(stringsMissingFromCmdOut).To(BeEmpty())
		Expect(stringsMissingFromFile).To(BeEmpty())
	})

	It("should check latest versioned devfile output", func() {
		out := helper.Cmd("odo", "init", "--devfile", "go", "--name", "my-go-app", "--devfile-version", "latest").ShouldPass().Out()
		stringsMissingFromCmdOut, stringsMissingFromFile, err := helper.CompareDocOutput(out, filepath.Join(commonPath, "latest_versioned_devfile_output.mdx"))
		Expect(err).To(BeNil())
		Expect(stringsMissingFromCmdOut).To(BeEmpty())
		Expect(stringsMissingFromFile).To(BeEmpty())
	})

	It("should check devfile obtained from URL output", func() {
		out := helper.Cmd("odo", "init", "--devfile-path", "https://registry.devfile.io/devfiles/nodejs-angular", "--name", "my-nodejs-app", "--starter", "nodejs-angular-starter").ShouldPass().Out()
		stringsMissingFromCmdOut, stringsMissingFromFile, err := helper.CompareDocOutput(out, filepath.Join(commonPath, "devfile_from_url_output.mdx"))
		Expect(err).To(BeNil())
		Expect(stringsMissingFromCmdOut).To(BeEmpty())
		Expect(stringsMissingFromFile).To(BeEmpty())
	})

})
