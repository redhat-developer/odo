package integration

import (
	"fmt"
	"github.com/openshift/odo/tests/helper"
	"os"
	"path/filepath"
	"time"

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

		It("should build when a new file and a new folder is added in the directory", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			newFilePath := filepath.Join(context, "nodejs-ex", "new-example.html")
			if err := helper.CreateFileWithContent(newFilePath, "<html>Hello</html>"); err != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp(cmpName, project)

			// verify that the new file was pushed
			stdOut := oc.ExecListDir(podName, project)
			Expect(stdOut).To(ContainSubstring(("README.md")))

			helper.MakeDir(filepath.Join(context, "nodejs-ex", "exampleDir"))

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// verify that the new file was pushed
			stdOut = oc.ExecListDir(podName, project)
			Expect(stdOut).To(ContainSubstring(("exampleDir")))
		})

		/* TODO uncomment once https://github.com/openshift/odo/issues/1354 is resolved

		It("should build when a file is deleted from the directory", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			helper.DeleteDir(filepath.Join(context, "/nodejs-ex", "/tests"))

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp(cmpName, project)

			// verify that the new file was pushed
			stdOut := oc.ExecListDir(podName, project)
			Expect(stdOut).To(Not(ContainSubstring(("tests"))))
		})
		*/

		It("should build when a file and a folder is renamed in the directory", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			helper.RenameFile(filepath.Join(context, "/nodejs-ex", "README.md"), filepath.Join(context, "/nodejs-ex", "NEW-README.md"))

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp(cmpName, project)

			// verify that the new file was pushed
			stdOut := oc.ExecListDir(podName, project)

			// TODO enable once https://github.com/openshift/odo/issues/1354 is resolve
			//Expect(stdOut).To(Not(ContainSubstring("README.md")))

			Expect(stdOut).To(ContainSubstring("NEW-README.md"))

			helper.RenameFile(filepath.Join(context, "/nodejs-ex", "/tests"), filepath.Join(context, "/nodejs-ex", "/testing"))

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			// verify that the new file was pushed
			stdOut = oc.ExecListDir(podName, project)

			// TODO enable once https://github.com/openshift/odo/issues/1354 is resolve
			//Expect(stdOut).To(Not(ContainSubstring("tests")))

			Expect(stdOut).To(ContainSubstring("testing"))
		})

		It("should not build when changes are detected in a ignored file", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)

			ignoreFilePath := filepath.Join(context, "nodejs-ex", ".odoignore")
			if err := helper.CreateFileWithContent(ignoreFilePath, ".git\n*.md"); err != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			helper.ReplaceString(filepath.Join(context+"/nodejs-ex"+"/README.md"), "This example will serve a welcome page", "This is a example welcome page!")

			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			output = helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex", "--ignore", "*.md")

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			helper.CmdShouldPass("git", "clone", "https://github.com/openshift/nodejs-ex", context+"/nodejs-ex")

			helper.CmdShouldPass("odo", "component", "create", "nodejs", cmpName, "--project", project, "--context", context+"/nodejs-ex", "--app", appName)

			helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex")

			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", context+"/nodejs-ex")

			output := helper.CmdShouldPass("odo", "push", "--context", context+"/nodejs-ex", "-f")

			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
		})
	})
})
