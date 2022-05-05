package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo alizer command tests", func() {
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

	When("source files are in the directory", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})

		It("alizer should return correct value", func() {
			res := helper.Cmd("odo", "alizer", "-o", "json").ShouldPass()
			stdout, stderr := res.Out(), res.Err()
			Expect(stderr).To(BeEmpty())
			Expect(helper.IsJSON(stdout)).To(BeTrue())
			helper.JsonPathContentIs(stdout, "devfile", "nodejs")
			helper.JsonPathContentIs(stdout, "devfileRegistry", "DefaultDevfileRegistry")
		})
	})

	It("alizer should fail in an empty directory", func() {
		res := helper.Cmd("odo", "alizer", "-o", "json").ShouldFail()
		stdout, stderr := res.Out(), res.Err()
		Expect(stdout).To(BeEmpty())
		Expect(helper.IsJSON(stderr)).To(BeTrue())
		helper.JsonPathContentContain(stderr, "message", "No valid devfile found for project in")
	})

	It("alizer should fail without json output", func() {
		stderr := helper.Cmd("odo", "alizer").ShouldFail().Err()
		Expect(stderr).To(ContainSubstring("this command can be run with json output only"))
	})
})
