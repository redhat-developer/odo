package devfile

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	//We continued iterating on bracket pair guides. Horizontal lines now outline the scope of a bracket pair. Also, vertical lines now depend on the indentation of the code that is surrounded by the bracket pair.. "github.com/onsi/gomega"
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
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
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

	When("creating and pushing with --debug a nodejs component with debhug run", func() {
		var projectDir string
		BeforeEach(func() {
			projectDir = filepath.Join(commonVar.Context, "projectDir")
			helper.CopyExample(filepath.Join("source", "web-nodejs-sample"), projectDir)
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--context", projectDir, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml")).ShouldPass()
			helper.Cmd("pwd").ShouldPass()
			helper.Cmd("mv", fmt.Sprint(projectDir, "/devfile.yaml"), fmt.Sprint(projectDir, "/.devfile.yaml")).ShouldPass()
			helper.Cmd("odo", "push", "--debug", "--context", projectDir).ShouldPass()
		})
		It("should log debug command output", func() {
			output := helper.Cmd("odo", "log", "--debug", "--context", projectDir).ShouldPass().Out()
			Expect(output).To(ContainSubstring("ODO_COMMAND_DEBUG"))
		})
	})
})
