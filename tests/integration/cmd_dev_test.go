package integration

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	segment "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/onsi/gomega/gexec"

	"github.com/redhat-developer/odo/tests/helper"
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
			Expect(output).To(ContainSubstring("The current directory does not represent an odo component"))

		})
	})

	When("a component is bootstrapped and pushed", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})

		It("should add annotation to use ImageStreams", func() {
			// #6376
			err := helper.RunDevMode(helper.DevSessionOpts{}, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
				annotations := commonVar.CliRunner.GetAnnotationsDeployment(cmpName, "app", commonVar.Project)
				Expect(annotations["alpha.image.policy.openshift.io/resolve-names"]).To(Equal("*"))
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should show validation errors if the devfile is incorrect", func() {
			err := helper.RunDevMode(helper.DevSessionOpts{}, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "kind: run", "kind: build")
				helper.WaitForOutputToContain("Error occurred on Push", 180, 10, session)
			})
			Expect(err).ToNot(HaveOccurred())
		})

		for _, podman := range []bool{true, false} {
			podman := podman
			It("should use the index information from previous push operation", helper.LabelPodmanIf(podman, func() {
				// Create a new file A
				fileAPath, fileAText := helper.CreateSimpleFile(commonVar.Context, "my-file-", ".txt")
				// watch that project
				err := helper.RunDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				}, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					// Change some other file B
					helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "App started", "App is super started")

					// File should exist, and its content should match what we initially set it to
					component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					execResult, _ := component.Exec("runtime", []string{"cat", "/projects/" + filepath.Base(fileAPath)}, pointer.Bool(true))
					Expect(execResult).To(ContainSubstring(fileAText))
				})
				Expect(err).ToNot(HaveOccurred())
			}))
		}

		It("ensure that index information is updated", func() {
			err := helper.RunDevMode(helper.DevSessionOpts{}, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
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

		When("a state file is not writable", func() {
			BeforeEach(func() {
				stateFile := filepath.Join(commonVar.Context, ".odo", "devstate.json")
				helper.MakeDir(filepath.Dir(stateFile))
				Expect(helper.CreateFileWithContent(stateFile, "")).ToNot(HaveOccurred())
				Expect(os.Chmod(stateFile, 0400)).ToNot(HaveOccurred())
			})
			It("should fail running odo dev", func() {
				res := helper.Cmd("odo", "dev", "--random-ports").ShouldFail()
				stdout := res.Out()
				stderr := res.Err()
				Expect(stdout).To(ContainSubstring("Cleaning"))
				Expect(stderr).To(ContainSubstring("unable to save forwarded ports to state file"))
			})
		})

		for _, podman := range []bool{true, false} {
			podman := podman
			When("recording telemetry data", helper.LabelPodmanIf(podman, func() {
				BeforeEach(func() {
					helper.EnableTelemetryDebug()
					session, _, _, _, _ := helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					})
					session.Stop()
					session.WaitEnd()
				})
				AfterEach(func() {
					helper.ResetTelemetry()
				})
				It("should record the telemetry data correctly", func() {
					td := helper.GetTelemetryDebugData()
					Expect(td.Event).To(ContainSubstring("odo dev"))
					if !podman {
						// TODO(feloy) what should be the correct exit code for odo dev after pressing ctrl-c?
						Expect(td.Properties.Success).To(BeFalse())
					}
					Expect(td.Properties.Error).ToNot(ContainSubstring("user interrupted"))
					Expect(td.Properties.CmdProperties[segment.ComponentType]).To(ContainSubstring("nodejs"))
					Expect(td.Properties.CmdProperties[segment.Language]).To(ContainSubstring("nodejs"))
					Expect(td.Properties.CmdProperties[segment.ProjectType]).To(ContainSubstring("nodejs"))
					Expect(td.Properties.CmdProperties).Should(HaveKey(segment.Caller))
					Expect(td.Properties.CmdProperties[segment.Caller]).To(BeEmpty())
					experimentalValue := false
					if podman {
						experimentalValue = true
					}
					Expect(td.Properties.CmdProperties[segment.ExperimentalMode]).To(Equal(experimentalValue))
					if podman {
						Expect(td.Properties.CmdProperties[segment.Platform]).To(Equal("podman"))
						Expect(td.Properties.CmdProperties[segment.PlatformVersion]).ToNot(BeEmpty())
					} else if os.Getenv("KUBERNETES") == "true" {
						Expect(td.Properties.CmdProperties[segment.Platform]).To(Equal("kubernetes"))
						serverVersion := commonVar.CliRunner.GetVersion()
						Expect(td.Properties.CmdProperties[segment.PlatformVersion]).To(ContainSubstring(serverVersion))
					} else {
						Expect(td.Properties.CmdProperties[segment.Platform]).To(Equal("openshift"))
						serverVersion := commonVar.CliRunner.GetVersion()
						if serverVersion == "" {
							Expect(td.Properties.CmdProperties[segment.PlatformVersion]).To(BeNil())
						} else {
							Expect(td.Properties.CmdProperties[segment.PlatformVersion]).To(ContainSubstring(serverVersion))
						}
					}
				})
			}))

			When("odo dev is executed", helper.LabelPodmanIf(podman, func() {

				var devSession helper.DevSession

				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					})
					Expect(err).ToNot(HaveOccurred())
				})

				When("odo dev is stopped", func() {
					BeforeEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})

					It("should delete component from the cluster", func() {
						component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						component.ExpectIsNotDeployed()
					})
				})
			}))
		}

		When("an env.yaml file contains a non-current Project", func() {
			BeforeEach(func() {
				odoDir := filepath.Join(commonVar.Context, ".odo", "env")
				helper.MakeDir(odoDir)
				err := helper.CreateFileWithContent(filepath.Join(odoDir, "env.yaml"), `
ComponentSettings:
  Project: another-project
`)
				Expect(err).ShouldNot(HaveOccurred())

			})

			When("odo dev is executed", func() {

				var devSession helper.DevSession

				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					devSession.Kill()
					devSession.WaitEnd()
				})

				It("should not have modified env.yaml, and use current namespace", func() {
					helper.FileShouldContainSubstring(".odo/env/env.yaml", "Project: another-project")

					deploymentName := fmt.Sprintf("%s-%s", cmpName, "app")
					out := commonVar.CliRunner.Run("get", "deployments", deploymentName, "-n", commonVar.Project).Out.Contents()
					Expect(out).To(ContainSubstring(deploymentName))
				})
			})
		})

		When("odo dev is executed and Ephemeral is set to false", func() {

			var devSession helper.DevSession
			BeforeEach(func() {
				if os.Getenv("KUBERNETES") == "true" {
					Skip("This is a OpenShift specific scenario, skipping")
				}
				helper.Cmd("odo", "preference", "set", "-f", "Ephemeral", "false").ShouldPass()
				var err error
				devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})

			When("killing odo dev and running odo delete component --wait", func() {
				BeforeEach(func() {
					devSession.Kill()
					devSession.WaitEnd()
					helper.Cmd("odo", "delete", "component", "--wait", "-f").ShouldPass()
				})

				It("should have deleted all resources before returning", func() {
					By("deleting the service", func() {
						services := commonVar.CliRunner.GetServices(commonVar.Project)
						Expect(services).To(BeEmpty())
					})
					By("deleting the PVC", func() {
						pvcs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
						Expect(pvcs).To(BeEmpty())
					})
					By("deleting the pod", func() {
						pods := commonVar.CliRunner.GetAllPodNames(commonVar.Project)
						Expect(pods).To(BeEmpty())
					})
				})
			})

			When("stopping odo dev normally", func() {
				BeforeEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should have deleted all resources before returning", func() {
					By("deleting the service", func() {
						services := commonVar.CliRunner.GetServices(commonVar.Project)
						Expect(services).To(BeEmpty())
					})
					By("deleting the PVC", func() {
						pvcs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
						Expect(pvcs).To(BeEmpty())
					})
					By("deleting the pod", func() {
						pods := commonVar.CliRunner.GetAllPodNames(commonVar.Project)
						Expect(pods).To(BeEmpty())
					})
				})
			})
		})

		When("odo dev is executed and Ephemeral is set to false", func() {

			var devSession helper.DevSession
			BeforeEach(func() {
				if os.Getenv("KUBERNETES") == "true" {
					Skip("This is a OpenShift specific scenario, skipping")
				}
				helper.Cmd("odo", "preference", "set", "-f", "Ephemeral", "false").ShouldPass()
				var err error
				devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				if os.Getenv("KUBERNETES") == "true" {
					return
				}
				// We stop the process so the process does not remain after the end of the tests
				devSession.Kill()
				devSession.WaitEnd()
			})

			It("should have created resources", func() {
				By("creating a service", func() {
					services := commonVar.CliRunner.GetServices(commonVar.Project)
					Expect(services).To(SatisfyAll(
						Not(BeEmpty()),
						ContainSubstring(fmt.Sprintf("%s-app", cmpName)),
					))
				})
				By("creating a PVC", func() {
					pvcs := commonVar.CliRunner.GetAllPVCNames(commonVar.Project)
					Expect(strings.Join(pvcs, "\n")).To(SatisfyAll(
						Not(BeEmpty()),
						ContainSubstring(fmt.Sprintf("%s-app", cmpName)),
					))
				})
				By("creating a pod", func() {
					pods := commonVar.CliRunner.GetAllPodNames(commonVar.Project)
					Expect(strings.Join(pods, "\n")).To(SatisfyAll(
						Not(BeEmpty()),
						ContainSubstring(fmt.Sprintf("%s-app-", cmpName)),
					))
				})

				// Returned pvc yaml contains ownerreference
				By("creating a pvc with ownerreference", func() {
					output := commonVar.CliRunner.Run("get", "pvc", "--namespace", commonVar.Project, "-o", `jsonpath='{range .items[*].metadata.ownerReferences[*]}{@..kind}{"/"}{@..name}{"\n"}{end}'`).Out.Contents()
					Expect(string(output)).To(ContainSubstring(fmt.Sprintf("Deployment/%s-app", cmpName)))
				})
			})
		})

		for _, podman := range []bool{true, false} {
			podman := podman
			When("odo is executed with --no-watch flag", helper.LabelPodmanIf(podman, func() {

				var devSession helper.DevSession

				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
						CmdlineArgs: []string{"--no-watch"},
						RunOnPodman: podman,
					})
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				When("a file in component directory is modified", func() {

					BeforeEach(func() {
						helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "App started", "App is super started")
					})

					It("should not trigger a push", func() {
						component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						execResult, _ := component.Exec("runtime", []string{"cat", "/projects/server.js"}, pointer.Bool(true))
						Expect(execResult).To(ContainSubstring("App started"))
						Expect(execResult).ToNot(ContainSubstring("App is super started"))

					})

					When("p is pressed", func() {

						BeforeEach(func() {
							if os.Getenv("SKIP_KEY_PRESS") == "true" {
								Skip("This is a unix-terminal specific scenario, skipping")
							}

							devSession.PressKey('p')
						})

						It("should trigger a push", func() {
							_, _, _, err := devSession.WaitSync()
							Expect(err).ToNot(HaveOccurred())
							component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
							execResult, _ := component.Exec("runtime", []string{"cat", "/projects/server.js"}, pointer.Bool(true))
							Expect(execResult).To(ContainSubstring("App is super started"))
						})
					})
				})
			}))
		}

		When("a delay is necessary for the component to start and running odo dev", func() {

			var devSession helper.DevSession
			var ports map[string]string

			BeforeEach(func() {
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm start", "sleep 20 ; npm start")

				var err error
				devSession, _, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				devSession.Kill()
				devSession.WaitEnd()
			})

			It("should first fail then succeed querying endpoint", func() {
				url := fmt.Sprintf("http://%s", ports["3000"])
				_, err := http.Get(url)
				Expect(err).To(HaveOccurred())

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				Eventually(func() bool {
					logs := helper.GetCliRunner().GetLogs(podName)
					return strings.Contains(logs, "App started on PORT")
				}, 180, 10).Should(Equal(true))

				// Get new random port after restart
				_, _, ports, err = devSession.GetInfo()
				Expect(err).ToNot(HaveOccurred())
				url = fmt.Sprintf("http://%s", ports["3000"])

				resp, err := http.Get(url)
				Expect(err).ToNot(HaveOccurred())
				body, _ := io.ReadAll(resp.Body)
				helper.MatchAllInOutput(string(body), []string{"Hello from Node.js Starter Application!"})
			})
		})
	})
	Context("checking if odo dev matches local Devfile K8s resources and remote resources", func() {
		for _, devfile := range []struct {
			title             string
			devfileName       string
			envvars           []string
			deploymentName    []string
			newDeploymentName []string
		}{
			{
				title:             "without apply command",
				devfileName:       "devfile-with-k8s-resource.yaml",
				envvars:           nil,
				deploymentName:    []string{"my-component"},
				newDeploymentName: []string{"my-new-component"},
			},
			{
				title:             "with apply command",
				devfileName:       "devfile-composite-apply-commands.yaml",
				envvars:           []string{"PODMAN_CMD=echo"},
				deploymentName:    []string{"my-k8s-component", "my-openshift-component"},
				newDeploymentName: []string{"my-new-k8s-component", "my-new-openshift-component"},
			},
		} {
			devfile := devfile
			When(fmt.Sprintf("odo dev is executed to run a devfile containing a k8s resource %s", devfile.title), func() {
				var (
					devSession    helper.DevSession
					err           error
					getDeployArgs = []string{"get", "deployments", "-n", commonVar.Project}
				)

				BeforeEach(
					func() {
						helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
						helper.CopyExampleDevFile(
							filepath.Join("source", "devfiles", "nodejs", devfile.devfileName),
							filepath.Join(commonVar.Context, "devfile.yaml"),
							helper.DevfileMetadataNameSetter(cmpName))
						devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
							EnvVars: devfile.envvars,
						})
						Expect(err).To(BeNil())

						// ensure the deployment is created by `odo dev`
						out := string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents())
						helper.MatchAllInOutput(out, devfile.deploymentName)
						// we fake the new deployment creation by changing the old deployment's name

						helper.ReplaceStrings(filepath.Join(commonVar.Context, "devfile.yaml"), devfile.deploymentName, devfile.newDeploymentName)

						_, _, _, err := devSession.WaitSync()
						Expect(err).To(BeNil())
					})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should have deleted the old resource and created the new resource", func() {
					getDeployments := string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents())
					for i := range devfile.deploymentName {
						Expect(getDeployments).ToNot(ContainSubstring(devfile.deploymentName[i]))
						Expect(getDeployments).To(ContainSubstring(devfile.newDeploymentName[i]))
					}
				})
			})
		}
	})

	for _, ctx := range []struct {
		title          string
		devfile        string
		matchResources []string
	}{
		{
			title:          "odo dev is executed to run a devfile containing a k8s resource",
			devfile:        "devfile-with-k8s-resource.yaml",
			matchResources: []string{"my-component"},
		},
		{
			title:          "odo dev is executed to run a devfile containing multiple k8s resource defined under a single Devfile component",
			devfile:        "devfile-with-multiple-k8s-resources-in-single-component.yaml",
			matchResources: []string{"my-component", "my-component-2"},
		},
	} {
		ctx := ctx
		When(ctx.title, func() {
			var (
				devSession    helper.DevSession
				out           []byte
				err           error
				getDeployArgs = []string{"get", "deployments", "-n", commonVar.Project}
			)

			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", ctx.devfile),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(cmpName))
				devSession, out, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				devSession.Stop()
				devSession.WaitEnd()
			})

			It("should have created the necessary k8s resources", func() {
				By("checking the output for the resources", func() {
					helper.MatchAllInOutput(string(out), ctx.matchResources)
				})
				By("fetching the resources from the cluster", func() {
					helper.MatchAllInOutput(string(commonVar.CliRunner.Run(getDeployArgs...).Out.Contents()), ctx.matchResources)
				})
			})
		})
	}

	for _, manual := range []bool{false, true} {
		for _, podman := range []bool{false, true} {
			manual := manual
			podman := podman
			Context("port-forwarding for the component", helper.LabelPodmanIf(podman, func() {
				When("devfile has no endpoint", func() {
					BeforeEach(func() {
						if !podman {
							helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
						}
						helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
						helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-no-endpoint.yaml")).ShouldPass()
					})

					When("running odo dev", func() {
						var devSession helper.DevSession
						var ports map[string]string
						BeforeEach(func() {
							var err error
							opts := []string{}
							if manual {
								opts = append(opts, "--no-watch")
							}
							devSession, _, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
								CmdlineArgs: opts,
								RunOnPodman: podman,
							})
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							devSession.Stop()
							devSession.WaitEnd()
						})

						It("should have no endpoint forwarded", func() {
							Expect(ports).To(BeEmpty())
						})
					})
				})

				When("devfile has single endpoint", func() {
					BeforeEach(func() {
						if !podman {
							helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
						}
						helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
						helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
					})

					When("running odo dev", func() {
						var devSession helper.DevSession
						var ports map[string]string
						BeforeEach(func() {
							var err error
							opts := []string{}
							if manual {
								opts = append(opts, "--no-watch")
							}
							devSession, _, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
								CmdlineArgs: opts,
								RunOnPodman: podman,
							})
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							devSession.Stop()
							devSession.WaitEnd()
						})

						It("should expose the endpoint on localhost", func() {
							url := fmt.Sprintf("http://%s", ports["3000"])
							resp, err := http.Get(url)
							Expect(err).ToNot(HaveOccurred())
							defer resp.Body.Close()

							body, _ := io.ReadAll(resp.Body)
							helper.MatchAllInOutput(string(body), []string{"Hello from Node.js Starter Application!"})
							Expect(err).ToNot(HaveOccurred())
						})

						When("modifying memoryLimit for container in Devfile", func() {
							var stdout string
							var stderr string
							BeforeEach(func() {
								src := "memoryLimit: 1024Mi"
								dst := "memoryLimit: 1023Mi"
								helper.ReplaceString("devfile.yaml", src, dst)
								if manual {
									if os.Getenv("SKIP_KEY_PRESS") == "true" {
										Skip("This is a unix-terminal specific scenario, skipping")
									}

									devSession.PressKey('p')
								}
								var err error
								var stdoutBytes []byte
								var stderrBytes []byte
								stdoutBytes, stderrBytes, ports, err = devSession.WaitSync()
								Expect(err).Should(Succeed())
								stdout = string(stdoutBytes)
								stderr = string(stderrBytes)
							})

							It("should react on the Devfile modification", func() {
								if podman {
									By("warning users that odo dev needs to be restarted", func() {
										Expect(stdout).To(ContainSubstring(
											"Detected changes in the Devfile, but this is not supported yet on Podman. Please restart 'odo dev' for such changes to be applied."))
									})
								} else {
									By("not warning users that odo dev needs to be restarted", func() {
										warning := "Please restart 'odo dev'"
										Expect(stdout).ShouldNot(ContainSubstring(warning))
										Expect(stderr).ShouldNot(ContainSubstring(warning))
									})
									By("updating the pod", func() {
										podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
										bufferOutput := commonVar.CliRunner.Run("get", "pods", podName, "-o", "jsonpath='{.spec.containers[0].resources.requests.memory}'").Out.Contents()
										output := string(bufferOutput)
										Expect(output).To(ContainSubstring("1023Mi"))
									})
								}

								By("exposing the endpoint", func() {
									Eventually(func(g Gomega) {
										url := fmt.Sprintf("http://%s", ports["3000"])
										resp, err := http.Get(url)
										g.Expect(err).ToNot(HaveOccurred())
										defer resp.Body.Close()

										body, _ := io.ReadAll(resp.Body)
										for _, i := range []string{"Hello from Node.js Starter Application!"} {
											g.Expect(string(body)).To(ContainSubstring(i))
										}
										g.Expect(err).ToNot(HaveOccurred())
									}).WithPolling(1 * time.Second).WithTimeout(20 * time.Second).Should(Succeed())
								})
							})
						})
					})
				})

				When("devfile has multiple endpoints", func() {
					BeforeEach(func() {
						if !podman {
							helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
						}
						helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-multiple-endpoints"), commonVar.Context)
						helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-multiple-endpoints.yaml")).ShouldPass()
					})

					When("running odo dev", func() {
						var devSession helper.DevSession
						var ports map[string]string
						BeforeEach(func() {
							opts := []string{}
							if manual {
								opts = append(opts, "--no-watch")
							}
							var err error
							devSession, _, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
								CmdlineArgs: opts,
								RunOnPodman: podman,
							})
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							devSession.Stop()
							devSession.WaitEnd()
						})

						It("should expose all endpoints on localhost regardless of exposure", func() {
							By("not exposing debug endpoints", func() {
								for _, p := range []int{5005, 5006} {
									_, found := ports[strconv.Itoa(p)]
									Expect(found).To(BeFalse(), fmt.Sprintf("debug port %d should not be forwarded", p))
								}
							})

							getServerResponse := func(p int) (string, error) {
								resp, err := http.Get(fmt.Sprintf("http://%s", ports[strconv.Itoa(p)]))
								if err != nil {
									return "", err
								}
								defer resp.Body.Close()

								body, _ := io.ReadAll(resp.Body)
								return string(body), nil
							}
							containerPorts := []int{3000, 4567, 7890}
							for _, p := range containerPorts {
								By(fmt.Sprintf("exposing a port targeting container port %d", p), func() {
									r, err := getServerResponse(p)
									Expect(err).ShouldNot(HaveOccurred())
									helper.MatchAllInOutput(r, []string{"Hello from Node.js Starter Application!"})
								})
							}

							helper.ReplaceString("server.js", "Hello from Node.js", "H3110 from Node.js")

							if manual {
								if os.Getenv("SKIP_KEY_PRESS") == "true" {
									Skip("This is a unix-terminal specific scenario, skipping")
								}

								devSession.PressKey('p')
							}

							var stdout, stderr []byte
							var err error
							stdout, stderr, _, err = devSession.WaitSync()
							Expect(err).Should(Succeed())

							By("not warning users that odo dev needs to be restarted because the Devfile has not changed", func() {
								warning := "Please restart 'odo dev'"
								if podman {
									warning = "Detected changes in the Devfile, but this is not supported yet on Podman. Please restart 'odo dev' for such changes to be applied."
								}
								Expect(stdout).ShouldNot(ContainSubstring(warning))
								Expect(stderr).ShouldNot(ContainSubstring(warning))
							})

							for _, p := range containerPorts {
								By(fmt.Sprintf("returning the right response when querying port forwarded for container port %d", p),
									func() {
										Eventually(func(g Gomega) string {
											r, err := getServerResponse(p)
											g.Expect(err).ShouldNot(HaveOccurred())
											return r
										}, 180, 10).Should(Equal("H3110 from Node.js Starter Application!"))
									})
							}
						})
					})

				})
			})...)
		}
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
		for _, podman := range []bool{true, false} {
			podman := podman
			When("Devfile 2.1.0 is used - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var devfileCmpName string
				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile-variables.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
				})

				When("doing odo dev", func() {
					var session helper.DevSession
					BeforeEach(func() {
						var err error
						session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
							RunOnPodman: podman,
						})
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						session.Stop()
						session.WaitEnd()
					})

					It("3. should check if the env variable has a correct value", func() {
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						envVars := component.GetEnvVars("runtime")
						// check if the env variable has a correct value. This value was substituted from in devfile from variable
						Expect(envVars["FOO"]).To(Equal("bar"))
					})
				})

				When("doing odo dev with --var flag", func() {
					var session helper.DevSession
					BeforeEach(func() {
						var err error
						session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
							CmdlineArgs: []string{"--var", "VALUE_TEST=baz"},
							RunOnPodman: podman,
						})
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						session.Stop()
						session.WaitEnd()
					})

					It("should check if the env variable has a correct value", func() {
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						envVars := component.GetEnvVars("runtime")
						// check if the env variable has a correct value. This value was substituted from in devfile from variable
						Expect(envVars["FOO"]).To(Equal("baz"))
					})
				})

				When("doing odo dev with --var-file flag", func() {
					var session helper.DevSession
					varfilename := "vars.txt"
					BeforeEach(func() {
						var err error
						err = helper.CreateFileWithContent(varfilename, "VALUE_TEST=baz")
						Expect(err).ToNot(HaveOccurred())
						session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
							CmdlineArgs: []string{"--var-file", "vars.txt"},
							RunOnPodman: podman,
						})
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						session.Stop()
						session.WaitEnd()
						helper.DeleteFile(varfilename)
					})

					It("should check if the env variable has a correct value", func() {
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						envVars := component.GetEnvVars("runtime")
						// check if the env variable has a correct value. This value was substituted from in devfile from variable
						Expect(envVars["FOO"]).To(Equal("baz"))
					})
				})

				When("doing odo dev with --var-file flag and setting value in env", func() {
					var session helper.DevSession
					varfilename := "vars.txt"
					BeforeEach(func() {
						var err error
						_ = os.Setenv("VALUE_TEST", "baz")
						err = helper.CreateFileWithContent(varfilename, "VALUE_TEST")
						Expect(err).ToNot(HaveOccurred())
						session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
							CmdlineArgs: []string{"--var-file", "vars.txt"},
							RunOnPodman: podman,
						})
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						session.Stop()
						session.WaitEnd()
						helper.DeleteFile(varfilename)
						_ = os.Unsetenv("VALUE_TEST")
					})

					It("should check if the env variable has a correct value", func() {
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						envVars := component.GetEnvVars("runtime")
						// check if the env variable has a correct value. This value was substituted from in devfile from variable
						Expect(envVars["FOO"]).To(Equal("baz"))
					})
				})
			}))

			When("running odo dev and single env var is set - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var devfileCmpName string
				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-single-env.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
				})

				It("should be able to exec command", func() {
					err := helper.RunDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					}, func(session *gexec.Session, out, err []byte, ports map[string]string) {
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						output, _ := component.Exec("runtime", []string{"ls", "-lai", "/projects"}, pointer.Bool(true))
						helper.MatchAllInOutput(output, []string{"test_env_variable", "test_build_env_variable"})
					})
					Expect(err).ToNot(HaveOccurred())
				})
			}))

			When("running odo dev and multiple env variables are set - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var devfileCmpName string
				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-multiple-envs.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
				})

				It("should be able to exec command", func() {
					err := helper.RunDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					}, func(session *gexec.Session, out, err []byte, ports map[string]string) {
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						output, _ := component.Exec("runtime", []string{"ls", "-lai", "/projects"}, pointer.Bool(true))
						helper.MatchAllInOutput(output, []string{"test_build_env_variable1", "test_build_env_variable2", "test_env_variable1", "test_env_variable2"})
					})
					Expect(err).ToNot(HaveOccurred())
				})
			}))

			When("doing odo dev and there is a env variable with spaces - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var devfileCmpName string
				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-env-with-space.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
				})

				It("should be able to exec command", func() {
					err := helper.RunDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					}, func(session *gexec.Session, out, err []byte, ports map[string]string) {
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						output, _ := component.Exec("runtime", []string{"ls", "-lai", "/projects"}, pointer.Bool(true))
						helper.MatchAllInOutput(output, []string{"build env variable with space", "env with space"})
					})
					Expect(err).ToNot(HaveOccurred())
				})
			}))
		}

		When("creating local files and dir and running odo dev - "+devfileHandlerCtx.name, func() {
			var newDirPath, newFilePath, stdOut, podName string
			var session helper.DevSession
			var devfileCmpName string
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				newFilePath = filepath.Join(commonVar.Context, "foobar.txt")
				newDirPath = filepath.Join(commonVar.Context, "testdir")
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}
				// Create a new file that we plan on deleting later...
				if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
					fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
				}
				// Create a new directory
				helper.MakeDir(newDirPath)
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
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
					_, _, _, err := session.WaitSync()
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

		When("Starting a PostgreSQL service", func() {
			BeforeEach(func() {
				// Ensure that the operators are installed
				commonVar.CliRunner.EnsureOperatorIsInstalled("service-binding-operator")
				commonVar.CliRunner.EnsureOperatorIsInstalled("cloud-native-postgresql")
				Eventually(func() string {
					out, _ := commonVar.CliRunner.GetBindableKinds()
					return out
				}, 120, 3).Should(ContainSubstring("Cluster"))
				addBindableKind := commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "bindablekind-instance.yaml"))
				Expect(addBindableKind.ExitCode()).To(BeEquivalentTo(0))
				commonVar.CliRunner.EnsurePodIsUp(commonVar.Project, "cluster-sample-1")
			})

			When("creating local files and dir and running odo dev - "+devfileHandlerCtx.name, func() {
				var newDirPath, newFilePath, stdOut, podName string
				var session helper.DevSession
				var devfileCmpName string
				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					newFilePath = filepath.Join(commonVar.Context, "foobar.txt")
					newDirPath = filepath.Join(commonVar.Context, "testdir")
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
					// Create a new file that we plan on deleting later...
					if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
						fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
					}
					// Create a new directory
					helper.MakeDir(newDirPath)
					var err error
					session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
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
						_, _, _, err := session.WaitSync()
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
		})

		When("adding local files to gitignore and running odo dev", func() {
			var gitignorePath, newDirPath, newFilePath1, newFilePath2, newFilePath3, stdOut, podName string
			var session helper.DevSession
			var devfileCmpName string
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				gitignorePath = filepath.Join(commonVar.Context, ".gitignore")
				newFilePath1 = filepath.Join(commonVar.Context, "foobar.txt")
				newDirPath = filepath.Join(commonVar.Context, "testdir")
				newFilePath2 = filepath.Join(newDirPath, "foobar.txt")
				newFilePath3 = filepath.Join(newDirPath, "baz.txt")
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}
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
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			checkSyncedFiles := func(podName string) {
				stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				helper.MatchAllInOutput(stdOut, []string{"testdir"})
				helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt"})
				stdOut = commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects/testdir")
				helper.MatchAllInOutput(stdOut, []string{"baz.txt"})
				helper.DontMatchAllInOutput(stdOut, []string{"foobar.txt"})
			}

			It("should not sync ignored files to the container", func() {
				podName = commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				checkSyncedFiles(podName)
			})

			When("modifying /testdir/baz.txt file", func() {
				BeforeEach(func() {
					helper.ReplaceString(newFilePath3, "hello world", "hello world!!!")
				})

				It("should synchronize it only", func() {
					_, _, _, _ = session.WaitSync()
					podName = commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
					checkSyncedFiles(podName)
				})
			})

			When("modifying /foobar.txt file", func() {
				BeforeEach(func() {
					helper.ReplaceString(newFilePath1, "hello world", "hello world!!!")
				})

				It("should not synchronize it", func() {
					session.CheckNotSynced(10 * time.Second)
				})
			})

			When("modifying /testdir/foobar.txt file", func() {
				BeforeEach(func() {
					helper.ReplaceString(newFilePath2, "hello world", "hello world!!!")
				})

				It("should not synchronize it", func() {
					session.CheckNotSynced(10 * time.Second)
				})
			})
		})

		When("devfile has sourcemappings and running odo dev - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfileSourceMapping.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())

			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
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

		When("project and clonePath is present in devfile and running odo dev - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				// devfile with clonePath set in project field
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}

				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
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
				helper.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/apps/webapp", "/apps", commonVar.CliRunner)
			})
		})

		When("devfile project field is present and running odo dev - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}

				// reset clonePath and change the workdir accordingly, it should sync to project name
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "clonePath: webapp/", "# clonePath: webapp/")
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should sync to the correct dir in container", func() {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/apps/nodeshift")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				helper.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/apps/nodeshift", "/apps", commonVar.CliRunner)
			})
		})

		When("multiple projects are present - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-projects.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}

				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should sync to the correct dir in container", func() {
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				// for devfile-with-multiple-projects.yaml source mapping is not set so $PROJECTS_ROOT is /projects
				// multiple projects, so source code would sync to the first project /projects/webapp
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects/webapp")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				helper.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/projects/webapp", "/projects", commonVar.CliRunner)
			})
		})

		When("no project is present - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should sync to the correct dir in container", func() {

				podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
				output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
				helper.MatchAllInOutput(output, []string{"package.json"})

				// Verify the sync env variables are correct
				helper.VerifyContainerSyncEnv(podName, "runtime", commonVar.Project, "/projects", "/projects", commonVar.CliRunner)
			})
		})

		When("running odo dev with devfile contain volume - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-with-volumes.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
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

		When("running odo dev with devfile containing volume-component - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				if os.Getenv("KUBERNETES") == "true" {
					Skip("This is a OpenShift specific scenario, skipping")
				}
				devfileCmpName = helper.RandString(6)
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(devfileCmpName))
				if devfileHandlerCtx.sourceHandler != nil {
					devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				if os.Getenv("KUBERNETES") == "true" {
					return
				}
				session.Stop()
				session.WaitEnd()
			})

			It("should successfully use the volume components in container components", func() {

				// Verify the pvc size for firstvol
				storageSize := commonVar.CliRunner.GetPVCSize(devfileCmpName, "firstvol", commonVar.Project)
				// should be the default size
				Expect(storageSize).To(ContainSubstring("1Gi"))

				// Verify the pvc size for secondvol
				storageSize = commonVar.CliRunner.GetPVCSize(devfileCmpName, "secondvol", commonVar.Project)
				// should be the specified size in the devfile volume component
				Expect(storageSize).To(ContainSubstring("200Mi"))
			})
		})
	}

	Describe("1. devfile contains composite apply command", func() {
		const (
			k8sDeploymentName       = "my-k8s-component"
			openshiftDeploymentName = "my-openshift-component"
			DEVFILEPORT             = "3000"
		)
		var session helper.DevSession
		var sessionOut, sessionErr []byte
		var err error
		var ports map[string]string
		BeforeEach(func() {
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-composite-apply-commands.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				helper.DevfileMetadataNameSetter(cmpName))
		})
		When("odo dev is running", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				session, sessionOut, sessionErr, ports, err = helper.StartDevMode(helper.DevSessionOpts{
					EnvVars: []string{"PODMAN_CMD=echo"},
				})
				Expect(err).ToNot(HaveOccurred())
			})
			It("should execute the composite apply commands successfully", func() {
				checkDeploymentsExist := func() {
					out := commonVar.CliRunner.Run("get", "deployments", k8sDeploymentName).Out.Contents()
					Expect(string(out)).To(ContainSubstring(k8sDeploymentName))
					out = commonVar.CliRunner.Run("get", "deployments", openshiftDeploymentName).Out.Contents()
					Expect(string(out)).To(ContainSubstring(openshiftDeploymentName))
				}
				checkImageBuilt := func() {
					Expect(string(sessionOut)).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
					Expect(string(sessionOut)).To(ContainSubstring("push quay.io/unknown-account/myimage"))
				}
				checkEndpointAccessible := func(message []string) {
					url := fmt.Sprintf("http://%s", ports[DEVFILEPORT])
					resp, e := http.Get(url)
					Expect(e).ToNot(HaveOccurred())
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					helper.MatchAllInOutput(string(body), message)
				}
				By("checking is the image was successfully built", func() {
					checkImageBuilt()
				})

				By("checking the endpoint accessibility", func() {
					checkEndpointAccessible([]string{"Hello from Node.js Starter Application!"})
				})

				By("checking the deployment was created successfully", func() {
					checkDeploymentsExist()
				})
				By("ensuring multiple deployments exist for selector error is not occurred", func() {
					Expect(string(sessionErr)).ToNot(ContainSubstring("multiple Deployments exist for the selector"))
				})
				By("checking odo dev watches correctly", func() {
					// making changes to the project again
					helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js Starter Application", "from the new Node.js Starter Application")
					_, _, _, err = session.WaitSync()
					Expect(err).ToNot(HaveOccurred())
					checkDeploymentsExist()
					checkImageBuilt()
					checkEndpointAccessible([]string{"Hello from the new Node.js Starter Application!"})
				})

				By("cleaning up the resources on ending the session", func() {
					session.Stop()
					session.WaitEnd()
					out := commonVar.CliRunner.Run("get", "deployments").Out.Contents()
					Expect(string(out)).ToNot(ContainSubstring(k8sDeploymentName))
					Expect(string(out)).ToNot(ContainSubstring(openshiftDeploymentName))
				})
			})
		})

		Context("the devfile contains an image component that uses a remote Dockerfile", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			})
			for _, env := range [][]string{
				{"PODMAN_CMD=echo"},
				{
					"PODMAN_CMD=a-command-not-found-for-podman-should-make-odo-fallback-to-docker",
					"DOCKER_CMD=echo",
				},
			} {
				env := env
				When(fmt.Sprintf("%v remote server returns a valid file when odo dev is run", env), func() {
					var buildRegexp string
					var server *httptest.Server
					var url string

					BeforeEach(func() {
						buildRegexp = regexp.QuoteMeta("build -t quay.io/unknown-account/myimage -f ") +
							".*\\.dockerfile " + regexp.QuoteMeta(commonVar.Context)
						server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							fmt.Fprintf(w, `# Dockerfile
FROM node:8.11.1-alpine
COPY . /app
WORKDIR /app
RUN npm install
CMD ["npm", "start"]
`)
						}))
						url = server.URL

						helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "./Dockerfile", url)
						session, sessionOut, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
							EnvVars: env,
						})
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						session.Stop()
						session.WaitEnd()
						server.Close()
					})

					It("should build and push image when odo dev is run", func() {
						lines, _ := helper.ExtractLines(string(sessionOut))
						_, ok := helper.FindFirstElementIndexMatchingRegExp(lines, buildRegexp)
						Expect(ok).To(BeTrue(), "build regexp not found in output: "+buildRegexp)
						Expect(string(sessionOut)).To(ContainSubstring("push quay.io/unknown-account/myimage"))
					})
				})
				When(fmt.Sprintf("%v remote server returns an error when odo dev is run", env), func() {
					var server *httptest.Server
					var url string
					BeforeEach(func() {
						server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusNotFound)
						}))
						url = server.URL

						helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "./Dockerfile", url)
					})

					AfterEach(func() {
						server.Close()
					})

					It("should not build images when odo dev is run", func() {
						_, sessionOut, _, err := helper.DevModeShouldFail(
							helper.DevSessionOpts{
								EnvVars: env,
							},
							"failed to retrieve "+url)
						Expect(err).To(BeNil())
						Expect(sessionOut).NotTo(ContainSubstring("build -t quay.io/unknown-account/myimage -f "))
						Expect(sessionOut).NotTo(ContainSubstring("push quay.io/unknown-account/myimage"))
					})
				})
			}
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

		for _, podman := range []bool{true, false} {
			podman := podman

			When("running odo dev and devfile with composite command - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var devfileCmpName string
				var session helper.DevSession
				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommands.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
					var err error
					session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					})
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
				})

				It("should execute all commands in composite command", func() {
					// Verify the command executed successfully
					component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					dir := "/projects/testfolder"
					out, _ := component.Exec("runtime", []string{"stat", dir}, pointer.Bool(true))
					Expect(out).To(ContainSubstring(dir))
				})
			}))

			When("running odo dev and composite command is marked as parallel:true - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var devfileCmpName string
				var session helper.DevSession
				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommandsParallel.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
					var err error
					session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman})
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
				})

				It("should execute all commands in composite command", func() {
					// Verify the command executed successfully
					component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					dir := "/projects/testfolder"
					out, _ := component.Exec("runtime", []string{"stat", dir}, pointer.Bool(true))
					Expect(out).To(ContainSubstring(dir))
				})
			}))

			When("running odo dev and composite command are nested - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var devfileCmpName string
				var session helper.DevSession
				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfileNestedCompCommands.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
					var err error
					session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman})
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
				})

				It("should execute all commands in composite commmand", func() {
					// Verify the command executed successfully

					component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					dir := "/projects/testfolder"
					out, _ := component.Exec("runtime", []string{"stat", dir}, pointer.Bool(true))
					Expect(out).To(ContainSubstring(dir))
				})
			}))

			When("running odo dev and composite command is used as a run command - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var session helper.DevSession
				var stdout []byte
				var stderr []byte
				var devfileCmpName string
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
					session, stdout, stderr, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
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
							"Executing the application (command: start)",
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
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						dir := "/projects/testfolder"
						out, _ := component.Exec("runtime", []string{"stat", dir}, pointer.Bool(true))
						Expect(out).To(ContainSubstring(dir))
					})
				})
			}))

			// This test does not pass on podman. There are flaky permissions issues on source volume mounted by both components sleeper-run and runtime
			When("running build and run commands as composite in different containers and a shared volume - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var session helper.DevSession
				var stdout []byte
				var stderr []byte
				var devfileCmpName string
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
					session, stdout, stderr, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
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
							"Executing the application (command: start)",
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
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						dir := "/projects/testfolder"
						out, _ := component.Exec("runtime", []string{"stat", dir}, pointer.Bool(true))
						Expect(out).To(ContainSubstring(dir))
					})
				})
			}))
		}
	}

	for _, podman := range []bool{false, true} {
		podman := podman
		When("running odo dev and prestart events are defined", helper.LabelPodmanIf(podman, func() {
			BeforeEach(func() {
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-with-preStart.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(cmpName))
			})

			It("should not correctly execute PreStart commands", func() {
				args := []string{"dev", "--random-ports"}
				if podman {
					args = append(args, "--platform", "podman")
				}
				cmd := helper.Cmd("odo", args...)
				if podman {
					cmd = cmd.AddEnv("ODO_EXPERIMENTAL_MODE=true")
				}
				output := cmd.ShouldFail().Err()
				// This is expected to fail for now.
				// see https://github.com/redhat-developer/odo/issues/4187 for more info
				helper.MatchAllInOutput(output, []string{"myprestart should either map to an apply command or a composite command with apply commands\n"})
			})
		}))
	}

	for _, podman := range []bool{false, true} {
		podman := podman
		When("running odo dev and run command throws an error", helper.LabelPodmanIf(podman, func() {
			var session helper.DevSession
			var initErr []byte
			BeforeEach(func() {
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(cmpName))
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm start", "npm starts")
				var err error
				session, _, initErr, _, err = helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should error out with some log", func() {
				helper.MatchAllInOutput(string(initErr), []string{
					"exited with an error status in",
					"Did you mean one of these?",
				})
			})
		}))
	}

	for _, podman := range []bool{false, true} {
		podman := podman
		When("running odo dev --no-watch and build command throws an error", helper.LabelPodmanIf(podman, func() {
			var stderr string
			BeforeEach(func() {
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(cmpName))
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm install", "npm install-does-not-exist")
				args := []string{"dev", "--no-watch", "--random-ports"}
				if podman {
					args = append(args, "--platform", "podman")
				}
				cmd := helper.Cmd("odo", args...)
				if podman {
					cmd = cmd.AddEnv("ODO_EXPERIMENTAL_MODE=true")
				}
				stderr = cmd.ShouldFail().Err()
			})

			It("should error out with some log", func() {
				helper.MatchAllInOutput(stderr, []string{
					"unable to exec command",
					"Usage: npm <command>",
					"Did you mean one of these?",
				})
			})
		}))
	}

	for _, podman := range []bool{false, true} {
		podman := podman
		When("Create and dev java-springboot component", helper.LabelPodmanIf(podman, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				devfileCmpName = "javaspringboot-" + helper.RandString(6)
				helper.Cmd("odo", "init", "--name", devfileCmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile.yaml")).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should execute default build and run commands correctly", func() {

				cmp := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				cmdOutput, _ := cmp.Exec("runtime",
					[]string{
						"bash", "-c",
						// [s] to not match the current command: https://unix.stackexchange.com/questions/74185/how-can-i-prevent-grep-from-showing-up-in-ps-results
						"grep [s]pring-boot:run /proc/*/cmdline",
					},
					pointer.Bool(true),
				)
				Expect(cmdOutput).To(MatchRegexp("Binary file .* matches"))
			})
		}))
	}

	for _, podman := range []bool{false, true} {
		podman := podman
		When("setting git config and running odo dev", func() {
			remoteURL := "https://github.com/odo-devfiles/nodejs-ex"
			devfileCmpName := "nodejs"
			BeforeEach(func() {
				if podman {
					Skip("Not implemented yet on Podman - see https://github.com/redhat-developer/odo/issues/6493")
				}
				helper.Cmd("git", "init").ShouldPass()
				remote := "origin"
				helper.Cmd("git", "remote", "add", remote, remoteURL).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.Cmd("odo", "init", "--name", devfileCmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			})

			It("should create vcs-uri annotation for the deployment when running odo dev",
				helper.LabelPodmanIf(podman, func() {
					err := helper.RunDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					}, func(session *gexec.Session, outContents []byte, errContents []byte, ports map[string]string) {
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
				}))
		})
	}

	for _, podman := range []bool{true, false} {
		podman := podman

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
			When("running odo dev with alternative commands - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {

				var devfileCmpName string
				type testCase struct {
					buildCmd          string
					runCmd            string
					devAdditionalOpts []string
					checkFunc         func(stdout, stderr string)
				}
				testForCmd := func(tt testCase) {
					err := helper.RunDevMode(helper.DevSessionOpts{
						CmdlineArgs: tt.devAdditionalOpts,
						RunOnPodman: podman,
					}, func(session *gexec.Session, outContents []byte, errContents []byte, ports map[string]string) {
						stdout := string(outContents)
						stderr := string(errContents)

						By("checking the output of the command", func() {
							helper.MatchAllInOutput(stdout, []string{
								fmt.Sprintf("Building your application in container (command: %s)", tt.buildCmd),
								fmt.Sprintf("Executing the application (command: %s)", tt.runCmd),
							})
						})

						if tt.checkFunc != nil {
							tt.checkFunc(stdout, stderr)
						}

						By("verifying the exposed application endpoint", func() {
							url := fmt.Sprintf("http://%s", ports["3000"])
							resp, err := http.Get(url)
							Expect(err).ToNot(HaveOccurred())
							defer resp.Body.Close()

							body, _ := io.ReadAll(resp.Body)
							helper.MatchAllInOutput(string(body), []string{"Hello from Node.js Starter Application!"})
						})

					})
					Expect(err).ToNot(HaveOccurred())
				}

				remoteFileChecker := func(path string) {
					component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					out, _ := component.Exec("runtime", []string{"stat", path}, pointer.Bool(true))
					Expect(out).To(ContainSubstring(path))
				}

				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile-with-alternative-commands.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
				})

				When("running odo dev with a build command", func() {
					buildCmdTestFunc := func(buildCmd string, checkFunc func(stdout, stderr string)) {
						testForCmd(
							testCase{
								buildCmd:          buildCmd,
								runCmd:            "devrun",
								devAdditionalOpts: []string{"--build-command", buildCmd},
								checkFunc:         checkFunc,
							},
						)
					}

					It("should error out on an invalid command", func() {
						By("calling with an invalid build command", func() {
							args := []string{"dev", "--random-ports", "--build-command", "build-command-does-not-exist"}
							if podman {
								args = append(args, "--platform", "podman")
							}
							cmd := helper.Cmd("odo", args...)
							if podman {
								cmd.AddEnv("ODO_EXPERIMENTAL_MODE=true")
							}
							output := cmd.ShouldFail().Err()
							Expect(output).To(ContainSubstring("no build command with name \"build-command-does-not-exist\" found in Devfile"))
						})

						By("calling with a command of another kind (not build)", func() {
							// devrun is a valid run command, not a build command
							args := []string{"dev", "--random-ports", "--build-command", "devrun"}
							if podman {
								args = append(args, "--platform", "podman")
							}
							cmd := helper.Cmd("odo", args...)
							if podman {
								cmd.AddEnv("ODO_EXPERIMENTAL_MODE=true")
							}
							output := cmd.ShouldFail().Err()
							Expect(output).To(ContainSubstring("no build command with name \"devrun\" found in Devfile"))
						})
					})

					It("should execute the custom non-default build command successfully", func() {
						buildCmdTestFunc("my-custom-build", func(stdout, stderr string) {
							By("checking that it did not execute the default build command", func() {
								helper.DontMatchAllInOutput(stdout, []string{
									"Building your application in container (command: devbuild)",
								})
							})

							By("verifying that the custom command ran successfully", func() {
								remoteFileChecker("/projects/file-from-my-custom-build")
							})
						})
					})

					It("should execute the default build command successfully if specified explicitly", func() {
						// devbuild is the default build command
						buildCmdTestFunc("devbuild", func(stdout, stderr string) {
							By("checking that it did not execute the custom build command", func() {
								helper.DontMatchAllInOutput(stdout, []string{
									"Building your application in container (command: my-custom-build)",
								})
							})
						})
					})
				})

				When("running odo dev with a run command", func() {
					runCmdTestFunc := func(runCmd string, checkFunc func(stdout, stderr string)) {
						testForCmd(
							testCase{
								buildCmd:          "devbuild",
								runCmd:            runCmd,
								devAdditionalOpts: []string{"--run-command", runCmd},
								checkFunc:         checkFunc,
							},
						)
					}

					It("should error out on an invalid command", func() {
						By("calling with an invalid run command", func() {
							args := []string{"dev", "--random-ports", "--run-command", "run-command-does-not-exist"}
							if podman {
								args = append(args, "--platform", "podman")
							}
							cmd := helper.Cmd("odo", args...)
							if podman {
								cmd.AddEnv("ODO_EXPERIMENTAL_MODE=true")
							}
							output := cmd.ShouldFail().Err()
							Expect(output).To(ContainSubstring("no run command with name \"run-command-does-not-exist\" found in Devfile"))
						})

						By("calling with a command of another kind (not run)", func() {
							// devbuild is a valid build command, not a run command
							args := []string{"dev", "--random-ports", "--run-command", "devbuild"}
							if podman {
								args = append(args, "--platform", "podman")
							}
							cmd := helper.Cmd("odo", args...)
							if podman {
								cmd.AddEnv("ODO_EXPERIMENTAL_MODE=true")
							}
							output := cmd.ShouldFail().Err()
							Expect(output).To(ContainSubstring("no run command with name \"devbuild\" found in Devfile"))
						})
					})

					It("should execute the custom non-default run command successfully", func() {
						runCmdTestFunc("my-custom-run", func(stdout, stderr string) {
							By("checking that it did not execute the default run command", func() {
								helper.DontMatchAllInOutput(stdout, []string{
									"Executing the application (command: devrun)",
								})
							})

							By("verifying that the custom command ran successfully", func() {
								remoteFileChecker("/projects/file-from-my-custom-run")
							})
						})
					})

					It("should execute the default run command successfully if specified explicitly", func() {
						// devrun is the default run command
						runCmdTestFunc("devrun", func(stdout, stderr string) {
							By("checking that it did not execute the custom run command", func() {
								helper.DontMatchAllInOutput(stdout, []string{
									"Executing the application (command: my-custom-run)",
								})
							})
						})
					})
				})

				It("should execute the custom non-default build and run commands successfully", func() {
					buildCmd := "my-custom-build"
					runCmd := "my-custom-run"

					testForCmd(
						testCase{
							buildCmd:          buildCmd,
							runCmd:            runCmd,
							devAdditionalOpts: []string{"--build-command", buildCmd, "--run-command", runCmd},
							checkFunc: func(stdout, stderr string) {
								By("checking that it did not execute the default build and run commands", func() {
									helper.DontMatchAllInOutput(stdout, []string{
										"Building your application in container (command: devbuild)",
										"Executing the application (command: devrun)",
									})
								})

								By("verifying that the custom build command ran successfully", func() {
									remoteFileChecker("/projects/file-from-my-custom-build")
								})

								By("verifying that the custom run command ran successfully", func() {
									remoteFileChecker("/projects/file-from-my-custom-run")
								})
							},
						},
					)
				})
			}))
		}
	}

	// Tests https://github.com/redhat-developer/odo/issues/3838
	for _, podman := range []bool{true, false} {
		podman := podman
		When("java-springboot application is created and running odo dev", helper.LabelPodmanIf(podman, func() {
			var session helper.DevSession
			var component helper.Component
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-registry.yaml")).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
				component = helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
					CmdlineArgs: []string{"-v", "4"},
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			When("Update the devfile.yaml", func() {

				BeforeEach(func() {
					helper.ReplaceString("devfile.yaml", "memoryLimit: 768Mi", "memoryLimit: 767Mi")
					var err error
					_, _, _, err = session.WaitSync()
					Expect(err).ToNot(HaveOccurred())
				})

				It("Should build the application successfully", func() {
					podLogs := component.GetPodLogs()
					Expect(podLogs).To(ContainSubstring("BUILD SUCCESS"))
				})

				When("compare the local and remote files", func() {

					remoteFiles := []string{}
					localFiles := []string{}

					BeforeEach(func() {
						// commonVar.CliRunner.PodsShouldBeRunning(commonVar.Project, podName)
						output, _ := component.Exec("tools", []string{"find", "/projects"}, pointer.Bool(true))

						outputArr := []string{}
						sc := bufio.NewScanner(strings.NewReader(output))
						for sc.Scan() {
							outputArr = append(outputArr, sc.Text())
						}

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
		}))
	}

	for _, podman := range []bool{false, true} {
		podman := podman
		When("node-js application is created and deployed with devfile schema 2.2.0", helper.LabelPodmanIf(podman, func() {

			ensureResource := func(memorylimit, memoryrequest string) {
				component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				podDef := component.GetPodDef()

				By("check for memoryLimit", func() {
					memVal := podDef.Spec.Containers[0].Resources.Limits.Memory().String()
					Expect(memVal).To(Equal(memorylimit))
				})

				if !podman {
					// Resource Requests are not returned by podman generate kube (as of podman v4.3.1)
					// TODO(feloy) are they taken into account?

					By("check for memoryRequests", func() {
						memVal := podDef.Spec.Containers[0].Resources.Requests.Memory().String()
						Expect(memVal).To(Equal(memoryrequest))
					})
				}
			}

			var session helper.DevSession
			BeforeEach(func() {
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-with-MR-CL-CR.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(cmpName))
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should check memory Request and Limit", func() {
				ensureResource("1Gi", "512Mi")
			})

			if !podman {
				When("Update the devfile.yaml, and waiting synchronization", func() {

					BeforeEach(func() {
						helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-MR-CL-CR-modified.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
						var err error
						_, _, _, err = session.WaitSync()
						Expect(err).ToNot(HaveOccurred())
					})

					It("should check cpuLimit, cpuRequests, memoryRequests after restart", func() {
						ensureResource("1028Mi", "550Mi")
					})
				})
			}
		}))
	}

	for _, podman := range []bool{false, true} {
		podman := podman
		When("creating nodejs component, doing odo dev and run command has dev.odo.push.path attribute", helper.LabelPodmanIf(podman, func() {
			var session helper.DevSession
			var devStarted bool
			BeforeEach(func() {
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path",
					helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-remote-attributes.yaml")).ShouldPass()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)

				// create a folder and file which shouldn't be pushed
				helper.MakeDir(filepath.Join(commonVar.Context, "views"))
				_, _ = helper.CreateSimpleFile(filepath.Join(commonVar.Context, "views"), "view", ".html")

				helper.ReplaceString("package.json", "node server.js", "node server/server.js")
				var err error
				session, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())
				devStarted = true
			})
			AfterEach(func() {
				if devStarted {
					session.Stop()
					session.WaitEnd()
				}
			})

			It("should sync only the mentioned files at the appropriate remote destination", func() {
				component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				stdOut, _ := component.Exec("runtime", []string{"ls", "-lai", "/projects"}, pointer.Bool(true))

				helper.MatchAllInOutput(stdOut, []string{"package.json", "server"})
				helper.DontMatchAllInOutput(stdOut, []string{"test", "views", "devfile.yaml"})

				stdOut, _ = component.Exec("runtime", []string{"ls", "-lai", "/projects/server"}, pointer.Bool(true))
				helper.MatchAllInOutput(stdOut, []string{"server.js", "test"})
			})
		}))
	}

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

			session, _, errContents, _, err := helper.StartDevMode(helper.DevSessionOpts{})
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				session.Stop()
				session.WaitEnd()
			}()
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

			session, _, errContents, err := helper.StartDevMode(helper.DevSessionOpts{})
			Expect(err).ToNot(HaveOccurred())
			defer session.Stop()
			helper.MatchAllInOutput(string(errContents), []string{"odo may not work as expected in the default project"})
		})
	})
	*/

	// Test reused and adapted from the now-removed `cmd_devfile_delete_test.go`.
	// cf. https://github.com/redhat-developer/odo/blob/24fd02673d25eb4c7bb166ec3369554a8e64b59c/tests/integration/devfile/cmd_devfile_delete_test.go#L172-L238
	When("a component with endpoints is bootstrapped and pushed", func() {

		var devSession helper.DevSession

		BeforeEach(func() {
			cmpName = "nodejs-with-endpoints"
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path",
				helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-multiple-endpoints.yaml")).ShouldPass()

			var err error
			devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			devSession.Kill()
			devSession.WaitEnd()
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

		for _, podman := range []bool{true, false} {
			podman := podman
			When("a container component defines a Command or Args - "+devfileHandlerCtx.name, helper.LabelPodmanIf(podman, func() {
				var devfileCmpName string
				var stdoutBytes, stderrBytes []byte
				var devSession helper.DevSession
				var err error

				BeforeEach(func() {
					devfileCmpName = helper.RandString(6)
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "issue-5620-devfile-with-container-command-args.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(devfileCmpName))
					if devfileHandlerCtx.sourceHandler != nil {
						devfileHandlerCtx.sourceHandler(commonVar.Context, devfileCmpName)
					}
					devSession, stdoutBytes, stderrBytes, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					})
					Expect(err).ShouldNot(HaveOccurred())

				})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should run odo dev successfully (#5620)", func() {
					const errorMessage = "Failed to create the component:"
					Expect(string(stdoutBytes)).ToNot(ContainSubstring(errorMessage))
					Expect(string(stderrBytes)).ToNot(ContainSubstring(errorMessage))

					component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					component.Exec("runtime",
						[]string{
							remotecmd.ShellExecutable,
							"-c",
							fmt.Sprintf("kill -0 $(cat %s/.odo_cmd_run.pid) 2>/dev/null ; echo -n $?",
								strings.TrimSuffix(storage.SharedDataMountPath, "/")),
						},
						pointer.Bool(true),
					)
				})
			}))
		}
	}

	for _, podman := range []bool{true, false} {
		podman := podman
		When("a component with multiple endpoints is run", helper.LabelPodmanIf(podman, func() {
			stateFile := ".odo/devstate.json"
			var devSession helper.DevSession
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-multiple-endpoints"), commonVar.Context)
				if !podman {
					helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				}
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-multiple-endpoints.yaml")).ShouldPass()
				Expect(helper.VerifyFileExists(".odo/devstate.json")).To(BeFalse())
				var err error
				devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())
			})

			When("odo dev is stopped", func() {
				BeforeEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should remove forwarded ports from state file", func() {
					Expect(helper.VerifyFileExists(stateFile)).To(BeTrue())
					contentJSON, err := os.ReadFile(stateFile)
					Expect(err).ToNot(HaveOccurred())
					helper.JsonPathContentIs(string(contentJSON), "forwardedPorts", "")
				})
			})
		}))

		When("a component with multiple endpoints is run", helper.LabelPodmanIf(podman, func() {
			stateFile := ".odo/devstate.json"
			var devSession helper.DevSession
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-multiple-endpoints"), commonVar.Context)
				if !podman {
					helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
				}
				helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-multiple-endpoints.yaml")).ShouldPass()
				Expect(helper.VerifyFileExists(".odo/devstate.json")).To(BeFalse())
				var err error
				devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				// We stop the process so the process does not remain after the end of the tests
				devSession.Stop()
				devSession.WaitEnd()
			})

			It("should create a state file containing forwarded ports", func() {
				Expect(helper.VerifyFileExists(stateFile)).To(BeTrue())
				contentJSON, err := os.ReadFile(stateFile)
				Expect(err).ToNot(HaveOccurred())
				helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.0.containerName", "runtime")
				helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.1.containerName", "runtime")
				helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.0.localAddress", "127.0.0.1")
				helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.1.localAddress", "127.0.0.1")
				helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.0.containerPort", "3000")
				helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.1.containerPort", "4567")
				helper.JsonPathContentIsValidUserPort(string(contentJSON), "forwardedPorts.0.localPort")
				helper.JsonPathContentIsValidUserPort(string(contentJSON), "forwardedPorts.1.localPort")
			})
		}))

		// TODO: anandrkskd
		// not test as expected,
		// 1. git ignore should be modified before odo dev
		// 2. should not pass on podman
		When("a devfile with a local parent is used for odo dev and the parent is not synced", helper.LabelPodmanIf(podman, func() {
			var devSession helper.DevSession
			BeforeEach(func() {
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-child.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameSetter(cmpName))
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-parent.yaml"),
					filepath.Join(commonVar.Context, "devfile-parent.yaml"),
					helper.DevfileMetadataNameSetter(cmpName))
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				var err error
				devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())

				gitignorePath := filepath.Join(commonVar.Context, ".gitignore")
				err = helper.AppendToFile(gitignorePath, "\n/devfile-parent.yaml\n")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				// We stop the process so the process does not remain after the end of the tests
				devSession.Stop()
				devSession.WaitEnd()
			})

			When("updating the parent", func() {
				BeforeEach(func() {
					helper.ReplaceString("devfile-parent.yaml", "1024Mi", "1023Mi")
				})

				It("should update the component", func() {
					Eventually(func() string {
						stdout, _, _, err := devSession.GetInfo()
						Expect(err).ToNot(HaveOccurred())
						return string(stdout)
					}, 180, 10).Should(ContainSubstring("Updating Component"))
				})
			})
		}))
	}

	When("using devfile that contains K8s resource to run it on podman", Label(helper.LabelPodman), func() {
		const (
			imgName = "quay.io/unknown-account/myimage" // hard coded from the devfile-composite-apply-different-commandgk.yaml
		)
		var customImgName string

		var session helper.DevSession
		var outContents, errContents []byte
		BeforeEach(func() {
			customImgName = fmt.Sprintf("%s:%s", imgName, cmpName)
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-composite-apply-different-commandgk.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				helper.DevfileMetadataNameSetter(cmpName),
			)
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), imgName, customImgName)
			var err error
			session, outContents, errContents, _, err = helper.StartDevMode(
				helper.DevSessionOpts{RunOnPodman: true, EnvVars: []string{"ODO_PUSH_IMAGES=false"}},
			)
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
			session.WaitEnd()
		})
		It("should show warning about being unable to create the resource when running odo dev on podman", func() {
			Expect(string(errContents)).To(ContainSubstring("Kubernetes components are not supported on Podman. Skipping: "))
			Expect(string(errContents)).To(ContainSubstring("Apply Kubernetes components are not supported on Podman. Skipping: "))
			helper.MatchAllInOutput(string(errContents), []string{"deploy-k8s-resource", "deploy-a-third-k8s-resource"})
		})

		It("should build the images when running odo dev on podman", func() {
			// we do not test push because then it becomes complex to setup image registry credentials to pull the image
			// all pods created by odo have a `PullAlways` image policy.
			Expect(string(outContents)).To(ContainSubstring("Building Image: %s", customImgName))
			component := helper.NewPodmanComponent(cmpName, "app")
			Expect(component.ListImages()).To(ContainSubstring(customImgName))
		})
	})

	for _, podman := range []bool{true, false} {
		podman := podman
		When("a hotReload capable project is used with odo dev", helper.LabelPodmanIf(podman, func() {
			var devSession helper.DevSession
			var stdout []byte
			var executeRunCommand = "Executing the application (command: dev-run)"
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "java-quarkus"), commonVar.Context)
				helper.UpdateDevfileContent(filepath.Join(commonVar.Context, "devfile.yaml"), []helper.DevfileUpdater{helper.DevfileMetadataNameSetter(cmpName)})
				var err error
				devSession, stdout, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: podman,
				})
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				// We stop the process so the process does not remain after the end of the tests
				devSession.Stop()
				devSession.WaitEnd()
			})

			It("should execute the run command", func() {
				Expect(string(stdout)).To(ContainSubstring(executeRunCommand))
			})

			When("a source file is modified", func() {
				BeforeEach(func() {
					helper.ReplaceString(filepath.Join(commonVar.Context, "src", "main", "java", "org", "acme", "GreetingResource.java"), "Hello RESTEasy", "Hi RESTEasy")
					var err error
					stdout, _, _, err = devSession.WaitSync()
					Expect(err).Should(Succeed(), stdout)
				})

				It("should not re-execute the run command", func() {
					Expect(string(stdout)).ToNot(ContainSubstring(executeRunCommand))
				})
			})
		}))
	}

	for _, podman := range []bool{true, false} {
		podman := podman
		Describe("Devfile with no metadata.name", helper.LabelPodmanIf(podman, func() {

			BeforeEach(func() {
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-no-metadata-name.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					helper.DevfileMetadataNameRemover)
			})

			When("running odo dev against a component with no source code", func() {
				var devSession helper.DevSession
				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					})
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should use the directory as component name", func() {
					// when no further source code is available, directory name is returned by alizer.DetectName as component name;
					// and since it is all-numeric in our tests, an "x" prefix is added by util.GetDNS1123Name (called by alizer.DetectName)
					componentName := "x" + filepath.Base(commonVar.Context)

					component := helper.NewComponent(componentName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					component.Exec("runtime",
						[]string{
							remotecmd.ShellExecutable,
							"-c",
							fmt.Sprintf("cat %s/.odo_cmd_devrun.pid", strings.TrimSuffix(storage.SharedDataMountPath, "/")),
						},
						pointer.Bool(true),
					)
				})
			})
		}))
	}

	for _, t := range []struct {
		whenTitle   string
		devfile     string
		checkDev    func(cmpName string, podman bool)
		checkDeploy func(cmpName string)
	}{
		{
			whenTitle: "Devfile contains metadata.language",
			devfile:   "devfile-with-metadata-language.yaml",
			checkDev: func(cmpName string, podman bool) {
				component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				componentLabels := component.GetLabels()
				Expect(componentLabels["app.openshift.io/runtime"]).Should(Equal("javascript"))

				if !podman {
					commonVar.CliRunner.AssertContainsLabel(
						"service",
						commonVar.Project,
						cmpName,
						"app",
						labels.ComponentDevMode,
						"app.openshift.io/runtime",
						"javascript",
					)
				}
			},
			checkDeploy: func(cmpName string) {
				commonVar.CliRunner.AssertContainsLabel(
					"deployment",
					commonVar.Project,
					cmpName,
					"app",
					labels.ComponentDeployMode,
					"app.openshift.io/runtime",
					"javascript",
				)
			},
		},

		{
			whenTitle: "Devfile contains metadata.language invalid as a label value",
			devfile:   "devfile-with-metadata-language-as-invalid-label.yaml",
			checkDev: func(cmpName string, podman bool) {
				component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				componentLabels := component.GetLabels()
				Expect(componentLabels["app.openshift.io/runtime"]).Should(Equal("a-custom-language"))
				if !podman {
					commonVar.CliRunner.AssertContainsLabel(
						"service",
						commonVar.Project,
						cmpName,
						"app",
						labels.ComponentDevMode,
						"app.openshift.io/runtime",
						"a-custom-language",
					)
				}
			},
			checkDeploy: func(cmpName string) {
				commonVar.CliRunner.AssertContainsLabel(
					"deployment",
					commonVar.Project,
					cmpName,
					"app",
					labels.ComponentDeployMode,
					"app.openshift.io/runtime",
					"a-custom-language",
				)
			},
		},

		{
			whenTitle: "Devfile contains metadata.projectType",
			devfile:   "devfile-with-metadata-project-type.yaml",
			checkDev: func(cmpName string, podman bool) {
				component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				componentLabels := component.GetLabels()
				Expect(componentLabels["app.openshift.io/runtime"]).Should(Equal("nodejs"))
				if !podman {
					commonVar.CliRunner.AssertContainsLabel(
						"service",
						commonVar.Project,
						cmpName,
						"app",
						labels.ComponentDevMode,
						"app.openshift.io/runtime",
						"nodejs",
					)
				}
			},
			checkDeploy: func(cmpName string) {
				commonVar.CliRunner.AssertContainsLabel(
					"deployment",
					commonVar.Project,
					cmpName,
					"app",
					labels.ComponentDeployMode,
					"app.openshift.io/runtime",
					"nodejs",
				)
			},
		},

		{
			whenTitle: "Devfile contains metadata.projectType invalid as a label value",
			devfile:   "devfile-with-metadata-project-type-as-invalid-label.yaml",
			checkDev: func(cmpName string, podman bool) {
				component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				componentLabels := component.GetLabels()
				Expect(componentLabels["app.openshift.io/runtime"]).Should(Equal("dotnode"))
				if !podman {
					commonVar.CliRunner.AssertContainsLabel(
						"service",
						commonVar.Project,
						cmpName,
						"app",
						labels.ComponentDevMode,
						"app.openshift.io/runtime",
						"dotnode",
					)
				}
			},
			checkDeploy: func(cmpName string) {
				commonVar.CliRunner.AssertContainsLabel(
					"deployment",
					commonVar.Project,
					cmpName,
					"app",
					labels.ComponentDeployMode,
					"app.openshift.io/runtime",
					"dotnode",
				)
			},
		},

		{
			whenTitle: "Devfile contains neither metadata.language nor metadata.projectType",
			devfile:   "devfile-with-metadata-no-language-project-type.yaml",
			checkDev: func(cmpName string, podman bool) {
				component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				componentLabels := component.GetLabels()
				_, found := componentLabels["app.openshift.io/runtime"]
				Expect(found).Should(BeFalse(), "app.openshift.io/runtime label exists")

				if !podman {
					commonVar.CliRunner.AssertNoContainsLabel(
						"service",
						commonVar.Project,
						cmpName,
						"app",
						labels.ComponentDevMode,
						"app.openshift.io/runtime",
					)
				}
			},
			checkDeploy: func(cmpName string) {
				commonVar.CliRunner.AssertNoContainsLabel(
					"deployment",
					commonVar.Project,
					cmpName,
					"app",
					labels.ComponentDeployMode,
					"app.openshift.io/runtime",
				)
			},
		},
	} {

		t := t
		for _, podman := range []bool{true, false} {
			podman := podman

			When(t.whenTitle, helper.LabelPodmanIf(podman, func() {

				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", t.devfile),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(cmpName))
				})

				When("running odo dev", func() {
					var devSession helper.DevSession
					BeforeEach(func() {
						var err error
						devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
							RunOnPodman: podman,
						})
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})

					It("should set the correct value in labels of resources", func() {
						t.checkDev(cmpName, podman)
					})
				})

				if !podman { // not implement for podman
					When("odo deploy is executed", func() {
						BeforeEach(func() {

							helper.Cmd("odo", "deploy").ShouldPass()
						})

						AfterEach(func() {
							helper.Cmd("odo", "delete", "component", "--force")
						})

						It("should set the correct value in labels of deployed resources", func() {
							t.checkDeploy(cmpName)
						})
					})
				}
			}))
		}
	}

	for _, ctx := range []struct {
		devfile   string
		checkFunc func(podOut *corev1.Pod)
		podman    bool
	}{
		{
			devfile: "devfile-pod-container-overrides.yaml",
			checkFunc: func(podOut *corev1.Pod) {
				Expect(podOut.Spec.Containers[0].Resources.Limits.Memory().String()).To(ContainSubstring("512Mi"))
				Expect(podOut.Spec.Containers[0].Resources.Limits.Cpu().String()).To(ContainSubstring("250m"))
				Expect(podOut.Spec.ServiceAccountName).To(ContainSubstring("new-service-account"))
			},
			podman: false,
		},
		{
			devfile: "devfile-container-override-on-podman.yaml",
			checkFunc: func(podOut *corev1.Pod) {
				Expect(podOut.Spec.Containers[0].SecurityContext.RunAsUser).To(Equal(pointer.Int64(1001)))
				Expect(podOut.Spec.Containers[0].SecurityContext.RunAsGroup).To(Equal(pointer.Int64(1001)))
			},
			podman: true,
		},
	} {
		ctx := ctx
		Context("Devfile contains pod-overrides and container-overrides attributes", helper.LabelPodmanIf(ctx.podman, func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", ctx.devfile), filepath.Join(commonVar.Context, "devfile.yaml"), helper.DevfileMetadataNameSetter(cmpName))
			})
			It("should override the content in the pod it creates for the component on the cluster", func() {
				err := helper.RunDevMode(helper.DevSessionOpts{
					RunOnPodman: ctx.podman,
				}, func(session *gexec.Session, outContents, _ []byte, _ map[string]string) {
					component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					podOut := component.GetPodDef()
					ctx.checkFunc(podOut)
				})
				Expect(err).To(BeNil())
			})
		}))

	}

	Context("odo dev on podman when podman in unavailable", Label(helper.LabelPodman), func() {
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				helper.DevfileMetadataNameSetter(cmpName))
		})
		It("should fail to run odo dev", func() {
			errOut := helper.Cmd("odo", "dev", "--platform", "podman").WithEnv("PODMAN_CMD=echo", "ODO_EXPERIMENTAL_MODE=true").ShouldFail().Err()
			Expect(errOut).To(ContainSubstring("unable to access podman. Do you have podman client installed and configured correctly? cause: exec: \"echo\": executable file not found in $PATH"))
		})
	})
	Context("odo dev on podman with a devfile bound to fail", Label(helper.LabelPodman), func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				helper.DevfileMetadataNameSetter(cmpName))
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "registry.access.redhat.com/ubi8/nodejs", "registry.access.redhat.com/ubi8/nose")
		})
		It("should fail with an error and cleanup resources", func() {
			errContents := helper.Cmd("odo", "dev", "--platform=podman").AddEnv("ODO_EXPERIMENTAL_MODE=true").ShouldFail().Err()
			helper.MatchAllInOutput(errContents, []string{"Complete Podman output", "registry.access.redhat.com/ubi8/nose", "Repo not found"})
			component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
			component.ExpectIsNotDeployed()
		})
	})

	When("running applications listening on the container loopback interface", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-endpoint-on-loopback"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-endpoint-on-loopback.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				helper.DevfileMetadataNameSetter(cmpName))
		})

		haveHttpResponse := func(status int, body string) types.GomegaMatcher {
			return WithTransform(func(urlWithoutProto string) (*http.Response, error) {
				return http.Get("http://" + urlWithoutProto)
			}, SatisfyAll(HaveHTTPStatus(status), HaveHTTPBody(body)))
		}

		for _, plt := range []string{"", "cluster"} {
			plt := plt

			It("should error out if using --ignore-localhost on any platform other than Podman", func() {
				args := []string{"dev", "--ignore-localhost", "--random-ports"}
				var env []string
				if plt != "" {
					args = append(args, "--platform", plt)
					env = append(env, "ODO_EXPERIMENTAL_MODE=true")
				}
				stderr := helper.Cmd("odo", args...).AddEnv(env...).ShouldFail().Err()
				Expect(stderr).Should(ContainSubstring("--ignore-localhost cannot be used when running in cluster mode"))
			})

			It("should error out if using --forward-localhost on any platform other than Podman", func() {
				args := []string{"dev", "--forward-localhost", "--random-ports"}
				var env []string
				if plt != "" {
					args = append(args, "--platform", plt)
					env = append(env, "ODO_EXPERIMENTAL_MODE=true")
				}
				stderr := helper.Cmd("odo", args...).AddEnv(env...).ShouldFail().Err()
				Expect(stderr).Should(ContainSubstring("--forward-localhost cannot be used when running in cluster mode"))
			})
		}

		When("running on default cluster platform", func() {
			var devSession helper.DevSession
			var stdout, stderr string
			var ports map[string]string

			BeforeEach(func() {
				var bOut, bErr []byte
				var err error
				devSession, bOut, bErr, ports, err = helper.StartDevMode(helper.DevSessionOpts{})
				Expect(err).ShouldNot(HaveOccurred())
				stdout = string(bOut)
				stderr = string(bErr)
			})

			AfterEach(func() {
				devSession.Stop()
				devSession.WaitEnd()
			})

			It("should port-forward successfully", func() {
				By("not displaying warning message for loopback port", func() {
					Expect(stderr).ShouldNot(ContainSubstring("Detected that the following port(s) can be reached only via the container loopback interface"))
				})
				By("forwarding both loopback and non-loopback ports", func() {
					Expect(ports).Should(HaveLen(2))
					Expect(ports).Should(SatisfyAll(HaveKey("3000"), HaveKey("3001")))
				})
				By("displaying both loopback and non-loopback ports as forwarded", func() {
					Expect(stdout).Should(SatisfyAll(
						ContainSubstring("Forwarding from %s -> 3000", ports["3000"]),
						ContainSubstring("Forwarding from %s -> 3001", ports["3001"])))
				})
				By("reaching both loopback and non-loopback ports via port-forwarding", func() {
					for port, body := range map[int]string{
						3000: "Hello from Node.js Application!",
						3001: "Hello from Node.js Admin Application!",
					} {
						Eventually(func(g Gomega) {
							g.Expect(ports[strconv.Itoa(port)]).Should(haveHttpResponse(http.StatusOK, body))
						}).WithTimeout(60 * time.Second).WithPolling(3 * time.Second).Should(Succeed())
					}
				})
			})
		})

		Context("running on Podman", Label(helper.LabelPodman), func() {

			It("should error out if using both --ignore-localhost and --forward-localhost", func() {
				stderr := helper.Cmd("odo", "dev", "--random-ports", "--platform", "podman", "--ignore-localhost", "--forward-localhost").
					AddEnv("ODO_EXPERIMENTAL_MODE=true").
					ShouldFail().
					Err()
				Expect(stderr).Should(ContainSubstring("--ignore-localhost and --forward-localhost cannot be used together"))
			})

			It("should error out if not ignoring localhost", func() {
				stderr := helper.Cmd("odo", "dev", "--random-ports", "--platform", "podman").AddEnv("ODO_EXPERIMENTAL_MODE=true").ShouldFail().Err()
				Expect(stderr).Should(ContainSubstring("Detected that the following port(s) can be reached only via the container loopback interface: admin (3001)"))
			})

			When("ignoring localhost", func() {

				var devSession helper.DevSession
				var stderr string
				var ports map[string]string

				BeforeEach(func() {
					var bErr []byte
					var err error
					devSession, _, bErr, ports, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: true,
						CmdlineArgs: []string{"--ignore-localhost"},
					})
					Expect(err).ShouldNot(HaveOccurred())
					stderr = string(bErr)
				})

				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should port-forward successfully", func() {
					By("displaying warning message for loopback port", func() {
						Expect(stderr).Should(ContainSubstring("Detected that the following port(s) can be reached only via the container loopback interface: admin (3001)"))
					})
					By("creating a pod with a single container pod because --forward-localhost is false", func() {
						podDef := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner).GetPodDef()
						Expect(podDef.Spec.Containers).Should(HaveLen(1))
						Expect(podDef.Spec.Containers[0].Name).Should(Equal(fmt.Sprintf("%s-app-runtime", cmpName)))
					})
					By("reaching the local port for the non-loopback interface", func() {
						Eventually(func(g Gomega) {
							g.Expect(ports["3000"]).Should(haveHttpResponse(http.StatusOK, "Hello from Node.js Application!"))
						}).WithTimeout(60 * time.Second).WithPolling(3 * time.Second).Should(Succeed())
					})
					By("not succeeding to reach the local port for the loopback interface", func() {
						// By design, Podman will not forward to container apps listening on localhost.
						// See https://github.com/redhat-developer/odo/issues/6510 and https://github.com/containers/podman/issues/17353
						Consistently(func() error {
							_, err := http.Get("http://" + ports["3001"])
							return err
						}).Should(HaveOccurred())
					})
				})

				When("making changes in the project source code during the dev session", func() {
					BeforeEach(func() {
						helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "Hello from", "Hiya from the updated")
						var err error
						_, _, ports, err = devSession.WaitSync()
						Expect(err).ShouldNot(HaveOccurred())
					})

					It("should port-forward successfully", func() {
						By("reaching the local port for the non-loopback interface", func() {
							Eventually(func(g Gomega) {
								g.Expect(ports["3000"]).Should(haveHttpResponse(http.StatusOK, "Hiya from the updated Node.js Application!"))
							}).WithTimeout(60 * time.Second).WithPolling(3 * time.Second).Should(Succeed())
						})
						By("not succeeding to reach the local port for the loopback interface", func() {
							// By design, Podman will not forward to container apps listening on localhost.
							// See https://github.com/redhat-developer/odo/issues/6510 and https://github.com/containers/podman/issues/17353
							Consistently(func() error {
								_, err := http.Get("http://" + ports["3001"])
								return err
							}).Should(HaveOccurred())
						})
					})
				})
			})

			When("forwarding localhost", func() {
				var devSession helper.DevSession
				var stdout, stderr string
				var ports map[string]string

				BeforeEach(func() {
					var bOut, bErr []byte
					var err error
					devSession, bOut, bErr, ports, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: true,
						CmdlineArgs: []string{"--forward-localhost"},
					})
					Expect(err).ShouldNot(HaveOccurred())
					stdout = string(bOut)
					stderr = string(bErr)
				})

				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should port-forward successfully", func() {
					By("not displaying warning message for loopback port", func() {
						for _, output := range []string{stdout, stderr} {
							Expect(output).ShouldNot(ContainSubstring("Detected that the following port(s) can be reached only via the container loopback interface"))
						}
					})
					By("creating a pod with a two-containers pod because --forward-localhost is true", func() {
						podDef := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner).GetPodDef()
						Expect(podDef.Spec.Containers).Should(HaveLen(2))
						var found bool
						var pfHelperContainer corev1.Container
						for _, container := range podDef.Spec.Containers {
							if container.Name == fmt.Sprintf("%s-app-odo-helper-port-forwarding", cmpName) {
								pfHelperContainer = container
								found = true
								break
							}
						}
						Expect(found).Should(BeTrue(), fmt.Sprintf("Could not find container 'odo-helper-port-forwarding' in pod spec: %v", podDef))
						Expect(pfHelperContainer.Image).Should(HavePrefix("quay.io/devfile/base-developer-image"))
					})
					By("reaching the local port for the non-loopback interface", func() {
						Eventually(func(g Gomega) {
							g.Expect(ports["3000"]).Should(haveHttpResponse(http.StatusOK, "Hello from Node.js Application!"))
						}).WithTimeout(60 * time.Second).WithPolling(3 * time.Second).Should(Succeed())
					})
					By("reaching the local port for the loopback interface", func() {
						Eventually(func(g Gomega) {
							g.Expect(ports["3001"]).Should(haveHttpResponse(http.StatusOK, "Hello from Node.js Admin Application!"))
						}).WithTimeout(60 * time.Second).WithPolling(3 * time.Second).Should(Succeed())
					})
				})

				When("making changes in the project source code during the dev session", func() {
					BeforeEach(func() {
						helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "Hello from", "Hiya from the updated")
						var err error
						_, _, ports, err = devSession.WaitSync()
						Expect(err).ShouldNot(HaveOccurred())
					})

					It("should port-forward successfully", func() {
						By("reaching the local port for the non-loopback interface", func() {
							Eventually(func(g Gomega) {
								g.Expect(ports["3000"]).Should(haveHttpResponse(http.StatusOK, "Hiya from the updated Node.js Application!"))
							}).WithTimeout(60 * time.Second).WithPolling(3 * time.Second).Should(Succeed())
						})
						By("reaching the local port for the loopback interface", func() {
							Eventually(func(g Gomega) {
								g.Expect(ports["3001"]).Should(haveHttpResponse(http.StatusOK, "Hiya from the updated Node.js Admin Application!"))
							}).WithTimeout(60 * time.Second).WithPolling(3 * time.Second).Should(Succeed())
						})
					})
				})
			})
		})

	})
})
