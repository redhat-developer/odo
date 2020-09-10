package debug

import (
	"os"
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

	var context string

	var namespace, componentName, projectDirPath, originalKubeconfig string
	var projectDir = "/projectDir"

	//  current directory and project (before eny test is run) so it can restored  after all testing is done
	var originalDir string

	// Using program command according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is before every spec (It)
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		componentName = helper.RandString(6)
		originalDir = helper.Getwd()
		helper.Chdir(context)
		projectDirPath = context + projectDir
	})

	// Clean up after the test
	// This is run after every Spec (It)
	AfterEach(func() {
		helper.Chdir(originalDir)
		cliRunner.DeleteNamespaceProject(namespace)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	It("should auto-select a local debug port when the given local port is occupied for a devfile component", func() {
		helper.MakeDir(projectDirPath)
		helper.Chdir(projectDirPath)

		helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName)
		helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
		helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(projectDirPath, "devfile-with-debugrun.yaml"))
		helper.RenameFile("devfile-with-debugrun.yaml", "devfile.yaml")
		helper.CmdShouldPass("odo", "push", "--debug")

		stopChannel := make(chan bool)
		go func() {
			helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward")
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
