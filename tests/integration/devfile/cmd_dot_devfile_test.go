package devfile

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"

	// We continued iterating on bracket pair guides. Horizontal lines now outline the scope of a bracket pair. Also, vertical lines now depend on the indentation of the code that is surrounded by the bracket pair.. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("Test suits to check .devfile.yaml compatibility", func() {
	var cmpName string
	var commonVar helper.CommonVar

	BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
	})

	AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("Creating a nodejs component and replace devfile.yaml to .devfile.yaml", func() {
		var _ = BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-.yaml")).ShouldPass()
			helper.CreateLocalEnv(commonVar.Context, cmpName, commonVar.Project)
			helper.Cmd("mv", "devfile.yaml", ".devfile.yaml").ShouldPass()
		})

		When("Creating url and doing odo push", func() {
			var stdout, url1, host string

			BeforeEach(func() {
				url1 = helper.RandString(6)
				host = helper.RandString(6)
				helper.Cmd("odo", "url", "create", url1, "--port", "9090", "--host", host, "--secure", "--ingress").ShouldPass()
				helper.Cmd("odo", "push").ShouldPass()
			})

			It("should verify if url is created and pushed", func() {
				stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
				helper.MatchAllInOutput(stdout, []string{url1, "Pushed", "true", "ingress"})
			})
			When("Deleting url doing odo push", func() {

				BeforeEach(func() {
					helper.Cmd("odo", "url", "delete", url1, "-f").ShouldPass()
				})

				It("should verify if url is created and pushed", func() {
					stdout = helper.Cmd("odo", "url", "list").ShouldPass().Out()
					helper.MatchAllInOutput(stdout, []string{url1, "Locally Deleted", "true", "ingress"})
				})
			})
		})
	})
})
