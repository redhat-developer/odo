package devfile

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

	Context("Pushing devfile without an .odo folder", func() {

		It("should be able to push based on metadata.name in devfile WITH a dash in the name", func() {
			// This is the name that's contained within `devfile-with-metadataname-foobar.yaml`
			name := "foobar"
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile-with-metadataname-foobar.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Executing devfile commands for component " + name))
		})

		It("should be able to push based on name passed", func() {
			name := "springboot"
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "push", "--project", commonVar.Project, name).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Executing devfile commands for component " + name))
		})

		It("should error out on devfile flag", func() {
			helper.Cmd("odo", "push", "--project", commonVar.Project, "--devfile", "invalid.yaml").ShouldFail()
		})

	})

	Context("Verify devfile push works", func() {

		It("should have no errors when no endpoints within the devfile, should create a service when devfile has endpoints", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-no-endpoints.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()
			output := commonVar.CliRunner.GetServices(commonVar.Project)
			Expect(output).NotTo(ContainSubstring(cmpName))

			helper.RenameFile("devfile-old.yaml", "devfile.yaml")
			output = helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()

			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
			output = commonVar.CliRunner.GetServices(commonVar.Project)
			Expect(output).To(ContainSubstring(cmpName))
		})

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

		When("Devfile 2.1.0 is used", func() {
			It("should work with devfile variables", func() {
				helper.CmdShouldPass("odo", "create", "nodejs", "--project", commonVar.Project, cmpName)

				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-variables.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				output := helper.CmdShouldPass("odo", "push")
				Expect(output).To(ContainSubstring("Changes successfully pushed to component"))
				routeURL := helper.DetermineRouteURL(commonVar.Context)

				// Ping said URL
				helper.HttpWaitFor(routeURL, "Hello from Node.js", 10, 5)

			})
		})

		It("checks that odo push works with a devfile with sourcemapping set", func() {
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

		It("checks that odo push works with a devfile with composite commands", func() {
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

		It("checks that odo push works with a devfile with parallel composite commands", func() {
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

		It("checks that odo push works with a devfile with nested composite commands", func() {
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

		It("should throw a validation error for composite run commands", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeRun.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.Cmd("odo", "push").ShouldFail().Err()
			Expect(output).To(ContainSubstring("not supported currently"))
		})

		It("should throw a validation error for composite command referencing non-existent commands", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeNonExistent.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.Cmd("odo", "push").ShouldFail().Err()
			Expect(output).To(ContainSubstring("does not exist in the devfile"))
		})

		It("should throw a validation error for composite command indirectly referencing itself", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileIndirectNesting.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.Cmd("odo", "push").ShouldFail().Err()
			Expect(output).To(ContainSubstring("cannot indirectly reference itself"))
		})

		It("should throw a validation error for composite command that has invalid exec subcommand", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeInvalidComponent.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			// Verify odo push failed
			output := helper.Cmd("odo", "push").ShouldFail().Err()
			Expect(output).To(ContainSubstring("command does not map to a container component"))
		})

		It("checks that odo push works outside of the context directory", func() {
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

		It("should not build when no changes are detected in the directory and build when a file change is detected", func() {
			utils.ExecPushToTestFileChanges(commonVar.Context, cmpName, commonVar.Project)
		})

		It("checks that odo push with -o json displays machine readable JSON event output", func() {

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

		It("should be able to create a file, push, delete, then push again propagating the deletions", func() {
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

		It("should delete the files from the container if its removed locally", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()

			// Check to see if it's been pushed (foobar.txt abd directory testdir)
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

			var statErr error
			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"",
				commonVar.Project,
				[]string{"stat", "/projects/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).ToNot(HaveOccurred())
			Expect(os.Remove(filepath.Join(commonVar.Context, "server.js"))).NotTo(HaveOccurred())
			helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass()

			commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
				podName,
				"",
				commonVar.Project,
				[]string{"stat", "/projects/server.js"},
				func(cmdOp string, err error) bool {
					statErr = err
					return err == nil
				},
			)
			Expect(statErr).To(HaveOccurred())
			Expect(statErr.Error()).To(ContainSubstring("cannot stat '/projects/server.js': No such file or directory"))
		})

		It("should build when no changes are detected in the directory and force flag is enabled", func() {
			utils.ExecPushWithForceFlag(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should execute the default build and run command groups if present", func() {
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

		It("should execute PreStart commands if present during pod startup", func() {
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
				output = helper.CmdShouldPass("odo", "push", "--project", commonVar.Project, "-f")
				helper.MatchAllInOutput(output, []string{"PreStart commands have been added to the component"})

				secondPushPodName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)

				secondPushInitContainers := commonVar.CliRunner.GetPodInitContainers(cmpName, commonVar.Project)
				// 3 preStart events + 1 supervisord init containers
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

		It("should execute PostStart commands if present and not execute when component already exists", func() {
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

		It("should err out on an event not mentioned in the devfile commands", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-invalid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldFail().Err()
			helper.MatchAllInOutput(output, []string{"does not map to a valid devfile command"})
		})

		It("should err out on an event command not mapping to a devfile container component", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "runtime #wrongruntime", "wrongruntime")

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldFail().Err()
			helper.MatchAllInOutput(output, []string{"does not map to a container component"})
		})

		It("should err out on an event composite command mentioning an invalid child command", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-valid-events.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "secondprestop #secondprestopiswrong", "secondprestopiswrong")

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldFail().Err()
			helper.MatchAllInOutput(output, []string{"does not exist in the devfile"})
		})

		It("should be able to handle a missing build command group", func() {
			utils.ExecWithMissingBuildCommand(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should error out on a missing run command group", func() {
			utils.ExecWithMissingRunCommand(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should be able to push using the custom commands", func() {
			utils.ExecWithCustomCommand(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should error out on a wrong custom commands", func() {
			utils.ExecWithWrongCustomCommand(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should error out on multiple or no default commands", func() {
			utils.ExecWithMultipleOrNoDefaults(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should execute commands with flags if the command has no group kind", func() {
			utils.ExecCommandWithoutGroupUsingFlags(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should error out if the devfile has an invalid command group", func() {
			utils.ExecWithInvalidCommandGroup(commonVar.Context, cmpName, commonVar.Project)
		})

		It("should restart the application if it is not hot reload capable", func() {
			utils.ExecWithHotReload(commonVar.Context, cmpName, commonVar.Project, false)
		})

		It("should not restart the application if it is hot reload capable", func() {
			utils.ExecWithHotReload(commonVar.Context, cmpName, commonVar.Project, true)
		})

		It("should restart the application if run mode is changed, regardless of hotReloadCapable value", func() {
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

		It("should run odo push successfully after odo push --debug", func() {
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

		It("Ensure that push -f correctly removes local deleted files from the remote target sync folder", func() {

			// 1) Push a generic Java project
			helper.Cmd("odo", "create", "java-springboot", "--project", commonVar.Project, cmpName).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// 2) Rename the pom.xml, which should cause the build to fail if sync is working as expected
			err := os.Rename(filepath.Join(commonVar.Context, "pom.xml"), filepath.Join(commonVar.Context, "pom.xml.renamed"))
			Expect(err).NotTo(HaveOccurred())

			// 3) Ensure that the build fails due to missing 'pom.xml', which ensures that the sync operation
			// correctly renamed pom.xml to pom.xml.renamed.
			output = helper.Cmd("odo", "push", "-f", "--project", commonVar.Project).ShouldFail().Err()
			helper.MatchAllInOutput(output, []string{"no POM in this directory"})
		})

	})

	Context("Verify files are correctly synced", func() {

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

		It("Should ensure that files are correctly synced on pod redeploy, with force push specified", func() {
			ensureFilesSyncedTest(commonVar.Project, true)
		})

		It("Should ensure that files are correctly synced on pod redeploy, without force push specified", func() {
			ensureFilesSyncedTest(commonVar.Project, false)
		})

	})

	Context("Verify devfile volume components work", func() {

		It("should error out when duplicate volume components exist", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "secondvol", "firstvol")

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldFail().Err()
			Expect(output).To(ContainSubstring("duplicate key: firstvol"))
		})

		It("should error out when a wrong volume size is used", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			helper.ReplaceString("devfile.yaml", "3Gi", "3Garbage")

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldFail().Err()
			Expect(output).To(ContainSubstring("quantities must match the regular expression"))
		})

		It("should error out if a container component has volume mount that does not refer a valid volume component", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.RenameFile("devfile.yaml", "devfile-old.yaml")
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-invalid-volmount.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "push", "--project", commonVar.Project).ShouldFail().Err()
			helper.MatchAllInOutput(output, []string{"unable to find the following volume mounts", "invalidvol1", "invalidvol2"})
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

		It("should throw a validation error for v1 devfiles", func() {
			helper.Cmd("odo", "create", "java-springboot", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "springboot", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.ReplaceString("devfile.yaml", "schemaVersion: 2.0.0", "apiVersion: 1.0.0")

			// Verify odo push failed
			output := helper.Cmd("odo", "push").ShouldFail().Err()
			Expect(output).To(ContainSubstring("schemaVersion not present in devfile"))
		})

	})

	Context("when .gitignore file exists", func() {
		It("checks that .odo/env exists in gitignore", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			ignoreFilePath := filepath.Join(commonVar.Context, ".gitignore")

			helper.FileShouldContainSubstring(ignoreFilePath, filepath.Join(".odo", "env"))

		})
	})

	Context("exec commands with environment variables", func() {
		It("Should be able to exec command with single environment variable", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-envs.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			output := helper.Cmd("odo", "push", "--build-command", "buildwithenv", "--run-command", "singleenv").ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{"mkdir $ENV1", "mkdir $BUILD_ENV1"})

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
			helper.MatchAllInOutput(output, []string{"test_env_variable", "test_build_env_variable"})

		})

		It("Should be able to exec command with multiple environment variables", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-envs.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			output := helper.Cmd("odo", "push", "--build-command", "buildwithmultipleenv", "--run-command", "multipleenv").ShouldPass().Out()
			helper.MatchAllInOutput(output, []string{"mkdir $ENV1 $ENV2", "mkdir $BUILD_ENV1 $BUILD_ENV2"})

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			output = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, sourcePath)
			helper.MatchAllInOutput(output, []string{"test_build_env_variable1", "test_build_env_variable2", "test_env_variable1", "test_env_variable2"})

		})

		It("Should be able to exec command with environment variable with spaces", func() {
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

	Context("Verify source code sync location", func() {

		It("Should sync to the correct dir in container if project and clonePath is present", func() {
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

		It("Should sync to the correct dir in container if project present", func() {
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

		It("Should sync to the correct dir in container if multiple project is present", func() {
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

		It("Should sync to the correct dir in container if no project is present", func() {
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

	Context("push with listing the devfile component", func() {
		It("checks components in a specific app and all apps", func() {

			// component created in "app" application
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			output := helper.Cmd("odo", "list").ShouldPass().Out()
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sNotPushed", "app", cmpName, commonVar.Project, "nodejs"))))

			output = helper.Cmd("odo", "push").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// component created in different application
			context2 := helper.CreateNewContext()
			cmpName2 := helper.RandString(6)
			appName := helper.RandString(6)

			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, "--app", appName, "--context", context2, cmpName2).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), context2)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(context2, "devfile.yaml"))

			output = helper.Cmd("odo", "list", "--context", context2).ShouldPass().Out()
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sNotPushed", appName, cmpName2, commonVar.Project, "nodejs"))))
			output2 := helper.Cmd("odo", "push", "--context", context2).ShouldPass().Out()
			Expect(output2).To(ContainSubstring("Changes successfully pushed to component"))

			output = helper.Cmd("odo", "list", "--project", commonVar.Project).ShouldPass().Out()
			// this test makes sure that a devfile component doesn't show up as an s2i component as well
			Expect(helper.Suffocate(output)).To(Equal(helper.Suffocate(fmt.Sprintf(`
			Devfile Components:
			APP        NAME       PROJECT        TYPE       STATE
			app        %[1]s     %[2]s           nodejs     Pushed
			`, cmpName, commonVar.Project))))

			output = helper.Cmd("odo", "list", "--all-apps", "--project", commonVar.Project).ShouldPass().Out()

			Expect(output).To(ContainSubstring(cmpName))
			Expect(output).To(ContainSubstring(cmpName2))

			helper.DeleteDir(context2)

		})

		It("checks devfile and s2i components together", func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("Skipping test because s2i image is not supported on Kubernetes cluster")
			}

			// component created in "app" application
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project, cmpName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			output := helper.Cmd("odo", "list").ShouldPass().Out()
			Expect(helper.Suffocate(output)).To(ContainSubstring(helper.Suffocate(fmt.Sprintf("%s%s%s%sNotPushed", "app", cmpName, commonVar.Project, "nodejs"))))

			output = helper.Cmd("odo", "push").ShouldPass().Out()
			Expect(output).To(ContainSubstring("Changes successfully pushed to component"))

			// component created in different application
			context2 := helper.CreateNewContext()
			cmpName2 := helper.RandString(6)
			appName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), context2)
			helper.Cmd("odo", "create", "--s2i", "nodejs", "--project", commonVar.Project, "--app", appName, "--context", context2, cmpName2).ShouldPass()

			output2 := helper.Cmd("odo", "push", "--context", context2).ShouldPass().Out()
			Expect(output2).To(ContainSubstring("Changes successfully pushed to component"))

			output = helper.Cmd("odo", "list", "--all-apps", "--project", commonVar.Project).ShouldPass().Out()

			Expect(output).To(ContainSubstring(cmpName))
			Expect(output).To(ContainSubstring(cmpName2))

			output = helper.Cmd("odo", "list", "--app", appName, "--project", commonVar.Project).ShouldPass().Out()
			Expect(output).To(Not(ContainSubstring(cmpName))) // cmpName component hasn't been created under appName
			Expect(output).To(ContainSubstring(cmpName2))

			helper.DeleteDir(context2)
		})

	})

	Context("Handle devfiles with parent", func() {
		var server *http.Server
		var freePort int
		var parentTmpFolder string

		var _ = BeforeEach(func() {
			// get a free port
			var err error
			freePort, err = util.HTTPGetFreePort()
			Expect(err).NotTo(HaveOccurred())

			// move the parent devfiles to a tmp folder
			parentTmpFolder = helper.CreateNewContext()
			helper.CopyExample(filepath.Join("source", "devfiles", "parentSupport"), parentTmpFolder)
			// update the port in the required devfile with the free port
			helper.ReplaceString(filepath.Join(parentTmpFolder, "devfile-middle-layer.yaml"), "(-1)", strconv.Itoa(freePort))

			// start the server and serve from the tmp folder of the devfiles
			server = helper.HttpFileServer(freePort, parentTmpFolder)

			// wait for the server to be respond with the desired result
			helper.HttpWaitFor("http://localhost:"+strconv.Itoa(freePort), "devfile", 10, 1)
		})

		var _ = AfterEach(func() {
			helper.DeleteDir(parentTmpFolder)
			err := server.Close()
			Expect(err).To(BeNil())
		})

		It("should handle a devfile with a parent and add a extra command", func() {
			utils.ExecPushToTestParent(commonVar.Context, cmpName, commonVar.Project)
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			listDir := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/project/")
			Expect(listDir).To(ContainSubstring("blah.js"))
		})

		It("should handle a devfile with a parent and override a composite command", func() {
			utils.ExecPushWithCompositeOverride(commonVar.Context, cmpName, commonVar.Project, freePort)
			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			listDir := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
			Expect(listDir).To(ContainSubstring("testfile"))
		})

		It("should handle a parent and override/append it's envs", func() {
			utils.ExecPushWithParentOverride(commonVar.Context, cmpName, "app", commonVar.Project, freePort)

			envMap := commonVar.CliRunner.GetEnvsDevFileDeployment(cmpName, "app", commonVar.Project)

			value, ok := envMap["ODO_TEST_ENV_0"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_0"))

			value, ok = envMap["ODO_TEST_ENV_1"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_1_1"))
		})

		It("should handle a multi layer parent", func() {
			utils.ExecPushWithMultiLayerParent(commonVar.Context, cmpName, "app", commonVar.Project, freePort)

			podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
			listDir := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/project")
			helper.MatchAllInOutput(listDir, []string{"blah.js", "new-blah.js"})

			envMap := commonVar.CliRunner.GetEnvsDevFileDeployment(cmpName, "app", commonVar.Project)

			value, ok := envMap["ODO_TEST_ENV_1"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_1_1"))

			value, ok = envMap["ODO_TEST_ENV_2"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_2"))

			value, ok = envMap["ODO_TEST_ENV_3"]
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("ENV_VALUE_3"))

		})
	})

	Context("when the run command throws an error", func() {
		It("should wait and error out with some log", func() {
			helper.Cmd("odo", "create", "nodejs", "--project", commonVar.Project).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm start", "npm starts")

			_, output := helper.Cmd("odo", "push").ShouldPass().OutAndErr()
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

	Context("Testing Push for OpenShift specific scenarios", func() {
		JustBeforeEach(func() {
			if os.Getenv("KUBERNETES") == "true" {
				Skip("This is a OpenShift specific scenario, skipping")
			}
		})

		It("throw an error when the project value is default", func() {
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "nodejs", "--project", "default", componentName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			stdout := helper.Cmd("odo", "push").ShouldFail().Err()
			helper.MatchAllInOutput(stdout, []string{"odo may not work as expected in the default project, please run the odo component in a non-default project"})
		})
	})

	Context("Testing Push for Kubernetes specific scenarios", func() {
		JustBeforeEach(func() {
			if os.Getenv("KUBERNETES") != "true" {
				Skip("This is a Kubernetes specific scenario, skipping")
			}
		})

		It("should push successfully project value is default", func() {
			componentName := helper.RandString(6)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "create", "nodejs", "--project", "default", componentName).ShouldPass()

			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

			stdout := helper.Cmd("odo", "push").ShouldPass().Out()
			helper.DontMatchAllInOutput(stdout, []string{"odo may not work as expected in the default project"})
		})
	})

	Context("Testing Push with remote attributes", func() {
		It("should push only the mentioned files at the appropriate remote destination", func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
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
})
