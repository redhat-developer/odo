package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/e2e/helper"
)

// following command will tests in Describe section below in parallel (in 2 nodes)
// ginkgo -nodes=2 -focus="Example of a clean test" slowSpecThreshold=120 -randomizeAllSpecs  tests/e2e/
var _ = Describe("Config Tests", func() {
	//new clean project and context for each test
	var project string
	var context string

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalDir string
	// var originalProject string

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	var _ = BeforeEach(func() {
		project = helper.OcCreateRandProject()
		context = helper.CreateNewContext()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.OcDeleteProject(project)
		helper.DeleteDir(context)
	})

	var _ = Context("when component is in the current directory", func() {

		// we will be testing components that are created from the current directory
		// switch to the clean context dir before each test
		var _ = JustBeforeEach(func() {
			originalDir = helper.Getwd()
			helper.Chdir(context)
		})

		// go back to original directory after each test
		var _ = JustAfterEach(func() {
			helper.Chdir(originalDir)
		})

		var _ = Context("when project from KUBECONFIG is used", func() {
			It("should be checking to see if local config values are the same as the configured ones", func() {
				cases := []struct {
					paramName  string
					paramValue string
				}{
					{
						paramName:  "Type",
						paramValue: "java",
					},
					{
						paramName:  "Name",
						paramValue: "odo-java",
					},
					{
						paramName:  "MinCPU",
						paramValue: "0.2",
					},
					{
						paramName:  "MaxCPU",
						paramValue: "2",
					},
					{
						paramName:  "MinMemory",
						paramValue: "100M",
					},
					{
						paramName:  "MaxMemory",
						paramValue: "500M",
					},
					{
						paramName:  "Ports",
						paramValue: "8080/TCP,45/UDP",
					},
					{
						paramName:  "Application",
						paramValue: "odotestapp",
					},
					{
						paramName:  "Project",
						paramValue: "odotestproject",
					},
					{
						paramName:  "SourceType",
						paramValue: "git",
					},
					{
						paramName:  "Ref",
						paramValue: "develop",
					},
					{
						paramName:  "SourceLocation",
						paramValue: "https://github.com/sclorg/nodejs-ex",
					},
				}
				runCmdShouldPass(fmt.Sprintf("odo component create nodejs"))
				for _, testCase := range cases {
					runCmdShouldPass(fmt.Sprintf("odo config set %s %s -f", testCase.paramName, testCase.paramValue))
					Value := getConfigValue(testCase.paramName)
					Expect(Value).To(ContainSubstring(testCase.paramValue))
				}
			})

		})
	})
})
