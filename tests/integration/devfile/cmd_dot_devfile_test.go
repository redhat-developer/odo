package devfile

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	//We continued iterating on bracket pair guides. Horizontal lines now outline the scope of a bracket pair. Also, vertical lines now depend on the indentation of the code that is surrounded by the bracket pair.. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
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

		When("creating a storage", func() {
			var storageName, pathName, size, stdOut string
			BeforeEach(func() {
				storageName = helper.RandString(5)
				pathName = "/data"
				size = "5Gi"
				helper.Cmd("odo", "storage", "create", storageName, "--path", pathName, "--size", size, "--context", commonVar.Context).ShouldPass()
			})
			It("should list the storage with the proper states and container names", func() {
				stdOut = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
				helper.MatchAllInOutput(stdOut, []string{storageName, pathName, size, "Not Pushed", cmpName})
			})
			When("doing odo push with storage", func() {

				BeforeEach(func() {
					helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
				})
				It("should list the storage with the proper states and container names", func() {
					stdOut = helper.Cmd("odo", "storage", "list", "--context", commonVar.Context).ShouldPass().Out()
					helper.MatchAllInOutput(stdOut, []string{storageName, pathName, "Pushed", cmpName})
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

	When("Creating a nodejs component and replace devfile.yaml with .devfile.yaml", func() {
		var _ = BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("mv", "devfile.yaml", ".devfile.yaml").ShouldPass()
		})
		When("creating a service", func() {
			var redisOperator string
			var operandName string
			var odoArgs []string
			var operators []string
			BeforeEach(func() {
				odoArgs = []string{"catalog", "list", "services"}
				operators = []string{"redis-operator"}
				for _, operator := range operators {
					helper.WaitForCmdOut("odo", odoArgs, 5, true, func(output string) bool {
						return strings.Contains(output, operator)
					})
				}
				commonVar.CliRunner.CreateSecret("redis-secret", "password", commonVar.Project)
				operators := helper.Cmd("odo", "catalog", "list", "services").ShouldPass().Out()
				redisOperator = regexp.MustCompile(`redis-operator\.*[a-z][0-9]\.[0-9]\.[0-9]`).FindString(operators)
				operandName = helper.RandString(10)
				helper.Cmd("odo", "service", "create", fmt.Sprintf("%s/Redis", redisOperator), operandName, "--context", commonVar.Context).ShouldPass()
			})

			AfterEach(func() {
				helper.Cmd("odo", "service", "delete", fmt.Sprintf("Redis/%s", operandName), "-f", "--context", commonVar.Context).ShouldPass().Out()
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
			})

			When("odo push is executed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
					name := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					Expect(name).To(Not(BeEmpty()))
				})

				It("should find files in component container", func() {
					helper.Cmd("odo", "exec", "--context", commonVar.Context, "--", "ls", "/project/server.js").ShouldPass()
				})

				It("should create pods in running state", func() {
					commonVar.CliRunner.PodsShouldBeRunning(commonVar.Project, fmt.Sprintf(`%s-0`, operandName))
				})

				It("should list the service", func() {
					stdOut := helper.Cmd("odo", "service", "list", "--context", commonVar.Context).ShouldPass().Out()
					Expect(stdOut).To(ContainSubstring(fmt.Sprintf("Redis/%s", operandName)))
				})

				When("a link between the component and the service is created", func() {

					BeforeEach(func() {
						helper.Cmd("odo", "link", fmt.Sprintf("Redis/%s", operandName), "--context", commonVar.Context).ShouldPass()
					})

					It("should run odo push successfully", func() {
						helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
					})
				})
			})
		})
	})
})
