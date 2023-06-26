package integration

import (
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
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	for _, label := range []string{
		helper.LabelNoCluster, helper.LabelUnauth,
	} {
		label := label
		Context("label "+label, Label(label), func() {
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
		})
	}

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

	Context("executing odo version command", func() {
		const (
			reOdoVersion        = `^odo\s*v[0-9]+.[0-9]+.[0-9]+(?:-\w+)?\s*\(\w+\)`
			reKubernetesVersion = `Kubernetes:\s*v[0-9]+.[0-9]+.[0-9]+((-\w+\.[0-9]+)?\+\w+)?`
			rePodmanVersion     = `Podman Client:\s*[0-9]+.[0-9]+.[0-9]+((-\w+\.[0-9]+)?\+\w+)?`
			reJSONVersion       = `^v{0,1}[0-9]+.[0-9]+.[0-9]+((-\w+\.[0-9]+)?\+\w+)?`
		)
		When("executing the complete command with server info", func() {
			var odoVersion string
			BeforeEach(func() {
				odoVersion = helper.Cmd("odo", "version").ShouldPass().Out()
			})
			for _, podman := range []bool{true, false} {
				podman := podman
				It("should show the version of odo major components including server login URL", helper.LabelPodmanIf(podman, func() {
					By("checking the human readable output", func() {
						Expect(odoVersion).Should(MatchRegexp(reOdoVersion))

						// odo tests setup (CommonBeforeEach) is designed in a way that if a test is labelled with 'podman', it will not have cluster configuration
						// so we only test podman info on podman labelled test, and clsuter info otherwise
						// TODO (pvala): Change this behavior when we write tests that should be tested on both podman and cluster simultaneously
						// Ref: https://github.com/redhat-developer/odo/issues/6719
						if podman {
							Expect(odoVersion).Should(MatchRegexp(rePodmanVersion))
							Expect(odoVersion).To(ContainSubstring(helper.GetPodmanVersion()))
						} else {
							Expect(odoVersion).Should(MatchRegexp(reKubernetesVersion))
							serverURL := oc.GetCurrentServerURL()
							Expect(odoVersion).Should(ContainSubstring("Server: " + serverURL))
							if !helper.IsKubernetesCluster() {
								ocpMatcher := ContainSubstring("OpenShift: ")
								if serverVersion := commonVar.CliRunner.GetVersion(); serverVersion == "" {
									// Might indicate a user permission error on certain clusters (observed with a developer account on Prow nightly jobs)
									ocpMatcher = Not(ocpMatcher)
								}
								Expect(odoVersion).Should(ocpMatcher)
							}
						}
					})

					By("checking the JSON output", func() {
						odoVersion = helper.Cmd("odo", "version", "-o", "json").ShouldPass().Out()
						Expect(helper.IsJSON(odoVersion)).To(BeTrue())
						helper.JsonPathSatisfies(odoVersion, "version", MatchRegexp(reJSONVersion))
						helper.JsonPathExist(odoVersion, "gitCommit")
						if podman {
							helper.JsonPathSatisfies(odoVersion, "podman.client.version", MatchRegexp(reJSONVersion), Equal(helper.GetPodmanVersion()))
						} else {
							helper.JsonPathSatisfies(odoVersion, "cluster.kubernetes.version", MatchRegexp(reJSONVersion))
							serverURL := oc.GetCurrentServerURL()
							helper.JsonPathContentIs(odoVersion, "cluster.serverURL", serverURL)
							if !helper.IsKubernetesCluster() {
								m := BeEmpty()
								if serverVersion := commonVar.CliRunner.GetVersion(); serverVersion != "" {
									// A blank serverVersion might indicate a user permission error on certain clusters (observed with a developer account on Prow nightly jobs)
									m = Not(m)
								}
								helper.JsonPathSatisfies(odoVersion, "cluster.openshift", m)
							}
						}
					})
				}))
			}

			for _, label := range []string{helper.LabelNoCluster, helper.LabelUnauth} {
				label := label
				It("should show the version of odo major components", Label(label), func() {
					Expect(odoVersion).Should(MatchRegexp(reOdoVersion))
				})
			}
		})

		When("podman client is bound to delay and odo version is run", Label(helper.LabelPodman), func() {
			var odoVersion string
			BeforeEach(func() {
				delayer := helper.GenerateDelayedPodman(commonVar.Context, 2)
				odoVersion = helper.Cmd("odo", "version").WithEnv("PODMAN_CMD="+delayer, "PODMAN_CMD_INIT_TIMEOUT=1s").ShouldPass().Out()
			})
			It("should not print podman version if podman cmd timeout has been reached", func() {
				Expect(odoVersion).Should(MatchRegexp(reOdoVersion))
				Expect(odoVersion).ToNot(ContainSubstring("Podman Client:"))
			})
		})
		It("should only print client info when using --client flag", func() {
			By("checking human readable output", func() {
				odoVersion := helper.Cmd("odo", "version", "--client").ShouldPass().Out()
				Expect(odoVersion).Should(MatchRegexp(reOdoVersion))
				Expect(odoVersion).ToNot(SatisfyAll(ContainSubstring("Server"), ContainSubstring("Kubernetes"), ContainSubstring("Podman Client")))
			})

			By("checking JSON output", func() {
				odoVersion := helper.Cmd("odo", "version", "--client", "-o", "json").ShouldPass().Out()
				Expect(helper.IsJSON(odoVersion)).To(BeTrue())
				helper.JsonPathSatisfies(odoVersion, "version", MatchRegexp(reJSONVersion))
				helper.JsonPathExist(odoVersion, "gitCommit")
				helper.JsonPathSatisfies(odoVersion, "cluster", BeEmpty())
				helper.JsonPathSatisfies(odoVersion, "podman", BeEmpty())
			})
		})
	})

	Describe("Experimental Mode", Label(helper.LabelNoCluster), func() {
		AfterEach(func() {
			helper.ResetExperimentalMode()
		})

		When("experimental mode is enabled", func() {
			BeforeEach(func() {
				helper.EnableExperimentalMode()
			})

			AfterEach(func() {
				helper.ResetExperimentalMode()
			})

			It("should display warning message", func() {
				out := helper.Cmd("odo", "version", "--client").ShouldPass().Out()
				Expect(out).Should(ContainSubstring("Experimental mode enabled. Use at your own risk."))
			})
		})
	})

})
