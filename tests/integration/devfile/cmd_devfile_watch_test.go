package devfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/pkg/watch"
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"
)

var _ = Describe("odo devfile watch command tests", func() {
	var namespace, context, cmpName, currentWorkingDirectory, originalKubeconfig string

	// Using program commmand according to cliRunner in devfile
	cliRunner := helper.GetCliRunner()

	// Setup up state for each test spec
	// create new project (not set as active) and new context directory for each test spec
	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		originalKubeconfig = os.Getenv("KUBECONFIG")
		helper.LocalKubeconfigSet(context)
		namespace = cliRunner.CreateRandNamespaceProject()
		currentWorkingDirectory = helper.Getwd()
		cmpName = helper.RandString(6)
		helper.Chdir(context)
	})

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

	Context("when running help for watch command", func() {
		It("should display the help", func() {
			appHelp := helper.CmdShouldPass("odo", "watch", "-h")
			helper.MatchAllInOutput(appHelp, []string{"Watch for changes", "git components"})
		})
	})

	Context("when executing watch without pushing a devfile component", func() {
		It("should fail", func() {
			helper.Chdir(currentWorkingDirectory)
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, "--context", context, cmpName)
			output := helper.CmdShouldFail("odo", "watch", "--context", context)
			Expect(output).To(ContainSubstring("component does not exist. Please use `odo push` to create your component"))
		})

		It("should error out on devfile flag", func() {
			helper.CmdShouldFail("odo", "watch", "--devfile", "invalid.yaml")
		})
	})

	Context("when executing odo watch after odo push", func() {
		It("should listen for file changes", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			watchFlag := ""
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing devrun command"},
			}
			// odo watch and validate
			utils.OdoWatch(utils.OdoV1Watch{}, odoV2Watch, namespace, context, watchFlag, cliRunner, "kube")
		})
	})

	Context("when executing odo watch after odo push with flag commands", func() {
		It("should listen for file changes", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--build-command", "build", "--run-command", "run", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			watchFlag := "--build-command build --run-command run"
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing build command", "Executing run command"},
			}
			// odo watch and validate
			utils.OdoWatch(utils.OdoV1Watch{}, odoV2Watch, namespace, context, watchFlag, cliRunner, "kube")
		})
	})

	Context("when executing odo watch", func() {
		It("should show validation errors if the devfile is incorrect", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			session := helper.CmdRunner("odo", "watch", "--context", context)
			defer session.Kill()

			helper.WaitForOutputToContain("Waiting for something to change", 180, 10, session)

			helper.ReplaceString(filepath.Join(context, "devfile.yaml"), "kind: build", "kind: run")

			helper.WaitForOutputToContain(watch.PushErrorString, 180, 10, session)

		})
	})

	Context("when executing odo watch", func() {
		It("should use the index information from previous push operation", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			// 1) Push a generic project
			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// 2) Create a new file A
			fileAPath, fileAText := createSimpleFile(context)

			// 3) Odo watch that project
			session := helper.CmdRunner("odo", "watch", "--context", context)
			defer session.Kill()

			helper.WaitForOutputToContain("Waiting for something to change", 180, 10, session)

			// 4) Change some other file B
			helper.ReplaceString(filepath.Join(context, "server.js"), "App started", "App is super started")
			helper.WaitForOutputToContain("Executing devrun command", 180, 10, session)

			podName := cliRunner.GetRunningPodNameByComponent(cmpName, namespace)

			// File should exist, and its content should match what we initially set it to
			execResult := cliRunner.Exec(podName, namespace, "cat", "/projects/"+filepath.Base(fileAPath))
			Expect(execResult).To(ContainSubstring(fileAText))

		})

		It("should listen for file changes with delay set to 0", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			watchFlag := "--delay 0"
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing devrun command"},
			}
			// odo watch and validate
			utils.OdoWatch(utils.OdoV1Watch{}, odoV2Watch, namespace, context, watchFlag, cliRunner, "kube")
		})

	})

	Context("when executing odo watch after odo push with debug flag", func() {
		It("should be able to start a debug session after push with debug flag using odo watch and revert back after normal push", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(context, "devfile.yaml"))

			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// push with debug flag
			output = helper.CmdShouldPass("odo", "push", "--debug", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			watchFlag := ""
			// check if the normal debugRun command was executed
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing debugrun command"},
			}
			// odo watch and validate if we can port forward successfully
			utils.OdoWatchWithDebug(odoV2Watch, context, watchFlag)

			// revert to normal odo push
			output = helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// check if the normal run command was executed
			odoV2Watch = utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing devrun command"},
			}
			utils.OdoWatch(utils.OdoV1Watch{}, odoV2Watch, namespace, context, watchFlag, cliRunner, "kube")
		})
	})

	Context("when executing odo watch", func() {
		It("ensure that index information is updated by watch", func() {
			helper.CmdShouldPass("odo", "create", "nodejs", "--project", namespace, cmpName)

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context, "devfile.yaml"))

			// 1) Push a generic project
			output := helper.CmdShouldPass("odo", "push", "--project", namespace)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			indexAfterPush, err := util.ReadFileIndex(filepath.Join(context, ".odo", "odo-file-index.json"))
			Expect(err).ToNot(HaveOccurred())

			// 2) Odo watch that project
			session := helper.CmdRunner("odo", "watch", "--context", context)
			defer session.Kill()

			helper.WaitForOutputToContain("Waiting for something to change", 180, 10, session)

			// 3) Create a new file A
			fileAPath, _ := createSimpleFile(context)

			// 4) Wait for the new file to exist in the index
			Eventually(func() bool {

				newIndexAfterPush, err := util.ReadFileIndex(filepath.Join(context, ".odo", "odo-file-index.json"))
				if err != nil {
					fmt.Fprintln(GinkgoWriter, "New index not found or could not be read", err)
					return false
				}

				_, exists := newIndexAfterPush.Files[filepath.Base(fileAPath)]
				if !exists {
					fmt.Fprintln(GinkgoWriter, "path", fileAPath, "not found.", err)
				}
				return exists

			}, 180, 10).Should(Equal(true))

			// 5) Delete file A and verify that it disappears from the index
			err = os.Remove(fileAPath)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {

				newIndexAfterPush, err := util.ReadFileIndex(filepath.Join(context, ".odo", "odo-file-index.json"))
				if err != nil {
					fmt.Fprintln(GinkgoWriter, "New index not found or could not be read", err)
					return false
				}

				// Santity test: at least one file should be present
				if len(newIndexAfterPush.Files) == 0 {
					return false
				}

				// The fileA file should NOT be found
				match := false
				for relativeFilePath := range newIndexAfterPush.Files {

					if strings.Contains(relativeFilePath, filepath.Base(fileAPath)) {
						match = true
					}
				}
				return !match

			}, 180, 10).Should(Equal(true))

			// 6) Change server.js
			helper.ReplaceString(filepath.Join(context, "server.js"), "App started", "App is super started")
			helper.WaitForOutputToContain("server.js", 180, 10, session)

			// 7) Wait for the size values in the old and new index files to differ, indicating that watch has updated the index
			Eventually(func() bool {

				newIndexAfterPush, err := util.ReadFileIndex(filepath.Join(context, ".odo", "odo-file-index.json"))
				if err != nil {
					fmt.Fprintln(GinkgoWriter, "New index not found or could not be read", err)
					return false
				}

				beforePushValue, exists := indexAfterPush.Files["server.js"]
				if !exists {
					fmt.Fprintln(GinkgoWriter, "server.js not found in old index file")
					return false
				}

				afterPushValue, exists := newIndexAfterPush.Files["server.js"]
				if !exists {
					fmt.Fprintln(GinkgoWriter, "server.js not found in new index file")
					return false
				}

				fmt.Fprintln(GinkgoWriter, "comparing old and new file sizes", beforePushValue.Size, afterPushValue.Size)

				return beforePushValue.Size != afterPushValue.Size

			}, 180, 10).Should(Equal(true))

		})
	})

})

func createSimpleFile(context string) (string, string) {

	textFilePath := filepath.Join(context, "my-file-"+helper.RandString(10)+".txt")
	textOne := []byte(helper.RandString(10))
	err := ioutil.WriteFile(textFilePath, textOne, 0644)
	Expect(err).NotTo(HaveOccurred())

	return textFilePath, string(textOne)
}
