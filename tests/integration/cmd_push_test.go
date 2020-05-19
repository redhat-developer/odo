package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/openshift/odo/tests/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo push command tests", func() {

	var globals helper.Globals

	appName := "app"
	cmpName := "nodejs"

	oc := helper.NewOcRunner("oc")

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		globals = helper.CommonBeforeEach()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEeach(globals)
	})

	Context("Check pod timeout", func() {

		It("Check that pod timeout works and we time out immediately..", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "preference", "set", "PushTimeout", "1")
			output := helper.CmdShouldFail("odo", "push", "--context", globals.Context+"/nodejs-ex")
			Expect(output).To(ContainSubstring("waited 1s but couldn't find running pod matching selector"))
		})

	})

	Context("Check for label propagation after pushing", func() {

		It("Check for labels", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context, "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			// Check for all the labels
			oc.VerifyLabelExistsOfComponent(cmpName, globals.Project, "app:"+appName)
			oc.VerifyLabelExistsOfComponent(cmpName, globals.Project, "app.kubernetes.io/part-of:"+appName)
			oc.VerifyLabelExistsOfComponent(cmpName, globals.Project, "app.kubernetes.io/managed-by:odo")

			// Check for the version
			versionInfo := helper.CmdShouldPass("odo", "version")
			re := regexp.MustCompile(`v[0-9]\S*`)
			odoVersionString := re.FindStringSubmatch(versionInfo)
			oc.VerifyLabelExistsOfComponent(cmpName, globals.Project, "app.kubernetes.io/managed-by-version:"+odoVersionString[0])
		})
	})

	Context("Test push outside of the current working direcory", func() {

		var _ = BeforeEach(func() {
			helper.Chdir(globals.Context)
		})

		It("Push, modify a file and then push outside of the working directory", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context, "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			// Create a new file to test propagating changes
			newFilePath := filepath.Join(globals.Context, "nodejs-ex", "foobar.txt")
			if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}

			// Test propagating changes
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			// Delete the file and check that the file is deleted
			helper.DeleteDir(newFilePath)

			// Test propagating deletions
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
		})

	})

	Context("when push command is executed", func() {
		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", globals.Context+"/nodejs-ex")
			output := helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			url := oc.GetFirstURL(cmpName, appName, globals.Project)
			helper.ReplaceString(filepath.Join(globals.Context, "nodejs-ex", "views", "index.html"), "Welcome to your Node.js application on OpenShift", "UPDATED!")

			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)
		})

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)

			// Create a new file that we plan on deleting later...
			newFilePath := filepath.Join(globals.Context, "nodejs-ex", "foobar.txt")
			if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}

			// Create a new directory
			newDirPath := filepath.Join(globals.Context, "nodejs-ex", "testdir")
			helper.MakeDir(newDirPath)

			// Push
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameOfComp(cmpName, globals.Project)

			envs := oc.GetEnvs(cmpName, appName, globals.Project)
			dir := envs["ODO_S2I_DEPLOYMENT_DIR"]

			stdOut := oc.ExecListDir(podName, globals.Project, dir)
			Expect(stdOut).To(ContainSubstring(("foobar.txt")))
			Expect(stdOut).To(ContainSubstring(("testdir")))

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex", "-v4")

			// Then check to see if it's truly been deleted
			stdOut = oc.ExecListDir(podName, globals.Project, dir)
			Expect(stdOut).To(Not(ContainSubstring(("foobar.txt"))))
			Expect(stdOut).To(Not(ContainSubstring(("testdir"))))
		})

		It("should build when a new file and a new folder is added in the directory", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			output := helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			newFilePath := filepath.Join(globals.Context, "nodejs-ex", "new-example.html")
			if err := helper.CreateFileWithContent(newFilePath, "<html>Hello</html>"); err != nil {
				fmt.Printf("the new-example.html file was not created, reason %v", err.Error())
			}

			output = helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp(cmpName, globals.Project)

			// verify that the new file was pushed
			envs := oc.GetEnvs(cmpName, appName, globals.Project)
			dir := envs["ODO_S2I_DEPLOYMENT_DIR"]
			stdOut := oc.ExecListDir(podName, globals.Project, dir)
			Expect(stdOut).To(ContainSubstring(("README.md")))

			// make a new folder and push
			helper.MakeDir(filepath.Join(globals.Context, "nodejs-ex", "exampleDir"))
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")

			// verify that the new file was pushed
			stdOut = oc.ExecListDir(podName, globals.Project, dir)
			Expect(stdOut).To(ContainSubstring(("exampleDir")))
		})

		It("should build when a file and a folder is renamed in the directory", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			output := helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// rename a file and push
			helper.RenameFile(filepath.Join(globals.Context, "nodejs-ex", "README.md"), filepath.Join(globals.Context, "nodejs-ex", "NEW-FILE.md"))
			output = helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp(cmpName, globals.Project)

			envs := oc.GetEnvs(cmpName, appName, globals.Project)
			dir := envs["ODO_S2I_DEPLOYMENT_DIR"]

			// verify that the new file was pushed
			stdOut := oc.ExecListDir(podName, globals.Project, dir)

			Expect(stdOut).To(Not(ContainSubstring("README.md")))

			Expect(stdOut).To(ContainSubstring("NEW-FILE.md"))

			// rename a folder and push
			helper.RenameFile(filepath.Join(globals.Context, "nodejs-ex", "tests"), filepath.Join(globals.Context, "nodejs-ex", "testing"))
			output = helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// verify that the new file was pushed
			stdOut = oc.ExecListDir(podName, globals.Project, dir)

			Expect(stdOut).To(Not(ContainSubstring("tests")))

			Expect(stdOut).To(ContainSubstring("testing"))
		})

		It("should not build when changes are detected in a ignored file", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)

			// create the .odoignore file and push
			ignoreFilePath := filepath.Join(globals.Context, "nodejs-ex", ".odoignore")
			if err := helper.CreateFileWithContent(ignoreFilePath, ".git\n*.md"); err != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")

			// modify a ignored file and push
			helper.ReplaceString(filepath.Join(globals.Context, "nodejs-ex", "README.md"), "This example will serve a welcome page", "This is a example welcome page!")
			output := helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// test ignores using the flag
			output = helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex", "--ignore", "*.md")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", globals.Context+"/nodejs-ex")

			// use the force build flag and push
			output := helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex", "-f")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
		})

		It("should push only the modified files", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs:latest", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")

			url := oc.GetFirstURL(cmpName, appName, globals.Project)
			// Wait for running app before getting info about files.
			// During the startup sequence there is something that will modify the access time of a source file.
			helper.HttpWaitFor("http://"+url, "Welcome to your Node.js", 30, 1)

			envs := oc.GetEnvs(cmpName, appName, globals.Project)
			dir := envs["ODO_S2I_SRC_BIN_PATH"]

			earlierCatServerFile := ""
			oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				globals.Project,
				[]string{"stat", filepath.Join(dir, "src", "server.js")},
				func(cmdOp string, err error) bool {
					earlierCatServerFile = cmdOp
					return true
				},
			)

			earlierCatViewFile := ""
			oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				globals.Project,
				[]string{"stat", filepath.Join(dir, "src", "views", "index.html")},
				func(cmdOp string, err error) bool {
					earlierCatViewFile = cmdOp
					return true
				},
			)

			helper.ReplaceString(filepath.Join(globals.Context, "nodejs-ex", "views", "index.html"), "Welcome to your Node.js application on OpenShift", "UPDATED!")
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)

			modifiedCatViewFile := ""
			oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				globals.Project,
				[]string{"stat", filepath.Join(dir, "src", "views", "index.html")},
				func(cmdOp string, err error) bool {
					modifiedCatViewFile = cmdOp
					return true
				},
			)

			modifiedCatServerFile := ""
			oc.CheckCmdOpInRemoteCmpPod(
				cmpName,
				appName,
				globals.Project,
				[]string{"stat", filepath.Join(dir, "src", "server.js")},
				func(cmdOp string, err error) bool {
					modifiedCatServerFile = cmdOp
					return true
				},
			)

			Expect(modifiedCatViewFile).NotTo(Equal(earlierCatViewFile))
			Expect(modifiedCatServerFile).To(Equal(earlierCatServerFile))
		})

		It("should delete the files from the container if its removed locally", func() {
			oc.ImportJavaIS(globals.Project)
			cmpName := "backend"
			helper.CopyExample(filepath.Join("source", "openjdk"), globals.Context)
			helper.CmdShouldPass("odo", "create", "java:8", "backend", "--project", globals.Project, "--context", globals.Context, "--app", appName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", globals.Context)
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			envs := oc.GetEnvs(cmpName, appName, globals.Project)
			dir := envs["ODO_S2I_SRC_BIN_PATH"]

			var statErr error
			oc.CheckCmdOpInRemoteCmpPod(
				"backend",
				appName,
				globals.Project,
				[]string{"stat", filepath.Join(dir, "src", "src", "main", "java", "AnotherMessageProducer.java")},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(os.Remove(filepath.Join(globals.Context, "src", "main", "java", "AnotherMessageProducer.java"))).NotTo(HaveOccurred())
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)

			oc.CheckCmdOpInRemoteCmpPod(
				"backend",
				appName,
				globals.Project,
				[]string{"stat", filepath.Join(dir, "src", "src", "main", "java", "AnotherMessageProducer.java")},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)

			Expect(statErr).To(HaveOccurred())
			path := filepath.Join(dir, "src", "src", "main", "java", "AnotherMessageProducer.java")
			Expect(statErr.Error()).To(ContainSubstring("cannot stat '" + path + "': No such file or directory"))
		})

	})

	Context("when .odoignore file exists", func() {
		It("should create and push the contents of a named component excluding the contents in .odoignore file", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")

			ignoreFilePath := filepath.Join(globals.Context, "nodejs-ex", ".odoignore")
			if err := helper.CreateFileWithContent(ignoreFilePath, ".git\ntests/\nREADME.md"); err != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}

			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--project", globals.Project, "--context", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "push", "--context", globals.Context+"/nodejs-ex")

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp("nodejs", globals.Project)

			envs := oc.GetEnvs(cmpName, appName, globals.Project)
			dir := envs["ODO_S2I_DEPLOYMENT_DIR"]

			// verify that the views folder got pushed
			stdOut1 := oc.ExecListDir(podName, globals.Project, dir)
			Expect(stdOut1).To(ContainSubstring("views"))

			// verify that the tests was not pushed
			stdOut2 := oc.ExecListDir(podName, globals.Project, dir)
			Expect(stdOut2).To(Not(ContainSubstring(("tests"))))

			// verify that the README.md file was not pushed
			stdOut3 := oc.ExecListDir(podName, globals.Project, dir)
			Expect(stdOut3).To(Not(ContainSubstring(("README.md"))))
		})
	})

	Context("when .gitignore file exists", func() {
		It("should create and push the contents of a named component and include odo-file-index.json path to .gitignore file to exclude the contents", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context+"/nodejs-ex", "--app", appName)

			// push and include the odo-file-index.json path to .gitignore file
			helper.CmdShouldPass("odo", "push", "--context", filepath.Join(globals.Context, "nodejs-ex"))
			ignoreFilePath := filepath.Join(globals.Context, "nodejs-ex", ".gitignore")
			helper.FileShouldContainSubstring(ignoreFilePath, filepath.Join(".odo", "odo-file-index.json"))
		})
	})

	Context("when .gitignore file does not exist", func() {
		It("should create and push the contents of a named component and also create .gitignore then include odo-file-index.json path to .gitignore file to exclude the contents", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), globals.Context)
			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", globals.Project, "--context", globals.Context, "--app", appName)

			// push and include the odo-file-index.json path to .gitignore file
			helper.CmdShouldPass("odo", "push", "--context", globals.Context)
			ignoreFilePath := filepath.Join(globals.Context, ".gitignore")
			helper.FileShouldContainSubstring(ignoreFilePath, filepath.Join(".odo", "odo-file-index.json"))
		})
	})

	Context("when running odo push with flag --show-log", func() {
		It("should be able to spam odo push without anything breaking", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", globals.Context+"/nodejs-ex")
			helper.CmdShouldPass("odo", "create", "nodejs", "nodejs", "--project", globals.Project, "--context", globals.Context+"/nodejs-ex")
			// Iteration 1
			output := helper.CmdShouldPass("odo", "push", "--show-log", "--context", globals.Context+"/nodejs-ex")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
			// Iteration 2
			output = helper.CmdShouldPass("odo", "push", "--show-log", "--context", globals.Context+"/nodejs-ex")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
			// Iteration 3
			output = helper.CmdShouldPass("odo", "push", "--show-log", "--context", globals.Context+"/nodejs-ex")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
		})
	})
})
