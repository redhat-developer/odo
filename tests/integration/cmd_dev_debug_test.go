package integration

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/odo/tests/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("odo dev debug command tests", func() {
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

	for _, podman := range []bool{false, true} {
		podman := podman
		When("a component is bootstrapped", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-debugrun.yaml")).ShouldPass()
				Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
			})

			When("running odo dev with debug flag", helper.LabelPodmanIf(podman, func() {
				var devSession helper.DevSession
				var ports map[string]string
				BeforeEach(func() {
					var err error
					devSession, _, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
						CmdlineArgs: []string{"--debug"},
						RunOnPodman: podman,
					})
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should connect to relevant ports forwarded", func() {
					By("connecting to the application port", func() {
						helper.HttpWaitForWithStatus("http://"+ports["3000"], "Hello from Node.js Starter Application!", 12, 5, 200)
					})
					By("expecting a ws connection when tried to connect on default debug port locally", func() {
						// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
						// We are just using this to validate if nodejs agent is listening on the other side
						helper.HttpWaitForWithStatus("http://"+ports["5858"], "WebSockets request was expected", 12, 5, 400)
					})
				})

				// #6056
				It("should not add a DEBUG_PORT variable to the container", func() {
					cmp := helper.NewComponent(cmpName, "app", "runtime", commonVar.Project, commonVar.CliRunner)
					stdout := cmp.Exec("runtime", "sh", "-c", "echo -n ${DEBUG_PORT}")
					Expect(stdout).To(BeEmpty())
				})
			}))
		})
	}

	for _, devfileHandlerCtx := range []struct {
		name          string
		sourceHandler func(path string, originalCmpName string)
	}{
		{
			name: "with metadata.name",
		},
		{
			name: "without metadata.name",
			sourceHandler: func(path string, originalCmpName string) {
				helper.UpdateDevfileContent(filepath.Join(path, "devfile.yaml"), []helper.DevfileUpdater{helper.DevfileMetadataNameRemover})
				helper.ReplaceString(filepath.Join(path, "package.json"), "nodejs-starter", originalCmpName)
			},
		},
	} {
		devfileHandlerCtx := devfileHandlerCtx
		When("a composite command is used as debug command - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			var stdout []byte
			var stderr []byte
			var ports map[string]string
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfileCompositeRunAndDebug.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}
				var err error
				session, stdout, stderr, ports, err = helper.StartDevMode(helper.DevSessionOpts{
					CmdlineArgs: []string{"--debug"},
				})
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should run successfully", func() {
				By("verifying from the output that all commands have been executed", func() {
					helper.MatchAllInOutput(string(stdout), []string{
						"Building your application in container",
						"Executing the application (command: mkdir)",
						"Executing the application (command: echo)",
						"Executing the application (command: install)",
						"Executing the application (command: start-debug)",
					})
				})

				By("verifying that any command that did not succeed in the middle has logged such information correctly", func() {
					helper.MatchAllInOutput(string(stderr), []string{
						"Devfile command \"echo\" exited with an error status",
						"intentional-error-message",
					})
				})

				By("building the application only once", func() {
					// Because of the Spinner, the "Building your application in container" is printed twice in the captured stdout.
					// The bracket allows to match the last occurrence with the command execution timing information.
					Expect(strings.Count(string(stdout), "Building your application in container (command: install) [")).
						To(BeNumerically("==", 1), "\nOUTPUT: "+string(stdout)+"\n")
				})

				By("verifying that the command did run successfully", func() {
					// Verify the command executed successfully
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
					res := commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
						podName,
						"runtime",
						commonVar.Project,
						[]string{"stat", "/projects/testfolder"},
						func(cmdOp string, err error) bool {
							return err == nil
						},
					)
					Expect(res).To(BeTrue())
				})

				By("expecting a ws connection when tried to connect on default debug port locally", func() {
					// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
					// We are just using this to validate if nodejs agent is listening on the other side
					helper.HttpWaitForWithStatus("http://"+ports["5858"], "WebSockets request was expected", 12, 5, 400)
				})
			})
		})
	}

	When("a composite apply command is used as debug command", func() {
		deploymentName := "my-component"
		var session helper.DevSession
		var sessionOut []byte
		var err error
		var ports map[string]string
		const (
			DEVFILE_DEBUG_PORT = "5858"
		)

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-composite-apply-commands.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				helper.DevfileMetadataNameSetter(cmpName))
			session, sessionOut, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
				EnvVars:     []string{"PODMAN_CMD=echo"},
				CmdlineArgs: []string{"--debug"},
			})
			Expect(err).ToNot(HaveOccurred())
		})
		It("should execute the composite apply commands successfully", func() {
			checkDeploymentExists := func() {
				out := commonVar.CliRunner.Run("get", "deployments", deploymentName).Out.Contents()
				Expect(out).To(ContainSubstring(deploymentName))
			}
			checkImageBuilt := func() {
				Expect(string(sessionOut)).To(ContainSubstring("Building & Pushing Container"))
				Expect(string(sessionOut)).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
				Expect(string(sessionOut)).To(ContainSubstring("push quay.io/unknown-account/myimage"))
			}

			checkWSConnection := func() {
				// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
				// We are just using this to validate if nodejs agent is listening on the other side
				helper.HttpWaitForWithStatus("http://"+ports[DEVFILE_DEBUG_PORT], "WebSockets request was expected", 12, 5, 400)
			}
			By("expecting a ws connection when tried to connect on default debug port locally", func() {
				checkWSConnection()
			})

			By("checking is the image was successfully built", func() {
				checkImageBuilt()
			})

			By("checking the deployment was created successfully", func() {
				checkDeploymentExists()
			})

			By("checking odo dev watches correctly", func() {
				// making changes to the project again
				helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js Starter Application", "from the new Node.js Starter Application")
				_, _, _, err = session.WaitSync()
				Expect(err).ToNot(HaveOccurred())
				checkDeploymentExists()
				checkImageBuilt()
				checkWSConnection()
			})

			By("cleaning up the resources on ending the session", func() {
				session.Stop()
				session.WaitEnd()
				out := commonVar.CliRunner.Run("get", "deployments").Out.Contents()
				Expect(out).ToNot(ContainSubstring(deploymentName))
			})
		})
	})

	for _, devfileHandlerCtx := range []struct {
		name          string
		sourceHandler func(path string, originalCmpName string)
	}{
		{
			name: "with metadata.name",
		},
		{
			name: "without metadata.name",
			sourceHandler: func(path string, originalCmpName string) {
				helper.UpdateDevfileContent(filepath.Join(path, "devfile.yaml"), []helper.DevfileUpdater{helper.DevfileMetadataNameRemover})
				helper.ReplaceString(filepath.Join(path, "package.json"), "nodejs-starter", originalCmpName)
			},
		},
	} {
		devfileHandlerCtx := devfileHandlerCtx
		When("running build and debug commands as composite in different containers and a shared volume - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			var stdout []byte
			var stderr []byte
			var ports map[string]string
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfileCompositeBuildRunDebugInMultiContainersAndSharedVolume.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}
				var err error
				session, stdout, stderr, ports, err = helper.StartDevMode(helper.DevSessionOpts{
					CmdlineArgs: []string{"--debug"},
				})
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should run successfully", func() {
				By("verifying from the output that all commands have been executed", func() {
					helper.MatchAllInOutput(string(stdout), []string{
						"Building your application in container (command: mkdir)",
						"Building your application in container (command: sleep-cmd-build)",
						"Building your application in container (command: build-cmd)",
						"Executing the application (command: sleep-cmd-run)",
						"Executing the application (command: echo-with-error)",
						"Executing the application (command: check-build-result)",
						"Executing the application (command: start-debug)",
					})
				})

				By("verifying that any command that did not succeed in the middle has logged such information correctly", func() {
					helper.MatchAllInOutput(string(stderr), []string{
						"Devfile command \"echo-with-error\" exited with an error status",
						"intentional-error-message",
					})
				})

				By("building the application only once per exec command in the build command", func() {
					// Because of the Spinner, the "Building your application in container" is printed twice in the captured stdout.
					// The bracket allows to match the last occurrence with the command execution timing information.
					out := string(stdout)
					for _, cmd := range []string{"mkdir", "sleep-cmd-build", "build-cmd"} {
						Expect(strings.Count(out, fmt.Sprintf("Building your application in container (command: %s) [", cmd))).
							To(BeNumerically("==", 1), "\nOUTPUT: "+string(stdout)+"\n")
					}
				})

				By("verifying that the command did run successfully", func() {
					// Verify the command executed successfully
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
					res := commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
						podName,
						"runtime",
						commonVar.Project,
						[]string{"stat", "/projects/testfolder"},
						func(cmdOp string, err error) bool {
							return err == nil
						},
					)
					Expect(res).To(BeTrue())
				})

				By("expecting a ws connection when tried to connect on default debug port locally", func() {
					// 400 response expected because the endpoint expects a websocket request and we are doing a HTTP GET
					// We are just using this to validate if nodejs agent is listening on the other side
					helper.HttpWaitForWithStatus("http://"+ports["5858"], "WebSockets request was expected", 12, 5, 400)
				})
			})
		})
	}

	When("a component without debug command is bootstrapped", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-without-debugrun.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})

		It("should fail running odo dev --debug", func() {
			output := helper.Cmd("odo", "dev", "--debug").ShouldFail().Err()
			Expect(output).To(ContainSubstring("no command of kind \"debug\" found in the devfile"))
		})
	})
})
