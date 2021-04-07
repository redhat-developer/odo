package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo debug command tests", func() {
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	Context("odo debug on a nodejs:latest component", func() {
		It("check that machine output debug information works", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs:latest", "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort, "--context", commonVar.Context)
			}()

			// Make sure that the debug information output, outputs correctly.
			// We do *not* check the json output since the debugProcessID will be different each time.
			helper.WaitForCmdOut("odo", []string{"debug", "info", "--context", commonVar.Context, "-o", "json"}, 1, false, func(output string) bool {
				if strings.Contains(output, `"kind": "OdoDebugInfo"`) &&
					strings.Contains(output, `"localPort": `+freePort) {
					return true
				}
				return false
			})

			stopChannel <- true
		})

		It("should expect a ws connection when tried to connect on different debug port locally and remotely", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs:latest", "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "config", "set", "--force", "DebugPort", "9292", "--context", commonVar.Context)
			dbgPort := helper.GetConfigValueWithContext("DebugPort", commonVar.Context)
			Expect(dbgPort).To(Equal("9292"))
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", "5050", "--context", commonVar.Context)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:5050", "WebSockets request was expected", 12, 5, 400)
			stopChannel <- true
		})

	})

	Context("odo debug info should work on a odo component", func() {
		It("should start a debug session and run debug info on a running debug session", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs:latest", "nodejs-cmp-"+commonVar.Project, "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort, "--context", commonVar.Context)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
			runningString := helper.CmdShouldPass("odo", "debug", "info", "--context", commonVar.Context)
			Expect(runningString).To(ContainSubstring(freePort))
			Expect(helper.ListFilesInDir(os.TempDir())).To(ContainElement(commonVar.Project + "-app" + "-nodejs-cmp-" + commonVar.Project + "-odo-debug.json"))
			stopChannel <- true
		})

		It("should start a debug session and run debug info on a closed debug session", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs:latest", "nodejs-cmp-"+commonVar.Project, "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort, "--context", commonVar.Context)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
			runningString := helper.CmdShouldPass("odo", "debug", "info", "--context", commonVar.Context)
			Expect(runningString).To(ContainSubstring(freePort))
			stopChannel <- true
			failString := helper.CmdShouldFail("odo", "debug", "info", "--context", commonVar.Context)
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
})
