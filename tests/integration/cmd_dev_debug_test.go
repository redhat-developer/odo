package integration

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo dev debug command tests", func() {
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("a component is bootstrapped", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})

		When("running odo dev with debug flag", func() {
			var devSession helper.DevSession
			var ports map[string]string
			BeforeEach(func() {
				var err error
				devSession, _, _, ports, err = helper.StartDevMode("--debug")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				devSession.Kill()
				devSession.WaitEnd()
			})
			It("should expect a ws connection when tried to connect on default debug port locally", func() {
				// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
				// We are just using this to validate if nodejs agent is listening on the other side
				helper.HttpWaitForWithStatus("http://"+ports["5858"], "WebSockets request was expected", 12, 5, 400)
			})
		})
	})

	When("a composite command is used as debug command", func() {
		const devfileCmpName = "nodejs"
		var session helper.DevSession
		var stdout []byte
		var stderr []byte
		var ports map[string]string
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeRunAndDebug.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			var err error
			session, stdout, stderr, ports, err = helper.StartDevMode("--debug")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			session.Stop()
			session.WaitEnd()
		})

		It("should run successfully", func() {
			By("verifying from the output that all commands have been executed", func() {
				helper.MatchAllInOutput(string(stdout), []string{
					"Building your application in container on cluster",
					"Executing the application (command: mkdir)",
					"Executing the application (command: echo)",
					"Executing the application (command: install)",
					"Executing the application (command: start-debug)",
				})
			})

			By("verifying that any command that did not succeed in the middle has logged such information correctly", func() {
				helper.MatchAllInOutput(string(stderr), []string{
					"Devfile command \"echo\" exited with an error status",
					"intentional-error-message",
				})
			})

			By("building the application only once", func() {
				// Because of the Spinner, the "Building your application in container on cluster" is printed twice in the captured stdout.
				// The bracket allows to match the last occurrence with the command execution timing information.
				Expect(strings.Count(string(stdout), "Building your application in container on cluster (command: install) [")).
					To(BeNumerically("==", 1))
			})

			By("verifying that the command did run successfully", func() {
				// Verify the command executed successfully
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				res := commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
					podName,
					"runtime",
					commonVar.Project,
					[]string{"stat", "/projects/testfolder"},
					func(cmdOp string, err error) bool {
						return err == nil
					},
				)
				Expect(res).To(BeTrue())
			})

			By("expecting a ws connection when tried to connect on default debug port locally", func() {
				// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
				// We are just using this to validate if nodejs agent is listening on the other side
				helper.HttpWaitForWithStatus("http://"+ports["5858"], "WebSockets request was expected", 12, 5, 400)
			})
		})
	})

	When("running build and debug commands as composite in different containers and a shared volume", func() {
		const devfileCmpName = "nodejs"
		var session helper.DevSession
		var stdout []byte
		var stderr []byte
		var ports map[string]string
		BeforeEach(func() {
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfileCompositeBuildRunDebugInMultiContainersAndSharedVolume.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			var err error
			session, stdout, stderr, ports, err = helper.StartDevMode("--debug")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			session.Stop()
			session.WaitEnd()
		})

		It("should run successfully", func() {
			By("verifying from the output that all commands have been executed", func() {
				helper.MatchAllInOutput(string(stdout), []string{
					"Building your application in container on cluster (command: mkdir)",
					"Building your application in container on cluster (command: sleep-cmd-build)",
					"Building your application in container on cluster (command: build-cmd)",
					"Executing the application (command: sleep-cmd-run)",
					"Executing the application (command: echo-with-error)",
					"Executing the application (command: check-build-result)",
					"Executing the application (command: start-debug)",
				})
			})

			By("verifying that any command that did not succeed in the middle has logged such information correctly", func() {
				helper.MatchAllInOutput(string(stderr), []string{
					"Devfile command \"echo-with-error\" exited with an error status",
					"intentional-error-message",
				})
			})

			By("building the application only once per exec command in the build command", func() {
				// Because of the Spinner, the "Building your application in container on cluster" is printed twice in the captured stdout.
				// The bracket allows to match the last occurrence with the command execution timing information.
				out := string(stdout)
				for _, cmd := range []string{"mkdir", "sleep-cmd-build", "build-cmd"} {
					Expect(strings.Count(out, fmt.Sprintf("Building your application in container on cluster (command: %s) [", cmd))).
						To(BeNumerically("==", 1))
				}
			})

			By("verifying that the command did run successfully", func() {
				// Verify the command executed successfully
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				res := commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
					podName,
					"runtime",
					commonVar.Project,
					[]string{"stat", "/projects/testfolder"},
					func(cmdOp string, err error) bool {
						return err == nil
					},
				)
				Expect(res).To(BeTrue())
			})

			By("expecting a ws connection when tried to connect on default debug port locally", func() {
				// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
				// We are just using this to validate if nodejs agent is listening on the other side
				helper.HttpWaitForWithStatus("http://"+ports["5858"], "WebSockets request was expected", 12, 5, 400)
			})
		})
	})

	When("a component without debug command is bootstrapped", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-without-debugrun.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})

		It("should fail running odo dev --debug", func() {
			output := helper.Cmd("odo", "dev", "--debug").ShouldFail().Err()
			Expect(output).To(ContainSubstring("no command of kind \"debug\" found in the devfile"))
		})
	})
})
