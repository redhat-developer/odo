package integration

import (
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo generic", func() {
	// TODO: A neater way to provide odo path. Currently we assume \
	// odo and oc in $PATH already
	var oc helper.OcRunner
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("Check the help usage for odo", func() {

		It("Makes sure that we have the long-description when running odo and we dont error", func() {
			output := helper.Cmd("odo").ShouldPass().Out()
			Expect(output).To(ContainSubstring("To see a full list of commands, run 'odo --help'"))
		})

		It("Make sure we have the full description when performing odo --help", func() {
			output := helper.Cmd("odo", "--help").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Use \"odo [command] --help\" for more information about a command."))
		})

		It("Fail when entering an incorrect name for a component", func() {
			output := helper.Cmd("odo", "delete", "foobar").ShouldFail().Err()
			Expect(output).To(ContainSubstring("Subcommand not found, use one of the available commands"))
		})

		It("Fail with showing help only once for incorrect command", func() {
			output := helper.Cmd("odo", "hello").ShouldFail().Err()
			Expect(strings.Count(output, "odo [flags]")).Should(Equal(1))
		})

	})

	Context("When deleting two project one after the other", func() {
		It("should be able to delete sequentially", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()

			helper.DeleteProject(project2)
			helper.DeleteProject(project1)
		})
	})

	Context("When deleting three project one after the other in opposite order", func() {
		It("should be able to delete", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()
			project3 := helper.CreateRandProject()

			helper.DeleteProject(project1)
			helper.DeleteProject(project2)
			helper.DeleteProject(project3)

		})
	})

	Context("when executing odo version command", func() {
		It("should show the version of odo major components including server login URL", func() {
			odoVersion := helper.Cmd("odo", "version").ShouldPass().Out()
			reOdoVersion := regexp.MustCompile(`^odo\s*v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?\s*\(\w+\)`)
			odoVersionStringMatch := reOdoVersion.MatchString(odoVersion)
			rekubernetesVersion := regexp.MustCompile(`Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+((-\w+\.[0-9]+)?\+\w+)?`)
			kubernetesVersionStringMatch := rekubernetesVersion.MatchString(odoVersion)
			Expect(odoVersionStringMatch).Should(BeTrue())
			Expect(kubernetesVersionStringMatch).Should(BeTrue())
			serverURL := oc.GetCurrentServerURL()
			Expect(odoVersion).Should(ContainSubstring("Server: " + serverURL))
		})
	})
})
