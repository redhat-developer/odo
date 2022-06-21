package devfile

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo logs command tests", func() {
	var componentName string
	var commonVar helper.CommonVar

	areAllPodsRunning := func() bool {
		allPodsRunning := true
		status := string(commonVar.CliRunner.Run("get", "pods", "-n", commonVar.Project, "-o", "jsonpath=\"{.items[*].status.phase}\"").Out.Contents())
		// value of status would be a string decorated by double quotes; so we ignore the first and last character
		// this could have been done using strings.TrimPrefix and strings.TrimSuffix, but that's two lines/calls.
		status = status[1 : len(status)-1]
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

	When("directory is empty", func() {

		BeforeEach(func() {
			Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
		})

		It("should error", func() {
			output := helper.Cmd("odo", "logs").ShouldFail().Err()
			Expect(output).To(ContainSubstring("this command cannot run in an empty directory"))
		})
	})

	When("component is created and odo logs is executed", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", componentName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-deploy-functional-pods.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})
		When("running in Dev mode", func() {
			var devSession helper.DevSession
			var err error

			BeforeEach(func() {
				devSession, _, _, _, err = helper.StartDevMode()
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				devSession.Kill()
				devSession.WaitEnd()
			})
			It("should successfully show logs of the running component", func() {
				// `odo logs`
				out := helper.Cmd("odo", "logs").ShouldPass().Out()
				helper.MatchAllInOutput(out, []string{"runtime:", "main:"})

				// `odo logs --dev`
				out = helper.Cmd("odo", "logs", "--dev").ShouldPass().Out()
				helper.MatchAllInOutput(out, []string{"runtime:", "main:"})

				// `odo logs --deploy`
				out = helper.Cmd("odo", "logs", "--deploy").ShouldPass().Out()
				Expect(out).To(ContainSubstring("no containers running in the specified mode for the component"))
			})
		})

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
		})

		When("running in both Dev and Deploy mode", func() {
			var devSession helper.DevSession
			var err error
			BeforeEach(func() {
				devSession, _, _, _, err = helper.StartDevMode()
				Expect(err).ToNot(HaveOccurred())
				helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
				Eventually(func() bool {
					return areAllPodsRunning()
				}).Should(Equal(true))
			})
			AfterEach(func() {
				devSession.Kill()
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
		})
	})
})
