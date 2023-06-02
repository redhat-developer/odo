package integration

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	segment "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/state"
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

	When("a component is bootstrapped", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeFalse())
		})

		It("should fail to run odo dev when not connected to any cluster", Label(helper.LabelNoCluster), func() {
			errOut := helper.Cmd("odo", "dev").ShouldFail().Err()
			Expect(errOut).To(ContainSubstring("unable to access the cluster"))
		})
		It("should fail to run odo dev when podman is nil", Label(helper.LabelPodman), func() {
			errOut := helper.Cmd("odo", "dev", "--platform", "podman").WithEnv("PODMAN_CMD=echo").ShouldFail().Err()
			Expect(errOut).To(ContainSubstring("unable to access podman"))
		})

		It("should start on cluster even if Podman client takes long to initialize", func() {
			if runtime.GOOS == "windows" {
				Skip("skipped on Windows as it requires Unix permissions")
			}
			_, err := os.Stat("/bin/bash")
			if errors.Is(err, fs.ErrNotExist) {
				Skip("skipped because bash executable not found")
			}

			// odo dev on cluster should not wait for the Podman client to initialize properly, if this client takes very long.
			// See https://github.com/redhat-developer/odo/issues/6575.
			// StartDevMode will time out if Podman client takes too long to initialize.
			delayer := filepath.Join(commonVar.Context, "podman-cmd-delayer")
			err = helper.CreateFileWithContentAndPerm(delayer, `#!/bin/bash

echo Delaying command execution... >&2
sleep 10
echo "$@"
`, 0755)
			Expect(err).ShouldNot(HaveOccurred())

			var devSession helper.DevSession
			var stderrBytes []byte
			devSession, _, stderrBytes, _, err = helper.StartDevMode(helper.DevSessionOpts{
				RunOnPodman: false,
				CmdlineArgs: []string{"-v", "3"},
				EnvVars: []string{
					"PODMAN_CMD=" + delayer,
					"PODMAN_CMD_INIT_TIMEOUT=1s",
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			defer func() {
				devSession.Kill()
				devSession.WaitEnd()
			}()

			Expect(string(stderrBytes)).Should(MatchRegexp("timeout \\([^()]+\\) while waiting for Podman version"))
		})

		When("using a default namespace", func() {
			BeforeEach(func() {
				commonVar.CliRunner.SetProject("default")
			})
			AfterEach(func() {
				commonVar.CliRunner.SetProject(commonVar.Project)
			})
			It("should print warning about default namespace when running odo dev", func() {
				namespace := "project"
				if helper.IsKubernetesCluster() {
					namespace = "namespace"
				}
				// Resources might not pass the security requirements on the default namespace on certain clusters (case of OpenShift 4.14),
				// but this is not important here, as we just want to make sure that the warning message is displayed (even if the Dev Session does not start correctly).
				devSession, _, stderr, err := helper.WaitForDevModeToContain(helper.DevSessionOpts{}, "Running on the cluster in Dev mode", false, false)
				Expect(err).ShouldNot(HaveOccurred())
				defer func() {
					devSession.Stop()
					devSession.WaitEnd()
				}()
				Expect(string(stderr)).To(ContainSubstring(fmt.Sprintf("You are using \"default\" %[1]s, odo may not work as expected in the default %[1]s.", namespace)))
			})
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

			It("should fail when using --random-ports and --port-forward together", helper.LabelPodmanIf(podman, func() {
				args := []string{"dev", "--random-ports", "--port-forward=8000:3000"}
				if podman {
					args = append(args, "--platform", "podman")
				}
				errOut := helper.Cmd("odo", args...).ShouldFail().Err()
				Expect(errOut).To(ContainSubstring("--random-ports and --port-forward cannot be used together"))
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

		It("should not set securitycontext for podsecurity admission", func() {
			if os.Getenv("KUBERNETES") != "true" {
				Skip("This is a Kubernetes specific scenario, skipping")
			}
			err := helper.RunDevMode(helper.DevSessionOpts{}, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
				component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
				podDef := component.GetPodDef()
				Expect(podDef.Spec.SecurityContext.RunAsNonRoot).To(BeNil())
				Expect(podDef.Spec.SecurityContext.SeccompProfile).To(BeNil())
			})
			Expect(err).ToNot(HaveOccurred())
		})

		When("pod security is enforced as restricted", func() {
			BeforeEach(func() {
				commonVar.CliRunner.SetLabelsOnNamespace(
					commonVar.Project,
					"pod-security.kubernetes.io/enforce=restricted",
					"pod-security.kubernetes.io/enforce-version=latest",
				)
			})

			It("should set securitycontext for podsecurity admission", func() {
				if os.Getenv("KUBERNETES") != "true" {
					Skip("This is a Kubernetes specific scenario, skipping")
				}
				err := helper.RunDevMode(helper.DevSessionOpts{}, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
					component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
					podDef := component.GetPodDef()
					Expect(*podDef.Spec.SecurityContext.RunAsNonRoot).To(BeTrue())
					Expect(string(podDef.Spec.SecurityContext.SeccompProfile.Type)).To(Equal("RuntimeDefault"))
				})
				Expect(err).ToNot(HaveOccurred())
			})
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
				Expect(stderr).To(ContainSubstring("unable to save state file"))
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
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"),
					"npm start",
					// odo dev now waits some time until the app is ready or a timeout (current set to 1m) expires before starting port-forwarding.
					// So we are sleeping more than the timeout.
					// See https://github.com/redhat-developer/odo/issues/6667
					"sleep 80 ; npm start")

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

		When("Automount volumes are present in the namespace", func() {

			BeforeEach(func() {
				commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "config-automount/"))
			})

			When("odo dev is executed", func() {

				var devSession helper.DevSession

				BeforeEach(func() {
					var err error
					devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{})
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should mount the volumes", func() {
					component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)

					// Check volumes are mounted
					for _, path := range []string{
						"/tmp/automount-default-pvc",
						"/etc/config/automount-default-configmap",
						"/etc/secret/automount-default-secret",

						"/tmp/automount-readonly-pvc",

						"/mnt/mount-path/pvc",
						"/mnt/mount-path/configmap",
						"/mnt/mount-path/secret",

						"/etc/config/automount-access-mode-configmap",
						"/etc/config/automount-access-mode-configmap-decimal",
						"/etc/secret/automount-access-mode-secret",
					} {
						var output string
						Eventually(func() bool {
							output, _ = component.Exec("runtime", []string{"df", path}, nil)
							return len(output) > 0
						}).WithPolling(1 * time.Second).WithTimeout(60 * time.Second).Should(BeTrue())
						// This checks this is really a mount
						Expect(output).ToNot(ContainSubstring("overlay"))
					}

					// Check files are present for configmap / secret
					// and have expected access mode (by default 0644)
					files := map[string]struct {
						content    string
						accessMode string
					}{
						"/etc/config/automount-default-configmap/foo1": {
							content:    "bar1",
							accessMode: "rw-r--r--",
						},
						"/etc/config/automount-default-configmap/ping1": {
							content:    "pong1",
							accessMode: "rw-r--r--",
						},
						"/etc/secret/automount-default-secret/code1": {
							content:    "1234",
							accessMode: "rw-r--r--",
						},
						"/etc/secret/automount-default-secret/secret1": {
							content:    "PassWd1",
							accessMode: "rw-r--r--",
						},

						"/mnt/mount-path/configmap/foo2": {
							content:    "bar2",
							accessMode: "rw-r--r--",
						},
						"/mnt/mount-path/configmap/ping2": {
							content:    "pong2",
							accessMode: "rw-r--r--",
						},
						"/mnt/mount-path/secret/code2": {
							content:    "2345",
							accessMode: "rw-r--r--",
						},
						"/mnt/mount-path/secret/secret2": {
							content:    "PassWd2",
							accessMode: "rw-r--r--",
						},

						"/mnt/subpaths/foo5": {
							content:    "bar5",
							accessMode: "rw-r--r--",
						},
						"/mnt/subpaths/ping5": {
							content:    "pong5",
							accessMode: "rw-r--r--",
						},
						"/mnt/subpaths/code5": {
							content:    "5678",
							accessMode: "rw-r--r--",
						},
						"/mnt/subpaths/secret5": {
							content:    "PassWd5",
							accessMode: "rw-r--r--",
						},

						"/etc/config/automount-access-mode-configmap/config0444": {
							content:    "foo",
							accessMode: "r--r--r--",
						},
						"/etc/config/automount-access-mode-configmap-decimal/config292": {
							content:    "foo-decimal",
							accessMode: "r--r--r--",
						},
						"/etc/secret/automount-access-mode-secret/secret0444": {
							content:    "1234",
							accessMode: "r--r--r--",
						},
						"/etc/config0444": {
							content:    "foo",
							accessMode: "r--r--r--",
						},
						"/etc/secret0444": {
							content:    "5ecr3t",
							accessMode: "r--r--r--",
						},
					}
					for file, desc := range files {
						output, _ := component.Exec("runtime", []string{"cat", file}, pointer.Bool(true))
						Expect(output).To(Equal(desc.content))

						// -L follows symlinks, to get the mode of the targeted file, as files reside on a ..data directory
						output, _ = component.Exec("runtime", []string{"ls", "-lL", file}, pointer.Bool(true))
						Expect(output).To(ContainSubstring(desc.accessMode))
					}

					envVars := map[string]string{
						"foo4":  "bar4",
						"ping4": "pong4",

						"code4":   "4567",
						"secret4": "PassWd4",
					}
					for name, value := range envVars {
						output, _ := component.Exec("runtime", []string{"bash", "-c", "echo -n $" + name}, pointer.Bool(true))
						Expect(output).To(Equal(value))
					}

					// Default PVC is not read-only
					component.Exec("runtime", []string{"touch", "/tmp/automount-default-pvc/newfile"}, pointer.Bool(true))

					// Read-only PVC is read-only
					_, stderr := component.Exec("runtime", []string{"touch", "/tmp/automount-readonly-pvc/newfile"}, pointer.Bool(false))
					Expect(stderr).To(ContainSubstring("Read-only file system"))

				})
			})
		})

		for _, podman := range []bool{true, false} {
			podman := podman

			When("build command takes really long to start", helper.LabelPodmanIf(podman, func() {
				BeforeEach(func() {
					helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"),
						"npm install",
						"echo Will execute command after 20m ... && sleep 1200 && npm install")
				})

				It("should cancel build command and return if odo dev is stopped", func() {
					opts := helper.DevSessionOpts{
						RunOnPodman: podman,
					}
					devSession, _, _, err := helper.WaitForDevModeToContain(opts, "Building your application in container", false, false)
					Expect(err).ShouldNot(HaveOccurred())

					// Build is taking long => it should be cancellable
					devSession.Stop()
					// WaitEnd will timeout after some time, less than the execution of the build command above
					devSession.WaitEnd()
				})
			}))

			When("run command takes really long to start", helper.LabelPodmanIf(podman, func() {
				BeforeEach(func() {
					helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"),
						"npm start",
						"echo Will execute command after 20m ... && sleep 1200 && npm start")
				})

				It("should cancel run command and return if odo dev is stopped", func() {
					opts := helper.DevSessionOpts{
						RunOnPodman: podman,
					}
					// Run command is launched in the background
					devSession, _, _, err := helper.WaitForDevModeToContain(opts, "Waiting for the application to be ready", false, false)
					Expect(err).ShouldNot(HaveOccurred())

					// Build is taking long => it should be cancellable
					devSession.Stop()
					// WaitEnd will timeout after some time, less than the execution of the build command above
					devSession.WaitEnd()
				})
			}))
		}
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
							cmpName)
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
					cmpName)
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

	for _, podman := range []bool{true, false} {
		podman := podman
		Context(fmt.Sprintf("multiple dev sessions with different project are running on same platform (podman=%v), same port", podman), helper.LabelPodmanIf(podman, func() {
			const (
				nodejsContainerPort = "3000"
				goContainerPort     = "8080"
				nodejsCustomAddress = "127.0.10.3"
				goCustomAddress     = "127.0.10.1"
			)
			var (
				nodejsProject, goProject       string
				nodejsDevSession, goDevSession helper.DevSession
				nodejsPorts, goPorts           map[string]string
				nodejsLocalPort                = helper.GetCustomStartPort()
				goLocalPort                    = nodejsLocalPort + 1

				nodejsURL = fmt.Sprintf("%s:%d", nodejsCustomAddress, nodejsLocalPort)
				goURL     = fmt.Sprintf("%s:%d", goCustomAddress, goLocalPort)
			)
			BeforeEach(func() {
				if runtime.GOOS == "darwin" {
					Skip("cannot run this test out of the box on macOS because the test uses a custom address in the range 127.0.0/8 and for macOS we need to ensure the addresses are open for request before using them; Ref: https://superuser.com/questions/458875/how-do-you-get-loopback-addresses-other-than-127-0-0-1-to-work-on-os-x#458877")
				}
				nodejsProject = helper.CreateNewContext()
				goProject = helper.CreateNewContext()
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), nodejsProject)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					filepath.Join(nodejsProject, "devfile.yaml"),
					cmpName+"-nodejs",
				)
				helper.CopyExample(filepath.Join("source", "go"), goProject)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "go-devfiles", "devfile.yaml"),
					filepath.Join(goProject, "devfile.yaml"),
					cmpName+"-go",
				)
			})
			AfterEach(func() {
				helper.DeleteDir(nodejsProject)
				helper.DeleteDir(goProject)
			})
			When("odo dev session is run for nodejs component", func() {
				BeforeEach(func() {
					helper.Chdir(nodejsProject)
					var err error
					nodejsDevSession, _, _, nodejsPorts, err = helper.StartDevMode(helper.DevSessionOpts{
						CmdlineArgs:      []string{"--port-forward", fmt.Sprintf("%d:%s", nodejsLocalPort, nodejsContainerPort)},
						RunOnPodman:      podman,
						TimeoutInSeconds: 0,
						NoRandomPorts:    true,
						CustomAddress:    nodejsCustomAddress,
					})
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					nodejsDevSession.Stop()
					nodejsDevSession.WaitEnd()
					helper.Chdir(commonVar.Context)
				})
				When("odo dev session is run for go project on the same port but different address", func() {
					BeforeEach(func() {
						helper.Chdir(goProject)
						var err error
						goDevSession, _, _, goPorts, err = helper.StartDevMode(helper.DevSessionOpts{
							CmdlineArgs:      []string{"--port-forward", fmt.Sprintf("%d:%s", goLocalPort, goContainerPort)},
							RunOnPodman:      podman,
							TimeoutInSeconds: 0,
							NoRandomPorts:    true,
							CustomAddress:    goCustomAddress,
						})
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						goDevSession.Stop()
						goDevSession.WaitEnd()
						helper.Chdir(commonVar.Context)
					})
					It("should be able to run both the sessions", func() {
						Expect(nodejsPorts[nodejsContainerPort]).To(BeEquivalentTo(nodejsURL))
						Expect(goPorts[goContainerPort]).To(BeEquivalentTo(goURL))
						helper.HttpWaitForWithStatus(fmt.Sprintf("http://%s", nodejsURL), "Hello from Node.js Starter Application!", 1, 0, 200)
						helper.HttpWaitForWithStatus(fmt.Sprintf("http://%s", goURL), "Hello, !", 1, 0, 200)
					})
					When("go and nodejs files are modified", func() {
						BeforeEach(func() {
							var wg sync.WaitGroup
							wg.Add(2)
							go func() {
								defer wg.Done()
								_, _, _, err := nodejsDevSession.WaitSync()
								Expect(err).ToNot(HaveOccurred())
							}()
							go func() {
								defer wg.Done()
								_, _, _, err := goDevSession.WaitSync()
								Expect(err).ToNot(HaveOccurred())
							}()
							helper.ReplaceString(filepath.Join(goProject, "main.go"), "Hello, %s!", "H3110, %s!")
							helper.ReplaceString(filepath.Join(nodejsProject, "server.js"), "Hello from Node.js", "H3110 from Node.js")
							wg.Wait()
						})
						It("should be possible to access both the projects on same address and port", func() {
							Expect(nodejsPorts[nodejsContainerPort]).To(BeEquivalentTo(nodejsURL))
							Expect(goPorts[goContainerPort]).To(BeEquivalentTo(goURL))
							helper.HttpWaitForWithStatus(fmt.Sprintf("http://%s", nodejsURL), "H3110 from Node.js Starter Application!", 1, 0, 200)
							helper.HttpWaitForWithStatus(fmt.Sprintf("http://%s", goURL), "H3110, !", 1, 0, 200)
						})
					})
				})
			})
		}))
	}
	for _, podman := range []bool{true, false} {
		podman := podman
		Context("port-forwarding for the component", helper.LabelPodmanIf(podman, func() {
			for _, manual := range []bool{true, false} {
				manual := manual
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

						It(fmt.Sprintf("should have no endpoint forwarded (podman=%v, manual=%v)", podman, manual), func() {
							Expect(ports).To(BeEmpty())
						})
					})
				})
				for _, customAddress := range []bool{true, false} {
					customAddress := customAddress
					var localAddress string
					if customAddress {
						localAddress = "0.0.0.0"
					}
					for _, customPortForwarding := range []bool{true, false} {
						customPortForwarding := customPortForwarding
						var NoRandomPorts bool
						if customPortForwarding {
							NoRandomPorts = true
						}
						When("devfile has single endpoint", func() {
							var (
								localPort int
							)
							const (
								containerPort = "3000"
							)
							BeforeEach(func() {
								localPort = helper.GetCustomStartPort()
								helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
								helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile.yaml")).ShouldPass()
							})

							When("running odo dev", func() {
								var devSession helper.DevSession
								var ports map[string]string
								BeforeEach(func() {
									var err error
									opts := []string{}
									if customPortForwarding {
										opts = []string{fmt.Sprintf("--port-forward=%d:%s", localPort, containerPort)}
									}
									if manual {
										opts = append(opts, "--no-watch")
									}
									devSession, _, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
										CmdlineArgs:   opts,
										NoRandomPorts: NoRandomPorts,
										RunOnPodman:   podman,
										CustomAddress: localAddress,
									})
									Expect(err).ToNot(HaveOccurred())

								})

								AfterEach(func() {
									devSession.Stop()
									devSession.WaitEnd()
								})

								It(fmt.Sprintf("should expose the endpoint on localhost (podman=%v, manual=%v, customPortForwarding=%v, customAddress=%v)", podman, manual, customPortForwarding, customAddress), func() {
									url := fmt.Sprintf("http://%s", ports[containerPort])
									if customPortForwarding {
										Expect(url).To(ContainSubstring(strconv.Itoa(localPort)))
									}
									resp, err := http.Get(url)
									Expect(err).ToNot(HaveOccurred())
									defer resp.Body.Close()

									body, _ := io.ReadAll(resp.Body)
									helper.MatchAllInOutput(string(body), []string{"Hello from Node.js Starter Application!"})
									Expect(err).ToNot(HaveOccurred())
								})

								When("modifying name for container in Devfile", func() {
									var stdout string
									var stderr string
									BeforeEach(func() {
										if manual {
											if os.Getenv("SKIP_KEY_PRESS") == "true" {
												Skip("This is a unix-terminal specific scenario, skipping")
											}
										}
										var (
											wg          sync.WaitGroup
											err         error
											stdoutBytes []byte
											stderrBytes []byte
										)
										wg.Add(1)
										go func() {
											defer wg.Done()
											stdoutBytes, stderrBytes, ports, err = devSession.WaitSync()
											Expect(err).Should(Succeed())
											stdout = string(stdoutBytes)
											stderr = string(stderrBytes)
										}()
										src := "runtime"
										dst := "other"
										helper.ReplaceString("devfile.yaml", src, dst)
										if manual {
											devSession.PressKey('p')
										}
										wg.Wait()
									})

									It(fmt.Sprintf("should react on the Devfile modification (podman=%v, manual=%v, customPortForwarding=%v, customAddress=%v)", podman, manual, customPortForwarding, customAddress), func() {
										By("not warning users that odo dev needs to be restarted", func() {
											warning := "Please restart 'odo dev'"
											Expect(stdout).ShouldNot(ContainSubstring(warning))
											Expect(stderr).ShouldNot(ContainSubstring(warning))
										})
										By("updating the pod", func() {
											component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
											podDef := component.GetPodDef()
											containerName := podDef.Spec.Containers[0].Name
											Expect(containerName).To(ContainSubstring("other"))
										})

										By("exposing the endpoint", func() {
											Eventually(func(g Gomega) {
												url := fmt.Sprintf("http://%s", ports[containerPort])
												if customPortForwarding {
													Expect(url).To(ContainSubstring(strconv.Itoa(localPort)))
												}
												if customAddress {
													Expect(url).To(ContainSubstring(localAddress))
												}
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
							var (
								localPort1, localPort2, localPort3 int
							)
							const (
								// ContainerPort<N> are hard-coded from devfile-with-multiple-endpoints.yaml
								// Note 1:	Debug endpoints will not be exposed for this instance, so we do not add custom mapping for them.
								// Note 2: We add custom mapping for all the endpoints so that none of them are assigned random ports from the 20001-30001 range;
								// Note 2(contd.): this is to avoid a race condition where a test running in parallel is also assigned similar ranged port the one here, and we fail to access either of them.
								containerPort1 = "3000"
								containerPort2 = "4567"
								containerPort3 = "7890"
							)
							BeforeEach(func() {
								localPort1 = helper.GetCustomStartPort()
								localPort2 = localPort1 + 1
								localPort3 = localPort1 + 2
								helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-multiple-endpoints"), commonVar.Context)
								helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-multiple-endpoints.yaml")).ShouldPass()
							})

							When("running odo dev", func() {
								var devSession helper.DevSession
								var ports map[string]string
								BeforeEach(func() {
									opts := []string{}
									if customPortForwarding {
										opts = []string{fmt.Sprintf("--port-forward=%d:%s", localPort1, containerPort1), fmt.Sprintf("--port-forward=%d:%s", localPort2, containerPort2), fmt.Sprintf("--port-forward=%d:%s", localPort3, containerPort3)}
									}
									if manual {
										opts = append(opts, "--no-watch")
									}
									var err error
									devSession, _, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
										CmdlineArgs:   opts,
										NoRandomPorts: NoRandomPorts,
										RunOnPodman:   podman,
										CustomAddress: localAddress,
									})
									Expect(err).ToNot(HaveOccurred())
								})

								AfterEach(func() {
									devSession.Stop()
									devSession.WaitEnd()
								})

								It(fmt.Sprintf("should expose all endpoints on localhost regardless of exposure(podman=%v, manual=%v, customPortForwarding=%v, customAddress=%v)", podman, manual, customPortForwarding, customAddress), func() {
									By("not exposing debug endpoints", func() {
										for _, p := range []int{5005, 5006} {
											_, found := ports[strconv.Itoa(p)]
											Expect(found).To(BeFalse(), fmt.Sprintf("debug port %d should not be forwarded", p))
										}
									})

									getServerResponse := func(containerPort, localPort string) (string, error) {
										url := fmt.Sprintf("http://%s", ports[containerPort])
										if customPortForwarding {
											Expect(url).To(ContainSubstring(localPort))
										}
										if customAddress {
											Expect(url).To(ContainSubstring(localAddress))
										}
										resp, err := http.Get(url)
										if err != nil {
											return "", err
										}
										defer resp.Body.Close()

										body, _ := io.ReadAll(resp.Body)
										return string(body), nil
									}
									containerPorts := []string{containerPort1, containerPort2, containerPort3}
									localPorts := []int{localPort1, localPort2, localPort3}

									for i := range containerPorts {
										containerPort := containerPorts[i]
										localPort := localPorts[i]
										By(fmt.Sprintf("exposing a port targeting container port %s", containerPort), func() {
											r, err := getServerResponse(containerPort, strconv.Itoa(localPort))
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
										Expect(stdout).ShouldNot(ContainSubstring(warning))
										Expect(stderr).ShouldNot(ContainSubstring(warning))
									})

									for i := range containerPorts {
										containerPort := containerPorts[i]
										localPort := localPorts[i]
										By(fmt.Sprintf("returning the right response when querying port forwarded for container port %s", containerPort),
											func() {
												Eventually(func(g Gomega) string {
													r, err := getServerResponse(containerPort, strconv.Itoa(localPort))
													g.Expect(err).ShouldNot(HaveOccurred())
													return r
												}, 180, 10).Should(Equal("H3110 from Node.js Starter Application!"))
											})
									}
								})
							})

						})

					}
				}
			}
		}))
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
						devfileCmpName)
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
						devfileCmpName)
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
						devfileCmpName)
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
						devfileCmpName)
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
					devfileCmpName)
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
						devfileCmpName)
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
			var gitignorePath, newDirPath, newFilePath1, newFilePath2, newFilePath3, newFilePath4, newFilePath5, stdOut, podName string
			var session helper.DevSession
			var devfileCmpName string
			BeforeEach(func() {
				devfileCmpName = helper.RandString(6)
				gitignorePath = filepath.Join(commonVar.Context, ".gitignore")
				newFilePath1 = filepath.Join(commonVar.Context, "foobar.txt")
				newDirPath = filepath.Join(commonVar.Context, "testdir")
				newFilePath2 = filepath.Join(newDirPath, "foobar.txt")
				newFilePath3 = filepath.Join(newDirPath, "baz.txt")
				newFilePath4 = filepath.Join(newDirPath, "ignore.css")
				newFilePath5 = filepath.Join(newDirPath, "main.css")
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					devfileCmpName)
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
				if err := helper.CreateFileWithContent(newFilePath4, "div {}"); err != nil {
					fmt.Printf("the %s file was not created, reason %v", newFilePath4, err.Error())
				}
				if err := helper.CreateFileWithContent(newFilePath5, "div {}"); err != nil {
					fmt.Printf("the %s file was not created, reason %v", newFilePath5, err.Error())
				}
				if err := helper.CreateFileWithContent(gitignorePath, `foobar.txt
*.css
!main.css`); err != nil {
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
				helper.MatchAllInOutput(stdOut, []string{"main.css"})
				helper.DontMatchAllInOutput(stdOut, []string{"ignore.css"})
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
					devfileCmpName)
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
					devfileCmpName)
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
					devfileCmpName)
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
					devfileCmpName)
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
					devfileCmpName)
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
					devfileCmpName)
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
					devfileCmpName)
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
			DEVFILEPORT             = "8080"
		)
		var session helper.DevSession
		var sessionOut, sessionErr []byte
		var err error
		var ports map[string]string

		BeforeEach(func() {
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-composite-apply-commands.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
		})

		for _, tt := range []struct {
			name                            string
			containerBackendGlobalExtraArgs []string
			imageBuildExtraArgs             []string
			containerRunExtraArgs           []string
		}{
			{
				name: "odo dev is running",
			},
			{
				name: "odo dev is running with image build extra args",
				imageBuildExtraArgs: []string{
					"--platform=linux/amd64",
					"--build-arg=MY_ARG=my_value",
				},
			},
			{
				name: "odo dev is running with container backend global extra args",
				containerBackendGlobalExtraArgs: []string{
					"--log-level=error",
				},
			},
			{
				name: "odo dev is running with container run extra args",
				containerRunExtraArgs: []string{
					"--quiet",
					"--tls-verify=false",
				},
			},
			{
				name: "odo dev is running with both image build and container run extra args",
				imageBuildExtraArgs: []string{
					"--platform=linux/amd64",
					"--build-arg=MY_ARG=my_value",
				},
				containerBackendGlobalExtraArgs: []string{
					"--log-level=panic",
				},
				containerRunExtraArgs: []string{
					"--quiet",
					"--tls-verify=false",
				},
			},
		} {
			tt := tt
			for _, podman := range []bool{false, true} {
				podman := podman
				When(tt.name, helper.LabelPodmanIf(podman, func() {
					BeforeEach(func() {
						helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
						var env []string
						if podman {
							env = append(env, "ODO_PUSH_IMAGES=false")
						} else {
							env = append(env, "PODMAN_CMD=echo")
						}
						if len(tt.containerBackendGlobalExtraArgs) != 0 {
							env = append(env, "ODO_CONTAINER_BACKEND_GLOBAL_ARGS="+strings.Join(tt.containerBackendGlobalExtraArgs, ";"))
						}
						if len(tt.imageBuildExtraArgs) != 0 {
							env = append(env, "ODO_IMAGE_BUILD_ARGS="+strings.Join(tt.imageBuildExtraArgs, ";"))
						}
						var cmdLineArgs []string
						if len(tt.containerRunExtraArgs) != 0 {
							env = append(env, "ODO_CONTAINER_RUN_ARGS="+strings.Join(tt.containerRunExtraArgs, ";"))
						}
						if podman {
							// Increasing verbosity to check that extra args are being passed to the "podman" commands
							cmdLineArgs = append(cmdLineArgs, "-v=4")
						}
						session, sessionOut, sessionErr, ports, err = helper.StartDevMode(helper.DevSessionOpts{
							RunOnPodman: podman,
							EnvVars:     env,
							CmdlineArgs: cmdLineArgs,
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
							var substring string
							if len(tt.containerBackendGlobalExtraArgs) != 0 {
								substring = strings.Join(tt.containerBackendGlobalExtraArgs, " ") + " "
							}
							substring += "build "
							if len(tt.imageBuildExtraArgs) != 0 {
								substring += strings.Join(tt.imageBuildExtraArgs, " ") + " "
							}

							substring += fmt.Sprintf("-t quay.io/unknown-account/myimage -f %s %s",
								filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context)

							if podman {
								Expect(string(sessionErr)).To(ContainSubstring(substring))
							} else {
								Expect(string(sessionOut)).To(ContainSubstring(substring))
								Expect(string(sessionOut)).To(ContainSubstring("push quay.io/unknown-account/myimage"))
							}
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

						if podman {
							expected := "podman "
							if len(tt.containerBackendGlobalExtraArgs) != 0 {
								expected += fmt.Sprintf("%s ", strings.Join(tt.containerBackendGlobalExtraArgs, " "))
							}
							expected += "play kube "
							if len(tt.containerRunExtraArgs) != 0 {
								expected += fmt.Sprintf("%s ", strings.Join(tt.containerRunExtraArgs, " "))
							}
							expected += "-"
							By("checking that extra args are passed to the podman play kube command", func() {
								Expect(string(sessionErr)).Should(ContainSubstring(expected))
							})
						}

						By("checking the endpoint accessibility", func() {
							checkEndpointAccessible([]string{"Hello world from node.js!"})
						})

						if !podman {
							By("checking the deployment was created successfully", func() {
								checkDeploymentsExist()
							})
							By("ensuring multiple deployments exist for selector error is not occurred", func() {
								Expect(string(sessionErr)).ToNot(ContainSubstring("multiple Deployments exist for the selector"))
							})
						}

						By("checking odo dev watches correctly", func() {
							// making changes to the project again
							helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "world from node.js", "from the new Node.js Starter Application")
							_, _, _, err = session.WaitSync()
							Expect(err).ToNot(HaveOccurred())
							if !podman {
								checkDeploymentsExist()
							}
							checkImageBuilt()
							checkEndpointAccessible([]string{"Hello from the new Node.js Starter Application!"})
						})

						By("cleaning up the resources on ending the session", func() {
							session.Stop()
							session.WaitEnd()
							if !podman {
								out := commonVar.CliRunner.Run("get", "deployments").Out.Contents()
								Expect(string(out)).ToNot(ContainSubstring(k8sDeploymentName))
								Expect(string(out)).ToNot(ContainSubstring(openshiftDeploymentName))
							}
						})
					})
				}))
			}
		}

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
						_, sessionOut, _, err := helper.WaitForDevModeToContain(
							helper.DevSessionOpts{
								EnvVars: env,
							},
							"failed to retrieve "+url,
							true,
							false)
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
						devfileCmpName)
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
						devfileCmpName)
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
						devfileCmpName)
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
						devfileCmpName)
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
					By("telling the user that odo is synchronizing the files", func() {
						Expect(string(stdout)).Should(ContainSubstring("Syncing files into the container"))
					})
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
						devfileCmpName)
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
					cmpName)
			})

			It("should not correctly execute PreStart commands", func() {
				args := []string{"dev", "--random-ports"}
				if podman {
					args = append(args, "--platform", "podman")
				}
				cmd := helper.Cmd("odo", args...)
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
					cmpName)
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
		for _, noWatch := range []bool{false, true} {
			podman := podman
			noWatch := noWatch
			noWatchFlag := ""
			if noWatch {
				noWatchFlag = " --no-watch"
			}
			title := fmt.Sprintf("running odo dev%s and build command throws an error", noWatchFlag)
			When(title, helper.LabelPodmanIf(podman, func() {
				var session helper.DevSession
				var stdout, stderr []byte
				BeforeEach(func() {
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						cmpName)
					helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm install", "npm install-does-not-exist")

					var err error
					session, stdout, stderr, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
						NoWatch:     noWatch,
					})
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
				})

				It("should error out with some log", func() {
					helper.MatchAllInOutput(string(stdout), []string{
						"unable to exec command",
					})
					helper.MatchAllInOutput(string(stderr), []string{
						"Usage: npm <command>",
						"Did you mean one of these?",
					})
				})
			}))
		}
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
					version := helper.GetPodmanVersion()
					if strings.HasPrefix(version, "3.") {
						Skip("Getting annotations is not available with Podman v3")
					}
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
						component := helper.NewComponent(devfileCmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
						annotations := component.GetAnnotations()
						var valueFound bool
						for key, value := range annotations {
							// Pdoman adds a suffix to the annotation key with the name of the container
							if strings.HasPrefix(key, "app.openshift.io/vcs-uri") && value == remoteURL {
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
								"Syncing files into the container",
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
						devfileCmpName)
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

							session, stdout, _, _, err := helper.StartDevMode(helper.DevSessionOpts{
								RunOnPodman: podman,
								CmdlineArgs: []string{"--build-command", "build-command-does-not-exist"},
							})
							Expect(err).ToNot(HaveOccurred())
							defer func() {
								session.Stop()
								session.WaitEnd()
							}()

							Expect(string(stdout)).To(ContainSubstring("no build command with name \"build-command-does-not-exist\" found in Devfile"))
						})

						By("calling with a command of another kind (not build)", func() {
							// devrun is a valid run command, not a build command
							session, stdout, _, _, err := helper.StartDevMode(helper.DevSessionOpts{
								RunOnPodman: podman,
								CmdlineArgs: []string{"--build-command", "devrun"},
							})
							Expect(err).ToNot(HaveOccurred())
							defer func() {
								session.Stop()
								session.WaitEnd()
							}()

							Expect(string(stdout)).To(ContainSubstring("no build command with name \"devrun\" found in Devfile"))
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
							session, stdout, _, _, err := helper.StartDevMode(helper.DevSessionOpts{
								RunOnPodman: podman,
								CmdlineArgs: []string{"--run-command", "run-command-does-not-exist"},
							})
							Expect(err).ToNot(HaveOccurred())
							defer func() {
								session.Stop()
								session.WaitEnd()
							}()

							Expect(string(stdout)).To(ContainSubstring("no run command with name \"run-command-does-not-exist\" found in Devfile"))
						})

						By("calling with a command of another kind (not run)", func() {
							// devbuild is a valid build command, not a run command
							session, stdout, _, _, err := helper.StartDevMode(helper.DevSessionOpts{
								RunOnPodman: podman,
								CmdlineArgs: []string{"--run-command", "devbuild"},
							})
							Expect(err).ToNot(HaveOccurred())
							defer func() {
								session.Stop()
								session.WaitEnd()
							}()

							Expect(string(stdout)).To(ContainSubstring("no run command with name \"devbuild\" found in Devfile"))
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
					cmpName)
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
						helper.CopyExampleDevFile(
							filepath.Join("source", "devfiles", "nodejs", "devfile-with-MR-CL-CR-modified.yaml"),
							filepath.Join(commonVar.Context, "devfile.yaml"),
							cmpName)
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
						devfileCmpName)
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

			It("should fail running a second session on the same platform", func() {
				_, _, _, err := helper.WaitForDevModeToContain(helper.DevSessionOpts{
					RunOnPodman: podman,
				}, "unable to save state file", true, true)
				Expect(err).To(HaveOccurred())
			})

			It("should create state files containing information, including forwarded ports", func() {
				var pid int
				var contentJSON []byte
				By("creating a devsate.json file", func() {
					Expect(helper.VerifyFileExists(stateFile)).To(BeTrue())
					var err error
					contentJSON, err = os.ReadFile(stateFile)
					Expect(err).ToNot(HaveOccurred())
					helper.JsonPathExist(string(contentJSON), "pid")
					platform := "cluster"
					if podman {
						platform = "podman"
					}
					helper.JsonPathContentIs(string(contentJSON), "platform", platform)
					helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.0.containerName", "runtime")
					helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.1.containerName", "runtime")
					helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.0.localAddress", "127.0.0.1")
					helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.1.localAddress", "127.0.0.1")
					helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.0.containerPort", "3000")
					helper.JsonPathContentIs(string(contentJSON), "forwardedPorts.1.containerPort", "4567")
					helper.JsonPathContentIsValidUserPort(string(contentJSON), "forwardedPorts.0.localPort")
					helper.JsonPathContentIsValidUserPort(string(contentJSON), "forwardedPorts.1.localPort")
					var content state.Content
					err = json.Unmarshal(contentJSON, &content)
					Expect(err).ShouldNot(HaveOccurred())
					pid = content.PID
					fmt.Printf("PID: %d\n", pid)
				})
				By("creating a devsate.$PID.json file with same content as devstate.json", func() {
					pidStateFile := fmt.Sprintf(".odo/devstate.%d.json", pid)
					pidContentJSON, err := os.ReadFile(pidStateFile)
					Expect(err).ToNot(HaveOccurred())
					Expect(pidContentJSON).To(Equal(contentJSON))
				})
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
					cmpName)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-parent.yaml"),
					filepath.Join(commonVar.Context, "devfile-parent.yaml"),
					cmpName+"-parent")
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
				cmpName,
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
			Expect(string(errContents)).To(ContainSubstring("Apply Kubernetes/Openshift components are not supported on Podman. Skipping: "))
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
		When("a hotReload capable Run command is used with odo dev", helper.LabelPodmanIf(podman, func() {
			var devSession helper.DevSession
			var stdout []byte
			var executeRunCommand = "Executing the application (command: dev-run)"
			var executeBuildCommand = "Building your application"
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

			It("should execute the build and run commands", func() {
				Expect(string(stdout)).To(ContainSubstring(executeBuildCommand))
				Expect(string(stdout)).To(ContainSubstring(executeRunCommand))

				By("telling the user that odo is synchronizing the files", func() {
					Expect(string(stdout)).Should(ContainSubstring("Syncing files into the container"))
				})
			})

			When("a source file is modified", func() {
				BeforeEach(func() {
					helper.ReplaceString(filepath.Join(commonVar.Context, "src", "main", "java", "org", "acme", "GreetingResource.java"), "Hello RESTEasy", "Hi RESTEasy")
					var err error
					stdout, _, _, err = devSession.WaitSync()
					Expect(err).Should(Succeed(), stdout)
				})

				It("should not re-execute the run command", func() {
					Expect(string(stdout)).To(ContainSubstring(executeBuildCommand))
					Expect(string(stdout)).ToNot(ContainSubstring(executeRunCommand))

					By("telling the user that odo is synchronizing the files", func() {
						Expect(string(stdout)).Should(ContainSubstring("Syncing files into the container"))
					})
				})
			})
		}))

		When("hotReload capable Build and Run commands are used with odo dev", helper.LabelPodmanIf(podman, func() {
			var devSession helper.DevSession
			var stdout []byte
			var executeRunCommand = "Executing the application (command: run)"
			var executeBuildCommand = "Building your application"
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "angular"), commonVar.Context)
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

			It("should execute the build and run commands", func() {
				Expect(string(stdout)).To(ContainSubstring(executeBuildCommand))
				Expect(string(stdout)).To(ContainSubstring(executeRunCommand))

				By("telling the user that odo is synchronizing the files", func() {
					Expect(string(stdout)).Should(ContainSubstring("Syncing files into the container"))
				})
			})

			When("a source file is modified", func() {
				BeforeEach(func() {
					helper.ReplaceString(filepath.Join(commonVar.Context, "src", "index.html"), "DevfileStackNodejsAngular", "Devfile Stack Nodejs Angular")
					var err error
					stdout, _, _, err = devSession.WaitSync()
					Expect(err).Should(Succeed(), stdout)
				})

				It("should not re-execute the run command", func() {
					Expect(string(stdout)).ToNot(ContainSubstring(executeBuildCommand))
					Expect(string(stdout)).ToNot(ContainSubstring(executeRunCommand))

					By("telling the user that odo is synchronizing the files", func() {
						Expect(string(stdout)).Should(ContainSubstring("Syncing files into the container"))
					})
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
					cmpName,
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
						cmpName)
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
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", ctx.devfile),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
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
				cmpName)
		})
		It("should fail to run odo dev", func() {
			errOut := helper.Cmd("odo", "dev", "--platform", "podman").WithEnv("PODMAN_CMD=echo").ShouldFail().Err()
			Expect(errOut).To(ContainSubstring("unable to access podman. Do you have podman client installed and configured correctly? cause: exec: \"echo\": executable file not found in $PATH"))
		})
	})
	Context("odo dev on podman with a devfile bound to fail", Label(helper.LabelPodman), func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "registry.access.redhat.com/ubi8/nodejs", "registry.access.redhat.com/ubi8/nose")
		})
		It("should fail with an error", func() {
			session, stdout, _, _, err := helper.StartDevMode(helper.DevSessionOpts{
				RunOnPodman: true,
			})
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				session.Stop()
				session.WaitEnd()
			}()

			helper.MatchAllInOutput(string(stdout), []string{"Complete Podman output", "registry.access.redhat.com/ubi8/nose", "Repo not found"})
		})
	})

	When("running applications listening on the container loopback interface", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-endpoint-on-loopback"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-endpoint-on-loopback.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
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
				if plt != "" {
					args = append(args, "--platform", plt)
				}
				stderr := helper.Cmd("odo", args...).ShouldFail().Err()
				Expect(stderr).Should(ContainSubstring("--ignore-localhost cannot be used when running in cluster mode"))
			})

			It("should error out if using --forward-localhost on any platform other than Podman", func() {
				args := []string{"dev", "--forward-localhost", "--random-ports"}
				if plt != "" {
					args = append(args, "--platform", plt)
				}
				stderr := helper.Cmd("odo", args...).ShouldFail().Err()
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
					ShouldFail().
					Err()
				Expect(stderr).Should(ContainSubstring("--ignore-localhost and --forward-localhost cannot be used together"))
			})

			It("should error out if not ignoring localhost", func() {
				session, _, stderr, _, err := helper.StartDevMode(helper.DevSessionOpts{
					RunOnPodman: true,
				})
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					session.Stop()
					session.WaitEnd()
				}()
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

	for _, podman := range []bool{false, true} {
		podman := podman
		// More details on https://github.com/devfile/api/issues/852#issuecomment-1211928487
		When("starting with Devfile with autoBuild or deployByDefault components", helper.LabelPodmanIf(podman, func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-autobuild-deploybydefault.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
			})

			When("running odo dev with some components not referenced in the Devfile", func() {
				var devSession helper.DevSession
				var stdout, stderr string

				BeforeEach(func() {
					var bOut, bErr []byte
					var err error
					var envvars []string
					if podman {
						envvars = append(envvars, "ODO_PUSH_IMAGES=false")
					} else {
						envvars = append(envvars, "PODMAN_CMD=echo")
					}
					devSession, bOut, bErr, _, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
						EnvVars:     envvars,
					})
					Expect(err).ShouldNot(HaveOccurred())
					stdout = string(bOut)
					stderr = string(bErr)
				})

				AfterEach(func() {
					devSession.Stop()
					if podman {
						devSession.WaitEnd()
					}
				})

				It("should create the appropriate resources", func() {
					if podman {
						k8sOcComponents := helper.ExtractK8sAndOcComponentsFromOutputOnPodman(stderr)
						By("handling Kubernetes/OpenShift components that would have been created automatically", func() {
							Expect(k8sOcComponents).Should(ContainElements(
								"k8s-deploybydefault-true-and-referenced",
								"k8s-deploybydefault-true-and-not-referenced",
								"k8s-deploybydefault-not-set-and-not-referenced",
								"ocp-deploybydefault-true-and-referenced",
								"ocp-deploybydefault-true-and-not-referenced",
								"ocp-deploybydefault-not-set-and-not-referenced",
							))
						})
						By("not handling Kubernetes/OpenShift components with deployByDefault=false", func() {
							Expect(k8sOcComponents).ShouldNot(ContainElements(
								"k8s-deploybydefault-false-and-referenced",
								"k8s-deploybydefault-false-and-not-referenced",
								"ocp-deploybydefault-false-and-referenced",
								"ocp-deploybydefault-false-and-not-referenced",
							))
						})
						By("not handling referenced Kubernetes/OpenShift components with deployByDefault unset", func() {
							Expect(k8sOcComponents).ShouldNot(ContainElement("k8s-deploybydefault-not-set-and-referenced"))
						})
					} else {
						By("automatically applying Kubernetes/OpenShift components with deployByDefault=true", func() {
							for _, l := range []string{
								"k8s-deploybydefault-true-and-referenced",
								"k8s-deploybydefault-true-and-not-referenced",
								"ocp-deploybydefault-true-and-referenced",
								"ocp-deploybydefault-true-and-not-referenced",
							} {
								Expect(stdout).Should(ContainSubstring("Creating resource Pod/%s", l))
							}
						})
						By("automatically applying non-referenced Kubernetes/OpenShift components with deployByDefault not set", func() {
							for _, l := range []string{
								"k8s-deploybydefault-not-set-and-not-referenced",
								"ocp-deploybydefault-not-set-and-not-referenced",
							} {
								Expect(stdout).Should(ContainSubstring("Creating resource Pod/%s", l))
							}
						})
						By("not applying Kubernetes/OpenShift components with deployByDefault=false", func() {
							for _, l := range []string{
								"k8s-deploybydefault-false-and-referenced",
								"k8s-deploybydefault-false-and-not-referenced",
								"ocp-deploybydefault-false-and-referenced",
								"ocp-deploybydefault-false-and-not-referenced",
							} {
								Expect(stdout).ShouldNot(ContainSubstring("Creating resource Pod/%s", l))
							}
						})
						By("not applying referenced Kubernetes/OpenShift components with deployByDefault unset", func() {
							Expect(stdout).ShouldNot(ContainSubstring("Creating resource Pod/k8s-deploybydefault-not-set-and-referenced"))
						})
					}

					imageMessagePrefix := "Building & Pushing Image"
					if podman {
						imageMessagePrefix = "Building Image"
					}

					By("automatically applying image components with autoBuild=true", func() {
						for _, tag := range []string{
							"autobuild-true-and-referenced",
							"autobuild-true-and-not-referenced",
						} {
							Expect(stdout).Should(ContainSubstring("%s: localhost:5000/odo-dev/node:%s", imageMessagePrefix, tag))
						}
					})
					By("automatically applying non-referenced Image components with autoBuild not set", func() {
						Expect(stdout).Should(ContainSubstring("%s: localhost:5000/odo-dev/node:autobuild-not-set-and-not-referenced", imageMessagePrefix))
					})
					By("not applying image components with autoBuild=false", func() {
						for _, tag := range []string{
							"autobuild-false-and-referenced",
							"autobuild-false-and-not-referenced",
						} {
							Expect(stdout).ShouldNot(ContainSubstring("localhost:5000/odo-dev/node:%s", tag))
						}
					})
					By("not applying referenced Image components with deployByDefault unset", func() {
						Expect(stdout).ShouldNot(ContainSubstring("localhost:5000/odo-dev/node:autobuild-not-set-and-referenced"))
					})
				})
			})

			When("running odo dev with some components referenced in the Devfile", func() {
				var devSession helper.DevSession
				var stdout, stderr string

				BeforeEach(func() {
					var bOut, bErr []byte
					var err error
					var envvars []string
					if podman {
						envvars = append(envvars, "ODO_PUSH_IMAGES=false")
					} else {
						envvars = append(envvars, "PODMAN_CMD=echo")
					}
					devSession, bOut, bErr, _, err = helper.StartDevMode(helper.DevSessionOpts{
						CmdlineArgs: []string{"--run-command", "run-with-referenced-components"},
						EnvVars:     envvars,
						RunOnPodman: podman,
					})
					Expect(err).ShouldNot(HaveOccurred())
					stdout = string(bOut)
					stderr = string(bErr)
				})

				AfterEach(func() {
					devSession.Stop()
					if podman {
						devSession.WaitEnd()
					}
				})

				It("should create the appropriate resources", func() {
					if podman {
						k8sOcComponents := helper.ExtractK8sAndOcComponentsFromOutputOnPodman(stderr)
						By("handling Kubernetes/OpenShift components that would have been created automatically", func() {
							Expect(k8sOcComponents).Should(ContainElements(
								"k8s-deploybydefault-true-and-referenced",
								"k8s-deploybydefault-true-and-not-referenced",
								"k8s-deploybydefault-not-set-and-not-referenced",
								"ocp-deploybydefault-true-and-referenced",
								"ocp-deploybydefault-true-and-not-referenced",
								"ocp-deploybydefault-not-set-and-not-referenced",
							))
						})

						By("handling referenced Kubernetes/OpenShift components", func() {
							Expect(k8sOcComponents).Should(ContainElements(
								"k8s-deploybydefault-true-and-referenced",
								"k8s-deploybydefault-false-and-referenced",
								"k8s-deploybydefault-not-set-and-referenced",
								"ocp-deploybydefault-true-and-referenced",
								"ocp-deploybydefault-false-and-referenced",
								"ocp-deploybydefault-not-set-and-referenced",
							))
						})

						By("not handling non-referenced Kubernetes/OpenShift components with deployByDefault=false", func() {
							Expect(k8sOcComponents).ShouldNot(ContainElements(
								"k8s-deploybydefault-false-and-not-referenced",
								"ocp-deploybydefault-false-and-not-referenced",
							))
						})
					} else {
						By("applying referenced Kubernetes/OpenShift components", func() {
							for _, l := range []string{
								"k8s-deploybydefault-true-and-referenced",
								"k8s-deploybydefault-false-and-referenced",
								"k8s-deploybydefault-not-set-and-referenced",
								"ocp-deploybydefault-true-and-referenced",
								"ocp-deploybydefault-false-and-referenced",
								"ocp-deploybydefault-not-set-and-referenced",
							} {
								Expect(stdout).Should(ContainSubstring("Creating resource Pod/%s", l))
							}
						})

						By("automatically applying Kubernetes/OpenShift components with deployByDefault=true", func() {
							for _, l := range []string{
								"k8s-deploybydefault-true-and-referenced",
								"k8s-deploybydefault-true-and-not-referenced",
								"ocp-deploybydefault-true-and-referenced",
								"ocp-deploybydefault-true-and-not-referenced",
							} {
								Expect(stdout).Should(ContainSubstring("Creating resource Pod/%s", l))
							}
						})
						By("automatically applying non-referenced Kubernetes/OpenShift components with deployByDefault not set", func() {
							for _, l := range []string{
								"k8s-deploybydefault-not-set-and-not-referenced",
								"ocp-deploybydefault-not-set-and-not-referenced",
							} {
								Expect(stdout).Should(ContainSubstring("Creating resource Pod/%s", l))
							}
						})

						By("not applying non-referenced Kubernetes/OpenShift components with deployByDefault=false", func() {
							for _, l := range []string{
								"k8s-deploybydefault-false-and-not-referenced",
								"ocp-deploybydefault-false-and-not-referenced",
							} {
								Expect(stdout).ShouldNot(ContainSubstring("Creating resource Pod/%s", l))
							}
						})
					}

					imageMessagePrefix := "Building & Pushing Image"
					if podman {
						imageMessagePrefix = "Building Image"
					}

					By("applying referenced image components", func() {
						for _, tag := range []string{
							"autobuild-true-and-referenced",
							"autobuild-false-and-referenced",
							"autobuild-not-set-and-referenced",
						} {
							Expect(stdout).Should(ContainSubstring("%s: localhost:5000/odo-dev/node:%s", imageMessagePrefix, tag))
						}
					})
					By("automatically applying image components with autoBuild=true", func() {
						for _, tag := range []string{
							"autobuild-true-and-referenced",
							"autobuild-true-and-not-referenced",
						} {
							Expect(stdout).Should(ContainSubstring("%s: localhost:5000/odo-dev/node:%s", imageMessagePrefix, tag))
						}
					})
					By("automatically applying non-referenced Image components with autoBuild not set", func() {
						Expect(stdout).Should(ContainSubstring("%s: localhost:5000/odo-dev/node:autobuild-not-set-and-not-referenced", imageMessagePrefix))
					})
					By("not applying non-referenced image components with autoBuild=false", func() {
						Expect(stdout).ShouldNot(ContainSubstring("localhost:5000/odo-dev/node:autobuild-false-and-not-referenced"))
					})
				})
			})

		}))
	}

	for _, podman := range []bool{false, true} {
		podman := podman
		Context("image names as selectors", helper.LabelPodmanIf(podman, func() {

			When("starting with a Devfile with relative and absolute image names and Kubernetes resources", func() {

				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
					helper.CopyExample(
						filepath.Join("source", "devfiles", "nodejs", "kubernetes", "devfile-image-names-as-selectors"),
						filepath.Join(commonVar.Context, "kubernetes", "devfile-image-names-as-selectors"))
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-image-names-as-selectors.yaml"),
						filepath.Join(commonVar.Context, "devfile.yaml"),
						cmpName)
				})

				When("adding a local registry for images", func() {

					const imageRegistry = "ttl.sh"

					BeforeEach(func() {
						helper.Cmd("odo", "preference", "set", "ImageRegistry", imageRegistry, "--force").ShouldPass()
					})

					AfterEach(func() {
						helper.Cmd("odo", "preference", "unset", "ImageRegistry", "--force").ShouldPass()
					})

					extractContainerNameImageMapFn := func(resourceType, resourceName, jsonPath string) map[string]string {
						result := make(map[string]string)
						data := commonVar.CliRunner.Run("-n", commonVar.Project, "get", resourceType, resourceName,
							"-o", fmt.Sprintf("jsonpath=%s", jsonPath)).Out.Contents()
						scanner := bufio.NewScanner(bytes.NewReader(data))
						for scanner.Scan() {
							l := scanner.Text()
							name, image, found := strings.Cut(l, " ")
							if !found {
								continue
							}
							result[name] = image
						}
						return result
					}

					When("running odo dev", func() {
						var devSession helper.DevSession
						var stdout string

						BeforeEach(func() {
							var env []string
							if podman {
								env = append(env, "ODO_PUSH_IMAGES=false")
							} else {
								env = append(env, "PODMAN_CMD=echo")
							}
							var outB []byte
							var err error
							devSession, outB, _, _, err = helper.StartDevMode(helper.DevSessionOpts{
								RunOnPodman: podman,
								EnvVars:     env,
							})
							Expect(err).ShouldNot(HaveOccurred())
							stdout = string(outB)
						})

						AfterEach(func() {
							devSession.Stop()
							if podman {
								devSession.WaitEnd()
							}
						})

						It("should treat relative image names as selectors", func() {
							imageMessagePrefix := "Building & Pushing Image"
							if podman {
								imageMessagePrefix = "Building Image"
							}

							lines, err := helper.ExtractLines(stdout)
							Expect(err).ShouldNot(HaveOccurred())

							var replacementImageName string
							var imagesProcessed []string
							re := regexp.MustCompile(fmt.Sprintf(`(?:%s):\s*([^\n]+)`, imageMessagePrefix))
							replaceImageRe := regexp.MustCompile(fmt.Sprintf("%s/%s-nodejs-devtools:[^\n]+", imageRegistry, cmpName))
							for _, l := range lines {
								matches := re.FindStringSubmatch(l)
								if len(matches) > 1 {
									img := matches[1]
									imagesProcessed = append(imagesProcessed, img)
									if replaceImageRe.MatchString(img) {
										replacementImageName = img
									}
								}
							}

							By("building and optionally pushing relative image components", func() {
								Expect(replacementImageName).ShouldNot(BeEmpty(), "could not find image matching regexp %v", replaceImageRe)
								Expect(imagesProcessed).Should(ContainElement(
									MatchRegexp(fmt.Sprintf("%s/%s-nodejs-devtools:[^\n]+", imageRegistry, cmpName))))
							})

							By("building and optionally pushing absolute image components with no replacement", func() {
								for _, img := range []string{"ttl.sh/odo-dev-node:1h", "ttl.sh/nodejs-devtools2:1h"} {
									Expect(imagesProcessed).Should(ContainElement(img))
								}
							})

							if !podman {
								// On Podman, `odo dev` just warns if there are any Kubernetes/OpenShift components at the moment. But this is already tested elsewhere
								// and not useful to test here.
								// But we should definitely test it if we plan on supporting more K8s resources from those components.

								type resourceData struct {
									containers     map[string]string
									initContainers map[string]string
								}

								k8sResourcesDeployed := map[string]resourceData{
									"CronJob": {
										containers: extractContainerNameImageMapFn("CronJob", "my-ocp-cron-job",
											"{range .spec.jobTemplate.spec.template.spec.containers[*]}{.name} {.image}{\"\\n\"}{end}"),
										initContainers: extractContainerNameImageMapFn("CronJob", "my-ocp-cron-job",
											"{range .spec.jobTemplate.spec.template.spec.initContainers[*]}{.name} {.image}{\"\\n\"}{end}"),
									},
									"DaemonSet": {
										containers: extractContainerNameImageMapFn("DaemonSet", "my-k8s-daemonset",
											"{range .spec.template.spec.containers[*]}{.name} {.image}{\"\\n\"}{end}"),
										initContainers: extractContainerNameImageMapFn("DaemonSet", "my-k8s-daemonset",
											"{range .spec.template.spec.initContainers[*]}{.name} {.image}{\"\\n\"}{end}"),
									},
									"Deployment": {
										containers: extractContainerNameImageMapFn("Deployment", "my-k8s-deployment",
											"{range .spec.template.spec.containers[*]}{.name} {.image}{\"\\n\"}{end}"),
										initContainers: extractContainerNameImageMapFn("Deployment", "my-k8s-deployment",
											"{range .spec.template.spec.initContainers[*]}{.name} {.image}{\"\\n\"}{end}"),
									},
									"Job": {
										containers: extractContainerNameImageMapFn("Job", "my-ocp-job",
											"{range .spec.template.spec.containers[*]}{.name} {.image}{\"\\n\"}{end}"),
										initContainers: extractContainerNameImageMapFn("Job", "my-ocp-job",
											"{range .spec.template.spec.initContainers[*]}{.name} {.image}{\"\\n\"}{end}"),
									},
									"Pod": {
										containers: extractContainerNameImageMapFn("Pod", "my-k8s-pod",
											"{range .spec.containers[*]}{.name} {.image}{\"\\n\"}{end}"),
										initContainers: extractContainerNameImageMapFn("Pod", "my-k8s-pod",
											"{range .spec.initContainers[*]}{.name} {.image}{\"\\n\"}{end}"),
									},
									"ReplicaSet": {
										containers: extractContainerNameImageMapFn("ReplicaSet", "my-k8s-replicaset",
											"{range .spec.template.spec.containers[*]}{.name} {.image}{\"\\n\"}{end}"),
										initContainers: extractContainerNameImageMapFn("ReplicaSet", "my-k8s-replicaset",
											"{range .spec.template.spec.initContainers[*]}{.name} {.image}{\"\\n\"}{end}"),
									},
									"ReplicationController": {
										containers: extractContainerNameImageMapFn("ReplicationController", "my-k8s-replicationcontroller",
											"{range .spec.template.spec.containers[*]}{.name} {.image}{\"\\n\"}{end}"),
										initContainers: extractContainerNameImageMapFn("ReplicationController", "my-k8s-replicationcontroller",
											"{range .spec.template.spec.initContainers[*]}{.name} {.image}{\"\\n\"}{end}"),
									},
									"StatefulSet": {
										containers: extractContainerNameImageMapFn("StatefulSet", "my-k8s-statefulset",
											"{range .spec.template.spec.containers[*]}{.name} {.image}{\"\\n\"}{end}"),
										initContainers: extractContainerNameImageMapFn("StatefulSet", "my-k8s-statefulset",
											"{range .spec.template.spec.initContainers[*]}{.name} {.image}{\"\\n\"}{end}"),
									},
								}

								By("replacing matching image names in core Kubernetes components", func() {
									const mainCont1 = "my-main-cont1"
									for resType, data := range k8sResourcesDeployed {
										Expect(data.containers[mainCont1]).Should(
											Equal(replacementImageName),
											func() string {
												return fmt.Sprintf(
													"unexpected image for container %q in %q deployed from K8s or OCP component. All resources: %v",
													mainCont1, resType, k8sResourcesDeployed)
											})
									}
								})

								By("not replacing non-matching or absolute image names in core Kubernetes resources", func() {
									const (
										mainCont2 = "my-main-cont2"
										initCont1 = "my-init-cont1"
										initCont2 = "my-init-cont2"
									)
									for resType, data := range k8sResourcesDeployed {
										Expect(data.containers[mainCont2]).Should(
											Equal("ttl.sh/nodejs-devtools2:1h"),
											func() string {
												return fmt.Sprintf(
													"unexpected image for container %q in %q deployed from K8s or OCP component. All resources: %v",
													mainCont2, resType, k8sResourcesDeployed)
											})
										Expect(data.initContainers[initCont1]).Should(
											Equal("ttl.sh/odo-dev-node:1h"),
											func() string {
												return fmt.Sprintf(
													"unexpected image for init container %q in %q deployed from K8s or OCP component. All resources: %v",
													initCont1, resType, k8sResourcesDeployed)
											})
										Expect(data.initContainers[initCont2]).Should(
											Equal("nodejs-devtools007"),
											func() string {
												return fmt.Sprintf(
													"unexpected image for init container %q in %q deployed from K8s or OCP component. All resources: %v",
													initCont1, resType, k8sResourcesDeployed)
											})
									}
								})
							}
						})
					})

				})

			})
		}))

	}
})
