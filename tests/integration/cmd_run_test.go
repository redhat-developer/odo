package integration

import (
	"path/filepath"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo run command tests", func() {
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		_ = cmpName // TODO remove when used
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("directory is empty", Label(helper.LabelNoCluster), func() {
		BeforeEach(func() {
			Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
		})

		It("should error", func() {
			output := helper.Cmd("odo", "run", "my-command").ShouldFail().Err()
			Expect(output).To(ContainSubstring("The current directory does not represent an odo component"))
		})
	})

	When("a component is bootstrapped", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-for-run.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})

		It("should fail if command is not found in devfile", Label(helper.LabelNoCluster), func() {
			output := helper.Cmd("odo", "run", "unknown-command").ShouldFail().Err()
			Expect(output).To(ContainSubstring(`no command named "unknown-command" found in the devfile`))

		})

		It("should fail if platform is not available", Label(helper.LabelNoCluster), func() {
			By("failing when trying to run on default platform", func() {
				output := helper.Cmd("odo", "run", "build").ShouldFail().Err()
				Expect(output).To(ContainSubstring(`unable to access the cluster`))

			})
			By("failing when trying to run on cluster", func() {
				output := helper.Cmd("odo", "run", "build", "--platform", "cluster").ShouldFail().Err()
				Expect(output).To(ContainSubstring(`unable to access the cluster`))

			})
			By("failing when trying to run on podman", func() {
				output := helper.Cmd("odo", "run", "build", "--platform", "podman").ShouldFail().Err()
				Expect(output).To(ContainSubstring(`unable to access podman`))
			})
		})

		It("should fail if odo dev is not running", func() {
			output := helper.Cmd("odo", "run", "build").ShouldFail().Err()
			Expect(output).To(ContainSubstring(`unable to get pod for component`))
			Expect(output).To(ContainSubstring(`Please check the command 'odo dev' is running`))
		})

		for _, podman := range []bool{false, true} {
			podman := podman
			When("odo dev is executed and ready", helper.LabelPodmanIf(podman, func() {

				var devSession helper.DevSession

				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					})
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should execute commands", func() {
					platform := "cluster"
					if podman {
						platform = "podman"
					}

					By("executing an exec command", func() {
						output := helper.Cmd("odo", "run", "list-files", "--platform", platform).ShouldPass().Out()
						Expect(output).To(ContainSubstring("etc"))
					})

					By("executing an exec command in another container", func() {
						output := helper.Cmd("odo", "run", "list-files-in-other-container", "--platform", platform).ShouldPass().Out()
						Expect(output).To(ContainSubstring("etc"))
					})

					if !podman {
						By("executing apply command on Kubernetes component", func() {
							output := helper.Cmd("odo", "run", "deploy-config", "--platform", platform).ShouldPass().Out()
							Expect(output).To(ContainSubstring("Creating resource ConfigMap/my-config"))
							out := commonVar.CliRunner.Run("get", "configmap", "my-config", "-n",
								commonVar.Project).Wait().Out.Contents()
							Expect(out).To(ContainSubstring("my-config"))
						})
					}

					if podman {
						By("executing apply command on Image component", func() {
							// Will fail because Dockerfile is not present, but we just want to check the build is started
							// We cannot use PODMAN_CMD=echo with --platform=podman
							output := helper.Cmd("odo", "run", "build-image", "--platform", platform).ShouldFail().Out()
							Expect(output).To(ContainSubstring("Building image locally"))
						})
					} else {
						By("executing apply command on Image component", func() {
							output := helper.Cmd("odo", "run", "build-image", "--platform", platform).AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
							Expect(output).To(ContainSubstring("Building image locally"))
							Expect(output).To(ContainSubstring("Pushing image to container registry"))

						})
					}
				})
			}))
		}
	})
})
