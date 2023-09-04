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

	It("should fail, with default cluster mode", func() {
		By("running odo describe component -o json with an unknown name", func() {
			helper.CreateInvalidDevfile(commonVar.Context)
			res := helper.Cmd("odo", "describe", "component", "--name", "unknown-name", "-o", "json").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(helper.IsJSON(stderr)).To(BeTrue())
			Expect(stdout).To(BeEmpty())
			helper.JsonPathContentContain(stderr, "message", "no component found with name \"unknown-name\"")
		})

		By("running odo describe component with an unknown name", func() {
			helper.CreateInvalidDevfile(commonVar.Context)
			res := helper.Cmd("odo", "describe", "component", "--name", "unknown-name").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(stdout).To(BeEmpty())
			Expect(stderr).To(ContainSubstring("no component found with name \"unknown-name\""))
		})
	})

	It("should fail, with cluster", func() {
		By("running odo describe component -o json with an unknown name", func() {
			helper.CreateInvalidDevfile(commonVar.Context)
			res := helper.Cmd("odo", "describe", "component", "--name", "unknown-name", "--platform", "cluster", "-o", "json").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(helper.IsJSON(stderr)).To(BeTrue())
			Expect(stdout).To(BeEmpty())
			helper.JsonPathContentContain(stderr, "message", "no component found with name \"unknown-name\" in the namespace \""+commonVar.Project+"\"")
		})

		By("running odo describe component with an unknown name", func() {
			helper.CreateInvalidDevfile(commonVar.Context)
			res := helper.Cmd("odo", "describe", "component", "--name", "unknown-name", "--platform", "cluster").ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(stdout).To(BeEmpty())
			Expect(stderr).To(ContainSubstring("no component found with name \"unknown-name\" in the namespace \"" + commonVar.Project + "\""))
		})
	})

	It("should fail, with podman", Label(helper.LabelPodman), func() {
		By("running odo describe component -o json with an unknown name", func() {
			helper.CreateInvalidDevfile(commonVar.Context)
			res := helper.Cmd("odo", "describe", "component", "--name", "unknown-name", "--platform", "podman", "-o", "json").
				ShouldFail()
			stdout, stderr := res.Out(), res.Err()
			Expect(helper.IsJSON(stderr)).To(BeTrue())
			Expect(stdout).To(BeEmpty())
			helper.JsonPathContentContain(stderr, "message", "no component found with name \"unknown-name\"")
		})

		By("running odo describe component with an unknown name", func() {
			helper.CreateInvalidDevfile(commonVar.Context)
			stderr := helper.Cmd("odo", "describe", "component", "--name", "unknown-name", "--platform", "podman").
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

			helper.JsonPathContentHasLen(jsonContent, "devfileData.commands", 4)
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.0.name", "install")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.0.group", "build")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.0.commandLine", "npm install")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.1.name", "run")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.1.group", "run")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.1.commandLine", "npm start")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.2.name", "debug")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.2.group", "debug")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.2.commandLine", "npm run debug")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.3.name", "test")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.3.group", "test")
			helper.JsonPathContentIs(jsonContent, "devfileData.commands.3.commandLine", "npm test")
			for i := 0; i <= 3; i++ {
				helper.JsonPathContentIs(jsonContent, fmt.Sprintf("devfileData.commands.%d.type", i), "exec")
				helper.JsonPathContentIs(jsonContent, fmt.Sprintf("devfileData.commands.%d.isDefault", i), "true")
				helper.JsonPathContentIs(jsonContent, fmt.Sprintf("devfileData.commands.%d.component", i), "runtime")
				helper.JsonPathContentIs(jsonContent, fmt.Sprintf("devfileData.commands.%d.componentType", i), "container")
			}
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
				Expect(content).ShouldNot(ContainSubstring("Commands:"))
			} else {
				Expect(content).To(ContainSubstring("Display Name: "))
				Expect(content).To(ContainSubstring("Language: "))
				Expect(content).To(ContainSubstring("Version: "))
				Expect(content).To(ContainSubstring("Description: "))
				Expect(content).To(ContainSubstring("Tags: "))
				Expect(content).To(ContainSubstring("Dev: true"))
				Expect(content).To(ContainSubstring("Debug: true"))
				Expect(content).To(ContainSubstring("Deploy: false"))

				Expect(content).To(ContainSubstring("Commands:"))
				for _, c := range []string{"exec"} {
					Expect(content).To(ContainSubstring("Type: " + c))
				}
				for _, c := range []string{"runtime"} {
					Expect(content).To(ContainSubstring("Component: " + c))
				}
				for _, c := range []string{"container"} {
					Expect(content).To(ContainSubstring("Component Type: " + c))
				}
				for _, c := range []string{"install", "run", "debug", "test"} {
					Expect(content).To(ContainSubstring(c))
				}
				for _, c := range []string{"build", "run", "debug", "test"} {
					Expect(content).To(ContainSubstring("Group: %s", c))
				}
				for _, c := range []string{"npm install", "npm start", "npm run debug", "npm test"} {
					Expect(content).To(ContainSubstring("Command Line: %q", c))
				}
			}
		}

		for _, label := range []string{
			helper.LabelNoCluster, helper.LabelUnauth,
		} {
			label := label
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
		}

		It("should not describe the component from another directory, with default cluster mode", func() {
			By("running with json output", func() {
				otherDir := filepath.Join(commonVar.Context, "tmp")
				helper.MakeDir(otherDir)
				helper.Chdir(otherDir)
				helper.CreateInvalidDevfile(otherDir)
				res := helper.Cmd("odo", "describe", "component", "--name", cmpName, "-o", "json").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(helper.IsJSON(stderr)).To(BeTrue())
				Expect(stdout).To(BeEmpty())
				helper.JsonPathContentContain(stderr, "message", "no component found with name \""+cmpName+"\"")
			})

			By("running with default output", func() {
				otherDir := filepath.Join(commonVar.Context, "tmp")
				helper.MakeDir(otherDir)
				helper.Chdir(otherDir)
				helper.CreateInvalidDevfile(otherDir)
				res := helper.Cmd("odo", "describe", "component", "--name", cmpName).ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(stdout).To(BeEmpty())
				Expect(stderr).To(ContainSubstring("no component found with name %q", cmpName))
			})
		})

		It("should not describe the component from another directory, with cluster", func() {
			By("running with json output", func() {
				otherDir := filepath.Join(commonVar.Context, "tmp")
				helper.MakeDir(otherDir)
				helper.Chdir(otherDir)
				helper.CreateInvalidDevfile(otherDir)
				res := helper.Cmd("odo", "describe", "component", "--name", cmpName, "-o", "json", "--platform", "cluster").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(helper.IsJSON(stderr)).To(BeTrue())
				Expect(stdout).To(BeEmpty())
				helper.JsonPathContentContain(stderr, "message", "no component found with name \""+cmpName+"\" in the namespace \""+commonVar.Project+"\"")
			})

			By("running with default output", func() {
				otherDir := filepath.Join(commonVar.Context, "tmp")
				helper.MakeDir(otherDir)
				helper.Chdir(otherDir)
				helper.CreateInvalidDevfile(otherDir)
				res := helper.Cmd("odo", "describe", "component", "--name", cmpName, "--platform", "cluster").ShouldFail()
				stdout, stderr := res.Out(), res.Err()
				Expect(stdout).To(BeEmpty())
				Expect(stderr).To(ContainSubstring("no component found with name %q in the namespace %q", cmpName, commonVar.Project))
			})
		})

		for _, podman := range []bool{true, false} {
			podman := podman
			for _, debug := range []bool{false, true} {
				debug := debug
				When(fmt.Sprintf("running odo dev (podman=%s,debug=%s)", strconv.FormatBool(podman), strconv.FormatBool(debug)), helper.LabelPodmanIf(podman, func() {
					var devSession helper.DevSession

					BeforeEach(func() {
						opts := helper.DevSessionOpts{RunOnPodman: podman}
						if debug {
							opts.CmdlineArgs = []string{"--debug"}
							if podman {
								// TODO(rm3l): use forward-localhost when it is implemented
								opts.CmdlineArgs = append(opts.CmdlineArgs, "--ignore-localhost")
							}
						}
						var err error
						devSession, err = helper.StartDevMode(opts)
						Expect(err).NotTo(HaveOccurred())
					})

					AfterEach(func() {
						devSession.Stop()
						devSession.WaitEnd()
					})

					Context("Default output", func() {

						When("describing the component in dev mode", func() {
							var stdout string
							BeforeEach(func() {
								stdout = helper.Cmd("odo", "describe", "component").ShouldPass().Out()
							})

							It("should describe the component", func() {
								checkDevfileDescription(stdout, false)
								Expect(stdout).To(ContainSubstring("Running on:"))
								Expect(stdout).To(ContainSubstring("Forwarded ports"))
								if podman {
									Expect(stdout).To(ContainSubstring("[podman] 127.0.0.1:%s -> runtime:3000\n    Name: http-3000", devSession.Endpoints["3000"][len("127.0.0.1:"):]))
									if debug {
										Expect(stdout).To(
											ContainSubstring("127.0.0.1:%s -> runtime:5858\n    Name: debug\n    Exposure: none\n    Debug: true",
												devSession.Endpoints["5858"][len("127.0.0.1:"):]))
									}
									Expect(stdout).NotTo(ContainSubstring("[cluster] 127.0.0.1:"))
									Expect(stdout).To(ContainSubstring("podman: Dev"))
									Expect(stdout).NotTo(ContainSubstring("cluster: "))
								} else {
									Expect(stdout).To(ContainSubstring("[cluster] 127.0.0.1:%s -> runtime:3000\n    Name: http-3000", devSession.Endpoints["3000"][len("127.0.0.1:"):]))
									if debug {
										Expect(stdout).To(ContainSubstring("[cluster] 127.0.0.1:%s -> runtime:5858\n    Name: debug\n    Exposure: none\n    Debug: true", devSession.Endpoints["5858"][len("127.0.0.1:"):]))
									}
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

							When("describing the component from another directory", func() {
								var stdout string
								BeforeEach(func() {
									stdout = helper.Cmd("odo", "describe", "component", "--name", cmpName).
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

					Context("JSON output", func() {

						When("describing the component in dev mode", func() {
							var stdout, stderr string
							BeforeEach(func() {
								stdout, stderr = helper.Cmd("odo", "describe", "component", "-o", "json").
									ShouldPass().
									OutAndErr()
							})

							It("should describe the component", func() {
								Expect(helper.IsJSON(stdout)).To(BeTrue())
								Expect(stderr).To(BeEmpty())
								checkDevfileJSONDescription(stdout, "devfile.yaml")
								helper.JsonPathContentIs(stdout, "runningIn.dev", "true")
								helper.JsonPathContentIs(stdout, "runningIn.deploy", "false")
								if debug {
									helper.JsonPathContentIs(stdout, "devForwardedPorts.#", "2")
								} else {
									helper.JsonPathContentIs(stdout, "devForwardedPorts.#", "1")
								}
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.containerName", "runtime")
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.portName", "http-3000")
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.isDebug", "false")
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.localAddress", "127.0.0.1")
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.localPort", devSession.Endpoints["3000"][len("127.0.0.1:"):])
								helper.JsonPathContentIs(stdout, "devForwardedPorts.0.containerPort", "3000")
								helper.JsonPathDoesNotExist(stdout, "devForwardedPorts.0.exposure")
								if debug {
									helper.JsonPathContentIs(stdout, "devForwardedPorts.1.containerName", "runtime")
									helper.JsonPathContentIs(stdout, "devForwardedPorts.1.portName", "debug")
									helper.JsonPathContentIs(stdout, "devForwardedPorts.1.isDebug", "true")
									helper.JsonPathContentIs(stdout, "devForwardedPorts.1.localAddress", "127.0.0.1")
									helper.JsonPathContentIs(stdout, "devForwardedPorts.1.localPort", devSession.Endpoints["5858"][len("127.0.0.1:"):])
									helper.JsonPathContentIs(stdout, "devForwardedPorts.1.containerPort", "5858")
									helper.JsonPathContentIs(stdout, "devForwardedPorts.1.exposure", "none")
								}
								if podman {
									helper.JsonPathContentIs(stdout, "devForwardedPorts.0.platform", "podman")
									if debug {
										helper.JsonPathContentIs(stdout, "devForwardedPorts.1.platform", "podman")
									}
									helper.JsonPathContentIs(stdout, "runningOn.podman.dev", "true")
									helper.JsonPathContentIs(stdout, "runningOn.podman.deploy", "false")
									helper.JsonPathDoesNotExist(stdout, "runningOn.cluster")
								} else {
									helper.JsonPathContentIs(stdout, "devForwardedPorts.0.platform", "cluster")
									if debug {
										helper.JsonPathContentIs(stdout, "devForwardedPorts.1.platform", "cluster")
									}
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

							When("describing the component from another directory", func() {
								var stdout, stderr string
								BeforeEach(func() {
									stdout, stderr = helper.Cmd("odo", "describe", "component", "--name", cmpName, "-o", "json").
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
									helper.JsonPathDoesNotExist(stdout, "devfileData.commands")
								})
							})
						})
					})
				}))
			}
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
						componentName)
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
						helper.CreateInvalidDevfile(commonVar.Context)
						out := helper.Cmd("odo", "describe", "component", "--name", componentName).ShouldPass().Out()
						helper.MatchAllInOutput(out, ctx.matchOutput)
					})
					By("checking the machine readable output with component name", func() {
						helper.CreateInvalidDevfile(commonVar.Context)
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
						devSession, err = helper.StartDevMode(helper.DevSessionOpts{RunOnPodman: podman})
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
								cmd = helper.Cmd("odo", args...)
							}
							output := cmd.ShouldPass().Out()
							ctx.checker(output, false)
						})
						By("checking JSON output", func() {
							args := []string{"describe", "component", "-ojson"}
							cmd := helper.Cmd("odo", args...)
							if podman {
								args = append(args, "--platform=podman")
								cmd = helper.Cmd("odo", args...)
							}
							output := cmd.ShouldPass().Out()
							ctx.checker(output, true)
						})
					})
				}))
			}
		}
	})

	When("a non-odo application is present on the cluster", func() {
		var (
			// From manifests
			componentName = "example-deployment"
			ingressDomain = "example-deployment.example.com/"
		)

		BeforeEach(func() {
			commonVar.CliRunner.Run("create", "-f", helper.GetExamplePath("manifests", "deployment-app-label.yaml"))
			if helper.IsKubernetesCluster() {
				commonVar.CliRunner.Run("create", "-f", helper.GetExamplePath("manifests", "ingress-app-label.yaml"))
			} else {
				commonVar.CliRunner.Run("create", "-f", helper.GetExamplePath("manifests", "route-app-label.yaml"))
			}

		})
		AfterEach(func() {
			if helper.IsKubernetesCluster() {
				commonVar.CliRunner.Run("delete", "-f", helper.GetExamplePath("manifests", "ingress-app-label.yaml"))
			} else {
				commonVar.CliRunner.Run("delete", "-f", helper.GetExamplePath("manifests", "route-app-label.yaml"))
			}
			commonVar.CliRunner.Run("delete", "-f", helper.GetExamplePath("manifests", "deployment-app-label.yaml"))
		})

		It("should describe the component", func() {
			output := helper.Cmd("odo", "describe", "component", "--name", componentName).ShouldPass().Out()

			Expect(output).To(ContainSubstring("Name: " + componentName))

			if helper.IsKubernetesCluster() {
				helper.MatchAllInOutput(output, []string{
					"Kubernetes Ingresses",
					componentName + ": " + ingressDomain,
				})
			} else {
				helper.MatchAllInOutput(output, []string{
					"OpenShift Routes",
					componentName + ": ",
				})
			}
		})
	})

	Context("describe commands in Devfile", Label(helper.LabelUnauth), Label(helper.LabelNoCluster), func() {

		When("initializing a component with different types of commands", func() {

			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "devfiles", "nodejs", "project"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-deploy-functional-pods.yaml"),
					path.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
			})

			It("should describe the Devfile commands in human-readable form", func() {
				stdout := helper.Cmd("odo", "describe", "component").ShouldPass().Out()
				Expect(stdout).To(ContainSubstring("Commands:"))
				for _, c := range []string{"exec", "composite", "apply"} {
					Expect(stdout).To(ContainSubstring("Type: " + c))
				}
				for _, c := range []string{"runtime", "innerloop-pod", "prod-image", "outerloop-deploy"} {
					Expect(stdout).To(ContainSubstring("Component: " + c))
				}
				for _, c := range []string{"container", "kubernetes", "image"} {
					Expect(stdout).To(ContainSubstring("Component Type: " + c))
				}
				for _, c := range []string{
					"install",
					"innerloop-pod-command",
					"start",
					"run",
					"build-image",
					"deploy-deployment",
					"deploy-another-deployment",
					"outerloop-pod-command",
					"deploy",
				} {
					Expect(stdout).To(ContainSubstring(c))
				}
				for _, c := range []string{"build", "run", "deploy"} {
					Expect(stdout).To(ContainSubstring("Group: %s", c))
				}
				for _, c := range []string{"npm install", "npm start"} {
					Expect(stdout).To(ContainSubstring("Command Line: %q", c))
				}
				for _, c := range []string{"quay.io/tkral/devfile-nodejs-deploy:latest"} {
					Expect(stdout).To(ContainSubstring("Image Name: %s", c))
				}
			})

			It("should describe the Devfile commands in JSON output", func() {
				stdout := helper.Cmd("odo", "describe", "component", "-o", "json").ShouldPass().Out()
				Expect(helper.IsJSON(stdout)).To(BeTrue(), fmt.Sprintf("invalid JSON output: %q", stdout))

				helper.JsonPathContentHasLen(stdout, "devfileData.commands", 9)

				helper.JsonPathContentIs(stdout, "devfileData.commands.0.name", "install")
				helper.JsonPathContentIs(stdout, "devfileData.commands.0.group", "build")
				helper.JsonPathContentIs(stdout, "devfileData.commands.0.commandLine", "npm install")
				helper.JsonPathContentIs(stdout, "devfileData.commands.0.type", "exec")
				helper.JsonPathContentIs(stdout, "devfileData.commands.0.isDefault", "true")
				helper.JsonPathContentIs(stdout, "devfileData.commands.0.component", "runtime")
				helper.JsonPathContentIs(stdout, "devfileData.commands.0.componentType", "container")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.0.imageName")

				helper.JsonPathContentIs(stdout, "devfileData.commands.1.name", "innerloop-pod-command")
				helper.JsonPathContentIs(stdout, "devfileData.commands.1.type", "apply")
				helper.JsonPathContentIs(stdout, "devfileData.commands.1.component", "innerloop-pod")
				helper.JsonPathContentIs(stdout, "devfileData.commands.1.componentType", "kubernetes")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.1.group")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.1.commandLine")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.1.isDefault")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.1.imageName")

				helper.JsonPathContentIs(stdout, "devfileData.commands.4.name", "build-image")
				helper.JsonPathContentIs(stdout, "devfileData.commands.4.type", "apply")
				helper.JsonPathContentIs(stdout, "devfileData.commands.4.component", "prod-image")
				helper.JsonPathContentIs(stdout, "devfileData.commands.4.componentType", "image")
				helper.JsonPathContentIs(stdout, "devfileData.commands.4.imageName", "quay.io/tkral/devfile-nodejs-deploy:latest")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.4.group")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.4.commandLine")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.4.isDefault")

				helper.JsonPathContentIs(stdout, "devfileData.commands.8.name", "deploy")
				helper.JsonPathContentIs(stdout, "devfileData.commands.8.group", "deploy")
				helper.JsonPathContentIs(stdout, "devfileData.commands.8.type", "composite")
				helper.JsonPathContentIs(stdout, "devfileData.commands.8.isDefault", "true")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.8.imageName")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.8.commandLine")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.8.component")
				helper.JsonPathDoesNotExist(stdout, "devfileData.commands.8.componentType")
			})
		})
	})
})
