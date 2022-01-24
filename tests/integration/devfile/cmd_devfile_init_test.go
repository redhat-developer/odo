package devfile

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile init command tests", func() {

	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
	})

	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should fail when running odo init with incomplete flags", func() {
		helper.Cmd("odo", "init", "--name", "aname").ShouldFail()
	})

	It("should fail and keep an empty directory when running odo init with wrong starter name", func() {
		helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go", "--starter", "wrongname").ShouldFail()
		files := helper.ListFilesInDir(commonVar.Context)
		Expect(len(files)).To(Equal(0))
	})

	When("running odo init with valid flags", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", "aname", "--devfile", "go").ShouldPass()
		})

		It("should download a devfile.yaml file", func() {
			files := helper.ListFilesInDir(commonVar.Context)
			Expect(files).To(Equal([]string{"devfile.yaml"}))
		})
	})
})
