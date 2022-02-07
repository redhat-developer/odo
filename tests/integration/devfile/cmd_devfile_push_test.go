package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/tests/helper"
	"github.com/redhat-developer/odo/tests/integration/devfile/utils"

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

	When("creating a nodejs component", func() {
		output := ""
		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
		})
		When("setting git config and running odo push", func() {
			remoteURL := "https://github.com/odo-devfiles/nodejs-ex"
			BeforeEach(func() {
				helper.Cmd("git", "init").ShouldPass()
				remote := "origin"
				helper.Cmd("git", "remote", "add", remote, remoteURL).ShouldPass()
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})

			It("checks that odo push works with a devfile", func() {
				Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
			})

			It("check annotations from the deployment after odo push", func() {

				annotations := commonVar.CliRunner.GetAnnotationsDeployment(cmpName, "app", commonVar.Project)
				var valueFound bool
				for key, value := range annotations {
					if key == "app.openshift.io/vcs-uri" && value == remoteURL {
						valueFound = true
					}
				}
				Expect(valueFound).To(BeTrue())
			})

			When("updating a variable into devfile", func() {
				BeforeEach(func() {
					helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
				})

				It("should run odo push successfully", func() {
					helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
				})

			})
		})

		When("odo push is executed with json output", func() {

			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
			})

			It("should display push output in JSON format", func() {

				utils.AnalyzePushConsoleOutput(output)
			})
			When("update devfile and push again", func() {

				BeforeEach(func() {
					helper.ReplaceString("devfile.yaml", "name: FOO", "name: BAR")
					output = helper.Cmd("odo", "push", "-o", "json", "--project", commonVar.Project).ShouldPass().Out()
				})

				It("should display push updated output in JSON format", func() {

					utils.AnalyzePushConsoleOutput(output)
				})
			})
		})

		When("running odo push outside the context directory", func() {
			newContext := ""
			BeforeEach(func() {
				newContext = helper.CreateNewContext()
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				helper.Chdir(newContext)
				output = helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass().Out()
			})

			AfterEach(func() {
				helper.Chdir(commonVar.Context)
				helper.DeleteDir(newContext)
			})

			It("should push correct component based on --context flag", func() {

				Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
			})
		})

		When("Devfile 2.1.0 is used", func() {

			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-variables.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			})

			When("doing odo push", func() {

				BeforeEach(func() {
					output = helper.Cmd("odo", "push").ShouldPass().Out()
				})

				It("should pass", func() {

					Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
				})
				It("should check if the env variable has a correct value", func() {

					envVars := commonVar.CliRunner.GetEnvsDevFileDeployment(cmpName, "app", commonVar.Project)
					// check if the env variable has a correct value. This value was substituted from in devfile from variable
					Expect(envVars["FOO"]).To(Equal("bar"))
				})
			})
		})
		When("doing odo push", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
			})
			When("doing odo push again", func() {
				BeforeEach(func() {

					output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				})
				It("should not build when no changes are detected in the directory", func() {
					Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
				})
			})
			When("making changes in file and doing odo push again", func() {
				BeforeEach(func() {
					helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "Hello from Node.js", "UPDATED!")
					output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				})
				It("should build when a file change is detected", func() {
					Expect(output).To(ContainSubstring("Syncing files to the component"))
				})
			})
			When("doing odo push with -f flag", func() {
				BeforeEach(func() {
					output = helper.Cmd("odo", "push", "-f", "--project", commonVar.Project).ShouldPass().Out()
				})
				It("should build even when no changes are detected", func() {
					Expect(output).To(Not(ContainSubstring("No file changes detected, skipping build")))
				})
			})

			When("the pod is deleted and replaced", func() {
				BeforeEach(func() {
					oldPod := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					commonVar.CliRunner.DeletePod(oldPod, commonVar.Project)
					Eventually(func() bool {
						newPod := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
						return newPod != "" && newPod != oldPod
					}, 180, 10).Should(Equal(true))
					newPod := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					commonVar.CliRunner.PodsShouldBeRunning(commonVar.Project, newPod)
				})

				It("should execute run command on odo push", func() {
					output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
					Expect(output).To(ContainSubstring("Executing devrun command"))
				})
			})

			When("the Deployment's Replica count (pods) is scaled to 0", func() {
				BeforeEach(func() {
					commonVar.CliRunner.ScalePodToZero(cmpName, "app", commonVar.Project)
				})

				It("odo push should successfully recreate the pod", func() {
					helper.Cmd("odo", "push").ShouldPass()
				})
			})
		})

		When("creating local files and dir and doing odo push", func() {
			var newDirPath, newFilePath, stdOut, podName string
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
				helper.Cmd("odo", "push", "--project", commonVar.Project, "-v4").ShouldPass()
			})

			It("should correctly propagate changes to the container", func() {

				// Check to see if it's been pushed (foobar.txt abd directory testdir)
				podName = commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

				stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
				helper.MatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
			})
			When("deleting local files and dir and doing odo push again", func() {
				BeforeEach(func() {
					// Now we delete the file and dir and push
					helper.DeleteDir(newFilePath)
					helper.DeleteDir(newDirPath)
					helper.Cmd("odo", "push", "--project", commonVar.Project, "-v4").ShouldPass()
				})
				It("should not list deleted dir and file in container", func() {
					podName = commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					// Then check to see if it's truly been deleted
					stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
					helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt", "testdir"})
				})
			})
		})

		When("devfile has sourcemappings and doing odo push", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileSourceMapping.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})
			It("should sync files to the correct location", func() {

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

		When("project and clonePath is present in devfile and doing odo push", func() {
			BeforeEach(func() {
				// devfile with clonePath set in project field
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				helper.Cmd("odo", "push", "--v", "5").ShouldPass()
			})

			It("should sync to the correct dir in container", func() {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				// source code is synced to $PROJECTS_ROOT/clonePath
				// $PROJECTS_ROOT is /projects by default, if sourceMapping is set it is same as sourceMapping
				// for devfile-with-projects.yaml, sourceMapping is apps and clonePath is webapp
				// so source code would be synced to /apps/webapp
				output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/apps/webapp")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/apps/webapp", "/apps", commonVar.CliRunner)
			})
		})

		When("devfile project field is present and doing odo push", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				// reset clonePath and change the workdir accordingly, it should sync to project name
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "clonePath: webapp/", "# clonePath: webapp/")
				helper.Cmd("odo", "push").ShouldPass()
			})
			It("should sync to the correct dir in container", func() {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/apps/nodeshift")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/apps/nodeshift", "/apps", commonVar.CliRunner)
			})
		})

		When("multiple project is present", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.Cmd("odo", "push").ShouldPass()
			})
			It("should sync to the correct dir in container", func() {

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				// for devfile-with-multiple-projects.yaml source mapping is not set so $PROJECTS_ROOT is /projects
				// multiple projects, so source code would sync to the first project /projects/webapp
				output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects/webapp")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/projects/webapp", "/projects", commonVar.CliRunner)
			})
		})

		When("no project is present", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.Cmd("odo", "push").ShouldPass()
			})
			It("should sync to the correct dir in container", func() {

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				utils.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/projects", "/projects", commonVar.CliRunner)
			})
		})

		When("doing odo push with devfile contain volume", func() {
			var statErr error
			var cmdOutput string

			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volumes.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})
			It("should create pvc and reuse if it shares the same devfile volume name", func() {

				helper.MatchAllInOutput(output, []string{
					"Executing devbuild command",
					"Executing devrun command",
				})

				// Check to see if it's been pushed (foobar.txt abd directory testdir)
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

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

		})

		When("doing odo push with devfile containing volume-component", func() {
			BeforeEach(func() {
				helper.RenameFile("devfile.yaml", "devfile-old.yaml")
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})
			It("should successfully use the volume components in container components", func() {

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

		When("doing odo push --debug and devfile contain debugrun", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "--debug", "--project", commonVar.Project).ShouldPass().Out()
			})
			It("should execute debug commands", func() {
				helper.MatchAllInOutput(output, []string{
					"Executing devbuild command",
					"Executing debugrun command",
				})
			})
			When("doing odo push", func() {
				BeforeEach(func() {
					output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				})
				It("should execute dev commands", func() {
					helper.MatchAllInOutput(output, []string{
						"Executing devbuild command",
						"Executing devrun command",
					})
				})
			})
		})

		When("doing odo push and run command is not marked as hotReloadCapable", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push").ShouldPass().Out()
			})
			It("should restart the application", func() {
				// TODO: this is almost the same test as one below

				Expect(output).To(ContainSubstring("Executing devrun command \"npm start\""))

				helper.Cmd("odo", "push", "-f").ShouldPass()
			})
		})

		When("doing odo push and run command is marked as hotReloadCapable:true", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-hotReload.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push").ShouldPass().Out()
			})

			It("should not restart the application", func() {
				Expect(output).To(ContainSubstring("Executing devrun command \"npm start\""))

				helper.Cmd("odo", "push", "-f").ShouldPass()
			})

			When("doing odo push --debug ", func() {
				stdOut := ""
				BeforeEach(func() {
					stdOut = helper.Cmd("odo", "push", "--debug", "--project", commonVar.Project).ShouldPass().Out()
				})
				It("should restart the application regardless of hotReloadCapable value", func() {

					Expect(stdOut).To(Not(ContainSubstring("No file changes detected, skipping build")))
				})
			})
		})

		When("doing odo push and devfile with composite command", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push").ShouldPass().Out()
			})
			It("should execute all commands in composite commmand", func() {

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
		When("doing odo push and composite command is marked as paralell:true ", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommandsParallel.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "--build-command", "buildandmkdir").ShouldPass().Out()
			})
			It("should execute all commands in composite commmand", func() {

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

		When("doing odo push and composite command are nested", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileNestedCompCommands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push").ShouldPass().Out()
			})
			It("should execute all commands in composite commmand", func() {
				// Verify nested command was executed

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
		When("doing odo push and composite command is used as a run command", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeRun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push").ShouldFail().Err()
			})
			It("should throw a validation error for composite run commands", func() {
				Expect(output).To(ContainSubstring("not supported currently"))
			})
		})

		When("events are defined", func() {

			It("should correctly execute PreStart commands", func() {
				// expectedInitContainers := []string{"tools-myprestart-1", "tools-myprestart-2", "runtime-secondprestart-3"}

				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-preStart.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldFail().Err()
				// This is expected to fail for now.
				// see https://github.com/redhat-developer/odo/issues/4187 for more info
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

				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
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

		When("doing odo push and using correct custom commands (specified by flags)", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "--build-command", "build", "--run-command", "run").ShouldPass().Out()
			})
			It("should push successfully", func() {
				helper.MatchAllInOutput(output, []string{
					"Executing build command \"npm install\"",
					"Executing run command \"npm start\"",
				})

			})
		})
		When("doing odo push and using wrong custom commands (specified by flags)", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "--build-command", "buildgarbage").ShouldFail().Err()
			})
			It("should error out", func() {

				Expect(output).NotTo(ContainSubstring("Executing buildgarbage command"))
				Expect(output).To(ContainSubstring("the command \"%v\" is not found in the devfile", "buildgarbage"))
			})

		})

		When("command has no group kind and doing odo push", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-no-group-kind.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "--build-command", "devbuild", "--run-command", "devrun").ShouldPass().Out()
			})
			It("should execute commands with flags", func() {
				helper.MatchAllInOutput(output, []string{
					"Executing devbuild command \"npm install\"",
					"Executing devrun command \"npm start\"",
				})

			})
		})

		When("doing odo push and run command throws an error", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm start", "npm starts")
				_, output = helper.Cmd("odo", "push").ShouldPass().OutAndErr()
			})

			It("should wait and error out with some log", func() {

				helper.MatchAllInOutput(output, []string{
					"exited with error status within 1 sec",
					"Did you mean one of these?",
				})

				_, output = helper.Cmd("odo", "push", "-f", "--run-command", "run").ShouldPass().OutAndErr()
				helper.MatchAllInOutput(output, []string{
					"exited with error status within 1 sec",
					"Did you mean one of these?",
				})
			})
		})

		When("commands specify have env variables", func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-envs.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			})
			When("doing odo push and sigle env var is set", func() {
				BeforeEach(func() {
					output = helper.Cmd("odo", "push", "--build-command", "buildwithenv", "--run-command", "singleenv").ShouldPass().Out()
				})
				It("should be able to exec command", func() {

					helper.MatchAllInOutput(output, []string{"mkdir $ENV1", "mkdir $BUILD_ENV1"})

					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
					helper.MatchAllInOutput(output, []string{"test_env_variable", "test_build_env_variable"})
				})
			})
			When("doing odo push and multiple env variables are set", func() {
				BeforeEach(func() {
					output = helper.Cmd("odo", "push", "--build-command", "buildwithmultipleenv", "--run-command", "multipleenv").ShouldPass().Out()
				})
				It("should be able to exec command", func() {

					helper.MatchAllInOutput(output, []string{"mkdir $ENV1 $ENV2", "mkdir $BUILD_ENV1 $BUILD_ENV2"})

					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
					helper.MatchAllInOutput(output, []string{"test_build_env_variable1", "test_build_env_variable2", "test_env_variable1", "test_env_variable2"})
				})
			})
			When("doing odo push and there is a env variable with spaces", func() {
				BeforeEach(func() {
					output = helper.Cmd("odo", "push", "--build-command", "buildenvwithspace", "--run-command", "envwithspace").ShouldPass().Out()
				})
				It("should be able to exec command", func() {

					helper.MatchAllInOutput(output, []string{"mkdir \\\"$ENV1\\\"", "mkdir \\\"$BUILD_ENV1\\\""})

					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
					helper.MatchAllInOutput(output, []string{"build env variable with space", "env with space"})

				})
			})
		})
	})

	Context("pushing devfile without an .odo folder", func() {
		output := ""

		It("should error out on odo push and passing invalid devfile", func() {
			helper.Cmd("odo", "push", "--project", commonVar.Project, "--devfile", "invalid.yaml").ShouldFail()
		})

		When("doing odo push", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				output = helper.Cmd("odo", "push", "--project", commonVar.Project, "springboot").ShouldPass().Out()
			})
			It("should be able to push based on name passed", func() {

				Expect(output).To(ContainSubstring("Executing devfile commands for component springboot"))
			})
		})
	})

	When("Create and push java-springboot component", func() {

		var output string

		BeforeEach(func() {
			helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "springboot", "devfile.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
		})
		It("should execute default build and run commands correctly", func() {

			helper.MatchAllInOutput(output, []string{
				"Executing defaultbuild command",
				"mvn clean",
				"Executing defaultrun command",
				"spring-boot:run",
			})

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

	Context("devfile is modified", func() {
		// Tests https://github.com/redhat-developer/odo/issues/3838
		ensureFilesSyncedTest := func(namespace string, shouldForcePush bool) {
			output := ""
			When("java-springboot application is created and pushed", func() {
				BeforeEach(func() {
					helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-registry.yaml")).ShouldPass()
					helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)

					fmt.Fprintf(GinkgoWriter, "Testing with force push %v", shouldForcePush)
					output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				})

				It("should push the component successfully", func() {
					Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
				})

				When("Update the devfile.yaml, do odo push", func() {

					BeforeEach(func() {
						helper.ReplaceString("devfile.yaml", "memoryLimit: 768Mi", "memoryLimit: 769Mi")
						commands := []string{"push", "-v", "4", "--project", commonVar.Project}
						if shouldForcePush {
							commands = append(commands, "-f")
						}

						output = helper.Cmd("odo", commands...).ShouldPass().Out()
					})

					It("Ensure the build passes", func() {
						Expect(output).To(ContainSubstring("BUILD SUCCESS"))
					})

					When("compare the local and remote files", func() {

						remoteFiles := []string{}
						localFiles := []string{}

						BeforeEach(func() {
							podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, namespace)
							output = commonVar.CliRunner.Exec(podName, namespace, "find", sourcePath)
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
		}

		Context("odo push -f is executed", func() {
			ensureFilesSyncedTest(commonVar.Project, true)
		})
		Context("odo push (without -f) is executed", func() {
			ensureFilesSyncedTest(commonVar.Project, false)
		})

		When("node-js application is created and pushed with devfile schema 2.2.0", func() {

			var output string
			BeforeEach(func() {
				helper.Cmd("odo", "create", "--project", commonVar.Project, cmpName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-MR-CL-CR.yaml")).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			})

			ensureResource := func(output, cpulimit, cpurequest, memoryrequest string) {
				By("check for cpuLimit", func() {
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					bufferOutput := commonVar.CliRunner.Run("get", "pods", podName, "-o", "jsonpath='{.spec.containers[0].resources.limits.cpu}'").Out.Contents()
					output = string(bufferOutput)
					Expect(output).To(ContainSubstring(cpulimit))
				})

				By("check for cpuRequests", func() {
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					bufferOutput := commonVar.CliRunner.Run("get", "pods", podName, "-o", "jsonpath='{.spec.containers[0].resources.requests.cpu}'").Out.Contents()
					output = string(bufferOutput)
					Expect(output).To(ContainSubstring(cpurequest))
				})

				By("check for memoryRequests", func() {
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					bufferOutput := commonVar.CliRunner.Run("get", "pods", podName, "-o", "jsonpath='{.spec.containers[0].resources.requests.memory}'").Out.Contents()
					output = string(bufferOutput)
					Expect(output).To(ContainSubstring(memoryrequest))
				})
			}

			It("should check cpuLimit,cpuRequests,memoryRequests", func() {
				ensureResource(output, "1", "200m", "512Mi")
			})

			When("Update the devfile.yaml, do odo push", func() {

				BeforeEach(func() {
					helper.ReplaceString("devfile.yaml", "cpuLimit: \"1\"", "cpuLimit: 700m")
					helper.ReplaceString("devfile.yaml", "cpuRequest: 200m", "cpuRequest: 250m")
					helper.ReplaceString("devfile.yaml", "memoryRequest: 512Mi", "memoryRequest: 550Mi")
					output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
				})

				It("should check cpuLimit,cpuRequests,memoryRequests", func() {
					ensureResource(output, "700m", "250m", "550Mi")
				})
			})
		})
	})

	When("creating nodejs component, doing odo push and run command has dev.odo.push.path attribute", func() {
		BeforeEach(func() {
			helper.Cmd("odo", "create", cmpName, "--context", commonVar.Context, "--project", commonVar.Project, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-remote-attributes.yaml")).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

			// create a folder and file which shouldn't be pushed
			helper.MakeDir(filepath.Join(commonVar.Context, "views"))
			_, _ = helper.CreateSimpleFile(filepath.Join(commonVar.Context, "views"), "view", ".html")

			helper.ReplaceString("package.json", "node server.js", "node server/server.js")
			helper.Cmd("odo", "push", "--context", commonVar.Context).ShouldPass()
		})
		It("should push only the mentioned files at the appropriate remote destination", func() {

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

	Context("using OpenShift cluster", func() {
		BeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}
		})
		When("project with with 'default' name is used", func() {
			It("should throw an error", func() {
				componentName := helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.Cmd("odo", "create", "--project", "default", componentName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

				stdout := helper.Cmd("odo", "push").ShouldFail().Err()
				helper.MatchAllInOutput(stdout, []string{"odo may not work as expected in the default project, please run the odo component in a non-default project"})
			})
		})
	})

	Context("using Kubernetes cluster", func() {
		BeforeEach(func() {
			if os.Getenv("KUBERNETES") != "true" {
				Skip("This is a Kubernetes specific scenario, skipping")
			}
		})
		When("project with with 'default' name is used", func() {

			It("should push successfully", func() {
				componentName := helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.Cmd("odo", "create", "--project", "default", componentName, "--devfile", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

				stdout := helper.Cmd("odo", "push").ShouldPass().Out()
				helper.DontMatchAllInOutput(stdout, []string{"odo may not work as expected in the default project"})
				helper.Cmd("odo", "delete", "-f").ShouldPass()
			})
		})
	})

})
