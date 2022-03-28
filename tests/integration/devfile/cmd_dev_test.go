package devfile

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	segment "github.com/redhat-developer/odo/pkg/segment/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/onsi/gomega/gexec"
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
		Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	When("directory is empty", func() {

		BeforeEach(func() {
			Expect(helper.ListFilesInDir(commonVar.Context)).To(HaveLen(0))
		})

		It("should error", func() {
			output := helper.Cmd("odo", "dev", "--random-ports").ShouldFail().Err()
			Expect(output).To(ContainSubstring("this command cannot run in an empty directory"))

		})
	})

	When("a component is bootstrapped and pushed", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})
		It("should show validation errors if the devfile is incorrect", func() {
			err := helper.RunDevMode(func(session *gexec.Session, outContents, errContents []byte, urls []string) {
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "kind: run", "kind: build")
				helper.WaitForOutputToContain("Error occurred on Push", 180, 10, session)
			})
			Expect(err).ToNot(HaveOccurred())
		})
		It("should use the index information from previous push operation", func() {
			// Create a new file A
			fileAPath, fileAText := helper.CreateSimpleFile(commonVar.Context, "my-file-", ".txt")
			// watch that project
			err := helper.RunDevMode(func(session *gexec.Session, outContents, errContents []byte, urls []string) {
				// Change some other file B
				helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "App started", "App is super started")

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				// File should exist, and its content should match what we initially set it to
				execResult := commonVar.CliRunner.Exec(podName, commonVar.Project, "cat", "/projects/"+filepath.Base(fileAPath))
				Expect(execResult).To(ContainSubstring(fileAText))
			})
			Expect(err).ToNot(HaveOccurred())
		})
		It("ensure that index information is updated", func() {
			err := helper.RunDevMode(func(session *gexec.Session, outContents, errContents []byte, urls []string) {
				indexAfterPush, err := util.ReadFileIndex(filepath.Join(commonVar.Context, ".odo", "odo-file-index.json"))
				Expect(err).ToNot(HaveOccurred())

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

					// Sanity test: at least one file should be present
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
			Expect(err).ToNot(HaveOccurred())
		})

		When("odo dev is executed", func() {

			BeforeEach(func() {
				devSession, _, _, _, err := helper.StartDevMode()
				Expect(err).ToNot(HaveOccurred())
				defer devSession.Kill()
				// An ENV file should have been created indicating current namespace
				Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeTrue())
				helper.FileShouldContainSubstring(".odo/env/env.yaml", "Project: "+commonVar.Project)
			})

			When("deleting previous deployment and switching kubeconfig to another namespace", func() {
				var otherNS string
				BeforeEach(func() {
					helper.Cmd("odo", "delete", "component", "--name", cmpName, "-f").ShouldPass()
					output := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Err.Contents()
					Expect(string(output)).To(ContainSubstring("No resources found in " + commonVar.Project + " namespace."))

					Eventually(func() string {
						return string(commonVar.CliRunner.Run("get", "pods", "-n", commonVar.Project).Err.Contents())
					}).Should(ContainSubstring("No resources found"))

					otherNS = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
				})

				AfterEach(func() {
					commonVar.CliRunner.DeleteNamespaceProject(otherNS)
				})

				It("should run odo dev on initial namespace", func() {
					err := helper.RunDevMode(func(session *gexec.Session, outContents, errContents []byte, urls []string) {
						output := commonVar.CliRunner.Run("get", "deployment").Err.Contents()
						Expect(string(output)).To(ContainSubstring("No resources found in " + otherNS + " namespace."))

						output = commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
						Expect(string(output)).To(ContainSubstring(cmpName))
					})
					Expect(err).ToNot(HaveOccurred())
				})
			})
			When("recording telemetry data", func() {
				BeforeEach(func() {
					helper.EnableTelemetryDebug()
					session, _, _, _, _ := helper.StartDevMode()
					session.Stop()
				})
				AfterEach(func() {
					helper.ResetTelemetry()
				})
				It("should record the telemetry data correctly", func() {
					td := helper.GetTelemetryDebugData()
					Expect(td.Event).To(ContainSubstring("odo dev"))
					Expect(td.Properties.Success).To(BeFalse())
					Expect(td.Properties.Error).To(ContainSubstring("interrupt"))
					Expect(td.Properties.ErrorType == "*errors.errorString").To(BeTrue())
					Expect(td.Properties.CmdProperties[segment.ComponentType]).To(ContainSubstring("nodejs"))
					Expect(td.Properties.CmdProperties[segment.Language]).To(ContainSubstring("nodejs"))
					Expect(td.Properties.CmdProperties[segment.ProjectType]).To(ContainSubstring("nodejs"))
				})
			})
		})
	})

	Context("port-forwarding for the component", func() {
		When("devfile has single endpoint", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.Cmd("odo", "project", "set", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			})

			It("should expose the endpoint on localhost", func() {
				err := helper.RunDevMode(func(session *gexec.Session, outContents, errContents []byte, urls []string) {
					url := fmt.Sprintf("http://%s", urls[0])
					resp, err := http.Get(url)
					Expect(err).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), []string{"Hello from Node.js Starter Application!"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("devfile has multiple endpoints", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-multiple-endpoints"), commonVar.Context)
				helper.Cmd("odo", "project", "set", commonVar.Project).ShouldPass()
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-multiple-endpoints.yaml")).ShouldPass()
			})

			It("should expose two endpoints on localhost", func() {
				err := helper.RunDevMode(func(session *gexec.Session, outContents, errContents []byte, urls []string) {
					url1 := fmt.Sprintf("http://%s", urls[0])
					url2 := fmt.Sprintf("http://%s", urls[1])

					resp1, err := http.Get(url1)
					Expect(err).ToNot(HaveOccurred())
					defer resp1.Body.Close()

					resp2, err := http.Get(url2)
					Expect(err).ToNot(HaveOccurred())
					defer resp2.Body.Close()

					body1, _ := io.ReadAll(resp1.Body)
					helper.MatchAllInOutput(string(body1), []string{"Hello from Node.js Starter Application!"})

					body2, _ := io.ReadAll(resp2.Body)
					helper.MatchAllInOutput(string(body2), []string{"Hello from Node.js Starter Application!"})

					helper.ReplaceString("server.js", "Hello from Node.js", "H3110 from Node.js")
					helper.WaitForOutputToContain("Watching for changes in the current directory", 180, 10, session)

					Eventually(func() bool {
						resp3, err := http.Get(url1)
						if err != nil {
							return false
						}
						defer resp3.Body.Close()

						resp4, err := http.Get(url2)
						if err != nil {
							return false
						}
						defer resp4.Body.Close()

						body3, _ := io.ReadAll(resp3.Body)
						if string(body3) != "H3110 from Node.js Starter Application!" {
							return false
						}

						body4, _ := io.ReadAll(resp4.Body)
						return string(body4) == "H3110 from Node.js Starter Application!"
					}, 180, 10).Should(Equal(true))
				})
				Expect(err).ToNot(HaveOccurred())
			})

			When("an endpoint is added after first run of odo dev", func() {
				It("should print the message to run odo dev again", func() {
					err := helper.RunDevMode(func(session *gexec.Session, outContents, errContents []byte, urls []string) {
						helper.ReplaceString("devfile.yaml", "exposure: none", "exposure: public")
						helper.WaitForErroutToContain("devfile.yaml has been changed; please restart the `odo dev` command", 180, 10, session)
					})
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})

	When("Devfile 2.1.0 is used", func() {
		// from devfile
		devfileCmpName := "nodejs"
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-variables.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
		})

		When("doing odo dev", func() {
			var session helper.DevSession
			BeforeEach(func() {
				var err error
				session, _, _, _, err = helper.StartDevMode()
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
			})

			It("should check if the env variable has a correct value", func() {
				envVars := commonVar.CliRunner.GetEnvsDevFileDeployment(devfileCmpName, "app", commonVar.Project)
				// check if the env variable has a correct value. This value was substituted from in devfile from variable
				Expect(envVars["FOO"]).To(Equal("bar"))
			})
		})
	})

	When("running odo dev and single env var is set", func() {
		devfileCmpName := "nodejs"
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-single-env.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
		})

		It("should be able to exec command", func() {
			err := helper.RunDevMode(func(session *gexec.Session, out, err []byte, urls []string) {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				helper.MatchAllInOutput(output, []string{"test_env_variable", "test_build_env_variable"})
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("running odo dev and multiple env variables are set", func() {
		devfileCmpName := "nodejs"
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-multiple-envs.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
		})

		It("should be able to exec command", func() {
			err := helper.RunDevMode(func(session *gexec.Session, out, err []byte, urls []string) {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				helper.MatchAllInOutput(output, []string{"test_build_env_variable1", "test_build_env_variable2", "test_env_variable1", "test_env_variable2"})
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("doing odo dev and there is a env variable with spaces", func() {
		devfileCmpName := "nodejs"
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-env-with-space.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
		})

		It("should be able to exec command", func() {
			err := helper.RunDevMode(func(session *gexec.Session, out, err []byte, urls []string) {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				helper.MatchAllInOutput(output, []string{"build env variable with space", "env with space"})
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("creating local files and dir and running odo dev", func() {
		var newDirPath, newFilePath, stdOut, podName string
		var session helper.DevSession
		// from devfile
		devfileCmpName := "nodejs"
		BeforeEach(func() {
			newFilePath = filepath.Join(commonVar.Context, "foobar.txt")
			newDirPath = filepath.Join(commonVar.Context, "testdir")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			// Create a new file that we plan on deleting later...
			if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}
			// Create a new directory
			helper.MakeDir(newDirPath)
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			session.Stop()
		})

		It("should correctly propagate changes to the container", func() {

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName = commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)

			stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
			helper.MatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
		})

		When("deleting local files and dir and waiting for sync", func() {
			BeforeEach(func() {
				// Now we delete the file and dir and push
				helper.DeleteDir(newFilePath)
				helper.DeleteDir(newDirPath)
				_, _, err := session.WaitSync()
				Expect(err).ToNot(HaveOccurred())
			})
			It("should not list deleted dir and file in container", func() {
				podName = commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				// Then check to see if it's truly been deleted
				stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
			})
		})
	})

	When("adding local files to gitignore and running odo dev", func() {
		var gitignorePath, newDirPath, newFilePath1, newFilePath2, newFilePath3, stdOut, podName string
		var session helper.DevSession
		// from devfile
		devfileCmpName := "nodejs"
		BeforeEach(func() {
			gitignorePath = filepath.Join(commonVar.Context, ".gitignore")
			newFilePath1 = filepath.Join(commonVar.Context, "foobar.txt")
			newDirPath = filepath.Join(commonVar.Context, "testdir")
			newFilePath2 = filepath.Join(newDirPath, "foobar.txt")
			newFilePath3 = filepath.Join(newDirPath, "baz.txt")
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			if err := helper.CreateFileWithContent(newFilePath1, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}
			// Create a new directory
			helper.MakeDir(newDirPath)
			if err := helper.CreateFileWithContent(newFilePath2, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}
			if err := helper.CreateFileWithContent(newFilePath3, "hello world"); err != nil {
				fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
			}
			if err := helper.CreateFileWithContent(gitignorePath, "foobar.txt"); err != nil {
				fmt.Printf("the .gitignore file was not created, reason %v", err.Error())
			}
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should not sync ignored files to the container", func() {
			podName = commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)

			stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
			helper.MatchAllInOutput(stdOut, []string{"testdir"})
			helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt"})
			stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects/testdir")
			helper.MatchAllInOutput(stdOut, []string{"baz.txt"})
			helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt"})
		})
	})

	When("devfile has sourcemappings and running odo dev", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileSourceMapping.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())

		})
		AfterEach(func() {
			session.Stop()
		})

		It("should sync files to the correct location", func() {
			// Verify source code was synced to /test instead of /projects
			var statErr error
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				commonVar.Project,
				[]string{"stat", "/test/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})
	})

	When("project and clonePath is present in devfile and running odo dev", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			// devfile with clonePath set in project field
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should sync to the correct dir in container", func() {
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
			// source code is synced to $PROJECTS_ROOT/clonePath
			// $PROJECTS_ROOT is /projects by default, if sourceMapping is set it is same as sourceMapping
			// for devfile-with-projects.yaml, sourceMapping is apps and clonePath is webapp
			// so source code would be synced to /apps/webapp
			output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/apps/webapp")
			helper.MatchAllInOutput(output, []string{"package.json"})

			// Verify the sync env variables are correct
			utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/apps/webapp", "/apps", commonVar.CliRunner)
		})
	})

	When("devfile project field is present and running odo dev", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			// reset clonePath and change the workdir accordingly, it should sync to project name
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "clonePath: webapp/", "# clonePath: webapp/")
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should sync to the correct dir in container", func() {
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
			output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/apps/nodeshift")
			helper.MatchAllInOutput(output, []string{"package.json"})

			// Verify the sync env variables are correct
			utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/apps/nodeshift", "/apps", commonVar.CliRunner)
		})
	})

	When("multiple project are present", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should sync to the correct dir in container", func() {
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
			// for devfile-with-multiple-projects.yaml source mapping is not set so $PROJECTS_ROOT is /projects
			// multiple projects, so source code would sync to the first project /projects/webapp
			output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects/webapp")
			helper.MatchAllInOutput(output, []string{"package.json"})

			// Verify the sync env variables are correct
			utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/projects/webapp", "/projects", commonVar.CliRunner)
		})
	})

	When("no project is present", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should sync to the correct dir in container", func() {

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
			output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
			helper.MatchAllInOutput(output, []string{"package.json"})

			// Verify the sync env variables are correct
			utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/projects", "/projects", commonVar.CliRunner)
		})
	})

	When("running odo dev with devfile contain volume", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volumes.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should create pvc and reuse if it shares the same devfile volume name", func() {
			var statErr error
			var cmdOutput string
			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)

			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime2",
				commonVar.Project,
				[]string{"cat", "/myvol/myfile.log"},
				func(cmdOp string, err error) bool {
					cmdOutput = cmdOp
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(cmdOutput).To(ContainSubstring("hello"))

			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime2",
				commonVar.Project,
				[]string{"stat", "/data2"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})

		It("check the volume name and mount paths for the containers", func() {
			deploymentName, err := util.NamespaceKubernetesObject(devfileCmpName, "app")
			Expect(err).To(BeNil())

			volumesMatched := false

			// check the volume name and mount paths for the containers
			volNamesAndPaths := commonVar.CliRunner.GetVolumeMountNamesandPathsFromContainer(deploymentName, "runtime", commonVar.Project)
			volNamesAndPathsArr := strings.Fields(volNamesAndPaths)
			for _, volNamesAndPath := range volNamesAndPathsArr {
				volNamesAndPathArr := strings.Split(volNamesAndPath, ":")

				if strings.Contains(volNamesAndPathArr[0], "myvol") && volNamesAndPathArr[1] == "/data" {
					volumesMatched = true
				}
			}
			Expect(volumesMatched).To(Equal(true))
		})
	})

	When("running odo dev with devfile containing volume-component", func() {
		devfileCmpName := "test-devfile"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should successfully use the volume components in container components", func() {

			// Verify the pvc size for firstvol
			storageSize := commonVar.CliRunner.GetPVCSize(devfileCmpName, "firstvol", commonVar.Project)
			// should be the default size
			Expect(storageSize).To(ContainSubstring("1Gi"))

			// Verify the pvc size for secondvol
			storageSize = commonVar.CliRunner.GetPVCSize(devfileCmpName, "secondvol", commonVar.Project)
			// should be the specified size in the devfile volume component
			Expect(storageSize).To(ContainSubstring("3Gi"))
		})
	})

	When("running odo dev and devfile with composite command", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should execute all commands in composite commmand", func() {
			// Verify the command executed successfully
			var statErr error
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				commonVar.Project,
				[]string{"stat", "/projects/testfolder"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})
	})

	When("running odo dev and composite command is marked as paralell:true", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommandsParallel.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should execute all commands in composite commmand", func() {
			// Verify the command executed successfully
			var statErr error
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				commonVar.Project,
				[]string{"stat", "/projects/testfolder"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})
	})

	When("running odo dev and composite command are nested", func() {
		devfileCmpName := "nodejs"
		var session helper.DevSession
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileNestedCompCommands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should execute all commands in composite commmand", func() {
			// Verify the command executed successfully
			var statErr error
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				commonVar.Project,
				[]string{"stat", "/projects/testfolder"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
		})
	})

	When("running odo dev and composite command is used as a run command", func() {
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeRun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
		})

		It("should throw a validation error for composite run commands", func() {
			output := helper.Cmd("odo", "dev", "--random-ports").ShouldFail().Err()
			Expect(output).To(ContainSubstring("not supported currently"))
		})
	})

	When("running odo dev and prestart events are defined", func() {
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-preStart.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
		})

		It("should not correctly execute PreStart commands", func() {
			output := helper.Cmd("odo", "dev", "--random-ports").ShouldFail().Err()
			// This is expected to fail for now.
			// see https://github.com/redhat-developer/odo/issues/4187 for more info
			helper.MatchAllInOutput(output, []string{"myprestart should either map to an apply command or a composite command with apply commands\n"})
		})
	})

	When("running odo dev and run command throws an error", func() {
		var session helper.DevSession
		var initErr []byte
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm start", "npm starts")
			var err error
			session, _, initErr, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should error out with some log", func() {
			helper.MatchAllInOutput(string(initErr), []string{
				"exited with error status within 1 sec",
				"Did you mean one of these?",
			})
		})
	})

	When("Create and dev java-springboot component", func() {
		devfileCmpName := "java-spring-boot"
		var session helper.DevSession
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", devfileCmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should execute default build and run commands correctly", func() {

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)

			var statErr error
			var cmdOutput string
			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"runtime",
				commonVar.Project,
				// [s] to not match the current command: https://unix.stackexchange.com/questions/74185/how-can-i-prevent-grep-from-showing-up-in-ps-results
				[]string{"bash", "-c", "grep [s]pring-boot:run /proc/*/cmdline"},
				func(cmdOp string, err error) bool {
					cmdOutput = cmdOp
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(cmdOutput).To(MatchRegexp("Binary file .* matches"))
		})
	})

	When("setting git config and running odo dev", func() {
		remoteURL := "https://github.com/odo-devfiles/nodejs-ex"
		devfileCmpName := "nodejs"
		BeforeEach(func() {
			helper.Cmd("git", "init").ShouldPass()
			remote := "origin"
			helper.Cmd("git", "remote", "add", remote, remoteURL).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", devfileCmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
		})

		It("should create vcs-uri annotation for the deployment when running odo dev", func() {
			err := helper.RunDevMode(func(session *gexec.Session, outContents []byte, errContents []byte, urls []string) {
				annotations := commonVar.CliRunner.GetAnnotationsDeployment(devfileCmpName, "app", commonVar.Project)
				var valueFound bool
				for key, value := range annotations {
					if key == "app.openshift.io/vcs-uri" && value == remoteURL {
						valueFound = true
						break
					}
				}
				Expect(valueFound).To(BeTrue())
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	// Tests https://github.com/redhat-developer/odo/issues/3838
	When("java-springboot application is created and running odo dev", func() {
		var session helper.DevSession
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-registry.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			var err error
			session, _, _, _, err = helper.StartDevMode("-v", "4")
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		When("Update the devfile.yaml", func() {

			var outC []byte
			BeforeEach(func() {
				helper.ReplaceString("devfile.yaml", "memoryLimit: 1024Mi", "memoryLimit: 1023Mi")
				var err error
				outC, _, err = session.WaitSync()
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should build the application successfully", func() {
				Expect(string(outC)).To(ContainSubstring("BUILD SUCCESS"))
			})

			When("compare the local and remote files", func() {

				remoteFiles := []string{}
				localFiles := []string{}

				BeforeEach(func() {
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					commonVar.CliRunner.PodsShouldBeRunning(commonVar.Project, podName)
					output := commonVar.CliRunner.Exec(podName, commonVar.Project, "find", "/projects")
					outputArr := strings.Split(output, "\n")
					for _, line := range outputArr {

						if !strings.HasPrefix(line, "/projects"+"/") || strings.Contains(line, "lost+found") {
							continue
						}

						newLine, err := filepath.Rel("/projects", line)
						Expect(err).ToNot(HaveOccurred())

						newLine = filepath.ToSlash(newLine)
						if strings.HasPrefix(newLine, "target/") || newLine == "target" || strings.HasPrefix(newLine, ".") {
							continue
						}

						remoteFiles = append(remoteFiles, newLine)
					}

					// 5) Acquire file from local context, filtering out .*
					err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						newPath := filepath.ToSlash(path)

						if strings.HasPrefix(newPath, ".") {
							return nil
						}

						localFiles = append(localFiles, newPath)
						return nil
					})
					Expect(err).ToNot(HaveOccurred())
				})

				It("localFiles and remoteFiles should match", func() {
					sort.Strings(localFiles)
					sort.Strings(remoteFiles)
					Expect(localFiles).To(Equal(remoteFiles))
				})
			})
		})
	})

	When("node-js application is created and deployed with devfile schema 2.2.0", func() {

		ensureResource := func(cpulimit, cpurequest, memoryrequest string) {
			By("check for cpuLimit", func() {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				bufferOutput := commonVar.CliRunner.Run("get", "pods", podName, "-o", "jsonpath='{.spec.containers[0].resources.limits.cpu}'").Out.Contents()
				output := string(bufferOutput)
				Expect(output).To(ContainSubstring(cpulimit))
			})

			By("check for cpuRequests", func() {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				bufferOutput := commonVar.CliRunner.Run("get", "pods", podName, "-o", "jsonpath='{.spec.containers[0].resources.requests.cpu}'").Out.Contents()
				output := string(bufferOutput)
				Expect(output).To(ContainSubstring(cpurequest))
			})

			By("check for memoryRequests", func() {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				bufferOutput := commonVar.CliRunner.Run("get", "pods", podName, "-o", "jsonpath='{.spec.containers[0].resources.requests.memory}'").Out.Contents()
				output := string(bufferOutput)
				Expect(output).To(ContainSubstring(memoryrequest))
			})
		}

		var session helper.DevSession
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-MR-CL-CR.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should check cpuLimit, cpuRequests, memoryRequests", func() {
			ensureResource("1", "200m", "512Mi")
		})

		When("Update the devfile.yaml, and waiting synchronization", func() {

			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-MR-CL-CR-modified.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				var err error
				_, _, err = session.WaitSync()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should check cpuLimit, cpuRequests, memoryRequests after restart", func() {
				ensureResource("700m", "250m", "550Mi")
			})
		})
	})

	When("creating nodejs component, doing odo dev and run command has dev.odo.push.path attribute", func() {
		var session helper.DevSession
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-remote-attributes.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

			// create a folder and file which shouldn't be pushed
			helper.MakeDir(filepath.Join(commonVar.Context, "views"))
			_, _ = helper.CreateSimpleFile(filepath.Join(commonVar.Context, "views"), "view", ".html")

			helper.ReplaceString("package.json", "node server.js", "node server/server.js")
			var err error
			session, _, _, _, err = helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
		})

		It("should sync only the mentioned files at the appropriate remote destination", func() {
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			stdOut := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
			helper.MatchAllInOutput(stdOut, []string{"package.json", "server"})
			helper.DontMatchAllInOutput(stdOut, []string{"test", "views", "devfile.yaml"})

			stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects/server")
			helper.MatchAllInOutput(stdOut, []string{"server.js", "test"})
		})
	})

	Context("using Kubernetes cluster", func() {
		BeforeEach(func() {
			if os.Getenv("KUBERNETES") != "true" {
				Skip("This is a Kubernetes specific scenario, skipping")
			}
		})

		It("should run odo dev successfully on default namespace", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

			session, _, errContents, _, err := helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
			defer session.Stop()
			helper.DontMatchAllInOutput(string(errContents), []string{"odo may not work as expected in the default project"})
		})
	})

	/* TODO(feloy) Issue #5591
	Context("using OpenShift cluster", func() {
		BeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}
		})
		It("should run odo dev successfully on default namespace", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

			session, _, errContents, err := helper.StartDevMode()
			Expect(err).ToNot(HaveOccurred())
			defer session.Stop()
			helper.MatchAllInOutput(string(errContents), []string{"odo may not work as expected in the default project"})
		})
	})
	*/

	//Test reused and adapted from the now-removed `cmd_devfile_delete_test.go`.
	// cf. https://github.com/redhat-developer/odo/blob/24fd02673d25eb4c7bb166ec3369554a8e64b59c/tests/integration/devfile/cmd_devfile_delete_test.go#L172-L238
	When("a component with endpoints is bootstrapped and pushed", func() {

		BeforeEach(func() {
			cmpName = "nodejs-with-endpoints"
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path",
				helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-multiple-endpoints.yaml")).ShouldPass()

			devSession, _, _, _, err := helper.StartDevMode()
			Expect(err).ShouldNot(HaveOccurred())
			devSession.Kill()
		})

		It("should not create Ingress or Route resources in the cluster", func() {
			// Pod should exist
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			Expect(podName).NotTo(BeEmpty())
			services := commonVar.CliRunner.GetServices(commonVar.Project)
			Expect(services).To(SatisfyAll(
				Not(BeEmpty()),
				ContainSubstring(fmt.Sprintf("%s-app", cmpName)),
			))

			ingressesOut := commonVar.CliRunner.Run("get", "ingress",
				"-n", commonVar.Project,
				"-o", "custom-columns=NAME:.metadata.name",
				"--no-headers").Out.Contents()
			ingresses, err := helper.ExtractLines(string(ingressesOut))
			Expect(err).To(BeNil())
			Expect(ingresses).To(BeEmpty())

			if !helper.IsKubernetesCluster() {
				routesOut := commonVar.CliRunner.Run("get", "routes",
					"-n", commonVar.Project,
					"-o", "custom-columns=NAME:.metadata.name",
					"--no-headers").Out.Contents()
				routes, err := helper.ExtractLines(string(routesOut))
				Expect(err).To(BeNil())
				Expect(routes).To(BeEmpty())
			}
		})
	})
})
