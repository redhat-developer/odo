package devfile

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
	"github.com/tidwall/gjson"
)

var _ = Describe("odo devfile env command tests", func() {
	const (
		testName      = "testname"
		testProject   = "testproject"
		testDebugPort = "8888"
		fakeParameter = "fakeParameter"
	)

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		helper.SetDefaultDevfileRegistryAsStaging()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("creating a component", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", "nodejs", "acomponentname", "--project", commonVar.Project).ShouldPass()
		})

		It("Should fail to set and unset an invalid parameter", func() {
			helper.Cmd("odo", "env", "set", fakeParameter, fakeParameter, "-f").ShouldFail()
			helper.Cmd("odo", "env", "unset", fakeParameter, "-f").ShouldFail()
		})

		It("should show all default parameters with odo view info", func() {
			output := helper.Cmd("odo", "env", "view").ShouldPass().Out()
			wantOutput := []string{
				"PARAMETER NAME",
				"PARAMETER VALUE",
				"NAME",
				"acomponentname",
				"Project",
				commonVar.Project,
				"DebugPort",
				"Application",
				"app",
			}
			helper.MatchAllInOutput(output, wantOutput)
		})

		It("should show all default parameters with odo view info and JSON output ", func() {
			output := helper.Cmd("odo", "env", "view", "-o", "json").ShouldPass().Out()
			values := gjson.GetMany(output, "kind", "spec.name", "spec.project", "spec.appName")
			expected := []string{"EnvInfo", "acomponentname", commonVar.Project, "app"}
			Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
		})

		When("executing env set", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "env", "set", "Name", testName, "-f").ShouldPass()
				helper.Cmd("odo", "env", "set", "Project", testProject, "-f").ShouldPass()
				helper.Cmd("odo", "env", "set", "DebugPort", testDebugPort, "-f").ShouldPass()
			})

			It("should successfully view the parameters", func() {
				output := helper.Cmd("odo", "env", "view").ShouldPass().Out()
				wantOutput := []string{
					"PARAMETER NAME",
					"PARAMETER VALUE",
					"NAME",
					testName,
					"Project",
					testProject,
					"DebugPort",
					testDebugPort,
					"Application",
					"app",
				}
				helper.MatchAllInOutput(output, wantOutput)
			})

			It("should successfully view the parameters with JSON output", func() {
				output := helper.Cmd("odo", "env", "view", "-o", "json").ShouldPass().Out()
				values := gjson.GetMany(output, "kind", "spec.name", "spec.project", "spec.debugPort")
				expected := []string{"EnvInfo", testName, testProject, testDebugPort}
				Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
			})

			When("unsetting a parameter", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "env", "unset", "DebugPort", "-f").ShouldPass()
				})

				It("should not show the parameter", func() {
					output := helper.Cmd("odo", "env", "view").ShouldPass().Out()
					dontWantOutput := []string{
						testDebugPort,
					}
					helper.DontMatchAllInOutput(output, dontWantOutput)
					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
				})

				It("should not show the parameter in JSON output", func() {
					output := helper.Cmd("odo", "env", "view", "-o", "json").ShouldPass().Out()
					values := gjson.GetMany(output, "kind", "spec.debugPort")
					expected := []string{"EnvInfo", ""}
					Expect(helper.GjsonMatcher(values, expected)).To(Equal(true))
				})
			})
		})
	})
})
