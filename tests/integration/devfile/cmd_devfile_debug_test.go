package devfile

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/redhat-developer/odo/pkg/util"
	"github.com/tidwall/gjson"

	"github.com/redhat-developer/odo/tests/helper"

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
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("creating an application using devfile with debugrun", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, componentName, "--context", commonVar.Context, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})
		When("the debug command in the devfile is invalid", func() {
			BeforeEach(func() {
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm run debug", "npm run debugs")
			})
			It("should return an error message along with a log on odo push", func() {
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

		When("the application is pushed", func() {
			BeforeEach(func() {
				helper.Cmd("odo", "push", "--debug", "--context", commonVar.Context).ShouldPass()
			})
			It("checks that machine output debug information works", func() {
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

			It("should error out when using an invalid devfile", func() {
				By("listening on default port and forwarding to the default port", func() {
					helper.Cmd("odo", "debug", "port-forward", "--context", commonVar.Context, "--devfile", "invalid.yaml").ShouldFail()
				})
				By("getting debug session information", func() {
					helper.Cmd("odo", "debug", "info", "--context", commonVar.Context, "--devfile", "invalid.yaml").ShouldFail()
				})
			})

			It("should start a debug session and run debug info on a closed debug session", func() {
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
			It("should start a debug session and run debug info on a running debug session", func() {
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
				Expect(helper.ListFilesInDir(os.TempDir())).To(ContainElement(commonVar.Project + "-app-" + componentName + "-odo-debug.json"))
				defer helper.DeleteFile(filepath.Join(os.TempDir(), commonVar.Project+"-app-"+componentName+"-odo-debug.json"))
				stopChannel <- true
			})
		})
	})
})
