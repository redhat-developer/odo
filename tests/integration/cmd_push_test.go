package integration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/odo/pkg/devfile/convert"
	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo push command tests", func() {
	var oc helper.OcRunner
	var commonVar helper.CommonVar
	appName := "app"
	cmpName := "nodejs"

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		oc = helper.NewOcRunner("oc")
		commonVar = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	// Timeout not respected by devfile https://github.com/openshift/odo/issues/4529
	// Context("Check pod timeout", func() {

	// 	It("Check that pod timeout works and we time out immediately..", func() {
	// 		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
	// 		helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
	// 		helper.CmdShouldPass("odo", "preference", "set", "PushTimeout", "1")
	// 		output := helper.CmdShouldFail("odo", "push", "--context", commonVar.Context)
	// 		Expect(output).To(ContainSubstring("waited 1s but couldn't find running pod matching selector"))
	// 	})

	// })

	Context("Test push outside of the current working direcory", func() {

		// Change to "outside" the directory before running the below tests
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})

		It("Push, modify a file and then push outside of the working directory", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// Create a new file to test propagating changes
			newFilePath := filepath.Join(commonVar.Context, "foobar.txt")
			if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}

			// Test propagating changes
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// Delete the file and check that the file is deleted
			helper.DeleteDir(newFilePath)

			// Test propagating deletions
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
		})

	})

	Context("when push command is executed", func() {

		It("should be able to create a file, push, delete, then push again propagating the deletions and build", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			output := helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// Create a new file that we plan on deleting later...
			newFilePath := filepath.Join(commonVar.Context, "foobar.txt")
			if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}

			// Create a new directory
			newDirPath := filepath.Join(commonVar.Context, "testdir")
			helper.MakeDir(newDirPath)

			// Push
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameByComponent(cmpName, commonVar.Project)

			stdOut := oc.ExecListDir(podName, commonVar.Project, convert.DefaultSourceMappingS2i)
			helper.MatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.Cmd("odo", "push", "--context", commonVar.Context, "-v4").ShouldPass()

			// Then check to see if it's truly been deleted
			stdOut = oc.ExecListDir(podName, commonVar.Project, convert.DefaultSourceMappingS2i)
			helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
		})

		It("should build when a file and a folder is renamed in the directory", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName).ShouldPass()

			// create a file and folder then push
			err := os.MkdirAll(filepath.Join(commonVar.Context, "tests"), 0750)
			Expect(err).To(BeNil())
			testFile, err := os.Create(filepath.Join(commonVar.Context, "README.md"))
			testFile.Close()
			Expect(err).To(BeNil())

			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			output := helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// rename a file and push
			helper.RenameFile(filepath.Join(commonVar.Context, "README.md"), filepath.Join(commonVar.Context, "NEW-FILE.md"))
			output = helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// get the name of running pod
			podName := oc.GetRunningPodNameByComponent(cmpName, commonVar.Project)

			// verify that the new file was pushed
			stdOut := oc.ExecListDir(podName, commonVar.Project, convert.DefaultSourceMappingS2i)

			Expect(stdOut).To(Not(ContainSubstring("README.md")))

			Expect(stdOut).To(ContainSubstring("NEW-FILE.md"))

			// rename a folder and push
			helper.RenameFile(filepath.Join(commonVar.Context, "tests"), filepath.Join(commonVar.Context, "testing"))
			output = helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// verify that the new file was pushed
			stdOut = oc.ExecListDir(podName, commonVar.Project, convert.DefaultSourceMappingS2i)

			Expect(stdOut).To(Not(ContainSubstring("tests")))

			Expect(stdOut).To(ContainSubstring("testing"))
		})

		It("should push only the modified files", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs:latest", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			url := oc.GetFirstURL(cmpName, appName, commonVar.Project)

			// Wait for running app before getting info about files.
			// During the startup sequence there is something that will modify the access time of a source file.
			helper.HttpWaitFor("http://"+url, "Hello world from node.js!", 30, 1)
			earlierCatServerFile := ""
			earlierCatServerFile = helper.StatFileInPodContainer(oc, cmpName, convert.ContainerName, appName, commonVar.Project, filepath.ToSlash(filepath.Join(convert.DefaultSourceMappingS2i, "server.js")))

			earlierCatPackageFile := ""
			earlierCatPackageFile = helper.StatFileInPodContainer(oc, cmpName, convert.ContainerName, appName, commonVar.Project, filepath.ToSlash(filepath.Join(convert.DefaultSourceMappingS2i, "package.json")))

			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "Hello world from node.js!", "UPDATED!")
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)

			modifiedCatPackageFile := ""
			modifiedCatPackageFile = helper.StatFileInPodContainer(oc, cmpName, convert.ContainerName, appName, commonVar.Project, filepath.ToSlash(filepath.Join(convert.DefaultSourceMappingS2i, "package.json")))

			modifiedCatServerFile := ""
			modifiedCatServerFile = helper.StatFileInPodContainer(oc, cmpName, convert.ContainerName, appName, commonVar.Project, filepath.ToSlash(filepath.Join(convert.DefaultSourceMappingS2i, "server.js")))

			Expect(modifiedCatPackageFile).To(Equal(earlierCatPackageFile))
			Expect(modifiedCatServerFile).NotTo(Equal(earlierCatServerFile))
		})

	})

	Context("when .odoignore file exists", func() {
		// works
		It("should create and push the contents of a named component excluding the contents and changes detected in .odoignore file", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			ignoreFilePath := filepath.Join(commonVar.Context, ".odoignore")
			if err := helper.CreateFileWithContent(ignoreFilePath, ".git\n*.md"); err != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}
			_, err := os.Create(filepath.Join(commonVar.Context, "README.md"))
			Expect(err).To(BeNil())

			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

			// get the name of running pod
			podName := oc.GetRunningPodNameByComponent("nodejs", commonVar.Project)

			// verify that the server file got pushed
			stdOut1 := oc.ExecListDir(podName, commonVar.Project, convert.DefaultSourceMappingS2i)
			Expect(stdOut1).To(ContainSubstring("server.js"))

			// verify that the README.md file was not pushed
			stdOut3 := oc.ExecListDir(podName, commonVar.Project, convert.DefaultSourceMappingS2i)
			Expect(stdOut3).To(Not(ContainSubstring(("README.md"))))

			// modify a ignored file and push
			helper.ReplaceString(filepath.Join(commonVar.Context, "README.md"), "", "This is a example welcome page!")
			output := helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// test ignores using the flag
			output = helper.Cmd("odo", "push", "--context", commonVar.Context, "--ignore", "*.md").ShouldPass().Out()
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
		})
	})

	Context("when .gitignore file exists or not", func() {
		It("should create and push the contents of a named component and include odo-file-index.json path to .gitignore file to exclude the contents, if does not exists create one", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName).ShouldPass()

			// push and include the odo-file-index.json path to .gitignore file
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
			ignoreFilePath := filepath.Join(commonVar.Context, ".gitignore")
			helper.FileShouldContainSubstring(ignoreFilePath, filepath.Join(".odo", "odo-file-index.json"))
		})
	})

	Context("when running odo push with flag --show-log", func() {
		It("should be able to execute odo push consecutively without breaking anything", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context).ShouldPass()

			// Run odo push in consecutive iteration
			output := helper.Cmd("odo", "push", "--show-log", "--context", commonVar.Context).ShouldPass().Out()
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			for i := 0; i <= 1; i++ {
				output := helper.Cmd("odo", "push", "--show-log", "--context", commonVar.Context).ShouldPass().Out()
				Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
			}
		})
	})
})
