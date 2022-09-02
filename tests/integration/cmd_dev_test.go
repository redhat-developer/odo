package integration

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
		commonVar = helper.CommonBeforeEach(helper.SetupClusterTrue)
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
		It("should show validation errors if the devfile is incorrect", func() {
			err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "kind: run", "kind: build")
				helper.WaitForOutputToContain("Error occurred on Push", 180, 10, session)
			})
			Expect(err).ToNot(HaveOccurred())
		})
		It("should use the index information from previous push operation", func() {
			// Create a new file A
			fileAPath, fileAText := helper.CreateSimpleFile(commonVar.Context, "my-file-", ".txt")
			// watch that project
			err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
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
			err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents, errContents []byte, ports map[string]string) {
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

		When("recording telemetry data", func() {
			BeforeEach(func() {
				helper.EnableTelemetryDebug()
				session, _, _, _, _ := helper.StartDevMode(nil)
				session.Stop()
				session.WaitEnd()
			})
			AfterEach(func() {
				helper.ResetTelemetry()
			})
			It("should record the telemetry data correctly", func() {
				td := helper.GetTelemetryDebugData()
				Expect(td.Event).To(ContainSubstring("odo dev"))
				Expect(td.Properties.Success).To(BeFalse())
				Expect(td.Properties.Error).ToNot(ContainSubstring("user interrupted"))
				Expect(td.Properties.CmdProperties[segment.ComponentType]).To(ContainSubstring("nodejs"))
				Expect(td.Properties.CmdProperties[segment.Language]).To(ContainSubstring("nodejs"))
				Expect(td.Properties.CmdProperties[segment.ProjectType]).To(ContainSubstring("nodejs"))
			})
		})

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
					devSession, _, _, _, err = helper.StartDevMode(nil)
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

		When("odo dev is executed", func() {

			var devSession helper.DevSession

			BeforeEach(func() {
				var err error
				devSession, _, _, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				devSession.Kill()
				devSession.WaitEnd()
			})

			When("odo dev is stopped", func() {
				BeforeEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				It("should delete component from the cluster", func() {
					deploymentName := fmt.Sprintf("%s-%s", cmpName, "app")
					errout := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Err.Contents()
					Expect(string(errout)).ToNot(ContainSubstring(deploymentName))
				})
			})
		})

		When("odo dev is executed and Ephemeral is set to false", func() {

			var devSession helper.DevSession
			BeforeEach(func() {
				helper.Cmd("odo", "preference", "set", "-f", "Ephemeral", "false").ShouldPass()
				var err error
				devSession, _, _, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
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

		When("odo is executed with --no-watch flag", func() {

			var devSession helper.DevSession

			BeforeEach(func() {
				var err error
				devSession, _, _, _, err = helper.StartDevMode(nil, "--no-watch")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				devSession.Kill()
				devSession.WaitEnd()
			})

			When("a file in component directory is modified", func() {

				BeforeEach(func() {
					helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "App started", "App is super started")
				})

				It("should not trigger a push", func() {
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
					execResult := commonVar.CliRunner.Exec(podName, commonVar.Project, "cat", "/projects/server.js")
					Expect(execResult).To(ContainSubstring("App started"))
					Expect(execResult).ToNot(ContainSubstring("App is super started"))

				})

				When("p is pressed", func() {

					BeforeEach(func() {
						devSession.PressKey('p')
					})

					It("should trigger a push", func() {
						_, _, _, err := devSession.WaitSync()
						Expect(err).ToNot(HaveOccurred())
						podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
						execResult := commonVar.CliRunner.Exec(podName, commonVar.Project, "cat", "/projects/server.js")
						Expect(execResult).To(ContainSubstring("App is super started"))
					})
				})
			})
		})

		When("a delay is necessary for the component to start and running odo dev", func() {

			var devSession helper.DevSession
			var ports map[string]string

			BeforeEach(func() {
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm start", "sleep 20 ; npm start")

				var err error
				devSession, _, _, ports, err = helper.StartDevMode(nil)
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

	for _, manual := range []bool{false, true} {
		manual := manual
		Context("port-forwarding for the component", func() {
			When("devfile has single endpoint", func() {
				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
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
						devSession, _, _, ports, err = helper.StartDevMode(nil, opts...)
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
						BeforeEach(func() {
							src := "memoryLimit: 1024Mi"
							dst := "memoryLimit: 1023Mi"
							helper.ReplaceString("devfile.yaml", src, dst)
							if manual {
								devSession.PressKey('p')
							}
							var err error
							_, _, ports, err = devSession.WaitSync()
							Expect(err).Should(Succeed())
						})

						It("should expose the endpoint on localhost", func() {
							By("updating the pod", func() {
								podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
								bufferOutput := commonVar.CliRunner.Run("get", "pods", podName, "-o", "jsonpath='{.spec.containers[0].resources.requests.memory}'").Out.Contents()
								output := string(bufferOutput)
								Expect(output).To(ContainSubstring("1023Mi"))
							})

							By("exposing the endpoint", func() {
								url := fmt.Sprintf("http://%s", ports["3000"])
								resp, err := http.Get(url)
								Expect(err).ToNot(HaveOccurred())
								defer resp.Body.Close()

								body, _ := io.ReadAll(resp.Body)
								helper.MatchAllInOutput(string(body), []string{"Hello from Node.js Starter Application!"})
								Expect(err).ToNot(HaveOccurred())
							})
						})
					})
				})
			})

			When("devfile has multiple endpoints", func() {
				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-multiple-endpoints"), commonVar.Context)
					helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
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
						devSession, _, _, ports, err = helper.StartDevMode(nil, opts...)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})

					It("should expose two endpoints on localhost", func() {
						url1 := fmt.Sprintf("http://%s", ports["3000"])
						url2 := fmt.Sprintf("http://%s", ports["4567"])

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

						if manual {
							devSession.PressKey('p')
						}

						_, _, _, err = devSession.WaitSync()
						Expect(err).Should(Succeed())

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

					When("an endpoint is added after first run of odo dev", func() {

						BeforeEach(func() {
							helper.ReplaceString("devfile.yaml", "exposure: none", "exposure: public")

							if manual {
								devSession.PressKey('p')
							}

							var err error
							_, _, ports, err = devSession.WaitSync()
							Expect(err).Should(Succeed())

						})
						It("should expose three endpoints on localhost", func() {
							url1 := fmt.Sprintf("http://%s", ports["3000"])
							url2 := fmt.Sprintf("http://%s", ports["4567"])
							url3 := fmt.Sprintf("http://%s", ports["7890"])

							resp1, err := http.Get(url1)
							Expect(err).ToNot(HaveOccurred())
							defer resp1.Body.Close()

							resp2, err := http.Get(url2)
							Expect(err).ToNot(HaveOccurred())
							defer resp2.Body.Close()

							resp3, err := http.Get(url3)
							Expect(err).ToNot(HaveOccurred())
							defer resp3.Body.Close()

							body1, _ := io.ReadAll(resp1.Body)
							helper.MatchAllInOutput(string(body1), []string{"Hello from Node.js Starter Application!"})

							body2, _ := io.ReadAll(resp2.Body)
							helper.MatchAllInOutput(string(body2), []string{"Hello from Node.js Starter Application!"})

							body3, _ := io.ReadAll(resp3.Body)
							helper.MatchAllInOutput(string(body3), []string{"Hello from Node.js Starter Application!"})
						})
					})
				})

			})
		})
	}

	for _, devfileHandlerCtx := range []struct {
		name           string
		cmpName        string
		devfileHandler func(path string)
	}{
		{
			name: "with metadata.name",
			// cmpName from Devfile
			cmpName: "nodejs",
		},
		{
			name: "without metadata.name",
			// cmpName is returned by alizer.DetectName
			cmpName: "nodejs-starter",
			devfileHandler: func(path string) {
				helper.UpdateDevfileContent(path, []helper.DevfileUpdater{helper.DevfileMetadataNameRemover})
			},
		},
	} {
		devfileHandlerCtx := devfileHandlerCtx
		When("Devfile 2.1.0 is used - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-variables.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
			})

			When("doing odo dev", func() {
				var session helper.DevSession
				BeforeEach(func() {
					var err error
					session, _, _, _, err = helper.StartDevMode(nil)
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
				})

				It("should check if the env variable has a correct value", func() {
					envVars := commonVar.CliRunner.GetEnvsDevFileDeployment(devfileCmpName, "app", commonVar.Project)
					// check if the env variable has a correct value. This value was substituted from in devfile from variable
					Expect(envVars["FOO"]).To(Equal("bar"))
				})
			})

			When("doing odo dev with --var flag", func() {
				var session helper.DevSession
				BeforeEach(func() {
					var err error
					session, _, _, _, err = helper.StartDevMode(nil, "--var", "VALUE_TEST=baz")
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
				})

				It("should check if the env variable has a correct value", func() {
					envVars := commonVar.CliRunner.GetEnvsDevFileDeployment(devfileCmpName, "app", commonVar.Project)
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
					session, _, _, _, err = helper.StartDevMode(nil, "--var-file", "vars.txt")
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
					helper.DeleteFile(varfilename)
				})

				It("should check if the env variable has a correct value", func() {
					envVars := commonVar.CliRunner.GetEnvsDevFileDeployment(devfileCmpName, "app", commonVar.Project)
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
					session, _, _, _, err = helper.StartDevMode(nil, "--var-file", "vars.txt")
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					session.Stop()
					session.WaitEnd()
					helper.DeleteFile(varfilename)
					_ = os.Unsetenv("VALUE_TEST")
				})

				It("should check if the env variable has a correct value", func() {
					envVars := commonVar.CliRunner.GetEnvsDevFileDeployment(devfileCmpName, "app", commonVar.Project)
					// check if the env variable has a correct value. This value was substituted from in devfile from variable
					Expect(envVars["FOO"]).To(Equal("baz"))
				})
			})
		})

		When("running odo dev and single env var is set - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-single-env.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
			})

			It("should be able to exec command", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, out, err []byte, ports map[string]string) {
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
					output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
					helper.MatchAllInOutput(output, []string{"test_env_variable", "test_build_env_variable"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("running odo dev and multiple env variables are set - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-multiple-envs.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
			})

			It("should be able to exec command", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, out, err []byte, ports map[string]string) {
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
					output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
					helper.MatchAllInOutput(output, []string{"test_build_env_variable1", "test_build_env_variable2", "test_env_variable1", "test_env_variable2"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("doing odo dev and there is a env variable with spaces - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-command-env-with-space.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
			})

			It("should be able to exec command", func() {
				err := helper.RunDevMode(nil, nil, func(session *gexec.Session, out, err []byte, ports map[string]string) {
					podName := commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project)
					output := commonVar.CliRunner.ExecListDir(podName, commonVar.Project, "/projects")
					helper.MatchAllInOutput(output, []string{"build env variable with space", "env with space"})
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("creating local files and dir and running odo dev - "+devfileHandlerCtx.name, func() {
			var newDirPath, newFilePath, stdOut, podName string
			var session helper.DevSession
			var devfileCmpName string
			BeforeEach(func() {
				newFilePath = filepath.Join(commonVar.Context, "foobar.txt")
				newDirPath = filepath.Join(commonVar.Context, "testdir")
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				// Create a new file that we plan on deleting later...
				if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
					fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
				}
				// Create a new directory
				helper.MakeDir(newDirPath)
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
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
			})

			When("creating local files and dir and running odo dev - "+devfileHandlerCtx.name, func() {
				var newDirPath, newFilePath, stdOut, podName string
				var session helper.DevSession
				var devfileCmpName string
				BeforeEach(func() {
					newFilePath = filepath.Join(commonVar.Context, "foobar.txt")
					newDirPath = filepath.Join(commonVar.Context, "testdir")
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-service-binding-files.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
					devfileCmpName = devfileHandlerCtx.cmpName
					if devfileHandlerCtx.devfileHandler != nil {
						devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
					}
					// Create a new file that we plan on deleting later...
					if err := helper.CreateFileWithContent(newFilePath, "hello world"); err != nil {
						fmt.Printf("the foobar.txt file was not created, reason %v", err.Error())
					}
					// Create a new directory
					helper.MakeDir(newDirPath)
					var err error
					session, _, _, _, err = helper.StartDevMode(nil)
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
				gitignorePath = filepath.Join(commonVar.Context, ".gitignore")
				newFilePath1 = filepath.Join(commonVar.Context, "foobar.txt")
				newDirPath = filepath.Join(commonVar.Context, "testdir")
				newFilePath2 = filepath.Join(newDirPath, "foobar.txt")
				newFilePath3 = filepath.Join(newDirPath, "baz.txt")
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
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
				session, _, _, _, err = helper.StartDevMode(nil)
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
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileSourceMapping.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
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
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				// devfile with clonePath set in project field
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}

				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
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
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}

				// reset clonePath and change the workdir accordingly, it should sync to project name
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "clonePath: webapp/", "# clonePath: webapp/")
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
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
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-multiple-projects.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))

				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}

				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
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
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
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
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volumes.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
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
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-volume-components.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
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

	Describe("devfile contains composite apply command", func() {
		const (
			deploymentName = "my-component"
			DEVFILEPORT    = "3000"
		)
		var session helper.DevSession
		var sessionOut, sessionErr []byte
		var err error
		var ports map[string]string
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-composite-apply-commands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
		})
		When("odo dev is running", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				session, sessionOut, sessionErr, ports, err = helper.StartDevMode([]string{"PODMAN_CMD=echo"})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})
			It("should execute the composite apply commands successfully", func() {
				checkDeploymentExists := func() {
					out := commonVar.CliRunner.Run("get", "deployments", deploymentName).Out.Contents()
					Expect(out).To(ContainSubstring(deploymentName))
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
					checkDeploymentExists()
				})
				By("ensuring multiple deployments exist for selector error is not occurred", func() {
					Expect(string(sessionErr)).ToNot(ContainSubstring("multiple Deployments exist for the selector"))
				})
				By("checking odo dev watches correctly", func() {
					// making changes to the project again
					helper.ReplaceString(filepath.Join(commonVar.Context, "server.js"), "from Node.js Starter Application", "from the new Node.js Starter Application")
					_, _, _, err = session.WaitSync()
					Expect(err).ToNot(HaveOccurred())
					checkDeploymentExists()
					checkImageBuilt()
					checkEndpointAccessible([]string{"Hello from the new Node.js Starter Application!"})
				})

				By("cleaning up the resources on ending the session", func() {
					session.Stop()
					session.WaitEnd()
					out := commonVar.CliRunner.Run("get", "deployments").Out.Contents()
					Expect(out).ToNot(ContainSubstring(deploymentName))
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
						session, sessionOut, _, ports, err = helper.StartDevMode(env)
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
						_, sessionOut, _, err := helper.DevModeShouldFail(env, "failed to retrieve "+url)
						Expect(err).To(BeNil())
						Expect(sessionOut).NotTo(ContainSubstring("build -t quay.io/unknown-account/myimage -f "))
						Expect(sessionOut).NotTo(ContainSubstring("push quay.io/unknown-account/myimage"))
					})
				})
			}
		})
	})

	for _, devfileHandlerCtx := range []struct {
		name           string
		cmpName        string
		devfileHandler func(path string)
	}{
		{
			name: "with metadata.name",
			// cmpName from Devfile
			cmpName: "nodejs",
		},
		{
			name: "without metadata.name",
			// cmpName is returned by alizer.DetectName
			cmpName: "nodejs-starter",
			devfileHandler: func(path string) {
				helper.UpdateDevfileContent(path, []helper.DevfileUpdater{helper.DevfileMetadataNameRemover})
			},
		},
	} {
		devfileHandlerCtx := devfileHandlerCtx
		When("running odo dev and devfile with composite command - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should execute all commands in composite command", func() {
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

		When("running odo dev and composite command is marked as parallel:true - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeCommandsParallel.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should execute all commands in composite command", func() {
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

		When("running odo dev and composite command are nested - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			var session helper.DevSession
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileNestedCompCommands.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, _, _, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
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

		When("running odo dev and composite command is used as a run command - "+devfileHandlerCtx.name, func() {
			var session helper.DevSession
			var stdout []byte
			var stderr []byte
			var devfileCmpName string
			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfileCompositeRunAndDebug.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, stdout, stderr, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should run successfully", func() {
				By("verifying from the output that all commands have been executed", func() {
					helper.MatchAllInOutput(string(stdout), []string{
						"Building your application in container on cluster",
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
					// Because of the Spinner, the "Building your application in container on cluster" is printed twice in the captured stdout.
					// The bracket allows to match the last occurrence with the command execution timing information.
					Expect(strings.Count(string(stdout), "Building your application in container on cluster (command: install) [")).
						To(BeNumerically("==", 1), "\nOUTPUT: "+string(stdout)+"\n")
				})

				By("verifying that the command did run successfully", func() {
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
			})
		})

		When("running build and run commands as composite in different containers and a shared volume - "+devfileHandlerCtx.name, func() {
			var session helper.DevSession
			var stdout []byte
			var stderr []byte
			var devfileCmpName string
			BeforeEach(func() {
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfileCompositeBuildRunDebugInMultiContainersAndSharedVolume.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
				var err error
				session, stdout, stderr, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				session.Stop()
				session.WaitEnd()
			})

			It("should run successfully", func() {
				By("verifying from the output that all commands have been executed", func() {
					helper.MatchAllInOutput(string(stdout), []string{
						"Building your application in container on cluster (command: mkdir)",
						"Building your application in container on cluster (command: sleep-cmd-build)",
						"Building your application in container on cluster (command: build-cmd)",
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
					// Because of the Spinner, the "Building your application in container on cluster" is printed twice in the captured stdout.
					// The bracket allows to match the last occurrence with the command execution timing information.
					out := string(stdout)
					for _, cmd := range []string{"mkdir", "sleep-cmd-build", "build-cmd"} {
						Expect(strings.Count(out, fmt.Sprintf("Building your application in container on cluster (command: %s) [", cmd))).
							To(BeNumerically("==", 1), "\nOUTPUT: "+string(stdout)+"\n")
					}
				})

				By("verifying that the command did run successfully", func() {
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
			})
		})
	}

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
			session, _, initErr, _, err = helper.StartDevMode(nil)
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
	})

	When("running odo dev --no-watch and build command throws an error", func() {
		var stderr string
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), "npm install", "npm install-does-not-exist")
			stderr = helper.Cmd("odo", "dev", "--no-watch", "--random-ports").ShouldFail().Err()
		})

		It("should error out with some log", func() {
			helper.MatchAllInOutput(stderr, []string{
				"unable to exec command",
				"Usage: npm <command>",
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
			session, _, _, _, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
			session.WaitEnd()
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
			err := helper.RunDevMode(nil, nil, func(session *gexec.Session, outContents []byte, errContents []byte, ports map[string]string) {
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

	for _, devfileHandlerCtx := range []struct {
		name           string
		cmpName        string
		devfileHandler func(path string)
	}{
		{
			name: "with metadata.name",
			// cmpName from Devfile
			cmpName: "nodejs",
		},
		{
			name: "without metadata.name",
			// cmpName is returned by alizer.DetectName
			cmpName: "nodejs-starter",
			devfileHandler: func(path string) {
				helper.UpdateDevfileContent(path, []helper.DevfileUpdater{helper.DevfileMetadataNameRemover})
			},
		},
	} {
		devfileHandlerCtx := devfileHandlerCtx
		When("running odo dev with alternative commands - "+devfileHandlerCtx.name, func() {

			type testCase struct {
				buildCmd          string
				runCmd            string
				devAdditionalOpts []string
				checkFunc         func(stdout, stderr string)
			}
			testForCmd := func(tt testCase) {
				err := helper.RunDevMode(tt.devAdditionalOpts, nil, func(session *gexec.Session, outContents []byte, errContents []byte, ports map[string]string) {
					stdout := string(outContents)
					stderr := string(errContents)

					By("checking the output of the command", func() {
						helper.MatchAllInOutput(stdout, []string{
							fmt.Sprintf("Building your application in container on cluster (command: %s)", tt.buildCmd),
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

			remoteFileChecker := func(path string) bool {
				return commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
					commonVar.CliRunner.GetRunningPodNameByComponent(devfileHandlerCtx.cmpName, commonVar.Project),
					"runtime",
					commonVar.Project,
					[]string{"stat", path},
					func(cmdOp string, err error) bool {
						return err == nil
					},
				)
			}

			BeforeEach(func() {
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-with-alternative-commands.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"))
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
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

				It("should error out if called with an invalid command", func() {
					output := helper.Cmd("odo", "dev", "--random-ports", "--build-command", "build-command-does-not-exist").ShouldFail().Err()
					Expect(output).To(ContainSubstring("no build command with name \"build-command-does-not-exist\" found in Devfile"))
				})

				It("should error out if called with a command of another kind", func() {
					// devrun is a valid run command, not a build command
					output := helper.Cmd("odo", "dev", "--random-ports", "--build-command", "devrun").ShouldFail().Err()
					Expect(output).To(ContainSubstring("no build command with name \"devrun\" found in Devfile"))
				})

				It("should execute the custom non-default build command successfully", func() {
					buildCmdTestFunc("my-custom-build", func(stdout, stderr string) {
						By("checking that it did not execute the default build command", func() {
							helper.DontMatchAllInOutput(stdout, []string{
								"Building your application in container on cluster (command: devbuild)",
							})
						})

						By("verifying that the custom command ran successfully", func() {
							Expect(remoteFileChecker("/projects/file-from-my-custom-build")).To(BeTrue())
						})
					})
				})

				It("should execute the default build command successfully if specified explicitly", func() {
					// devbuild is the default build command
					buildCmdTestFunc("devbuild", func(stdout, stderr string) {
						By("checking that it did not execute the custom build command", func() {
							helper.DontMatchAllInOutput(stdout, []string{
								"Building your application in container on cluster (command: my-custom-build)",
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

				It("should error out if called with an invalid command", func() {
					output := helper.Cmd("odo", "dev", "--random-ports", "--run-command", "run-command-does-not-exist").ShouldFail().Err()
					Expect(output).To(ContainSubstring("no run command with name \"run-command-does-not-exist\" found in Devfile"))
				})

				It("should error out if called with a command of another kind", func() {
					// devbuild is a valid build command, not a run command
					output := helper.Cmd("odo", "dev", "--random-ports", "--run-command", "devbuild").ShouldFail().Err()
					Expect(output).To(ContainSubstring("no run command with name \"devbuild\" found in Devfile"))
				})

				It("should execute the custom non-default run command successfully", func() {
					runCmdTestFunc("my-custom-run", func(stdout, stderr string) {
						By("checking that it did not execute the default run command", func() {
							helper.DontMatchAllInOutput(stdout, []string{
								"Executing the application (command: devrun)",
							})
						})

						By("verifying that the custom command ran successfully", func() {
							Expect(remoteFileChecker("/projects/file-from-my-custom-run")).To(BeTrue())
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
									"Building your application in container on cluster (command: devbuild)",
									"Executing the application (command: devrun)",
								})
							})

							By("verifying that the custom build command ran successfully", func() {
								Expect(remoteFileChecker("/projects/file-from-my-custom-build")).To(BeTrue())
							})

							By("verifying that the custom run command ran successfully", func() {
								Expect(remoteFileChecker("/projects/file-from-my-custom-run")).To(BeTrue())
							})
						},
					},
				)
			})
		})
	}

	// Tests https://github.com/redhat-developer/odo/issues/3838
	When("java-springboot application is created and running odo dev", func() {
		var session helper.DevSession
		BeforeEach(func() {
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "springboot", "devfile-registry.yaml")).ShouldPass()
			helper.CopyExample(filepath.Join("source", "devfiles", "springboot", "project"), commonVar.Context)
			var err error
			session, _, _, _, err = helper.StartDevMode(nil, "-v", "4")
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
				podName := commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project)
				podLogs := commonVar.CliRunner.Run("-n", commonVar.Project, "logs", podName).Out.Contents()
				Expect(string(podLogs)).To(ContainSubstring("BUILD SUCCESS"))
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
			session, _, _, _, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
			session.WaitEnd()
		})

		It("should check cpuLimit, cpuRequests, memoryRequests", func() {
			ensureResource("1", "200m", "512Mi")
		})

		When("Update the devfile.yaml, and waiting synchronization", func() {

			BeforeEach(func() {
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-MR-CL-CR-modified.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
				var err error
				_, _, _, err = session.WaitSync()
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
			session, _, _, _, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			session.Stop()
			session.WaitEnd()
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

			session, _, errContents, _, err := helper.StartDevMode(nil)
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

			session, _, errContents, err := helper.StartDevMode(nil)
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
			devSession, _, _, _, err = helper.StartDevMode(nil)
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
		name           string
		cmpName        string
		devfileHandler func(path string)
	}{
		{
			name: "with metadata.name",
			// cmpName from Devfile
			cmpName: "nodejs",
		},
		{
			name: "without metadata.name",
			// cmpName is returned by alizer.DetectName
			cmpName: "nodejs-starter",
			devfileHandler: func(path string) {
				helper.UpdateDevfileContent(path, []helper.DevfileUpdater{helper.DevfileMetadataNameRemover})
			},
		},
	} {
		devfileHandlerCtx := devfileHandlerCtx
		When("a container component defines a Command or Args - "+devfileHandlerCtx.name, func() {
			var devfileCmpName string
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "issue-5620-devfile-with-container-command-args.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"))
				devfileCmpName = devfileHandlerCtx.cmpName
				if devfileHandlerCtx.devfileHandler != nil {
					devfileHandlerCtx.devfileHandler(filepath.Join(commonVar.Context, "devfile.yaml"))
				}
			})

			It("should run odo dev successfully (#5620)", func() {
				devSession, stdoutBytes, stderrBytes, _, err := helper.StartDevMode(nil)
				Expect(err).ShouldNot(HaveOccurred())
				defer devSession.Stop()
				const errorMessage = "Failed to create the component:"
				helper.DontMatchAllInOutput(string(stdoutBytes), []string{errorMessage})
				helper.DontMatchAllInOutput(string(stderrBytes), []string{errorMessage})

				// the command has been started directly in the background. Check the PID stored in a specific file.
				commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
					commonVar.CliRunner.GetRunningPodNameByComponent(devfileCmpName, commonVar.Project),
					"runtime",
					commonVar.Project,
					[]string{
						remotecmd.ShellExecutable, "-c",
						fmt.Sprintf("kill -0 $(cat %s/.odo_cmd_run.pid) 2>/dev/null ; echo -n $?",
							strings.TrimSuffix(storage.SharedDataMountPath, "/")),
					},
					func(stdout string, err error) bool {
						Expect(err).ShouldNot(HaveOccurred())
						Expect(stdout).To(Equal("0"))
						return err == nil
					})
			})
		})
	}

	When("a component with multiple endpoints is run", func() {
		stateFile := ".odo/devstate.json"
		var devSession helper.DevSession
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project-with-multiple-endpoints"), commonVar.Context)
			helper.Cmd("odo", "set", "project", commonVar.Project).ShouldPass()
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-with-multiple-endpoints.yaml")).ShouldPass()
			Expect(helper.VerifyFileExists(".odo/devstate.json")).To(BeFalse())
			var err error
			devSession, _, _, _, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			// We stop the process so the process does not remain after the end of the tests
			devSession.Kill()
			devSession.WaitEnd()
		})

		It("should create a state file containing forwarded ports", func() {
			Expect(helper.VerifyFileExists(stateFile)).To(BeTrue())
			contentJSON, err := ioutil.ReadFile(stateFile)
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

		When("odo dev is stopped", func() {
			BeforeEach(func() {
				devSession.Stop()
				devSession.WaitEnd()
			})

			It("should remove forwarded ports from state file", func() {
				Expect(helper.VerifyFileExists(stateFile)).To(BeTrue())
				contentJSON, err := ioutil.ReadFile(stateFile)
				Expect(err).ToNot(HaveOccurred())
				helper.JsonPathContentIs(string(contentJSON), "forwardedPorts", "")
			})
		})
	})

	When("a devfile with a local parent is used for odo dev and the parent is not synced", func() {
		var devSession helper.DevSession
		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-child.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"))
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-parent.yaml"), filepath.Join(commonVar.Context, "devfile-parent.yaml"))
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			var err error
			devSession, _, _, _, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())

			gitignorePath := filepath.Join(commonVar.Context, ".gitignore")
			err = helper.AppendToFile(gitignorePath, "\n/devfile-parent.yaml\n")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			// We stop the process so the process does not remain after the end of the tests
			devSession.Kill()
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
	})

	When("a hotReload capable project is used with odo dev", func() {
		var devSession helper.DevSession
		var stdout []byte
		var executeRunCommand = "Executing the application (command: dev-run)"
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "java-quarkus"), commonVar.Context)
			var err error
			devSession, stdout, _, _, err = helper.StartDevMode(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			// We stop the process so the process does not remain after the end of the tests
			devSession.Kill()
			devSession.WaitEnd()
		})

		It("should execute the run command", func() {
			Expect(stdout).To(ContainSubstring(executeRunCommand))
		})

		When("a source file is modified", func() {
			BeforeEach(func() {
				helper.ReplaceString(filepath.Join(commonVar.Context, "src", "main", "java", "org", "acme", "GreetingResource.java"), "Hello RESTEasy", "Hi RESTEasy")
				var err error
				stdout, _, _, err = devSession.WaitSync()
				Expect(err).Should(Succeed(), stdout)
			})

			It("should not re-execute the run command", func() {
				Expect(stdout).ToNot(ContainSubstring(executeRunCommand))
			})
		})
	})

	Describe("Devfile with no metadata.name", func() {

		BeforeEach(func() {
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-no-metadata-name.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"))
		})

		When("running odo dev against a component with no source code", func() {
			var devSession helper.DevSession
			BeforeEach(func() {
				var err error
				devSession, _, _, _, err = helper.StartDevMode(nil)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				devSession.Stop()
			})

			It("should use the directory as component name", func() {
				// when no further source code is available, directory name is returned by alizer.DetectName as component name;
				// and since it is all-numeric in our tests, an "x" prefix is added by util.GetDNS1123Name (called by alizer.DetectName)
				cmpName := "x" + filepath.Base(commonVar.Context)
				commonVar.CliRunner.CheckCmdOpInRemoteDevfilePod(
					commonVar.CliRunner.GetRunningPodNameByComponent(cmpName, commonVar.Project),
					"runtime",
					commonVar.Project,
					[]string{
						remotecmd.ShellExecutable, "-c",
						fmt.Sprintf("cat %s/.odo_cmd_devrun.pid", strings.TrimSuffix(storage.SharedDataMountPath, "/")),
					},
					func(stdout string, err error) bool {
						Expect(err).ShouldNot(HaveOccurred())
						Expect(stdout).NotTo(BeEmpty())
						return err == nil
					})
			})
		})
	})
})
