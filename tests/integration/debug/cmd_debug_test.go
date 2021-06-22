package debug

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// since during parallel runs of cmd debug, the port might be occupied by the other tests
// we execute these tests serially
var _ = Describe("odo debug command serial tests", func() {

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

	It("should auto-select a local debug port when the given local port is occupied", func() {
		projName := helper.GetCliRunner().CreateRandNamespaceProjectOfLength(5)
		defer func() {
			helper.GetCliRunner().DeleteNamespaceProject(projName)
		}()
		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
		helper.Cmd("odo", "component", "create", "--s2i", "nodejs:latest", "nodejs-cmp", "--project", projName, "--context", commonVar.Context).ShouldPass()
		helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

		stopChannel := make(chan bool)
		go func() {
			helper.Cmd("odo", "debug", "port-forward", "--context", commonVar.Context).WithTerminate(60*time.Second, stopChannel).ShouldRun()
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

		freePort := ""
		helper.WaitForCmdOut("odo", []string{"debug", "info", "--context", commonVar.Context}, 1, false, func(output string) bool {
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
