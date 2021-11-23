package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/pkg/watch"
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"
)

var _ = Describe("odo devfile watch command tests", func() {
	var cmpName string
	var commonVar helper.CommonVar

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		cmpName = helper.RandString(6)
		helper.Chdir(commonVar.Context)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("running help for watch command", func() {
		It("should display the help", func() {
			appHelp := helper.Cmd("odo", "watch", "-h").ShouldPass().Out()
			helper.MatchAllInOutput(appHelp, []string{"Watch for changes", "git components"})
		})
	})

	When("executing watch without pushing a devfile component", func() {
		BeforeEach(func() {
			helper.Chdir(commonVar.OriginalWorkingDirectory)
			helper.Cmd("odo", "create", "--project", commonVar.Project, "--context", commonVar.Context, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass()
		})
		It("should fail", func() {
			output := helper.Cmd("odo", "watch", "--context", commonVar.Context).ShouldFail().Err()
			Expect(output).To(ContainSubstring("component does not exist. Please use `odo push` to create your component"))
		})
		It("should error out on devfile flag", func() {
			helper.Cmd("odo", "watch", "--devfile", "invalid.yaml").ShouldFail()
		})
	})

	When("executing odo watch", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})
		It("should listen for file changes", func() {
			watchFlag := ""
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing devrun command"},
			}
			// odo watch and validate
			utils.OdoWatch(odoV2Watch, commonVar.Project, commonVar.Context, watchFlag, commonVar.CliRunner, "kube")
		})
		It("should show validation errors if the devfile is incorrect", func() {
			session := helper.CmdRunner("odo", "watch")
			defer session.Kill()
			helper.WaitForOutputToContain("Waiting for something to change", 180, 10, session)
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "kind: build", "kind: run")
			helper.WaitForOutputToContain(watch.PushErrorString, 180, 10, session)
		})
		It("should use the index information from previous push operation", func() {
			// Create a new file A
			fileAPath, fileAText := helper.CreateSimpleFile(commonVar.Context, "my-file-", ".txt")

			// Odo watch that project
			session := helper.CmdRunner("odo", "watch")
			defer session.Kill()

			helper.WaitForOutputToContain("Waiting for something to change", 180, 10, session)

			// Change some other file B
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "App started", "App is super started")
			helper.WaitForOutputToContain("Executing devrun command", 180, 10, session)

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

			// File should exist, and its content should match what we initially set it to
			execResult := commonVar.CliRunner.Exec(podName, commonVar.Project, "cat", "/projects/"+filepath.Base(fileAPath))
			Expect(execResult).To(ContainSubstring(fileAText))
		})
		It("should listen for file changes with delay set to 0", func() {

			watchFlag := "--delay 0"
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing devrun command"},
			}
			// odo watch and validate
			utils.OdoWatch(odoV2Watch, commonVar.Project, commonVar.Context, watchFlag, commonVar.CliRunner, "kube")
		})
		It("ensure that index information is updated by watch", func() {
			indexAfterPush, err := util.ReadFileIndex(filepath.Join(commonVar.Context, ".odo", "odo-file-index.json"))
			Expect(err).ToNot(HaveOccurred())

			// Odo watch that project
			session := helper.CmdRunner("odo", "watch")
			defer session.Kill()

			helper.WaitForOutputToContain("Waiting for something to change", 180, 10, session)

			// Create a new file A
			fileAPath, _ := helper.CreateSimpleFile(commonVar.Context, "my-file-", ".txt")

			// Wait for the new file to exist in the index
			Eventually(func() bool {

				newIndexAfterPush, readErr := util.ReadFileIndex(filepath.Join(commonVar.Context, ".odo", "odo-file-index.json"))
				if readErr != nil {
					fmt.Fprintln(GinkgoWriter, "New index not found or could not be read", readErr)
					return false
				}

				_, exists := newIndexAfterPush.Files[filepath.Base(fileAPath)]
				if !exists {
					fmt.Fprintln(GinkgoWriter, "path", fileAPath, "not found.", readErr)
				}
				return exists

			}, 180, 10).Should(Equal(true))

			// Delete file A and verify that it disappears from the index
			err = os.Remove(fileAPath)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {

				newIndexAfterPush, err := util.ReadFileIndex(filepath.Join(commonVar.Context, ".odo", "odo-file-index.json"))
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

			// Change server.js
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "App started", "App is super started")
			helper.WaitForOutputToContain("server.js", 180, 10, session)

			// Wait for the size values in the old and new index files to differ, indicating that watch has updated the index
			Eventually(func() bool {

				newIndexAfterPush, err := util.ReadFileIndex(filepath.Join(commonVar.Context, ".odo", "odo-file-index.json"))
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

	When("executing odo watch", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			output := helper.Cmd("odo", "push", "--build-command", "build", "--run-command", "run", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})
		It("should be able to ignore the specified file, .git and odo-file-index.json ", func() {
			watchFlag := "--ignore doignoreme.txt"
			odoV2Watch := utils.OdoV2Watch{
				CmpName:               cmpName,
				StringsToBeMatched:    []string{"donotignoreme.txt changed", "Executing devbuild command", "Executing devrun command"},
				StringsNotToBeMatched: []string{"doignoreme.txt changed", "odo-file-index.json changed", ".git/index changed"},
			}
			// odo watch and validate
			utils.OdoWatchWithIgnore(odoV2Watch, commonVar.Context, watchFlag)
		})

		It("should listen for file changes", func() {
			watchFlag := "--build-command build --run-command run"
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing build command", "Executing run command"},
			}
			// odo watch and validate
			utils.OdoWatch(odoV2Watch, commonVar.Project, commonVar.Context, watchFlag, commonVar.CliRunner, "kube")
		})
	})

	When("executing odo watch after odo push with debug flag", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "create", cmpName, "--project", commonVar.Project, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml")).ShouldPass()

			// push with debug flag
			output := helper.Cmd("odo", "push", "--debug", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})
		It("should be able to start a debug session after push with debug flag using odo watch and revert back after normal push", func() {
			watchFlag := ""
			// check if the normal debugRun command was executed
			odoV2Watch := utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing debugrun command"},
			}
			// odo watch and validate if we can port forward successfully
			utils.OdoWatchWithDebug(odoV2Watch, commonVar.Context, watchFlag)

			// check the --debug-command flag
			watchFlag = "--debug-command debug"
			odoV2Watch.StringsToBeMatched = []string{"Executing debug command"}

			// odo watch and validate if we can port forward successfully
			utils.OdoWatchWithDebug(odoV2Watch, commonVar.Context, watchFlag)

			// revert to normal odo push
			watchFlag = ""
			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// check if the normal run command was executed
			odoV2Watch = utils.OdoV2Watch{
				CmpName:            cmpName,
				StringsToBeMatched: []string{"Executing devbuild command", "Executing devrun command"},
			}
			utils.OdoWatch(odoV2Watch, commonVar.Project, commonVar.Context, watchFlag, commonVar.CliRunner, "kube")

			// check that the --debug-command fails when the component is not pushed using debug mode
			output = helper.Cmd("odo", "watch", "--debug-command", "debug").WithRetry(1, 1).ShouldFail().Err()
			Expect(output).To(ContainSubstring("please start the component in debug mode"))
		})

	})

})
