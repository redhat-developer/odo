package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/pkg/watch"
	"github.com/redhat-developer/odo/tests/helper"
	"github.com/redhat-developer/odo/tests/integration/devfile/utils"
)

var _ = Describe("odo dev command tests", func() {
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

	When("a component is created and pushed", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			output := helper.Cmd("odo", "dev", "--project", commonVar.Project).ShouldPass().Out()
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
})
