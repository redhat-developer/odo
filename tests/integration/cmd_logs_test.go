package integration

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

func getLogCommand(podman bool, otherArgs ...string) *helper.CmdWrapper {
	args := []string{"logs"}
	if len(otherArgs) > 0 {
		args = append(args, otherArgs...)
	}
	if podman {
		args = append(args, "--platform", "podman")
	}
	cmd := helper.Cmd("odo", args...)
	if podman {
		cmd = cmd.AddEnv("ODO_EXPERIMENTAL_MODE=true")
	}
	return cmd
}

var _ = Describe("odo logs command tests", func() {
	var componentName string
	var commonVar helper.CommonVar

	areAllPodsRunning := func() bool {
		allPodsRunning := true
		status := string(commonVar.CliRunner.Run("get", "pods", "-n", commonVar.Project, "-o", "jsonpath=\"{.items[*].status.phase}\"").Out.Contents())
		status = strings.Trim(status, "\"")
		split := strings.Split(status, " ")
		for i := 0; i < len(split); i++ {
			if split[i] != "Running" {
				allPodsRunning = false
			}
		}
		return allPodsRunning
	}

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("in a devfile directory", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", componentName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
		})
		When("not connected to any cluster or podman", Label(helper.LabelNoCluster), func() {
			It("odo logs should fail with an error message", func() {
				cmd := helper.Cmd("odo", "logs")
				stderr := cmd.ShouldFail().Err()
				Expect(stderr).To(ContainSubstring("you need access to a cluster"))
			})

			It("odo logs --platform podman should fail with an error message", func() {
				os.Setenv("PODMAN_CMD", "false")
				defer os.Unsetenv("PODMAN_CMD")
				cmd := getLogCommand(true)
				stderr := cmd.ShouldFail().Err()
				Expect(stderr).To(ContainSubstring("you need access to podman"))
			})
		})

		When("podman is not installed", func() {
		})
	})

	for _, podman := range []bool{false, true} {
		podman := podman
		When("directory is empty", helper.LabelPodmanIf(podman, func() {

			BeforeEach(func() {
				Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
			})

			It("should error", func() {
				cmd := getLogCommand(podman)
				output := cmd.ShouldFail().Err()
				Expect(output).To(ContainSubstring("this command cannot run in an empty directory"))
			})
		}))

		When("odo logs is executed for a component that's not running in any modes", helper.LabelPodmanIf(podman, func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.Cmd("odo", "init", "--name", componentName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-deploy-functional-pods.yaml")).ShouldPass()
				Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
			})
			It("should print that no containers are running", func() {
				noContainersRunning := "no containers running in the specified mode for the component"
				cmd := getLogCommand(podman)
				out := cmd.ShouldPass().Out()
				Expect(out).To(ContainSubstring(noContainersRunning))
				cmd = getLogCommand(podman, "--follow")
				out = cmd.ShouldPass().Out()
				Expect(out).To(ContainSubstring(noContainersRunning))
			})
		}))
	}

	When("component is created and odo logs is executed", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", componentName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-deploy-functional-pods.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})

		for _, podman := range []bool{false, true} {
			podman := podman
			When("running in Dev mode", helper.LabelPodmanIf(podman, func() {
				var devSession helper.DevSession
				var err error

				BeforeEach(func() {
					devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					})
					Expect(err).ToNot(HaveOccurred())
					if !podman {
						// We need to wait for the pod deployed as a Kubernetes component
						Eventually(func() bool {
							return areAllPodsRunning()
						}).Should(Equal(true))
					}
				})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})
				It("should successfully show logs of the running component", func() {
					expectedInDev := []string{"runtime:", "main:"}
					if podman {
						// TODO(feloy): Kubernetes components are not deployed on dev for now
						expectedInDev = []string{"runtime:"}
					}
					// `odo logs`
					cmd := getLogCommand(podman)
					out := cmd.ShouldPass().Out()
					helper.MatchAllInOutput(out, expectedInDev)

					// `odo logs --dev`
					cmd = getLogCommand(podman, "--dev")
					out = cmd.ShouldPass().Out()
					helper.MatchAllInOutput(out, expectedInDev)

					// `odo logs --deploy`
					cmd = getLogCommand(podman, "--deploy")
					out = cmd.ShouldPass().Out()
					Expect(out).To(ContainSubstring("no containers running in the specified mode for the component"))
				})
				When("--follow flag is specified", func() {
					var logsSession helper.LogsSession
					var err error

					BeforeEach(func() {
						logsSession, _, _, err = helper.StartLogsFollow(podman, "--dev")
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						logsSession.Kill()
					})
					It("should successfully follow logs of running component", func() {
						var linesOfLogs int
						Consistently(func() bool {
							logs := logsSession.OutContents()
							if len(string(logs)) < linesOfLogs {
								return false
							}
							linesOfLogs = len(logs)
							return true
						}, 20*time.Second, 5).Should(BeTrue())
					})
				})
			}))
		}

		When("running in Deploy mode", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
				Eventually(func() bool {
					return areAllPodsRunning()
				}).Should(Equal(true))
			})
			It("should successfully show logs of the running component", func() {
				// `odo logs`
				out := helper.Cmd("odo", "logs").ShouldPass().Out()
				helper.MatchAllInOutput(out, []string{"main:", "main[1]:", "main[2]:"})

				// `odo logs --dev`
				out = helper.Cmd("odo", "logs", "--dev").ShouldPass().Out()
				Expect(out).To(ContainSubstring("no containers running in the specified mode for the component"))

				// `odo logs --deploy`
				out = helper.Cmd("odo", "logs", "--deploy").ShouldPass().Out()
				helper.MatchAllInOutput(out, []string{"main:", "main[1]:", "main[2]:"})
			})
			When("--follow flag is specified", func() {
				var logsSession helper.LogsSession
				var err error

				BeforeEach(func() {
					logsSession, _, _, err = helper.StartLogsFollow(false, "--deploy")
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					logsSession.Kill()
				})
				It("should successfully follow logs of running component", func() {
					var linesOfLogs int
					Consistently(func() bool {
						logs := logsSession.OutContents()
						if len(string(logs)) < linesOfLogs {
							return false
						}
						linesOfLogs = len(logs)
						return true
					}, 20*time.Second, 5).Should(BeTrue())
				})
			})
		})

		When("running in both Dev and Deploy mode", func() {
			var devSession helper.DevSession
			var err error
			BeforeEach(func() {
				devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
				helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
				Eventually(func() bool {
					return areAllPodsRunning()
				}).Should(Equal(true))
			})
			AfterEach(func() {
				devSession.Stop()
				devSession.WaitEnd()
			})
			It("should successfully show logs of the running component", func() {
				// `odo logs`
				out := helper.Cmd("odo", "logs").ShouldPass().Out()
				helper.MatchAllInOutput(out, []string{"runtime", "main:", "main[1]:", "main[2]:", "main[3]:"})

				// `odo logs --dev`
				out = helper.Cmd("odo", "logs", "--dev").ShouldPass().Out()
				helper.MatchAllInOutput(out, []string{"runtime:", "main:"})

				// `odo logs --deploy`
				out = helper.Cmd("odo", "logs", "--deploy").ShouldPass().Out()
				helper.MatchAllInOutput(out, []string{"main:", "main[1]:", "main[2]:"})

				// `odo logs --dev --deploy`
				out = helper.Cmd("odo", "logs", "--deploy", "--dev").ShouldFail().Err()
				Expect(out).To(ContainSubstring("pass only one of --dev or --deploy flags; pass no flag to see logs for both modes"))
			})
			When("--follow flag is specified", func() {
				var logsSession helper.LogsSession
				var err error

				BeforeEach(func() {
					logsSession, _, _, err = helper.StartLogsFollow(false)
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					logsSession.Kill()
				})
				It("should successfully follow logs of running component", func() {
					var linesOfLogs int
					Consistently(func() bool {
						logs := logsSession.OutContents()
						if len(string(logs)) < linesOfLogs {
							return false
						}
						linesOfLogs = len(logs)
						return true
					}, 20*time.Second, 5).Should(BeTrue())
				})
			})
		})
	})
})
