package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile push command tests", func() {
	var project string
	var context string
	var currentWorkingDirectory string

	var sourcePath = "/projects"

	oc := helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
		currentWorkingDirectory = helper.Getwd()
		helper.Chdir(context)
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.Chdir(currentWorkingDirectory)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("Verify devfile push works", func() {

		It("Check that odo push works with a devfile", func() {
			// Devfile push requires experimental mode to be set
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)

			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project)
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
		})

	})

	Context("when devfile push command is executed", func() {

		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			// Devfile push requires experimental mode to be set
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)

			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project)
			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project)

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
			helper.ReplaceString(filepath.Join(context, "server.js"), "node listening on", "UPDATED!")

			helper.CmdShouldPass("odo", "push", "--context", filepath.Join(context, "nodejs-ex"), "--namespace", project)
		})

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
			// Devfile push requires experimental mode to be set
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

			// Create a new file that we plan on deleting later...
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs"), context)

			newFilePath := filepath.Join(context, "foobar.txt")
			if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}

			// Create a new directory
			newDirPath := filepath.Join(context, "testdir")
			helper.MakeDir(newDirPath)

			// Push
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project)

			// component name is currently equal to directory name until odo create for devfiles is implemented
			cmpName := filepath.Base(context)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameByComponent(cmpName, project)

			stdOut := oc.ExecListDir(podName, project, sourcePath)
			Expect(stdOut).To(ContainSubstring(("foobar.txt")))
			Expect(stdOut).To(ContainSubstring(("testdir")))

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project, "-v4")

			// Then check to see if it's truly been deleted
			stdOut = oc.ExecListDir(podName, project, sourcePath)
			Expect(stdOut).To(Not(ContainSubstring(("foobar.txt"))))
			Expect(stdOut).To(Not(ContainSubstring(("testdir"))))
		})

		It("should delete the files from the container if its removed locally", func() {
			// Devfile push requires experimental mode to be set
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs-multicontainer"), context)
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project)

			// component name is currently equal to directory name until odo create for devfiles is implemented
			cmpName := filepath.Base(context)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameByComponent(cmpName, project)

			var statErr error
			oc.CheckCmdOpInRemoteDevfilePod(
				podName,
				project,
				[]string{"stat", "/projects/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(os.Remove(filepath.Join(context, "server.js"))).NotTo(HaveOccurred())
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project)

			oc.CheckCmdOpInRemoteDevfilePod(
				podName,
				project,
				[]string{"stat", "/projects/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).To(HaveOccurred())
			Expect(statErr.Error()).To(ContainSubstring("cannot stat '/projects/server.js': No such file or directory"))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			// Devfile push requires experimental mode to be set
			helper.CmdShouldPass("odo", "preference", "set", "Experimental", "true")

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs-multicontainer"), context)
			helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project)

			// use the force build flag and push
			output := helper.CmdShouldPass("odo", "push", "--devfile", "devfile.yaml", "--namespace", project, "-f")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
		})
	})

})
