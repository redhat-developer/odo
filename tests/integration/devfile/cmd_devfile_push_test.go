package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/tests/helper"
	"github.com/openshift/odo/tests/integration/devfile/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo devfile push command tests", func() {
	var cmpName string
	var sourcePath = "/projects"
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

	Context("verify devfile push works", func() {

		It("checks that odo push works with a devfile", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("git", "init").ShouldPass()
			remote := "origin"
			remoteURL := "https://github.com/odo-devfiles/nodejs-ex"
			helper.Cmd("git", "remote", "add", remote, remoteURL).ShouldPass()

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			annotations := commonVar.CliRunner.GetAnnotationsDeployment(cmpName, "app", commonVar.Project)
			var valueFound bool
			for key, value := range annotations {
				if key == "app.openshift.io/vcs-uri" && value == remoteURL {
					valueFound = true
				}
			}
			Expect(valueFound).To(BeTrue())

			// update devfile and push again
			helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
			helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
		})

		When("JSON output is requested", func() {
			It("should display output in JSON format", func() {

				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
				utils.AnalyzePushConsoleOutput(output)

				// update devfile and push again
				helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
				output = helper.Cmd("odo", "push", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
				utils.AnalyzePushConsoleOutput(output)

			})
		})

		When("pushing devfile without an .odo folder", func() {

			It("should be able to push based on name passed", func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output := helper.Cmd("odo", "push", "--project", commonVar.Project, "springboot").ShouldPass().Out()
				Expect(output).To(ContainSubstring("Executing devfile commands for component springboot"))
			})

			It("should error out on devfile flag", func() {
				helper.Cmd("odo", "push", "--project", commonVar.Project, "--devfile", "invalid.yaml").ShouldFail()
			})

		})

		When("not in context directory", func() {
			It("should push correct component based on --context flag", func() {
				newContext := helper.CreateNewContext()
				defer helper.DeleteDir(newContext)
				helper.Chdir(newContext)

				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), newContext)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(newContext, "devfile.yaml"))

				helper.Chdir(commonVar.Context)
				output := helper.Cmd("odo", "push", "--context", newContext).ShouldPass().Out()
				Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
			})
		})

		When("Devfile 2.1.0 is used", func() {
			It("devfile variables should work", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-variables.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push").ShouldPass().Out()
				Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

				envVars := commonVar.CliRunner.GetEnvsDevFileDeployment(cmpName, "app", commonVar.Project)
				// check if the env variable has a correct value. This value was substituted from in devfile from variable
				Expect(envVars["FOO"]).To(Equal("bar"))

			})
		})
	})

	Context("verify files are correctly synced", func() {

		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			utils.ExecPushToTestFileChanges(commonVar.Context, cmpName, commonVar.Project)
		})

		When("local files are created and deleted ", func() {
			It("should correctly propagate changes to the container", func() {
				newFilePath := filepath.Join(commonVar.Context, "foobar.txt")
				newDirPath := filepath.Join(commonVar.Context, "testdir")
				utils.ExecPushWithNewFileAndDir(commonVar.Context, cmpName, commonVar.Project, newFilePath, newDirPath)

				// Check to see if it's been pushed (foobar.txt abd directory testdir)
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

				stdOut := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
				helper.MatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})

				// Now we delete the file and dir and push
				helper.DeleteDir(newFilePath)
				helper.DeleteDir(newDirPath)
				helper.Cmd("odo", "push", "--project", commonVar.Project, "-v4").ShouldPass()

				// Then check to see if it's truly been deleted
				stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
				helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
			})
		})

		When("devfile is modified", func() {
			// Tests https://github.com/openshift/odo/issues/3838
			ensureFilesSyncedTest := func(namespace string, shouldForcePush bool) {
				helper.Cmd("odo", "create", "java-springboot", "--project", commonVar.Project, cmpName).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)

				fmt.Fprintf(GinkgoWriter, "Testing with force push %v", shouldForcePush)

				// 1) Push a standard spring boot project
				output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

				// 2) Update the devfile.yaml, causing push to redeploy the component
				helper.ReplaceString("devfile.yaml", "memoryLimit: 768Mi", "memoryLimit: 769Mi")
				commands := []string{"push", "-v", "4", "--project", commonVar.Project}
				if shouldForcePush {
					// Test both w/ and w/o '-f'
					commands = append(commands, "-f")
				}

				// 3) Ensure the build passes, indicating that all files were correctly synced to the new pod
				output = helper.Cmd("odo", commands...).ShouldPass().Out()
				Expect(output).To(ContainSubstring("BUILD SUCCESS"))

				// 4) Acquire files from remote container, filtering out target/* and .*
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, namespace)
				output = commonVar.CliRunner.Exec(podName, namespace, "find", sourcePath)
				remoteFiles := []string{}
				outputArr := strings.Split(output, "\n")
				for _, line := range outputArr {

					if !strings.HasPrefix(line, sourcePath+"/") || strings.Contains(line, "lost+found") {
						continue
					}

					newLine, err := filepath.Rel(sourcePath, line)
					Expect(err).ToNot(HaveOccurred())

					newLine = filepath.ToSlash(newLine)
					if strings.HasPrefix(newLine, "target/") || newLine == "target" || strings.HasPrefix(newLine, ".") {
						continue
					}

					remoteFiles = append(remoteFiles, newLine)
				}

				// 5) Acquire file from local context, filtering out .*
				localFiles := []string{}
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

				// 6) Sort and compare the local and remote files; they should match
				sort.Strings(localFiles)
				sort.Strings(remoteFiles)
				Expect(localFiles).To(Equal(remoteFiles))
			}

			When("odo push -f is executed", func() {

				It("should correctly sync files after pod redeploy", func() {
					ensureFilesSyncedTest(commonVar.Project, true)
				})
			})
			When("odo push (without -f) is executed", func() {

				It("should correctly sync files after pod redeploy", func() {
					ensureFilesSyncedTest(commonVar.Project, false)
				})
			})
		})

		When("run command has dev.odo.push.path attribute", func() {
			It("should push only the mentioned files at the appropriate remote destination", func() {
				helper.Cmd("odo", "create", "nodejs", cmpName, "--context", commonVar.Context, "--project", commonVar.Project).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-remote-attributes.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				// create a folder and file which shouldn't be pushed
				helper.MakeDir(filepath.Join(commonVar.Context, "views"))
				_, _ = helper.CreateSimpleFile(filepath.Join(commonVar.Context, "views"), "view", ".html")

				helper.ReplaceString("package.json", "node server.js", "node server/server.js")
				helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				stdOut := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
				helper.MatchAllInOutput(stdOut, []string{"package.json", "server"})
				helper.DontMatchAllInOutput(stdOut, []string{"test", "views", "devfile.yaml"})

				stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath+"/server")
				helper.MatchAllInOutput(stdOut, []string{"server.js", "test"})

				stdOut = helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
				Expect(stdOut).To(ContainSubstring("No file changes detected"))
			})
		})

		When("devfile has sourcemappings", func() {
			It("should sync files to the correct location", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileSourceMapping.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

				// Verify source code was synced to /test instead of /projects
				var statErr error
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
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

		When("project and clonePath is present", func() {
			It("should sync to the correct dir in container", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

				// devfile with clonePath set in project field
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				helper.Cmd("odo", "push", "--v", "5").ShouldPass()
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
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

		When("devfile project field is present", func() {
			It("should sync to the correct dir in container", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				// reset clonePath and change the workdir accordingly, it should sync to project name
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "clonePath: webapp/", "# clonePath: webapp/")

				helper.Cmd("odo", "push").ShouldPass()

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/apps/nodeshift")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/apps/nodeshift", "/apps", commonVar.CliRunner)
			})
		})

		When("multiple project is present", func() {
			It("should sync to the correct dir in container", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.Cmd("odo", "push").ShouldPass()
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				// for devfile-with-multiple-projects.yaml source mapping is not set so $PROJECTS_ROOT is /projects
				// multiple projects, so source code would sync to the first project /projects/webapp
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects/webapp")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/projects/webapp", "/projects", commonVar.CliRunner)
			})
		})

		When("no project is present", func() {
			It("should sync to the correct dir in container", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.Cmd("odo", "push").ShouldPass()
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/projects", "/projects", commonVar.CliRunner)
			})
		})
	})

	Context("verify devfile volume components work", func() {

		It("should create pvc and reuse if it shares the same devfile volume name", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volumes.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{
				"Executing devbuild command",
				"Executing devrun command",
			})

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

			var statErr error
			var cmdOutput string

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

			deploymentName, err := util.NamespaceKubernetesObject(cmpName, "app")
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

		It("should successfully use the volume components in container components", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// Verify the pvc size for firstvol
			storageSize := commonVar.CliRunner.GetPVCSize(cmpName, "firstvol", commonVar.Project)
			// should be the default size
			Expect(storageSize).To(ContainSubstring("1Gi"))

			// Verify the pvc size for secondvol
			storageSize = commonVar.CliRunner.GetPVCSize(cmpName, "secondvol", commonVar.Project)
			// should be the specified size in the devfile volume component
			Expect(storageSize).To(ContainSubstring("3Gi"))
		})

	})

	Context("verify command executions", func() {

		When("odo push --debug is executed", func() {
			It("should execute debug commands", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push", "--debug", "--project", commonVar.Project).ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{
					"Executing devbuild command",
					"Executing debugrun command",
				})

				output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{
					"Executing devbuild command",
					"Executing devrun command",
				})

			})
		})

		When("odo push -f is executed", func() {
			It("should build even when no changes are detected", func() {
				utils.ExecPushWithForceFlag(commonVar.Context, cmpName, commonVar.Project)
			})
		})

		When("default build and run commands are defined", func() {
			It("should execute correct commands", func() {
				utils.ExecDefaultDevfileCommands(commonVar.Context, cmpName, commonVar.Project)

				// Check to see if it's been pushed (foobar.txt abd directory testdir)
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

				var statErr error
				var cmdOutput string
				commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
					podName,
					"runtime",
					commonVar.Project,
					[]string{"ps", "-ef"},
					func(cmdOp string, err error) bool {
						cmdOutput = cmdOp
						statErr = err
						return err == nil
					},
				)
				Expect(statErr).ToNot(HaveOccurred())
				Expect(cmdOutput).To(ContainSubstring("spring-boot:run"))
			})
		})

		When("run command is not marked as hotReloadCapable", func() {
			It("should restart the application", func() {
				// TODO: this is almost the same test as one below
				helper.Cmd("odo", "create", "nodejs", cmpName, "--project", commonVar.Project).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push").ShouldPass().Out()
				Expect(output).To(ContainSubstring("Executing devrun command \"npm start\""))

				helper.Cmd("odo", "push", "-f").ShouldPass()

				logs := helper.Cmd("odo", "log").ShouldPass().Out()
				Expect(logs).To(ContainSubstring("stop the program"))

			})
		})

		When("run command is marked as hotReloadCapable:true", func() {
			It("should not restart the application", func() {
				helper.Cmd("odo", "create", "nodejs", cmpName, "--project", commonVar.Project).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-hotReload.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push").ShouldPass().Out()
				Expect(output).To(ContainSubstring("Executing devrun command \"npm start\""))

				helper.Cmd("odo", "push", "-f").ShouldPass()

				logs := helper.Cmd("odo", "log").ShouldPass().Out()
				Expect(logs).To(ContainSubstring("Don't start program again, program is already started"))

			})

			When("run mode is changed to debug", func() {
				It("should restart the application regardless of hotReloadCapable value", func() {
					helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-hotReload.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()

					stdOut := helper.Cmd("odo", "push", "--debug", "--project", commonVar.Project).ShouldPass().Out()
					Expect(stdOut).To(Not(ContainSubstring("No file changes detected, skipping build")))

					logs := helper.Cmd("odo", "log").ShouldPass().Out()

					helper.MatchAllInOutput(logs, []string{
						"\"stop the program\" program=debugrun",
						"\"stop the program\" program=devrun",
					})

				})
			})
		})

		When("devfile with composite command", func() {
			It("should execute all commands in composite commmand", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push").ShouldPass().Out()
				Expect(output).To(ContainSubstring("Executing mkdir command"))

				// Verify the command executed successfully
				var statErr error
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
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

			When("composite command is marked as paralell:true ", func() {
				It("should execute all commands in composite commmand", func() {
					helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommandsParallel.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

					output := helper.Cmd("odo", "push", "--build-command", "buildandmkdir").ShouldPass().Out()
					Expect(output).To(ContainSubstring("Executing mkdir command"))

					// Verify the command executed successfully
					var statErr error
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
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

			When("composite command are nested", func() {
				It("should execute all commands in composite commmand", func() {
					helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileNestedCompCommands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

					// Verify nested command was executed
					output := helper.Cmd("odo", "push").ShouldPass().Out()
					Expect(output).To(ContainSubstring("Executing mkdir command"))

					// Verify the command executed successfully
					var statErr error
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
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
			When("composite command is used as a run command", func() {
				It("should throw a validation error for composite run commands", func() {
					helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeRun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

					// Verify odo push failed
					output := helper.Cmd("odo", "push").ShouldFail().Err()
					Expect(output).To(ContainSubstring("not supported currently"))
				})
			})
		})

		When("events are defined", func() {

			It("should correctly execute PreStart commands", func() {
				// expectedInitContainers := []string{"tools-myprestart-1", "tools-myprestart-2", "runtime-secondprestart-3"}

				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-preStart.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldFail().Err()
				// This is expected to fail for now.
				// see https://github.com/openshift/odo/issues/4187 for more info
				helper.MatchAllInOutput(output, []string{"myprestart should either map to an apply command or a composite command with apply commands\n"})

				/*
					helper.MatchAllInOutput(output, []string{"PreStart commands have been added to the component"})

					firstPushPodName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

					firstPushInitContainers := commonVar.CliRunner.GetPodInitContainers(cmpName, commonVar.Project)
					// 3 preStart events + 1 supervisord init containers
					Expect(len(firstPushInitContainers)).To(Equal(4))
					helper.MatchAllInOutput(strings.Join(firstPushInitContainers, ","), expectedInitContainers)

					// Need to force so build and run get triggered again with the component already created.
					output = helper.Cmd("odo", "push", "--project", commonVar.Project, "-f").ShouldPass().Out()
					helper.MatchAllInOutput(output, []string{"PreStart commands have been added to the component"})

					secondPushPodName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

					secondPushInitContainers := commonVar.CliRunner.GetPodInitContainers(cmpName, commonVar.Project)

					Expect(len(secondPushInitContainers)).To(Equal(4))
					helper.MatchAllInOutput(strings.Join(secondPushInitContainers, ","), expectedInitContainers)

					Expect(firstPushPodName).To(Equal(secondPushPodName))
					Expect(firstPushInitContainers).To(Equal(secondPushInitContainers))

					var statErr error
					commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
						firstPushPodName,
						"runtime",
						commonVar.Project,
						[]string{"cat", "/projects/test.txt"},
						func(cmdOp string, err error) bool {
							if err != nil {
								statErr = err
							} else if cmdOp == "" {
								statErr = fmt.Errorf("prestart event action error, expected: hello test2\nhello test2\nhello test\n, got empty string")
							} else {
								fileContents := strings.Split(cmdOp, "\n")
								if len(fileContents)-1 != 3 {
									statErr = fmt.Errorf("prestart event action count error, expected: 3 strings, got %d strings: %s", len(fileContents), strings.Join(fileContents, ","))
								} else if cmdOp != "hello test2\nhello test2\nhello test\n" {
									statErr = fmt.Errorf("prestart event action error, expected: hello test2\nhello test2\nhello test\n, got: %s", cmdOp)
								}
							}

							return true
						},
					)
					Expect(statErr).ToNot(HaveOccurred())
				*/
			})

			It("should correctly execute PostStart commands", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{"Executing mypoststart command \"echo I am a PostStart\"", "Executing secondpoststart command \"echo I am also a PostStart\""})

				// Need to force so build and run get triggered again with the component already created.
				output = helper.Cmd("odo", "push", "--project", commonVar.Project, "-f").ShouldPass().Out()
				helper.DontMatchAllInOutput(output, []string{"Executing mypoststart command \"echo I am a PostStart\"", "Executing secondpoststart command \"echo I am also a PostStart\""})
				helper.MatchAllInOutput(output, []string{
					"Executing devbuild command",
					"Executing devrun command",
				})
			})
		})

		When("using custom commands (specified by flags)", func() {
			It("should push successfully", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Project)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output := helper.Cmd("odo", "push", "--build-command", "build", "--run-command", "run").ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{
					"Executing build command \"npm install\"",
					"Executing run command \"npm start\"",
				})

			})

			It("should error out on a wrong custom commands", func() {
				helper.Cmd("odo", "create", "nodejs", cmpName, "--project", commonVar.Project).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output := helper.Cmd("odo", "push", "--build-command", "buildgarbage").ShouldFail().Err()
				Expect(output).NotTo(ContainSubstring("Executing buildgarbage command"))
				Expect(output).To(ContainSubstring("the command \"%v\" is not found in the devfile", "buildgarbage"))
			})

		})

		When("command has no group kind", func() {
			It("should execute commands with flags", func() {
				helper.Cmd("odo", "create", "nodejs", cmpName, "--project", commonVar.Project).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-no-group-kind.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output := helper.Cmd("odo", "push", "--build-command", "devbuild", "--run-command", "devrun").ShouldPass().Out()
				helper.MatchAllInOutput(output, []string{
					"Executing devbuild command \"npm install\"",
					"Executing devrun command \"npm start\"",
				})

			})
		})

		When("the run command throws an error", func() {
			It("should wait and error out with some log", func() {
				helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm start", "npm starts")

				_, output := helper.CmdShouldPassIncludeErrStream("odo", "push")
				helper.MatchAllInOutput(output, []string{
					"exited with error status within 1 sec",
					"Did you mean one of these?",
				})

				_, output = helper.CmdShouldPassIncludeErrStream("odo", "push", "-f", "--run-command", "run")
				helper.MatchAllInOutput(output, []string{
					"exited with error status within 1 sec",
					"Did you mean one of these?",
				})
			})
		})

		When("commands specify have env variables", func() {
			When("sigle env var is set", func() {
				It("should be able to exec command", func() {
					helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-envs.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
					output := helper.Cmd("odo", "push", "--build-command", "buildwithenv", "--run-command", "singleenv").ShouldPass().Out()
					helper.MatchAllInOutput(output, []string{"mkdir $ENV1", "mkdir $BUILD_ENV1"})

					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
					helper.MatchAllInOutput(output, []string{"test_env_variable", "test_build_env_variable"})
				})
			})
			When("multiple env variables are set", func() {
				It("should be able to exec command", func() {
					helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-envs.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
					output := helper.Cmd("odo", "push", "--build-command", "buildwithmultipleenv", "--run-command", "multipleenv").ShouldPass().Out()
					helper.MatchAllInOutput(output, []string{"mkdir $ENV1 $ENV2", "mkdir $BUILD_ENV1 $BUILD_ENV2"})

					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
					helper.MatchAllInOutput(output, []string{"test_build_env_variable1", "test_build_env_variable2", "test_env_variable1", "test_env_variable2"})
				})
			})
			When("there is a env variable with spaces", func() {
				It("should be able to exec command", func() {
					helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-envs.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
					output := helper.Cmd("odo", "push", "--build-command", "buildenvwithspace", "--run-command", "envwithspace").ShouldPass().Out()
					helper.MatchAllInOutput(output, []string{"mkdir \\\"$ENV1\\\"", "mkdir \\\"$BUILD_ENV1\\\""})

					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
					helper.MatchAllInOutput(output, []string{"build env variable with space", "env with space"})

				})
			})
		})
	})

	Context("using OpenShift cluster", func() {
		JustBeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}
		})
		When("project with with 'default' name is used", func() {
			It("should throw an error", func() {
				componentName := helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.Cmd("odo", "create", "nodejs", "--project", "default", componentName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				stdout := helper.Cmd("odo", "push").ShouldFail().Err()
				helper.MatchAllInOutput(stdout, []string{"odo may not work as expected in the default project, please run the odo component in a non-default project"})
			})
		})
	})

	Context("using Kubernetes cluster", func() {
		JustBeforeEach(func() {
			if os.Getenv("KUBERNETES") != "true" {
				Skip("This is a Kubernetes specific scenario, skipping")
			}
		})
		When("project with with 'default' name is used", func() {

			It("should push successfully", func() {
				componentName := helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.Cmd("odo", "create", "nodejs", "--project", "default", componentName).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				stdout := helper.Cmd("odo", "push").ShouldPass().Out()
				helper.DontMatchAllInOutput(stdout, []string{"odo may not work as expected in the default project"})
			})
		})
	})

})
