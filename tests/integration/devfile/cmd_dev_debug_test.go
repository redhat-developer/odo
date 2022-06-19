package devfile

import (
	"path/filepath"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo"
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

		It("should expect a ws connection when tried to connect on default debug port locally", func() {
			helper.MatchAllInOutput(string(stdout), []string{
				"Executing the application (command: mkdir)",
				"Executing the application (command: echo)",
				"Executing the application (command: install)",
				"Executing the application (command: start-debug)",
			})
			helper.MatchAllInOutput(string(stderr), []string{
				"Devfile command \"echo\" exited with an error status",
				"intentional-error-message",
			})

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

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://"+ports["5858"], "WebSockets request was expected", 12, 5, 400)
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
