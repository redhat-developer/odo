package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo logs command tests", func() {
	var componentName string
	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("directory is empty", func() {

		BeforeEach(func() {
			Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
		})

		It("should error", func() {
			output := helper.Cmd("odo", "logs").ShouldFail().Err()
			Expect(output).To(ContainSubstring("this command cannot run in an empty directory"))
		})
	})

	When("component is created and odo logs is executed", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", componentName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})
		It("should successfully show logs of the running component", func() {
			err := helper.RunDevMode(func(session *gexec.Session, outContents []byte, errContents []byte, ports map[string]string) {
				out := helper.Cmd("odo", "logs").ShouldPass().Out()
				Expect(out).To(ContainSubstring("runtime: App started on PORT 3000"))
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
