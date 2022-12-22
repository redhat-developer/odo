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

	FIt("should check good versioned devfile output", func() {
		out := helper.Cmd("odo", "init", "--devfile", "go", "--name", "my-go-app", "--devfile-version", "2.0.0").ShouldPass().Out()
		match, err := helper.CompareDocOutput(out, filepath.Join(commonPath, "versioned_devfile_output.mdx"))
		Expect(err).To(BeNil())
		Expect(match).To(BeEmpty())
	})
})
