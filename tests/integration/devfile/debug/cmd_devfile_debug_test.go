package debug

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// since during parallel runs of cmd devfile debug, the port might be occupied by the other tests
// we execute these tests serially
var _ = Describe("odo devfile debug command serial tests", func() {

	var componentName, projectDirPath string
	var projectDir = "/projectDir"

	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		componentName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
		projectDirPath = commonVar.Context + projectDir
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	It("should auto-select a local debug port when the given local port is occupied for a devfile component", func() {
		helper.MakeDir(projectDirPath)
		helper.Chdir(projectDirPath)

		helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, componentName).ShouldPass()
		helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(projectDirPath, "devfile-with-debugrun.yaml"))
		helper.RenameFile("devfile-with-debugrun.yaml", "devfile.yaml")
		helper.Cmd("odo", "push", "--debug").ShouldPass()

		stopChannel := make(chan bool)
		go func() {
			helper.Cmd("odo", "debug", "port-forward").WithTerminate(60*time.Second, stopChannel).ShouldRun()
		}()

		stopListenerChan := make(chan bool)
		startListenerChan := make(chan bool)
		listenerStarted := false
		go func() {
			defer GinkgoRecover()
			err := testingutil.FakePortListener(startListenerChan, stopListenerChan, envinfo.DefaultDebugPort)
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
		helper.WaitForCmdOut("odo", []string{"debug", "info"}, 1, false, func(output string) bool {
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
