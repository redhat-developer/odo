package integration

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo describe component command tests", func() {
	var commonVar helper.CommonVar
	var cmpName string

	// This is run before every Spec (It)
	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
		cmpName = helper.RandString(6)
	})

	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.CommonAfterEach(commonVar)
	})

	for _, label := range []string{
		helper.LabelNoCluster, helper.LabelUnauth,
	} {
		label := label
		It("should fail, without cluster", Label(label), func() {
			By("running odo describe component -o json with namespace flag without name flag", func() {
				res := helper.Cmd("odo", "describe", "component", "--namespace", "default", "-o", "json").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(helper.IsJSON(stderr)).To(BeTrue())
				Expect(stdout).To(BeEmpty())
				helper.JsonPathContentContain(stderr, "message", "--namespace can be used only with --name")
			})

			By("running odo describe component -o json without name and without devfile in the current directory", func() {
				res := helper.Cmd("odo", "describe", "component", "-o", "json").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(helper.IsJSON(stderr)).To(BeTrue())
				Expect(stdout).To(BeEmpty())
				helper.JsonPathContentContain(stderr, "message", "The current directory does not represent an odo component")
			})

			By("running odo describe component with namespace flag without name flag", func() {
				res := helper.Cmd("odo", "describe", "component", "--namespace", "default").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(stdout).To(BeEmpty())
				Expect(stderr).To(ContainSubstring("--namespace can be used only with --name"))
			})

			By("running odo describe component without name and without devfile in the current directory", func() {
				res := helper.Cmd("odo", "describe", "component").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(stdout).To(BeEmpty())
				Expect(stderr).To(ContainSubstring("The current directory does not represent an odo component"))
			})

		})
	}

	It("should fail, with cluster", func() {
		By("running odo describe component -o json with an unknown name", func() {
			res := helper.Cmd("odo", "describe", "component", "--name", "unknown-name", "-o", "json").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(helper.IsJSON(stderr)).To(BeTrue())
			Expect(stdout).To(BeEmpty())
			helper.JsonPathContentContain(stderr, "message", "no component found with name \"unknown-name\" in the namespace \""+commonVar.Project+"\"")
		})

		By("running odo describe component with an unknown name", func() {
			res := helper.Cmd("odo", "describe", "component", "--name", "unknown-name").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(stdout).To(BeEmpty())
			Expect(stderr).To(ContainSubstring("no component found with name \"unknown-name\" in the namespace \"" + commonVar.Project + "\""))
		})
	})

	It("should fail, with podman", Label(helper.LabelPodman), func() {
		By("running odo describe component -o json with an unknown name", func() {
			res := helper.Cmd("odo", "describe", "component", "--name", "unknown-name", "--platform", "podman", "-o", "json").
				AddEnv("ODO_EXPERIMENTAL_MODE=true").
				ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(helper.IsJSON(stderr)).To(BeTrue())
			Expect(stdout).To(BeEmpty())
			helper.JsonPathContentContain(stderr, "message", "no component found with name \"unknown-name\"")
		})

		By("running odo describe component with an unknown name", func() {
			stderr := helper.Cmd("odo", "describe", "component", "--name", "unknown-name", "--platform", "podman").
				AddEnv("ODO_EXPERIMENTAL_MODE=true").
				ShouldFail().Err()
			Expect(stderr).To(ContainSubstring("no component found with name \"unknown-name\""))
		})
	})

	When("creating a component", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-registry.yaml")).ShouldPass()
		})

		checkDevfileJSONDescription := func(jsonContent string, devfileName string) {
			helper.JsonPathContentIs(jsonContent, "devfilePath", filepath.Join(commonVar.Context, devfileName))
			helper.JsonPathContentIs(jsonContent, "devfileData.devfile.metadata.name", cmpName)
			helper.JsonPathContentIs(jsonContent, "devfileData.supportedOdoFeatures.dev", "true")
			helper.JsonPathContentIs(jsonContent, "devfileData.supportedOdoFeatures.deploy", "false")
			helper.JsonPathContentIs(jsonContent, "devfileData.supportedOdoFeatures.debug", "true")
			helper.JsonPathContentIs(jsonContent, "managedBy", "odo")
		}

		checkDevfileDescription := func(content string, withUnknown bool) {
			Expect(content).To(ContainSubstring("Name: " + cmpName))
			Expect(content).To(ContainSubstring("Project Type: nodejs"))
			Expect(content).To(ContainSubstring("Supported odo features:"))
			if withUnknown {
				for _, v := range []string{"Version", "Display Name", "Description", "Language"} {
					Expect(content).To(ContainSubstring(v + ": Unknown"))
				}
				Expect(content).To(ContainSubstring("Dev: Unknown"))
				Expect(content).To(ContainSubstring("Debug: Unknown"))
				Expect(content).To(ContainSubstring("Deploy: Unknown"))
			} else {
				Expect(content).To(ContainSubstring("Display Name: "))
				Expect(content).To(ContainSubstring("Language: "))
				Expect(content).To(ContainSubstring("Version: "))
				Expect(content).To(ContainSubstring("Description: "))
				Expect(content).To(ContainSubstring("Tags: "))
				Expect(content).To(ContainSubstring("Dev: true"))
				Expect(content).To(ContainSubstring("Debug: true"))
				Expect(content).To(ContainSubstring("Deploy: false"))
			}
		}

		for _, label := range []string{
			helper.LabelNoCluster, helper.LabelUnauth,
		} {
			label := label
			for _, experimental := range []bool{false, true} {
				experimental := experimental

				When("experimental mode="+strconv.FormatBool(experimental), func() {

					BeforeEach(func() {
						if experimental {
							helper.EnableExperimentalMode()
						}
					})
					AfterEach(func() {
						if experimental {
							helper.ResetExperimentalMode()
						}
					})

					It("should describe the component in the current directory", Label(label), func() {
						By("running with json output", func() {
							res := helper.Cmd("odo", "describe", "component", "-o", "json").ShouldPass()
							stdout, stderr := res.Out(), res.Err()
							Expect(helper.IsJSON(stdout)).To(BeTrue())
							Expect(stderr).To(BeEmpty())
							checkDevfileJSONDescription(stdout, "devfile.yaml")
							helper.JsonPathContentIs(stdout, "runningIn", "")
							helper.JsonPathContentIs(stdout, "devForwardedPorts", "")
							helper.JsonPathDoesNotExist(stdout, "runningOn") // Deprecated
							helper.JsonPathDoesNotExist(stdout, "platform")
						})

						By("running with default output", func() {
							res := helper.Cmd("odo", "describe", "component").ShouldPass()
							stdout := res.Out()
							checkDevfileDescription(stdout, false)
							Expect(stdout).To(ContainSubstring("Running in: None"))
							Expect(stdout).ToNot(ContainSubstring("Forwarded ports"))
							Expect(stdout).ToNot(ContainSubstring("Running on:"))
						})
					})

					When("renaming to hide devfile.yaml file", Label(label), func() {
						BeforeEach(func() {
							err := os.Rename("devfile.yaml", ".devfile.yaml")
							Expect(err).NotTo(HaveOccurred())
						})

						It("should describe the component in the current directory using the hidden devfile", func() {
							By("running with json output", func() {
								res := helper.Cmd("odo", "describe", "component", "-o", "json").ShouldPass()
								stdout, stderr := res.Out(), res.Err()
								Expect(helper.IsJSON(stdout)).To(BeTrue())
								Expect(stderr).To(BeEmpty())
								checkDevfileJSONDescription(stdout, ".devfile.yaml")
								helper.JsonPathContentIs(stdout, "runningIn", "")
								helper.JsonPathContentIs(stdout, "devForwardedPorts", "")
								helper.JsonPathDoesNotExist(stdout, "runningOn") // Deprecated
								helper.JsonPathDoesNotExist(stdout, "platform")
							})

							By("running with default output", func() {
								res := helper.Cmd("odo", "describe", "component").ShouldPass()
								stdout := res.Out()
								checkDevfileDescription(stdout, false)
								Expect(stdout).To(ContainSubstring("Running in: None"))
								Expect(stdout).ToNot(ContainSubstring("Forwarded ports"))
								Expect(stdout).ToNot(ContainSubstring("Running on:"))
							})
						})
					})
				})
			}
		}

		It("should not describe the component from another directory", func() {
			By("running with json output", func() {
				err := os.Chdir("/")
				Expect(err).NotTo(HaveOccurred())
				res := helper.Cmd("odo", "describe", "component", "--name", cmpName, "-o", "json").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(helper.IsJSON(stderr)).To(BeTrue())
				Expect(stdout).To(BeEmpty())
				helper.JsonPathContentContain(stderr, "message", "no component found with name \""+cmpName+"\" in the namespace \""+commonVar.Project+"\"")
			})

			By("running with default output", func() {
				err := os.Chdir("/")
				Expect(err).NotTo(HaveOccurred())
				res := helper.Cmd("odo", "describe", "component", "--name", cmpName).ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(stdout).To(BeEmpty())
				Expect(stderr).To(ContainSubstring("no component found with name \"" + cmpName + "\" in the namespace \"" + commonVar.Project + "\""))
			})
		})

		for _, podman := range []bool{true, false} {
			podman := podman
			When(fmt.Sprintf("running odo dev (podman=%s)", strconv.FormatBool(podman)), helper.LabelPodmanIf(podman, func() {
				var devSession helper.DevSession
				var ports map[string]string

				BeforeEach(func() {
					var err error
					devSession, _, _, ports, err = helper.StartDevMode(helper.DevSessionOpts{
						RunOnPodman: podman,
					})
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					devSession.Stop()
					devSession.WaitEnd()
				})

				Context(fmt.Sprintf("Default output (podman=%s)", strconv.FormatBool(podman)), func() {

					When("describing the component in dev mode and without the experimental mode", func() {
						var stdout string
						BeforeEach(func() {
							stdout = helper.Cmd("odo", "describe", "component").ShouldPass().Out()
						})

						It("should describe the component", func() {
							checkDevfileDescription(stdout, false)
							if podman {
								// Information available only when running under the experimental mode
								Expect(stdout).To(ContainSubstring("Running in: None"))
								Expect(stdout).NotTo(ContainSubstring("Forwarded ports"))
								Expect(stdout).NotTo(ContainSubstring("127.0.0.1:"))
							} else {
								Expect(stdout).To(ContainSubstring("Running in: Dev"))
								Expect(stdout).To(ContainSubstring("Forwarded ports"))
								Expect(stdout).To(ContainSubstring("127.0.0.1:" + ports["3000"][len("127.0.0.1:"):] + " -> runtime:3000"))
							}
							Expect(stdout).NotTo(ContainSubstring("[cluster] 127.0.0.1:"))
							Expect(stdout).NotTo(ContainSubstring("[podman] 127.0.0.1:"))
							Expect(stdout).NotTo(ContainSubstring("Running on:"))
							Expect(stdout).NotTo(ContainSubstring("podman: "))
							Expect(stdout).NotTo(ContainSubstring("cluster: "))
						})
					})

					When("describing the component in dev mode and with the experimental mode enabled", func() {
						var stdout string
						BeforeEach(func() {
							stdout = helper.Cmd("odo", "describe", "component").AddEnv("ODO_EXPERIMENTAL_MODE=true").ShouldPass().Out()
						})

						It("should describe the component", func() {
							checkDevfileDescription(stdout, false)
							Expect(stdout).To(ContainSubstring("Running on:"))
							Expect(stdout).To(ContainSubstring("Forwarded ports"))
							if podman {
								Expect(stdout).To(ContainSubstring("[podman] 127.0.0.1:" + ports["3000"][len("127.0.0.1:"):] + " -> runtime:3000"))
								Expect(stdout).NotTo(ContainSubstring("[cluster] 127.0.0.1:"))
								Expect(stdout).To(ContainSubstring("podman: Dev"))
								Expect(stdout).NotTo(ContainSubstring("cluster: "))
							} else {
								Expect(stdout).To(ContainSubstring("[cluster] 127.0.0.1:" + ports["3000"][len("127.0.0.1:"):] + " -> runtime:3000"))
								Expect(stdout).NotTo(ContainSubstring("[podman] 127.0.0.1:"))
								Expect(stdout).To(ContainSubstring("cluster: Dev"))
								Expect(stdout).NotTo(ContainSubstring("podman: "))
							}
						})

					})

					When("switching to another directory", func() {
						BeforeEach(func() {
							err := os.Chdir("/")
							Expect(err).NotTo(HaveOccurred())
						})

						When("describing the component from another directory and without the experimental mode", func() {
							var stdout, stderr string
							BeforeEach(func() {
								cmd := helper.Cmd("odo", "describe", "component", "--name", cmpName)
								if podman {
									cmd = cmd.ShouldFail()
								} else {
									cmd = cmd.ShouldPass()
								}
								stdout, stderr = cmd.OutAndErr()
							})

							if podman {
								It("should fail to describe the named component", func() {
									// Podman mode assumes the test does not require a cluster.
									// But running "odo describe component --name" in non-experimental mode attempts to get information from a cluster first.
									// TODO We need to think about how to test both Cluster and Podman modes.
									Expect(stderr).To(ContainSubstring("cluster is non accessible"))
									Expect(stdout).NotTo(ContainSubstring("Forwarded ports"))
									Expect(stdout).NotTo(ContainSubstring("Running on"))
									Expect(stdout).NotTo(ContainSubstring("podman:"))
									Expect(stdout).NotTo(ContainSubstring("cluster:"))
								})
							} else {
								It("should describe the named component", func() {
									checkDevfileDescription(stdout, true)
									Expect(stdout).NotTo(ContainSubstring("Forwarded ports"))
									Expect(stdout).NotTo(ContainSubstring("Running on"))
									Expect(stdout).NotTo(ContainSubstring("podman:"))
									Expect(stdout).NotTo(ContainSubstring("cluster:"))
								})
							}

						})

						When("describing the component from another directory and with the experimental mode enabled", func() {
							var stdout string
							BeforeEach(func() {
								stdout = helper.Cmd("odo", "describe", "component", "--name", cmpName).
									AddEnv("ODO_EXPERIMENTAL_MODE=true").
									ShouldPass().
									Out()
							})

							It("should describe the named component", func() {
								checkDevfileDescription(stdout, true)
								Expect(stdout).NotTo(ContainSubstring("Forwarded ports"))
								Expect(stdout).To(ContainSubstring("Running in: Dev"))
								Expect(stdout).To(ContainSubstring("Running on"))
								if podman {
									Expect(stdout).To(ContainSubstring("podman: Dev"))
									Expect(stdout).NotTo(ContainSubstring("cluster:"))
								} else {
									Expect(stdout).To(ContainSubstring("cluster: Dev"))
									Expect(stdout).NotTo(ContainSubstring("podman:"))
								}
							})
						})
					})
				})

				Context(fmt.Sprintf("JSON output (podman=%s)", strconv.FormatBool(podman)), func() {
					When("describing the component in dev mode and without the experimental mode", func() {
						var stdout, stderr string
						BeforeEach(func() {
							stdout, stderr = helper.Cmd("odo", "describe", "component", "-o", "json").ShouldPass().OutAndErr()
						})

						It("should describe the component", func() {
							Expect(helper.IsJSON(stdout)).To(BeTrue())
							Expect(stderr).To(BeEmpty())
							checkDevfileJSONDescription(stdout, "devfile.yaml")
							helper.JsonPathDoesNotExist(stdout, "runningOn")
							helper.JsonPathDoesNotExist(stdout, "platform") // Deprecated
							if podman {
								// Information available only when running under the experimental mode
								helper.JsonPathDoesNotExist(stdout, "devForwardedPorts")
								helper.JsonPathContentIs(stdout, "runningIn.dev", "")
							} else {
								helper.JsonPathContentIs(stdout, "runningIn.dev", "true")
								helper.JsonPathContentIs(stdout, "runningIn.deploy", "false")
								helper.JsonPathContentIs(stdout, "devForwardedPorts.#", "1")
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.containerName", "runtime")
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.localAddress", "127.0.0.1")
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.localPort", ports["3000"][len("127.0.0.1:"):])
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.containerPort", "3000")
							}
						})
					})

					When("describing the component in dev mode and with the experimental mode enabled", func() {
						var stdout, stderr string
						BeforeEach(func() {
							stdout, stderr = helper.Cmd("odo", "describe", "component", "-o", "json").
								AddEnv("ODO_EXPERIMENTAL_MODE=true").
								ShouldPass().
								OutAndErr()
						})

						It("should describe the component", func() {
							Expect(helper.IsJSON(stdout)).To(BeTrue())
							Expect(stderr).To(BeEmpty())
							checkDevfileJSONDescription(stdout, "devfile.yaml")
							helper.JsonPathContentIs(stdout, "runningIn.dev", "true")
							helper.JsonPathContentIs(stdout, "runningIn.deploy", "false")
							helper.JsonPathContentIs(stdout, "devForwardedPorts.#", "1")
							helper.JsonPathContentIs(stdout, "devForwardedPorts.0.containerName", "runtime")
							helper.JsonPathContentIs(stdout, "devForwardedPorts.0.localAddress", "127.0.0.1")
							helper.JsonPathContentIs(stdout, "devForwardedPorts.0.localPort", ports["3000"][len("127.0.0.1:"):])
							helper.JsonPathContentIs(stdout, "devForwardedPorts.0.containerPort", "3000")
							if podman {
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.platform", "podman")
								helper.JsonPathContentIs(stdout, "runningOn.podman.dev", "true")
								helper.JsonPathContentIs(stdout, "runningOn.podman.deploy", "false")
								helper.JsonPathDoesNotExist(stdout, "runningOn.cluster")
							} else {
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.platform", "cluster")
								helper.JsonPathContentIs(stdout, "runningOn.cluster.dev", "true")
								helper.JsonPathContentIs(stdout, "runningOn.cluster.deploy", "false")
								helper.JsonPathDoesNotExist(stdout, "runningOn.podman")
							}
						})
					})

					When("switching to another directory", func() {
						BeforeEach(func() {
							err := os.Chdir("/")
							Expect(err).NotTo(HaveOccurred())
						})

						When("describing the component from another directory and without the experimental mode", func() {
							var stdout, stderr string
							BeforeEach(func() {
								cmd := helper.Cmd("odo", "describe", "component", "--name", cmpName, "-o", "json")
								if podman {
									cmd = cmd.ShouldFail()
								} else {
									cmd = cmd.ShouldPass()
								}
								stdout, stderr = cmd.OutAndErr()
							})

							if podman {
								It("should fail to describe the named component", func() {
									// Podman mode assumes the test does not require a cluster.
									// But running "odo describe component --name" in non-experimental mode attempts to get information from a cluster first.
									// TODO We need to think about how to test both Cluster and Podman modes.
									Expect(helper.IsJSON(stderr)).To(BeTrue())
									Expect(stdout).To(BeEmpty())
									helper.JsonPathDoesNotExist(stderr, "runningOn") // Deprecated
									helper.JsonPathDoesNotExist(stderr, "platform")
									helper.JsonPathContentIs(stderr, "message", "cluster is non accessible")
									helper.JsonPathDoesNotExist(stderr, "devfilePath")
									helper.JsonPathDoesNotExist(stderr, "devForwardedPorts")
									helper.JsonPathDoesNotExist(stderr, "devfileData")
									helper.JsonPathDoesNotExist(stderr, "runningIn")
								})
							} else {
								It("should describe the named component", func() {
									Expect(helper.IsJSON(stdout)).To(BeTrue())
									Expect(stderr).To(BeEmpty())
									helper.JsonPathContentIs(stdout, "devfilePath", "")
									helper.JsonPathContentIs(stdout, "devfileData.devfile.metadata.name", cmpName)
									helper.JsonPathContentIs(stdout, "devfileData.devfile.metadata.projectType", "nodejs")
									for _, v := range []string{"version", "displayName", "description", "language"} {
										helper.JsonPathContentIs(stdout, "devfileData.devfile.metadata."+v, "Unknown")
									}
									helper.JsonPathContentIs(stdout, "devForwardedPorts", "")
									helper.JsonPathContentIs(stdout, "runningIn.dev", "true")
									helper.JsonPathContentIs(stdout, "runningIn.deploy", "false")
									helper.JsonPathDoesNotExist(stdout, "runningOn") // Deprecated
									helper.JsonPathDoesNotExist(stdout, "platform")
								})
							}
						})

						When("describing the component from another directory and with the experimental mode enabled", func() {
							var stdout, stderr string
							BeforeEach(func() {
								stdout, stderr = helper.Cmd("odo", "describe", "component", "--name", cmpName, "-o", "json").
									AddEnv("ODO_EXPERIMENTAL_MODE=true").
									ShouldPass().
									OutAndErr()
							})

							It("should describe the named component", func() {
								Expect(helper.IsJSON(stdout)).To(BeTrue())
								Expect(stderr).To(BeEmpty())
								helper.JsonPathContentIs(stdout, "runningIn.dev", "true")
								helper.JsonPathContentIs(stdout, "runningIn.deploy", "false")
								helper.JsonPathContentIs(stdout, "devfilePath", "")
								helper.JsonPathContentIs(stdout, "devfileData.devfile.metadata.name", cmpName)
								helper.JsonPathContentIs(stdout, "devfileData.devfile.metadata.projectType", "nodejs")
								for _, v := range []string{"version", "displayName", "description", "language"} {
									helper.JsonPathContentIs(stdout, "devfileData.devfile.metadata."+v, "Unknown")
								}
								helper.JsonPathContentIs(stdout, "devForwardedPorts", "")
								if podman {
									helper.JsonPathContentIs(stdout, "runningOn.podman.dev", "true")
									helper.JsonPathContentIs(stdout, "runningOn.podman.deploy", "false")
									helper.JsonPathDoesNotExist(stdout, "runningOn.cluster")
								} else {
									helper.JsonPathContentIs(stdout, "runningOn.cluster.dev", "true")
									helper.JsonPathContentIs(stdout, "runningOn.cluster.deploy", "false")
									helper.JsonPathDoesNotExist(stdout, "runningOn.podman")
								}
							})
						})
					})
				})
			}))
		}

		for _, ctx := range []struct {
			title           string
			devfile         string
			matchOutput     []string
			matchJSONOutput map[string]string
		}{
			{
				title: "ingress/routes",
				devfile: func() string {
					if helper.IsKubernetesCluster() {
						return "devfile-deploy-ingress.yaml"
					}
					return "devfile-deploy-route.yaml"
				}(),
				matchOutput: func() []string {
					if helper.IsKubernetesCluster() {
						return []string{"Kubernetes Ingresses", "nodejs.example.com/", "nodejs.example.com/foo"}
					}
					return []string{"OpenShift Routes", "/foo"}
				}(),
				matchJSONOutput: func() map[string]string {
					if helper.IsKubernetesCluster() {
						return map[string]string{"ingresses.0.name": "my-nodejs-app", "ingresses.0.rules.0.host": "nodejs.example.com", "ingresses.0.rules.0.paths.0": "/", "ingresses.0.rules.0.paths.1": "/foo"}
					}
					return map[string]string{"routes.0.name": "my-nodejs-app", "routes.0.rules.0.paths.0": "/foo"}
				}(),
			},
			{
				title:           "ingress with defaultBackend",
				devfile:         "devfile-deploy-defaultBackend-ingress.yaml",
				matchOutput:     []string{"Kubernetes Ingresses", "*/*"},
				matchJSONOutput: map[string]string{"ingresses.0.name": "my-nodejs-app", "ingresses.0.rules.0.host": "*", "ingresses.0.rules.0.paths.0": "/*"},
			},
		} {
			ctx := ctx
			When("running odo deploy to create ingress/routes", func() {
				var componentName string
				BeforeEach(func() {
					componentName = helper.RandString(6)
					helper.CopyExampleDevFile(
						filepath.Join("source", "devfiles", "nodejs", ctx.devfile),
						path.Join(commonVar.Context, "devfile.yaml"),
						helper.DevfileMetadataNameSetter(componentName))
					helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
				})
				It(fmt.Sprintf("should show the %s in odo describe component output", ctx.title), func() {
					By("checking the human readable output", func() {
						out := helper.Cmd("odo", "describe", "component").ShouldPass().Out()
						helper.MatchAllInOutput(out, ctx.matchOutput)
					})
					By("checking the machine readable output", func() {
						out := helper.Cmd("odo", "describe", "component", "-o", "json").ShouldPass().Out()
						for key, value := range ctx.matchJSONOutput {
							helper.JsonPathContentContain(out, key, value)
						}
					})
					By("checking the human readable output with component name", func() {
						out := helper.Cmd("odo", "describe", "component", "--name", componentName).ShouldPass().Out()
						helper.MatchAllInOutput(out, ctx.matchOutput)
					})
					By("checking the machine readable output with component name", func() {
						out := helper.Cmd("odo", "describe", "component", "--name", componentName, "-o", "json").ShouldPass().Out()
						for key, value := range ctx.matchJSONOutput {
							helper.JsonPathContentContain(out, key, value)
						}
					})
				})
			})
		}
	})
	Context("checking for remote source code location", func() {
		for _, podman := range []bool{true, false} {
			podman := podman
			for _, ctx := range []struct {
				devfile, title string
				checker        func(output string, isJSON bool)
				beforeEach     func()
			}{
				{
					title:   "devfile with sourceMapping",
					devfile: "devfileSourceMapping.yaml",
					checker: func(output string, isJSON bool) {
						const location = "/test"
						if isJSON {
							helper.JsonPathContentIs(output, "devfileData.devfile.components.#(name==runtime).container.sourceMapping", location)
							return
						}
						Expect(output).To(ContainSubstring(location))
					},
				},
				{
					devfile: "devfile.yaml",
					title:   "devfile with no sourceMapping, defaults to /projects",
					checker: func(output string, isJSON bool) {
						const location = "/projects"
						if isJSON {
							helper.JsonPathContentIs(output, "devfileData.devfile.components.#(name==runtime).container.sourceMapping", location)
							return
						}
						Expect(output).To(ContainSubstring(location))
					},
				},
				{
					devfile: "devfileCompositeBuildRunDebugInMultiContainersAndSharedVolume.yaml",
					title:   "devfile with containers that has mountSource set to false",
					checker: func(output string, isJSON bool) {
						if isJSON {
							helper.JsonPathContentIs(output, "devfileData.devfile.components.#(name==runtime).container.sourceMapping", "/projects")
							helper.JsonPathDoesNotExist(output, "devfileData.devfile.components.#(name==sleeper-run).container.sourceMapping")
							helper.JsonPathDoesNotExist(output, "devfileData.devfile.components.#(name==sleeper-build).container.sourceMapping")
							helper.JsonPathDoesNotExist(output, "devfileData.devfile.components.#(name==echo-er).container.sourceMapping")
							helper.JsonPathDoesNotExist(output, "devfileData.devfile.components.#(name==build-checker).container.sourceMapping")
							return
						}
						Expect(output).To(ContainSubstring("runtime\n    Source Mapping: /projects"))
						helper.DontMatchAllInOutput(output, []string{"sleeper-run\n    Source Mapping: /projects", "sleeper-build\n    Source Mapping:", "echo-er\n    Source Mapping:", "build-checker\n    Source Mapping:"})
					},
				},
			} {
				ctx := ctx
				When(fmt.Sprintf("using %s and starting an odo dev session", ctx.title), helper.LabelPodmanIf(podman, func() {
					var devSession helper.DevSession
					BeforeEach(func() {
						helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
						helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", ctx.devfile)).ShouldPass()
						var err error
						devSession, _, _, _, err = helper.StartDevMode(helper.DevSessionOpts{RunOnPodman: podman})
						Expect(err).ToNot(HaveOccurred())
					})
					AfterEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})
					It("should show remote source code location in odo describe component output", func() {
						By("checking human readable output", func() {
							args := []string{"describe", "component"}
							cmd := helper.Cmd("odo", args...)
							if podman {
								args = append(args, "--platform=podman")
								cmd = helper.Cmd("odo", args...).AddEnv("ODO_EXPERIMENTAL_MODE=true")
							}
							output := cmd.ShouldPass().Out()
							ctx.checker(output, false)
						})
						By("checking JSON output", func() {
							args := []string{"describe", "component", "-ojson"}
							cmd := helper.Cmd("odo", args...)
							if podman {
								args = append(args, "--platform=podman")
								cmd = helper.Cmd("odo", args...).AddEnv("ODO_EXPERIMENTAL_MODE=true")
							}
							output := cmd.ShouldPass().Out()
							ctx.checker(output, true)
						})
					})
				}))
			}
		}
	})
})
