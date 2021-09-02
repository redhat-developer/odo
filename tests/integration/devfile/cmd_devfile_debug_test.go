package devfile

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/openshift/odo/pkg/util"
	"github.com/tidwall/gjson"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile debug command tests", func() {
	var componentName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
		helper.SetDefaultDevfileRegistryAsStaging()
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("odo debug on a nodejs:latest component", func() {
		It("check that machine output debug information works", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, componentName, "--context", commonVar.Context).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.Cmd("odo", "push", "--debug", "--context", commonVar.Context).ShouldPass()

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.Cmd("odo", "debug", "port-forward", "--local-port", freePort, "--context", commonVar.Context).WithTerminate(60*time.Second, stopChannel).ShouldRun()
			}()

			// Make sure that the debug information output, outputs correctly.
			// We do *not* check the json output since the debugProcessID will be different each time.
			helper.WaitForCmdOut("odo", []string{"debug", "info", "-o", "json", "--context", commonVar.Context}, 1, false, func(output string) bool {
				values := gjson.GetMany(output, "kind", "spec.localPort")
				expected := []string{"OdoDebugInfo", freePort}
				return helper.GjsonMatcher(values, expected)
			})

			stopChannel <- true
		})

		It("should expect a ws connection when tried to connect on default debug port locally", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, componentName, "--context", commonVar.Context).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--debug", "--context", commonVar.Context).ShouldPass()

			// check the env for the runMode
			envOutput, err := helper.ReadFile(filepath.Join(commonVar.Context, ".odo/env/env.yaml"))
			Expect(err).To(BeNil())
			Expect(envOutput).To(ContainSubstring(" RunMode: debug"))

			stopChannel := make(chan bool)
			go func() {
				helper.Cmd("odo", "debug", "port-forward", "--context", commonVar.Context).WithTerminate(60*time.Second, stopChannel).ShouldRun()
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:5858", "WebSockets request was expected", 12, 5, 400)
			stopChannel <- true
		})

		It("should error out on devfile flag", func() {
			helper.Cmd("odo", "debug", "port-forward", "--context", commonVar.Context, "--devfile", "invalid.yaml").ShouldFail()
			helper.Cmd("odo", "debug", "info", "--context", commonVar.Context, "--devfile", "invalid.yaml").ShouldFail()
		})

	})

	Context("odo debug info should work on a odo component", func() {
		It("should start a debug session and run debug info on a running debug session", func() {
			helper.Cmd("odo", "create", "nodejs", "nodejs-cmp", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.Cmd("odo", "push", "--debug", "--context", commonVar.Context).ShouldPass()

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.Cmd("odo", "debug", "port-forward", "--local-port", freePort, "--context", commonVar.Context).WithTerminate(60*time.Second, stopChannel).ShouldRun()
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
			runningString := helper.Cmd("odo", "debug", "info", "--context", commonVar.Context).ShouldPass().Out()
			Expect(runningString).To(ContainSubstring(freePort))
			Expect(helper.ListFilesInDir(os.TempDir())).To(ContainElement(commonVar.Project + "-nodejs-cmp-odo-debug.json"))
			stopChannel <- true
		})

		It("should start a debug session and run debug info on a closed debug session", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, componentName, "--context", commonVar.Context).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.Cmd("odo", "push", "--debug", "--context", commonVar.Context).ShouldPass()

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.Cmd("odo", "debug", "port-forward", "--local-port", freePort, "--context", commonVar.Context).WithTerminate(60*time.Second, stopChannel).ShouldRun()
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
			runningString := helper.Cmd("odo", "debug", "info", "--context", commonVar.Context).ShouldPass().Out()
			Expect(runningString).To(ContainSubstring(freePort))
			stopChannel <- true
			failString := helper.Cmd("odo", "debug", "info", "--context", commonVar.Context).ShouldFail().Err()
			Expect(failString).To(ContainSubstring("not running"))

			// according to https://golang.org/pkg/os/#Signal On Windows, sending os.Interrupt to a process with os.Process.Signal is not implemented
			// discussion on the go repo https://github.com/golang/go/issues/6720
			// session.Interrupt() will not work as it internally uses syscall.SIGINT
			// thus debug port-forward won't stop running
			// the solution is to use syscall.SIGKILL for windows but this will kill the process immediately
			// and the cleaning and closing tasks for debug port-forward won't run and the debug info file won't be cleared
			// thus we skip this last check
			// CTRL_C_EVENTS from the terminal works fine https://github.com/golang/go/issues/6720#issuecomment-66087737
			// here's a hack to generate the event https://golang.org/cl/29290044
			// but the solution is unacceptable https://github.com/golang/go/issues/6720#issuecomment-66087749
			if runtime.GOOS != "windows" {
				Expect(helper.ListFilesInDir(os.TempDir())).To(Not(ContainElement(commonVar.Project + "-app" + "-nodejs-cmp-" + commonVar.Project + "-odo-debug.json")))
			}

		})
	})

	Context("when the debug command throws an error during push", func() {
		It("should wait and error out with some log", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm run debug", "npm run debugs")

			_, output := helper.Cmd("odo", "push", "--debug", "--context", commonVar.Context).ShouldPass().OutAndErr()
			helper.MatchAllInOutput(output, []string{
				"exited with error status within 1 sec",
				"Did you mean this?",
			})

			_, output = helper.Cmd("odo", "push", "--debug", "--context", commonVar.Context, "--debug-command", "debug", "-f").ShouldPass().OutAndErr()
			helper.MatchAllInOutput(output, []string{
				"exited with error status within 1 sec",
				"Did you mean this?",
			})
		})
	})
})
