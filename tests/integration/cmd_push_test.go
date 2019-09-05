package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo push command tests", func() {
	var project string
	var context string

	appName := "app"
	cmpName := "nodejs"

	oc = helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		project = helper.CreateRandProject()
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when push command is executed", func() {
		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context+"/nodejs-ex")
			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			url := oc.GetFirstURL(cmpName, appName, project)
			helper.ReplaceString(filepath.Join(context+"/nodejs-ex"+"/views/index.html"), "Welcome to your Node.js application on OpenShift", "UPDATED!")

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)
		})

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/ruby-ex", context+"/ruby-ex")
			helper.CmdShouldPass("odo", "component", "create", "ruby", cmpName, "--project", project, "--context", context+"/ruby-ex", "--app", appName)

			// Create a new file that we plan on deleting later...
			newFilePath := filepath.Join(context, "ruby-ex", "foobar.txt")
			if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}

			// Create a new directory
			newDirPath := filepath.Join(context, "ruby-ex", "testdir")
			helper.MakeDir(newDirPath)

			// Push
			helper.CmdShouldPass("odo", "push", "--context", context+"/ruby-ex")

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameOfComp(cmpName, project)

			envs := oc.GetEnvs(cmpName, appName, project)
			dir := fmt.Sprintf("%s/%s", envs["ODO_S2I_SRC_BIN_PATH"], "/src")

			stdOut := oc.ExecListDir(podName, project, dir)
			Expect(stdOut).To(ContainSubstring(("foobar.txt")))
			Expect(stdOut).To(ContainSubstring(("testdir")))

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push", "--context", context+"/ruby-ex", "-v4")

			// Then check to see if it's truly been deleted
			stdOut = oc.ExecListDir(podName, project, dir)
			Expect(stdOut).To(Not(ContainSubstring(("foobar.txt"))))
			Expect(stdOut).To(Not(ContainSubstring(("testdir"))))
		})

		It("should build when a new file and a new folder is added in the directory", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			newFilePath := filepath.Join(context, "nodejs-ex", "new-example.html")
			if err := helper.CreateFileWithContent(newFilePath, "<html>Hello</html>"); err != nil {
				fmt.Printf("the new-example.html file was not created, reason %v", err.Error())
			}

			output = helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp(cmpName, project)

			// verify that the new file was pushed
			envs := oc.GetEnvs(cmpName, appName, project)
			dir := fmt.Sprintf("%s/%s", envs["ODO_S2I_SRC_BIN_PATH"], "/src")
			stdOut := oc.ExecListDir(podName, project, dir)
			Expect(stdOut).To(ContainSubstring(("README.md")))

			// make a new folder and push
			helper.MakeDir(filepath.Join(context, "nodejs-ex", "exampleDir"))
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// verify that the new file was pushed
			stdOut = oc.ExecListDir(podName, project, dir)
			Expect(stdOut).To(ContainSubstring(("exampleDir")))
		})

		It("should build when a file and a folder is renamed in the directory", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// rename a file and push
			helper.RenameFile(filepath.Join(context, "/nodejs-ex", "README.md"), filepath.Join(context, "/nodejs-ex", "NEW-FILE.md"))
			output = helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp(cmpName, project)

			envs := oc.GetEnvs(cmpName, appName, project)
			dir := fmt.Sprintf("%s/%s", envs["ODO_S2I_SRC_BIN_PATH"], "/src")

			// verify that the new file was pushed
			stdOut := oc.ExecListDir(podName, project, dir)

			Expect(stdOut).To(Not(ContainSubstring("README.md")))

			Expect(stdOut).To(ContainSubstring("NEW-FILE.md"))

			// rename a folder and push
			helper.RenameFile(filepath.Join(context, "/nodejs-ex", "/tests"), filepath.Join(context, "/nodejs-ex", "/testing"))
			output = helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// verify that the new file was pushed
			stdOut = oc.ExecListDir(podName, project, dir)

			Expect(stdOut).To(Not(ContainSubstring("tests")))

			Expect(stdOut).To(ContainSubstring("testing"))
		})

		It("should not build when changes are detected in a ignored file", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)

			// create the .odoignore file and push
			ignoreFilePath := filepath.Join(context, "nodejs-ex", ".odoignore")
			if err := helper.CreateFileWithContent(ignoreFilePath, ".git\n*.md"); err != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// modify a ignored file and push
			helper.ReplaceString(filepath.Join(context+"/nodejs-ex"+"/README.md"), "This example will serve a welcome page", "This is a example welcome page!")
			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// test ignores using the flag
			output = helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex", "--ignore", "*.md")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context+"/nodejs-ex")

			// use the force build flag and push
			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex", "-f")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
		})

		It("should push only the modified files", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			url := oc.GetFirstURL(cmpName, appName, project)
			// Wait for running app before getting info about files.
			// During the startup sequence there is something that will modify the access time of a source file.
			helper.HttpWaitFor("http://"+url, "Welcome to your Node.js", 30, 1)

			earlierCatServerFile := ""
			oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				project,
				[]string{"stat", "/tmp/src/server.js"},
				func(cmdOp string, err error) bool {
					earlierCatServerFile = cmdOp
					return true
				},
			)

			earlierCatViewFile := ""
			oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				project,
				[]string{"stat", "/tmp/src/views/index.html"},
				func(cmdOp string, err error) bool {
					earlierCatViewFile = cmdOp
					return true
				},
			)

			helper.ReplaceString(filepath.Join(context+"/nodejs-ex"+"/views/index.html"), "Welcome to your Node.js application on OpenShift", "UPDATED!")
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)

			modifiedCatViewFile := ""
			oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				project,
				[]string{"stat", "/tmp/src/views/index.html"},
				func(cmdOp string, err error) bool {
					modifiedCatViewFile = cmdOp
					return true
				},
			)

			modifiedCatServerFile := ""
			oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				project,
				[]string{"stat", "/tmp/src/server.js"},
				func(cmdOp string, err error) bool {
					modifiedCatServerFile = cmdOp
					return true
				},
			)

			Expect(modifiedCatViewFile).NotTo(Equal(earlierCatViewFile))
			Expect(modifiedCatServerFile).To(Equal(earlierCatServerFile))
		})
	})

	Context("when .odoignore file exists", func() {
		It("should create and push the contents of a named component excluding the contents in .odoignore file", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			ignoreFilePath := filepath.Join(context, "nodejs-ex", ".odoignore")
			if err := helper.CreateFileWithContent(ignoreFilePath, ".git\ntests/\nREADME.md"); err != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}

			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--project", project, "--context", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp("nodejs", project)

			envs := oc.GetEnvs(cmpName, appName, project)
			dir := fmt.Sprintf("%s/%s", envs["ODO_S2I_SRC_BIN_PATH"], "/src")

			// verify that the views folder got pushed
			stdOut1 := oc.ExecListDir(podName, project, dir)
			Expect(stdOut1).To(ContainSubstring("views"))

			// verify that the tests was not pushed
			stdOut2 := oc.ExecListDir(podName, project, dir)
			Expect(stdOut2).To(Not(ContainSubstring(("tests"))))

			// verify that the README.md file was not pushed
			stdOut3 := oc.ExecListDir(podName, project, dir)
			Expect(stdOut3).To(Not(ContainSubstring(("README.md"))))

		})
	})

	Context("when running odo push with flag --show-log", func() {
		It("should be able to spam odo push without anything breaking", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--project", project, "--context", context+"/nodejs-ex")
			// Iteration 1
			output := helper.CmdShouldPass("odo", "push", "--show-log", "--context", context+"/nodejs-ex")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
			// Iteration 2
			output = helper.CmdShouldPass("odo", "push", "--show-log", "--context", context+"/nodejs-ex")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
			// Iteration 3
			output = helper.CmdShouldPass("odo", "push", "--show-log", "--context", context+"/nodejs-ex")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
		})
	})
})
