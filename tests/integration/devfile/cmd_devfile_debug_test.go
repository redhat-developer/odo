package devfile

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/util"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile debug command tests", func() {
	var namespace, context, componentName, currentWorkingDirectory, projectDirPath, originalKubeconfig string
	var projectDir = "/projectDir"

	// Using program command according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))

		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		componentName = helper.RandString(6)
		helper.Chdir(context)
		projectDirPath = context + projectDir
	})

	preSetup := func() {
		helper.MakeDir(projectDirPath)
		helper.Chdir(projectDirPath)
	}

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		cliRunner.DeleteNamespaceProject(namespace)
		helper.Chdir(currentWorkingDirectory)
		err := os.Setenv("KUBECONFIG", originalKubeconfig)
		Expect(err).NotTo(HaveOccurred())
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("odo debug on a nodejs:latest component", func() {
		JustBeforeEach(func() {
			preSetup()
		})

		It("check that machine output debug information works", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName, "--context", projectDirPath)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(projectDirPath, "devfile-with-debugrun.yaml"))
			helper.RenameFile("devfile-with-debugrun.yaml", "devfile.yaml")
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDirPath)

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort, "--context", projectDirPath)
			}()

			// Make sure that the debug information output, outputs correctly.
			// We do *not* check the json output since the debugProcessID will be different each time.
			helper.WaitForCmdOut("odo", []string{"debug", "info", "-o", "json", "--context", projectDirPath}, 1, false, func(output string) bool {
				if strings.Contains(output, `"kind": "OdoDebugInfo"`) &&
					strings.Contains(output, `"localPort": `+freePort) {
					return true
				}
				return false
			})

			stopChannel <- true
		})

		It("should expect a ws connection when tried to connect on default debug port locally", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName, "--context", projectDirPath)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(projectDirPath, "devfile-with-debugrun.yaml"))
			helper.RenameFile("devfile-with-debugrun.yaml", "devfile.yaml")
			helper.CmdShouldPass("odo", "push", "--context", projectDirPath)
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDirPath)

			// check the env for the runMode
			envOutput, err := helper.ReadFile(filepath.Join(projectDirPath, ".odo/env/env.yaml"))
			Expect(err).To(BeNil())
			Expect(envOutput).To(ContainSubstring(" RunMode: debug"))

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--context", projectDirPath)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:5858", "WebSockets request was expected", 12, 5, 400)
			stopChannel <- true
		})

		It("should error out on devfile flag", func() {
			helper.CmdShouldFail("odo", "debug", "port-forward", "--context", projectDirPath, "--devfile", "invalid.yaml")
			helper.CmdShouldFail("odo", "debug", "info", "--context", projectDirPath, "--devfile", "invalid.yaml")
		})

	})

	Context("odo debug info should work on a odo component", func() {
		JustBeforeEach(func() {
			preSetup()
		})

		It("should start a debug session and run debug info on a running debug session", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs-cmp-"+namespace, "--project", namespace, "--context", projectDirPath)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(projectDirPath, "devfile-with-debugrun.yaml"))
			helper.RenameFile("devfile-with-debugrun.yaml", "devfile.yaml")
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDirPath)

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort, "--context", projectDirPath)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
			runningString := helper.CmdShouldPass("odo", "debug", "info", "--context", projectDirPath)
			Expect(runningString).To(ContainSubstring(freePort))
			Expect(helper.ListFilesInDir(os.TempDir())).To(ContainElement(namespace + "-nodejs-cmp-" + namespace + "-odo-debug.json"))
			stopChannel <- true
		})

		It("should start a debug session and run debug info on a closed debug session", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, componentName, "--context", projectDirPath)
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), projectDirPath)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(projectDirPath, "devfile-with-debugrun.yaml"))
			helper.RenameFile("devfile-with-debugrun.yaml", "devfile.yaml")
			helper.CmdShouldPass("odo", "push", "--debug", "--context", projectDirPath)

			httpPort, err := util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())
			freePort := strconv.Itoa(httpPort)

			stopChannel := make(chan bool)
			go func() {
				helper.CmdShouldRunAndTerminate(60*time.Second, stopChannel, "odo", "debug", "port-forward", "--local-port", freePort, "--context", projectDirPath)
			}()

			// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
			// We are just using this to validate if nodejs agent is listening on the other side
			helper.HttpWaitForWithStatus("http://localhost:"+freePort, "WebSockets request was expected", 12, 5, 400)
			runningString := helper.CmdShouldPass("odo", "debug", "info", "--context", projectDirPath)
			Expect(runningString).To(ContainSubstring(freePort))
			stopChannel <- true
			failString := helper.CmdShouldFail("odo", "debug", "info", "--context", projectDirPath)
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
				Expect(helper.ListFilesInDir(os.TempDir())).To(Not(ContainElement(namespace + "-app" + "-nodejs-cmp-" + namespace + "-odo-debug.json")))
			}

		})
	})
})
