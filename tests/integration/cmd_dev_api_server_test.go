package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/tests/helper"
	"k8s.io/utils/pointer"
)

var _ = Describe("odo dev command with api server tests", func() {
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
		for _, customPort := range []bool{false, true} {
			customPort := customPort
			When("the component is bootstrapped", helper.LabelPodmanIf(podman, func() {
				BeforeEach(func() {
					helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
					helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"), cmpName)
				})
				When(fmt.Sprintf("odo dev is run with --api-server flag (custom api server port=%v)", customPort), func() {
					var (
						devSession helper.DevSession
						localPort  = helper.GetCustomStartPort()
					)
					BeforeEach(func() {
						opts := helper.DevSessionOpts{
							RunOnPodman:    podman,
							StartAPIServer: true,
							EnvVars:        []string{"ODO_EXPERIMENTAL_MODE=true"},
						}
						if customPort {
							opts.APIServerPort = localPort
						}
						var err error
						devSession, err = helper.StartDevMode(opts)
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})
					It("should start the Dev server when --api-server flag is passed", func() {
						if customPort {
							Expect(devSession.APIServerEndpoint).To(ContainSubstring(fmt.Sprintf("%d", localPort)))
						}
						url := fmt.Sprintf("http://%s/instance", devSession.APIServerEndpoint)
						resp, err := http.Get(url)
						Expect(err).ToNot(HaveOccurred())
						Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
					})
				})
			}))
		}

		When("the component is bootstrapped", helper.LabelPodmanIf(podman, func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile.yaml"), filepath.Join(commonVar.Context, "devfile.yaml"), cmpName)
			})
			When("odo dev is run with --api-server flag", func() {
				var (
					devSession helper.DevSession
				)
				BeforeEach(func() {
					opts := helper.DevSessionOpts{
						RunOnPodman:    podman,
						StartAPIServer: true,
						EnvVars:        []string{"ODO_EXPERIMENTAL_MODE=true"},
					}
					var err error
					devSession, err = helper.StartDevMode(opts)
					Expect(err).ToNot(HaveOccurred())
				})
				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})
				It("should serve endpoints", func() {
					By("GETting /instance", func() {
						url := fmt.Sprintf("http://%s/instance", devSession.APIServerEndpoint)
						resp, err := http.Get(url)
						Expect(err).ToNot(HaveOccurred())
						Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
						defer resp.Body.Close()
						body, err := io.ReadAll(resp.Body)
						Expect(err).ToNot(HaveOccurred())
						strBody := string(body)
						helper.JsonPathExist(strBody, "pid")
						helper.JsonPathContentIs(strBody, "componentDirectory", commonVar.Context)
					})
					By("GETting /component", func() {
						url := fmt.Sprintf("http://%s/component", devSession.APIServerEndpoint)
						resp, err := http.Get(url)
						Expect(err).ToNot(HaveOccurred())
						Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
						defer resp.Body.Close()
						body, err := io.ReadAll(resp.Body)
						Expect(err).ToNot(HaveOccurred())
						strBody := string(body)
						helper.JsonPathContentIs(strBody, "devfilePath", filepath.Join(commonVar.Context, "devfile.yaml"))
						helper.JsonPathContentIs(strBody, "devfileData.devfile.metadata.name", cmpName)
						helper.JsonPathContentIs(strBody, "devfileData.supportedOdoFeatures.dev", "true")
						helper.JsonPathContentIs(strBody, "devfileData.supportedOdoFeatures.deploy", "false")
						helper.JsonPathContentIs(strBody, "devfileData.supportedOdoFeatures.debug", "false")
						helper.JsonPathContentIs(strBody, "managedBy", "odo")
						if podman {
							helper.JsonPathDoesNotExist(strBody, "runningOn.cluster")
							helper.JsonPathExist(strBody, "runningOn.podman")
							helper.JsonPathContentIs(strBody, "runningOn.podman.dev", "true")
							helper.JsonPathContentIs(strBody, "runningOn.podman.deploy", "false")
						} else {
							helper.JsonPathDoesNotExist(strBody, "runningOn.podman")
							helper.JsonPathExist(strBody, "runningOn.cluster")
							helper.JsonPathContentIs(strBody, "runningOn.cluster.dev", "true")
							helper.JsonPathContentIs(strBody, "runningOn.cluster.deploy", "false")
						}
					})
				})

				When("/component/command endpoint is POSTed", func() {
					BeforeEach(func() {
						url := fmt.Sprintf("http://%s/component/command", devSession.APIServerEndpoint)
						resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(`{"name": "push"}`)))
						Expect(err).ToNot(HaveOccurred())
						Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
					})

					It("should trigger a push", func() {
						err := devSession.WaitSync()
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			When("odo is executed with --no-watch and --api-server flags", helper.LabelPodmanIf(podman, func() {

				var devSession helper.DevSession

				BeforeEach(func() {
					var err error
					args := []string{"--no-watch"}
					devSession, err = helper.StartDevMode(helper.DevSessionOpts{
						CmdlineArgs:    args,
						RunOnPodman:    podman,
						StartAPIServer: true,
						EnvVars:        []string{"ODO_EXPERIMENTAL_MODE=true"},
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
						devSession.CheckNotSynced(10 * time.Second)
					})

					When("/component/command endpoint is POSTed", func() {

						BeforeEach(func() {
							url := fmt.Sprintf("http://%s/component/command", devSession.APIServerEndpoint)
							resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(`{"name": "push"}`)))
							Expect(err).ToNot(HaveOccurred())
							Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
						})

						It("should trigger a push", func() {
							err := devSession.WaitSync()
							Expect(err).ToNot(HaveOccurred())
							component := helper.NewComponent(cmpName, "app", labels.ComponentDevMode, commonVar.Project, commonVar.CliRunner)
							execResult, _ := component.Exec("runtime", []string{"cat", "/projects/server.js"}, pointer.Bool(true))
							Expect(execResult).To(ContainSubstring("App is super started"))
						})
					})
				})
			}))

			When("odo is executed with --api-server flag", helper.LabelPodmanIf(podman, func() {

				var devSession helper.DevSession

				BeforeEach(func() {
					var err error
					devSession, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman:    podman,
						StartAPIServer: true,
						EnvVars:        []string{"ODO_EXPERIMENTAL_MODE=true"},
					})
					Expect(err).ToNot(HaveOccurred())
				})

				When("/instance endpoint is DELETEd", func() {

					BeforeEach(func() {
						url := fmt.Sprintf("http://%s/instance", devSession.APIServerEndpoint)
						req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
						Expect(err).ToNot(HaveOccurred())
						client := &http.Client{}
						resp, err := client.Do(req)
						Expect(err).ToNot(HaveOccurred())
						Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK))
					})

					It("should terminate the dev session", func() {
						devSession.WaitEnd()
					})
				})
			}))
		}))
	}
})
