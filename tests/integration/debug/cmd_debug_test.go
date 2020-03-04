package debug

import (
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/tests/helper"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// since during parallel runs of cmd debug, the port might be occupied by the other tests
// we execute these tests serially
var _ = Describe("odo debug command serial tests", func() {

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

	It("should auto-select a local debug port when the given local port is occupied", func() {
		helper.CopyExample(filepath.Join("source", "nodejs"), context)
		helper.CmdShouldPass("odo", "component", "create", "nodejs:latest", "nodejs-cmp-"+project, "--project", project, "--context", context)
		helper.CmdShouldPass("odo", "push", "--context", context)

		stopChannel := make(chan bool)
		go func() {
			helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--context", context)
		}()

		stopListenerChan := make(chan bool)
		startListenerChan := make(chan bool)
		listenerStarted := false
		go func() {
			err := testingutil.FakePortListener(startListenerChan, stopListenerChan, config.DefaultDebugPort)
			if err != nil {
				close(startListenerChan)
				Expect(err).Should(BeNil())
			}
		}()
		// wait for the test server to start listening
		if <-startListenerChan {
			listenerStarted = true
		}

		debugInfoString := helper.CmdShouldPass("odo", "debug", "info", "--context", context)
		Expect(debugInfoString).To(ContainSubstring(""))

		freePort := ""
		helper.WaitForCmdOut("odo", []string{"debug", "info", "--context", context}, 1, true, func(output string) bool {
			if strings.Contains(output, "Debug is running") {
				splits := strings.SplitN(output, ":", 2)
				Expect(len(splits)).To(Equal(2))
				freePort = strings.TrimSpace(splits[1])
				_, err := strconv.Atoi(freePort)
				Expect(err).NotTo(HaveOccurred())
				return true
			}
			return false
		})

		// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
		// We are just using this to validate if nodejs agent is listening on the other side
		helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
		stopChannel <- true
		if listenerStarted == true {
			stopListenerChan <- true
		} else {
			close(stopListenerChan)
		}
	})
})
