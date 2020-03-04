package integration

import (
	"github.com/openshift/odo/pkg/util"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo debug command tests", func() {
	var project string
	var context string

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		context = helper.CreateNewContext()
		project = helper.CreateRandProject()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
	})

	// Clean up after the test
	// This is run after every Spec (It)
	AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("odo debug on a nodejs:latest component", func() {
		It("should expect a ws connection when tried to connect on different debug port locally and remotely", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs:latest", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "config", "set", "--force", "DebugPort", "9292", "--context", context)
			dbgPort := helper.GetConfigValueWithContext("DebugPort", context)
			Expect(dbgPort).To(Equal("9292"))
			helper.CmdShouldPass("odo", "push", "--context", context)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", "5050", "--context", context)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:5050", "WebSockets request was expected", 12, 5, 400)
			stopChannel <- true
		})

		It("should expect a ws connection when tried to connect on default debug port locally", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs:latest", "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--context", context)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:5858", "WebSockets request was expected", 12, 5, 400)
			stopChannel <- true
		})

	})

	Context("odo debug info should work on a odo component", func() {
		It("should start a debug session and run debug info on a running debug session", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs:latest", "nodejs-cmp-"+project, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			httpPort, err := util.HttpGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort, "--context", context)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
			runningString := helper.CmdShouldPass("odo", "debug", "info", "--context", context)
			Expect(runningString).To(ContainSubstring(freePort))
			Expect(helper.ListFilesInDir(os.TempDir())).To(ContainElement(project + "-app" + "-nodejs-cmp-" + project + "-odo-debug.json"))
			stopChannel <- true
		})

		It("should start a debug session and run debug info on a closed debug session", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs:latest", "nodejs-cmp-"+project, "--project", project, "--context", context)
			helper.CmdShouldPass("odo", "push", "--context", context)

			httpPort, err := util.HttpGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort, "--context", context)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
			runningString := helper.CmdShouldPass("odo", "debug", "info", "--context", context)
			Expect(runningString).To(ContainSubstring(freePort))
			stopChannel <- true
			failString := helper.CmdShouldPass("odo", "debug", "info", "--context", context)
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
				Expect(helper.ListFilesInDir(os.TempDir())).To(Not(ContainElement(project + "-app" + "-nodejs-cmp-" + project + "-odo-debug.json")))
			}

		})
	})
})
