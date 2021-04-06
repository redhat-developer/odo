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

	// Context("Check pod timeout", func() {

	// 	It("Check that pod timeout works and we time out immediately..", func() {
	// 		helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
	// 		helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
	// 		helper.CmdShouldPass("odo", "preference", "set", "PushTimeout", "1")
	// 		output := helper.CmdShouldFail("odo", "push", "--context", commonVar.Context)
	// 		Expect(output).To(ContainSubstring("waited 1s but couldn't find running pod matching selector"))
	// 	})

	// })

	Context("Check memory and cpu config before odo push", func() {
		It("Should work when memory is set..", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)

			helper.CmdShouldPass("odo", "config", "set", "Memory", "300Mi", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
		})

		It("Should fail if minMemory is set but maxmemory is not set..", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)

			helper.CmdShouldPass("odo", "config", "set", "minmemory", "100Mi", "--context", commonVar.Context)
			output := helper.CmdShouldFail("odo", "push", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("`minmemory` should accompany `maxmemory` or use `odo config set memory` to use same value for both min and max"))
		})

		It("should fail if maxmemory is set but minmemory is not set..", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)

			helper.CmdShouldPass("odo", "config", "set", "maxmemory", "400Mi", "--context", commonVar.Context)
			output := helper.CmdShouldFail("odo", "push", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("`minmemory` should accompany `maxmemory` or use `odo config set memory` to use same value for both min and max"))
		})

		It("Should work when cpu is set", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)

			helper.CmdShouldPass("odo", "config", "set", "cpu", "0.4", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
		})

		It("Should fail if mincpu is set but maxcpu is not set..", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)

			helper.CmdShouldPass("odo", "config", "set", "mincpu", "0.4", "--context", commonVar.Context)
			output := helper.CmdShouldFail("odo", "push", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("`mincpu` should accompany `maxcpu` or use `odo config set cpu` to use same value for both min and max"))
		})

		It("should fail if maxcpu is set but mincpu is not set..", func() {

			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)

			helper.CmdShouldPass("odo", "config", "set", "maxcpu", "0.5", "--context", commonVar.Context)
			output := helper.CmdShouldFail("odo", "push", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("`mincpu` should accompany `maxcpu` or use `odo config set cpu` to use same value for both min and max"))
		})
	})

	Context("Check for label propagation after pushing", func() {

		It("Check for labels", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// Check for all the labels
			oc.VerifyLabelExistsOfComponent(cmpName, commonVar.Project, "app:"+appName)
			oc.VerifyLabelExistsOfComponent(cmpName, commonVar.Project, "app.kubernetes.io/part-of:"+appName)
			oc.VerifyLabelExistsOfComponent(cmpName, commonVar.Project, "app.kubernetes.io/managed-by:odo")

			// Check for the version
			versionInfo := helper.CmdShouldPass("odo", "version")
			re := regexp.MustCompile(`v[0-9]\S*`)
			odoVersionString := re.FindStringSubmatch(versionInfo)
			oc.VerifyLabelExistsOfComponent(cmpName, commonVar.Project, "app.kubernetes.io/managed-by-version:"+odoVersionString[0])
		})
	})

	Context("Test push outside of the current working direcory", func() {

		// Change to "outside" the directory before running the below tests
		JustBeforeEach(func() {
			helper.Chdir(commonVar.Context)
		})

		It("Push, modify a file and then push outside of the working directory", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// Create a new file to test propagating changes
			newFilePath := filepath.Join(commonVar.Context, "foobar.txt")
			if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}

			// Test propagating changes
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// Delete the file and check that the file is deleted
			helper.DeleteDir(newFilePath)

			// Test propagating deletions
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
		})

	})

	Context("when push command is executed", func() {
		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)
			output := helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			url := oc.GetFirstURL(cmpName, appName, commonVar.Project)
			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "Hello world from node.js!", "UPDATED!")

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)
		})

		It("should be able to create a file, push, delete, then push again propagating the deletions and build", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			output := helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
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
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := oc.GetRunningPodNameOfComp(cmpName, commonVar.Project)

			envs := oc.GetEnvs(cmpName, appName, commonVar.Project)
			dir := envs["ODO_S2I_DEPLOYMENT_DIR"]

			stdOut := oc.ExecListDir(podName, commonVar.Project, dir)
			helper.MatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})

			// Now we delete the file and dir and push
			helper.DeleteDir(newFilePath)
			helper.DeleteDir(newDirPath)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context, "-v4")

			// Then check to see if it's truly been deleted
			stdOut = oc.ExecListDir(podName, commonVar.Project, dir)
			helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
		})

		It("should build when a file and a folder is renamed in the directory", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)

			// create a file and folder then push
			err := os.MkdirAll(filepath.Join(commonVar.Context, "tests"), 0750)
			Expect(err).To(BeNil())
			_, err = os.Create(filepath.Join(commonVar.Context, "README.md"))
			Expect(err).To(BeNil())

			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			output := helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// rename a file and push
			helper.RenameFile(filepath.Join(commonVar.Context, "README.md"), filepath.Join(commonVar.Context, "NEW-FILE.md"))
			output = helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp(cmpName, commonVar.Project)

			envs := oc.GetEnvs(cmpName, appName, commonVar.Project)
			dir := envs["ODO_S2I_DEPLOYMENT_DIR"]

			// verify that the new file was pushed
			stdOut := oc.ExecListDir(podName, commonVar.Project, dir)

			Expect(stdOut).To(Not(ContainSubstring("README.md")))

			Expect(stdOut).To(ContainSubstring("NEW-FILE.md"))

			// rename a folder and push
			helper.RenameFile(filepath.Join(commonVar.Context, "tests"), filepath.Join(commonVar.Context, "testing"))
			output = helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			// verify that the new file was pushed
			stdOut = oc.ExecListDir(podName, commonVar.Project, dir)

			Expect(stdOut).To(Not(ContainSubstring("tests")))

			Expect(stdOut).To(ContainSubstring("testing"))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)

			// use the force build flag and push
			output := helper.CmdShouldPass("odo", "push", "--context", commonVar.Context, "-f")
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
		})

		It("should push only the modified files", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs:latest", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			url := oc.GetFirstURL(cmpName, appName, commonVar.Project)

			// Wait for running app before getting info about files.
			// During the startup sequence there is something that will modify the access time of a source file.
			helper.HttpWaitFor("http://"+url, "Hello world from node.js!", 30, 1)

			envs := oc.GetEnvs(cmpName, appName, commonVar.Project)
			dir := envs["ODO_S2I_SRC_BIN_PATH"]

			earlierCatServerFile := ""
			earlierCatServerFile = oc.StatFileInPod(cmpName, appName, commonVar.Project, filepath.ToSlash(filepath.Join(dir, "src", "server.js")))

			earlierCatPackageFile := ""
			earlierCatPackageFile = oc.StatFileInPod(cmpName, appName, commonVar.Project, filepath.ToSlash(filepath.Join(dir, "src", "package.json")))

			helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "Hello world from node.js!", "UPDATED!")
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			helper.HttpWaitFor("http://"+url, "UPDATED!", 30, 1)

			modifiedCatPackageFile := ""
			modifiedCatPackageFile = oc.StatFileInPod(cmpName, appName, commonVar.Project, filepath.ToSlash(filepath.Join(dir, "src", "package.json")))

			modifiedCatServerFile := ""
			modifiedCatServerFile = oc.StatFileInPod(cmpName, appName, commonVar.Project, filepath.ToSlash(filepath.Join(dir, "src", "server.js")))

			Expect(modifiedCatPackageFile).To(Equal(earlierCatPackageFile))
			Expect(modifiedCatServerFile).NotTo(Equal(earlierCatServerFile))
		})

		It("should delete the files from the container if its removed locally", func() {
			oc.ImportJavaIS(commonVar.Project)
			cmpName := "backend"
			helper.CopyExample(filepath.Join("source", "openjdk"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "--s2i", "java:8", "backend", "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)
			helper.CmdShouldPass("odo", "url", "create", "--port", "8080", "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			envs := oc.GetEnvs(cmpName, appName, commonVar.Project)
			dir := envs["ODO_S2I_SRC_BIN_PATH"]

			var statErr error
			oc.CheckCmdOpInRemoteCmpPod(
				"backend",
				appName,
				commonVar.Project,
				[]string{"stat", filepath.ToSlash(filepath.Join(dir, "src", "src", "main", "java", "AnotherMessageProducer.java"))},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(os.Remove(filepath.Join(commonVar.Context, "src", "main", "java", "AnotherMessageProducer.java"))).NotTo(HaveOccurred())
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			oc.CheckCmdOpInRemoteCmpPod(
				"backend",
				appName,
				commonVar.Project,
				[]string{"stat", filepath.ToSlash(filepath.Join(dir, "src", "src", "main", "java", "AnotherMessageProducer.java"))},
				func(cmdOp string, err error) bool {
					statErr = err
					return true
				},
			)

			Expect(statErr).To(HaveOccurred())
			path := filepath.ToSlash(filepath.Join(dir, "src", "src", "main", "java", "AnotherMessageProducer.java"))
			Expect(statErr.Error()).To(ContainSubstring("cannot stat '" + path + "': No such file or directory"))
		})

	})

	Context("when .odoignore file exists", func() {
		It("should create and push the contents of a named component excluding the contents and changes detected in .odoignore file", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)

			ignoreFilePath := filepath.Join(commonVar.Context, ".odoignore")
			if err := helper.CreateFileWithContent(ignoreFilePath, ".git\n*.md"); err != nil {
				fmt.Printf("the .odoignore file was not created, reason %v", err.Error())
			}
			_, err := os.Create(filepath.Join(commonVar.Context, "README.md"))
			Expect(err).To(BeNil())

			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context)
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)

			// get the name of running pod
			podName := oc.GetRunningPodNameOfComp("nodejs", commonVar.Project)

			envs := oc.GetEnvs(cmpName, appName, commonVar.Project)
			dir := envs["ODO_S2I_DEPLOYMENT_DIR"]

			// verify that the server file got pushed
			stdOut1 := oc.ExecListDir(podName, commonVar.Project, dir)
			Expect(stdOut1).To(ContainSubstring("server.js"))

			// verify that the README.md file was not pushed
			stdOut3 := oc.ExecListDir(podName, commonVar.Project, dir)
			Expect(stdOut3).To(Not(ContainSubstring(("README.md"))))

			// modify a ignored file and push
			helper.ReplaceString(filepath.Join(commonVar.Context, "README.md"), "", "This is a example welcome page!")
			output := helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))

			// test ignores using the flag
			output = helper.CmdShouldPass("odo", "push", "--context", commonVar.Context, "--ignore", "*.md")
			Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
		})
	})

	Context("when .gitignore file exists or not", func() {
		It("should create and push the contents of a named component and include odo-file-index.json path to .gitignore file to exclude the contents, if does not exists create one", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "component", "create", "--s2i", "nodejs", cmpName, "--project", commonVar.Project, "--context", commonVar.Context, "--app", appName)

			// push and include the odo-file-index.json path to .gitignore file
			helper.CmdShouldPass("odo", "push", "--context", commonVar.Context)
			ignoreFilePath := filepath.Join(commonVar.Context, ".gitignore")
			helper.FileShouldContainSubstring(ignoreFilePath, filepath.Join(".odo", "odo-file-index.json"))
		})
	})

	Context("when running odo push with flag --show-log", func() {
		It("should be able to execute odo push consecutively without breaking anything", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CmdShouldPass("odo", "create", "--s2i", "nodejs", "nodejs", "--project", commonVar.Project, "--context", commonVar.Context)

			// Run odo push in consecutive iteration
			output := helper.CmdShouldPass("odo", "push", "--show-log", "--context", commonVar.Context)
			Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))

			for i := 0; i <= 1; i++ {
				output := helper.CmdShouldPass("odo", "push", "--show-log", "--context", commonVar.Context)
				Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
			}
		})
	})
})
