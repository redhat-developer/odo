package integration

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
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
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("running odo --help", func() {
		var output string
		BeforeEach(func() {
			output = helper.Cmd("odo", "--help").ShouldPass().Out()
		})
		It("retuns full help contents including usage, examples, commands, utility commands, component shortcuts, and flags sections", func() {
			helper.MatchAllInOutput(output, []string{"Usage:", "Examples:", "Main Commands:", "OpenShift Commands:", "Utility Commands:", "Flags:"})
		})

	})

	When("running odo without subcommand and flags", func() {
		var output string
		BeforeEach(func() {
			output = helper.Cmd("odo").ShouldPass().Out()
		})
		It("a short vesion of help contents is returned, an error is not expected", func() {
			Expect(output).To(ContainSubstring("To see a full list of commands, run 'odo --help'"))
		})
	})

	It("returns error when using an invalid command", func() {
		output := helper.Cmd("odo", "hello").ShouldFail().Err()
		Expect(output).To(ContainSubstring("Invalid command - see available commands/subcommands above"))
	})

	It("returns JSON error", func() {
		By("using an invalid command with JSON output", func() {
			res := helper.Cmd("odo", "unknown-command", "-o", "json").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(stdout).To(BeEmpty())
			Expect(helper.IsJSON(stderr)).To(BeTrue())
		})

		By("using an invalid describe sub-command with JSON output", func() {
			res := helper.Cmd("odo", "describe", "unknown-sub-command", "-o", "json").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(stdout).To(BeEmpty())
			Expect(helper.IsJSON(stderr)).To(BeTrue())
		})

		By("using an invalid list sub-command with JSON output", func() {
			res := helper.Cmd("odo", "list", "unknown-sub-command", "-o", "json").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(stdout).To(BeEmpty())
			Expect(helper.IsJSON(stderr)).To(BeTrue())
		})

		By("omitting required subcommand with JSON output", func() {
			res := helper.Cmd("odo", "describe", "-o", "json").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(stdout).To(BeEmpty())
			Expect(helper.IsJSON(stderr)).To(BeTrue())
		})
	})

	It("returns error when using an invalid command with --help", func() {
		output := helper.Cmd("odo", "hello", "--help").ShouldFail().Err()
		Expect(output).To(ContainSubstring("unknown command 'hello', type --help for a list of all commands"))
	})

	Context("When deleting two project one after the other", func() {
		It("should be able to delete sequentially", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()

			helper.DeleteProject(project1)
			helper.DeleteProject(project2)
		})
		It("should be able to delete them in any order", func() {
			project1 := helper.CreateRandProject()
			project2 := helper.CreateRandProject()
			project3 := helper.CreateRandProject()

			helper.DeleteProject(project2)
			helper.DeleteProject(project1)
			helper.DeleteProject(project3)
		})
	})

	When("executing odo version command", func() {
		var odoVersion string
		BeforeEach(func() {
			odoVersion = helper.Cmd("odo", "version").ShouldPass().Out()
		})

		It("should show the version of odo major components including server login URL", func() {
			reOdoVersion := regexp.MustCompile(`^odo\s*v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?\s*\(\w+\)`)
			odoVersionStringMatch := reOdoVersion.MatchString(odoVersion)
			Expect(odoVersionStringMatch).Should(BeTrue())
			if !helper.IsKubernetesCluster() {
				rekubernetesVersion := regexp.MustCompile(`Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+((-\w+\.[0-9]+)?\+\w+)?`)
				kubernetesVersionStringMatch := rekubernetesVersion.MatchString(odoVersion)
				Expect(kubernetesVersionStringMatch).Should(BeTrue())
				serverURL := oc.GetCurrentServerURL()
				Expect(odoVersion).Should(ContainSubstring("Server: " + serverURL))
			}
		})
	})

})
