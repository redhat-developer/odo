package integration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
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

	When("running odo --help", Label(helper.LabelNoCluster), func() {
		var output string
		BeforeEach(func() {
			output = helper.Cmd("odo", "--help").ShouldPass().Out()
		})
		It("retuns full help contents including usage, examples, commands, utility commands, component shortcuts, and flags sections", func() {
			helper.MatchAllInOutput(output, []string{"Usage:", "Examples:", "Main Commands:", "OpenShift Commands:", "Utility Commands:", "Flags:"})
		})

	})

	When("running odo without subcommand and flags", Label(helper.LabelNoCluster), func() {
		var output string
		BeforeEach(func() {
			output = helper.Cmd("odo").ShouldPass().Out()
		})
		It("a short vesion of help contents is returned, an error is not expected", func() {
			Expect(output).To(ContainSubstring("To see a full list of commands, run 'odo --help'"))
		})
	})

	It("returns error when using an invalid command", Label(helper.LabelNoCluster), func() {
		output := helper.Cmd("odo", "hello").ShouldFail().Err()
		Expect(output).To(ContainSubstring("Invalid command - see available commands/subcommands above"))
	})

	It("returns JSON error", Label(helper.LabelNoCluster), func() {
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

	It("returns error when using an invalid command with --help", Label(helper.LabelNoCluster), func() {
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
			reOdoVersion := `^odo\s*v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?\s*\(\w+\)`
			rekubernetesVersion := `Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+((-\w+\.[0-9]+)?\+\w+)?`
			Expect(odoVersion).Should(SatisfyAll(MatchRegexp(reOdoVersion), MatchRegexp(rekubernetesVersion)))
			serverURL := oc.GetCurrentServerURL()
			Expect(odoVersion).Should(ContainSubstring("Server: " + serverURL))
		})

		It("should show the version of odo major components", Label(helper.LabelNoCluster), func() {
			reOdoVersion := `^odo\s*v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?\s*\(\w+\)`
			Expect(odoVersion).Should(MatchRegexp(reOdoVersion))
		})
	})

	Describe("Experimental Mode", func() {
		experimentalFlag := "--run-on"

		AfterEach(func() {
			helper.ResetExperimentalMode()
		})

		It("should not list experimental flags by default", func() {
			helpOutput := helper.Cmd("odo", "help").ShouldPass().Out()
			Expect(helpOutput).ShouldNot(ContainSubstring(experimentalFlag))
		})

		Context("experimental mode has an unknown value", func() {
			for _, val := range []string{"", "false"} {
				val := val
				It("should not list experimental flags if ODO_EXPERIMENTAL is not true", func() {
					helpOutput := helper.Cmd("odo", "help").AddEnv(feature.OdoExperimentalModeEnvVar + "=" + val).ShouldPass().Out()
					Expect(helpOutput).ShouldNot(ContainSubstring(experimentalFlag))
				})
			}
		})

		When("experimental mode is enabled", func() {
			BeforeEach(func() {
				helper.EnableExperimentalMode()
			})

			AfterEach(func() {
				helper.ResetExperimentalMode()
			})

			It("experimental flags should be usable", func() {
				By("via help output", func() {
					helpOutput := helper.Cmd("odo", "help").ShouldPass().Out()
					Expect(helpOutput).Should(ContainSubstring(experimentalFlag))
				})
			})
		})
	})

})
