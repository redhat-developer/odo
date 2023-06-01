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

		for _, podman := range []bool{false} { // TODO add true
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

				It("should execute a command", func() {
					platform := "cluster"
					if podman {
						platform = "podman"
					}
					output := helper.Cmd("odo", "run", "create-file", "--platform", platform).ShouldPass().Out()
					Expect(output).To(ContainSubstring("Executing command in container (command: create-file)"))
				})

			}))
		}
	})
})
