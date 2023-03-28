package integration

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo analyze command tests", func() {
	for _, label := range []string{
		helper.LabelNoCluster, helper.LabelUnauth,
	} {
		label := label
		var _ = Context("label "+label, Label(label), func() {
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
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				})

				It("analyze should return correct value", func() {
					res := helper.Cmd("odo", "analyze", "-o", "json").ShouldPass()
					stdout, stderr := res.Out(), res.Err()
					Expect(stderr).To(BeEmpty())
					Expect(helper.IsJSON(stdout)).To(BeTrue())
					helper.JsonPathContentIs(stdout, "0.devfile", "nodejs")
					helper.JsonPathExist(stdout, "0.devfileRegistry")
					helper.JsonPathContentIs(stdout, "0.name", "node-echo")
					helper.JsonPathExist(stdout, "0.ports")
				})
			})

			It("analyze should fail in an empty directory", func() {
				res := helper.Cmd("odo", "analyze", "-o", "json").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(stdout).To(BeEmpty())
				Expect(helper.IsJSON(stderr)).To(BeTrue())
				helper.JsonPathContentContain(stderr, "message", "No valid devfile found for project in")
			})

			It("analyze should fail without json output", func() {
				stderr := helper.Cmd("odo", "analyze").ShouldFail().Err()
				Expect(stderr).To(ContainSubstring("this command can be run with json output only"))
			})
		})

	}

})
