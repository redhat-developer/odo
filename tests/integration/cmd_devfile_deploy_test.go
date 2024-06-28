package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/labels"

	segment "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile deploy command tests", func() {

	var commonVar helper.CommonVar
	var cmpName string

	var _ = BeforeEach(func() {
		cmpName = helper.RandString(6)
		commonVar = helper.CommonBeforeEach()
		helper.Chdir(commonVar.Context)
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
			output := helper.Cmd("odo", "deploy").ShouldFail().Err()
			Expect(output).To(ContainSubstring("The current directory does not represent an odo component"))

		})
	})
	When("a component is bootstrapped", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", cmpName, "--devfile-path", helper.GetExamplePath("source", "devfiles", "nodejs", "devfile-deploy.yaml")).ShouldPass()
		})
		When("using a default namespace", func() {
			BeforeEach(func() {
				commonVar.CliRunner.SetProject("default")
			})
			AfterEach(func() {
				helper.Cmd("odo", "delete", "component", "-f").ShouldPass()
				commonVar.CliRunner.SetProject(commonVar.Project)
			})

			It("should display warning when running the deploy command", func() {
				errOut := helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldRun().Err()
				namespace := "project"
				if helper.IsKubernetesCluster() {
					namespace = "namespace"
				}
				Expect(errOut).To(ContainSubstring(fmt.Sprintf("You are using \"default\" %[1]s, odo may not work as expected in the default %[1]s.", namespace)))
			})
		})
		It("should fail to run odo deploy when not connected to any cluster", Label(helper.LabelNoCluster), func() {
			errOut := helper.Cmd("odo", "deploy").ShouldFail().Err()
			Expect(errOut).To(ContainSubstring("unable to access the cluster"))
		})
	})
	for _, ctx := range []struct {
		title       string
		devfileName string
		setupFunc   func()
	}{
		{
			title:       "using a devfile.yaml containing a deploy command",
			devfileName: "devfile-deploy.yaml",
			setupFunc:   nil,
		},
		{
			title:       "using a devfile.yaml containing an outer-loop Kubernetes component referenced via an URI",
			devfileName: "devfile-deploy-with-k8s-uri.yaml",
			setupFunc: func() {
				helper.CopyExample(
					filepath.Join("source", "devfiles", "nodejs", "kubernetes", "devfile-deploy-with-k8s-uri"),
					filepath.Join(commonVar.Context, "kubernetes", "devfile-deploy-with-k8s-uri"))
			},
		},
	} {
		// this is a workaround to ensure that the for loop works with `It` blocks
		ctx := ctx

		When(ctx.title, func() {
			// from devfile
			deploymentName := "my-component"
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				if ctx.setupFunc != nil {
					ctx.setupFunc()
				}
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", ctx.devfileName),
					path.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
			})

			for _, tt := range []struct {
				name                string
				imageBuildExtraArgs []string
			}{
				{
					name: "running odo deploy",
				},
				{
					name: "running odo deploy with image build extra args",
					imageBuildExtraArgs: []string{
						"--platform=linux/amd64",
						"--build-arg=MY_ARG=my_value",
					},
				},
			} {
				tt := tt
				When(tt.name, func() {
					var stdout string
					BeforeEach(func() {
						env := []string{"PODMAN_CMD=echo"}
						if len(tt.imageBuildExtraArgs) != 0 {
							env = append(env, "ODO_IMAGE_BUILD_ARGS="+strings.Join(tt.imageBuildExtraArgs, ";"))
						}
						stdout = helper.Cmd("odo", "deploy").AddEnv(env...).ShouldPass().Out()
					})
					It("should succeed", func() {
						By("building and pushing image to registry", func() {
							substring := fmt.Sprintf("build -t quay.io/unknown-account/myimage -f %s %s",
								filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context)
							if len(tt.imageBuildExtraArgs) != 0 {
								substring = fmt.Sprintf("build %s -t quay.io/unknown-account/myimage -f %s %s",
									strings.Join(tt.imageBuildExtraArgs, " "), filepath.Join(commonVar.Context, "Dockerfile"), commonVar.Context)
							}
							Expect(stdout).To(ContainSubstring(substring))
							Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
						})
						By("deploying a deployment with the built image", func() {
							out := commonVar.CliRunner.Run("get", "deployment", deploymentName, "-n",
								commonVar.Project, "-o", `jsonpath="{.spec.template.spec.containers[0].image}"`).Wait().Out.Contents()
							Expect(out).To(ContainSubstring("quay.io/unknown-account/myimage"))
						})
					})

					It("should run odo dev successfully", func() {
						devSession, err := helper.StartDevMode(helper.DevSessionOpts{})
						Expect(err).ToNot(HaveOccurred())
						devSession.Kill()
						devSession.WaitEnd()
					})

					When("running and stopping odo dev", func() {
						BeforeEach(func() {
							devSession, err := helper.StartDevMode(helper.DevSessionOpts{})
							Expect(err).ShouldNot(HaveOccurred())
							devSession.Stop()
							devSession.WaitEnd()
						})

						It("should not delete the resources created with odo deploy", func() {
							output := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
							Expect(string(output)).To(ContainSubstring(deploymentName))
						})
					})
				})
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

				When("running odo deploy", func() {
					var stdout string
					BeforeEach(func() {
						stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
					})
					It("should succeed", func() {
						By("building and pushing image to registry", func() {
							Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " +
								filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
							Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
						})
						By("deploying a deployment with the built image in current namespace", func() {
							out := commonVar.CliRunner.Run("get", "deployment", deploymentName, "-n",
								commonVar.Project, "-o", `jsonpath="{.spec.template.spec.containers[0].image}"`).Wait().Out.Contents()
							Expect(out).To(ContainSubstring("quay.io/unknown-account/myimage"))
						})
					})

					When("the env.yaml file still contains a non-current Project", func() {
						BeforeEach(func() {
							odoDir := filepath.Join(commonVar.Context, ".odo", "env")
							helper.MakeDir(odoDir)
							err := helper.CreateFileWithContent(filepath.Join(odoDir, "env.yaml"), `
ComponentSettings:
  Project: another-project
`)
							Expect(err).ShouldNot(HaveOccurred())

						})

						It("should delete the component in the current namespace", func() {
							out := helper.Cmd("odo", "delete", "component", "-f").ShouldPass().Out()
							Expect(out).To(ContainSubstring("Deployment: my-component"))
						})
					})
				})
			})
		})
	}

	When("using a devfile.yaml containing two deploy commands", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-with-two-deploy-commands.yaml"),
				path.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
		})
		It("should run odo deploy", func() {
			stdout := helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
			By("building and pushing image to registry", func() {
				Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " + filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
				Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
			})
			By("deploying a deployment with the built image", func() {
				out := commonVar.CliRunner.Run("get", "deployment", "my-component", "-n", commonVar.Project, "-o", `jsonpath="{.spec.template.spec.containers[0].image}"`).Wait().Out.Contents()
				Expect(out).To(ContainSubstring("quay.io/unknown-account/myimage"))
			})
		})
	})

	When("recording telemetry data", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"),
				path.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
			helper.EnableTelemetryDebug()
			helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
		})
		AfterEach(func() {
			helper.ResetTelemetry()
		})
		It("should record the telemetry data correctly", func() {
			td := helper.GetTelemetryDebugData()
			Expect(td.Event).To(ContainSubstring("odo deploy"))
			Expect(td.Properties.Success).To(BeTrue())
			Expect(td.Properties.Error == "").To(BeTrue())
			Expect(td.Properties.ErrorType == "").To(BeTrue())
			Expect(td.Properties.CmdProperties[segment.ComponentType]).To(ContainSubstring("nodejs"))
			Expect(td.Properties.CmdProperties[segment.Language]).To(ContainSubstring("javascript"))
			Expect(td.Properties.CmdProperties[segment.ProjectType]).To(ContainSubstring("nodejs"))
			Expect(td.Properties.CmdProperties[segment.Flags]).To(BeEmpty())
			Expect(td.Properties.CmdProperties).Should(HaveKey(segment.Caller))
			Expect(td.Properties.CmdProperties[segment.Caller]).To(BeEmpty())
			Expect(td.Properties.CmdProperties[segment.ExperimentalMode]).To(Equal(false))
			if os.Getenv("KUBERNETES") == "true" {
				Expect(td.Properties.CmdProperties[segment.Platform]).To(Equal("kubernetes"))
			} else {
				Expect(td.Properties.CmdProperties[segment.Platform]).To(Equal("openshift"))
			}
			serverVersion := commonVar.CliRunner.GetVersion()
			// Result may or may not be empty, because sometimes `oc version` may not return the OpenShift Server version
			if serverVersion != "" {
				Expect(td.Properties.CmdProperties[segment.PlatformVersion]).To(ContainSubstring(serverVersion))
			}
		})
	})

	When("using a devfile.yaml containing an Image component with a build context", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.Cmd("odo", "init", "--name", "aname",
				"--devfile-path",
				helper.GetExamplePath("source", "devfiles", "nodejs",
					"devfile-outerloop-project_source-in-docker-build-context.yaml")).ShouldPass()
		})

		for _, scope := range []struct {
			name    string
			envvars []string
		}{
			{
				name:    "Podman",
				envvars: []string{"PODMAN_CMD=echo"},
			},
			{
				name: "Docker",
				envvars: []string{
					"PODMAN_CMD=a-command-not-found-for-podman-should-make-odo-fallback-to-docker",
					"DOCKER_CMD=echo",
				},
			},
		} {
			// this is a workaround to ensure that the for loop works with `It` blocks
			scope := scope

			It(fmt.Sprintf("should build image via %s if build context references PROJECT_SOURCE env var", scope.name), func() {
				stdout := helper.Cmd("odo", "deploy").AddEnv(scope.envvars...).ShouldPass().Out()
				lines, err := helper.ExtractLines(stdout)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(lines).ShouldNot(BeEmpty())
				containerImage := "localhost:5000/devfile-nodejs-deploy:0.1.0" // from Devfile yaml file
				dockerfilePath := filepath.Join(commonVar.Context, "Dockerfile")
				buildCtx := commonVar.Context
				expected := fmt.Sprintf("build -t %s -f %s %s", containerImage, dockerfilePath, buildCtx)
				i, found := helper.FindFirstElementIndexByPredicate(lines, func(s string) bool {
					return s == expected
				})
				Expect(found).To(BeTrue(), "line not found: ["+expected+"]")
				Expect(i).ToNot(BeZero(), "line not found at non-zero index: ["+expected+"]")
			})
		}
	})

	for _, ctx := range []struct {
		componentType string
		devfile       string
	}{
		{
			componentType: "K8s",
			devfile:       "devfile-deploy-multiple-k8s-resources-in-single-component.yaml",
		},
		{
			componentType: "OpenShift",
			devfile:       "devfile-deploy-multiple-k8s-resources-in-single-openshift-component.yaml",
		},
	} {
		ctx := ctx
		When(fmt.Sprintf("deploying a Devfile %s component with multiple K8s resources defined", ctx.componentType), func() {
			var out string
			var resources []string

			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", ctx.devfile),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
				out = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
				resources = []string{"Deployment/my-component", "Service/my-component-svc"}
			})

			It(fmt.Sprintf("should have created all the resources defined in the Devfile %s component", ctx.componentType), func() {
				By("checking the output", func() {
					helper.MatchAllInOutput(out, resources)
				})
				By("fetching the resources from the cluster", func() {
					for _, resource := range resources {
						Expect(commonVar.CliRunner.Run("get", resource).Out.Contents()).ToNot(BeEmpty())
					}
				})
			})
		})
	}

	When("deploying a ServiceBinding k8s resource", Label(helper.LabelServiceBinding), Label(helper.LabelSkipOnOpenShift), func() {
		const serviceBindingName = "my-nodejs-app-cluster-sample" // hard-coded from devfile-deploy-with-SB.yaml
		BeforeEach(func() {
			skipLogin := os.Getenv("SKIP_SERVICE_BINDING_TESTS")
			if skipLogin == "true" {
				Skip("Skipping service binding tests as SKIP_SERVICE_BINDING_TESTS is true")
			}

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
		When("odo deploy is run", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-deploy-with-SB.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
				helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()
			})
			It("should successfully deploy the ServiceBinding resource", func() {
				out, err := commonVar.CliRunner.GetServiceBinding(serviceBindingName, commonVar.Project)
				Expect(out).ToNot(BeEmpty())
				Expect(err).To(BeEmpty())
			})
		})

	})

	When("using a devfile.yaml containing an Image component with no build context", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "issue-5600-devfile-with-image-component-and-no-buildContext.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
		})

		for _, scope := range []struct {
			name    string
			envvars []string
		}{
			{
				name:    "Podman",
				envvars: []string{"PODMAN_CMD=echo"},
			},
			{
				name: "Docker",
				envvars: []string{
					"PODMAN_CMD=a-command-not-found-for-podman-should-make-odo-fallback-to-docker",
					"DOCKER_CMD=echo",
				},
			},
		} {
			// this is a workaround to ensure that the for loop works with `It` blocks
			scope := scope

			It(fmt.Sprintf("should build image via %s by defaulting build context to devfile path", scope.name), func() {
				stdout := helper.Cmd("odo", "deploy").AddEnv(scope.envvars...).ShouldPass().Out()
				lines, err := helper.ExtractLines(stdout)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(lines).ShouldNot(BeEmpty())
				containerImage := "localhost:5000/devfile-nodejs-deploy:0.1.0" // from Devfile yaml file
				dockerfilePath := filepath.Join(commonVar.Context, "Dockerfile")
				buildCtx := commonVar.Context
				expected := fmt.Sprintf("build -t %s -f %s %s", containerImage, dockerfilePath, buildCtx)
				i, found := helper.FindFirstElementIndexByPredicate(lines, func(s string) bool {
					return s == expected
				})
				Expect(found).To(BeTrue(), "line not found: ["+expected+"]")
				Expect(i).ToNot(BeZero(), "line not found at non-zero index: ["+expected+"]")
			})
		}
	})

	for _, env := range [][]string{
		{"PODMAN_CMD=echo"},
		{
			"PODMAN_CMD=a-command-not-found-for-podman-should-make-odo-fallback-to-docker",
			"DOCKER_CMD=echo",
		},
	} {
		env := env
		Describe("using a Devfile with an image component using a remote Dockerfile", func() {

			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(
					filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"),
					path.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
			})

			When("remote server returns an error", func() {
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

				It("should not build images", func() {
					cmdWrapper := helper.Cmd("odo", "deploy").AddEnv(env...).ShouldFail()
					stderr := cmdWrapper.Err()
					stdout := cmdWrapper.Out()
					Expect(stderr).To(ContainSubstring("failed to retrieve " + url))
					Expect(stdout).NotTo(ContainSubstring("build -t quay.io/unknown-account/myimage -f "))
					Expect(stdout).NotTo(ContainSubstring("push quay.io/unknown-account/myimage"))
				})
			})

			When("remote server returns a valid file", func() {
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
				})

				AfterEach(func() {
					server.Close()
				})

				It("should run odo deploy", func() {
					stdout := helper.Cmd("odo", "deploy").AddEnv(env...).ShouldPass().Out()

					By("building and pushing images", func() {
						lines, _ := helper.ExtractLines(stdout)
						_, ok := helper.FindFirstElementIndexMatchingRegExp(lines, buildRegexp)
						Expect(ok).To(BeTrue(), "build regexp not found in output: "+buildRegexp)
						Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
					})
				})
			})

		})
	}
	Context("deploying devfile with exec", func() {
		BeforeEach(func() {
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-deploy-exec.yaml"),
				path.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
		})
		for _, ctx := range []struct {
			title, compName string
		}{
			{
				title:    "component name of at max(63) characters length",
				compName: "document-how-odo-translates-container-component-to-deploymentss",
			},
			{
				title:    "component name of a normal character length",
				compName: helper.RandString(6),
			},
		} {
			ctx := ctx
			When(fmt.Sprintf("using devfile that works; with %s", ctx.title), func() {
				BeforeEach(func() {
					helper.UpdateDevfileContent(filepath.Join(commonVar.Context, "devfile.yaml"), []helper.DevfileUpdater{helper.DevfileMetadataNameSetter(ctx.compName)})
				})
				It("should complete the command execution successfully", func() {
					out := helper.Cmd("odo", "deploy").ShouldPass().Out()
					Expect(out).To(ContainSubstring("Executing command in container (command: deploy-exec)"))
				})
			})

			// We check the following tests for character length as long as 63 and for normal character length because for 63 char,
			// the job name will be truncated, and we want to ensure the correct truncated name is used to delete the old job before running a new one so that `odo deploy` does not fail
			When(fmt.Sprintf("the deploy command terminates abruptly; %s", ctx.title), func() {
				BeforeEach(func() {
					helper.UpdateDevfileContent(filepath.Join(commonVar.Context, "devfile.yaml"), []helper.DevfileUpdater{helper.DevfileMetadataNameSetter(ctx.compName)})
					helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), `image: registry.access.redhat.com/ubi8/nodejs-14:latest`, `image: registry.access.redhat.com/ubi8/nodejs-does-not-exist-14:latest`)
					helper.Cmd("odo", "deploy").WithTimeout(10).ShouldFail()
				})
				When("odo deploy command is run again", func() {
					BeforeEach(func() {
						// Restore the Devfile; this is not a required step to test, but we do it to not abruptly terminate the command again
						helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), `image: registry.access.redhat.com/ubi8/nodejs-does-not-exist-14:latest`, `image: registry.access.redhat.com/ubi8/nodejs-14:latest`)
					})
					It("should run successfully", func() {
						helper.Cmd("odo", "deploy").ShouldPass()
					})
				})
			})

			It("should not set securitycontext for podsecurity admission on job's pod template", func() {
				if os.Getenv("KUBERNETES") != "true" {
					Skip("This is a Kubernetes specific scenario, skipping")
				}
				helper.Cmd("odo", "deploy").Should(func(session *gexec.Session) {
					component := helper.NewComponent(cmpName, "app", labels.ComponentDeployMode, commonVar.Project, commonVar.CliRunner)
					jobDef := component.GetJobDef()
					Expect(jobDef.Spec.Template.Spec.SecurityContext.RunAsNonRoot).To(BeNil())
					Expect(jobDef.Spec.Template.Spec.SecurityContext.SeccompProfile).To(BeNil())
				})
			})

		}

		When("using a devfile name with length more than 63", func() {
			const (
				unacceptableLongName = "document-how-odo-translates-container-component-to-deploymentsss"
			)
			BeforeEach(func() {
				helper.UpdateDevfileContent(filepath.Join(commonVar.Context, "devfile.yaml"), []helper.DevfileUpdater{helper.DevfileMetadataNameSetter(unacceptableLongName)})
			})
			It("should fail with invalid component name error", func() {
				errOut := helper.Cmd("odo", "deploy").ShouldFail().Err()
				Expect(errOut).To(SatisfyAll(ContainSubstring(fmt.Sprintf("component name %q is not valid", unacceptableLongName)),
					ContainSubstring("Contain at most 63 characters"),
					ContainSubstring("Start with an alphanumeric character"),
					ContainSubstring("End with an alphanumeric character"),
					ContainSubstring("Must not contain all numeric values")))
			})
		})

		When("using devfile with a long running command in exec", func() {
			BeforeEach(func() {
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), `commandLine: echo Hello world`, `commandLine: sleep 62; echo hello world`)
			})
			It("should print the tip to run odo logs after 1 minute of execution", func() {
				out := helper.Cmd("odo", "deploy").ShouldPass().Out()
				Expect(out).To(ContainSubstring("Tip: Run `odo logs --deploy --follow` to get the logs of the command output."))
			})
		})

		When("using devfile where the exec command is bound to fail", func() {
			BeforeEach(func() {
				// the following new commandLine ensures "counter $i counter" is printed on 99 lines of the output and the last line is a failure from running a non-existent binary
				helper.ReplaceString(filepath.Join(commonVar.Context, "devfile.yaml"), `commandLine: echo Hello world`, `commandLine: for i in {1..100}; do echo counter $i counter; done; run-non-existent-binary`)
			})

			It("should print the last 100 lines of the log to the output", func() {
				out, errOut := helper.Cmd("odo", "deploy").ShouldFail().OutAndErr()
				Expect(out).To(ContainSubstring("Execution output:"))
				// checking 'counter 1 counter' does not exist in the log output ensures that only the last 100 lines are printed
				Expect(errOut).ToNot(ContainSubstring("counter 1 counter"))
				Expect(errOut).To(ContainSubstring("/bin/sh: run-non-existent-binary: command not found"))
			})
		})

	})

	Context("deploying devfile with long-running exec", func() {
		BeforeEach(func() {
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs", "devfile-deploy-exec-long.yaml"),
				path.Join(commonVar.Context, "devfile.yaml"),
				cmpName)
		})

		When("pod security is enforced as restricted", func() {
			BeforeEach(func() {
				commonVar.CliRunner.SetLabelsOnNamespace(
					commonVar.Project,
					"pod-security.kubernetes.io/enforce=restricted",
					"pod-security.kubernetes.io/enforce-version=latest",
				)
			})

			It("should set securitycontext for podsecurity admission on job's pod template", func() {
				if os.Getenv("KUBERNETES") != "true" {
					Skip("This is a Kubernetes specific scenario, skipping")
				}
				helper.Cmd("odo", "deploy").Should(func(session *gexec.Session) {
					component := helper.NewComponent(cmpName, "app", labels.ComponentDeployMode, commonVar.Project, commonVar.CliRunner)
					jobDef := component.GetJobDef()
					Expect(*jobDef.Spec.Template.Spec.SecurityContext.RunAsNonRoot).To(BeTrue())
					Expect(string(jobDef.Spec.Template.Spec.SecurityContext.SeccompProfile.Type)).To(Equal("RuntimeDefault"))
				})
			})
		})

		When("Automount volumes are present in the namespace", func() {

			BeforeEach(func() {
				commonVar.CliRunner.Run("apply", "-f", helper.GetExamplePath("manifests", "config-automount/"))
			})

			It("should mount the volumes", func() {
				helper.Cmd("odo", "deploy").Should(func(session *gexec.Session) {
					component := helper.NewComponent(cmpName, "app", labels.ComponentDeployMode, commonVar.Project, commonVar.CliRunner)
					jobDef := component.GetJobDef()
					// We only check that at least one volume is automounted
					// More tests are executed on `odo dev`, see "Automount volumes are present in the namespace" on odo dev tests.
					Expect(jobDef.Spec.Template.Spec.Volumes[0].Name).To(Equal("auto-pvc-automount-default-pvc"))
				})
			})
		})
	})

	// More details on https://github.com/devfile/api/issues/852#issuecomment-1211928487
	Context("Devfile with autoBuild or deployByDefault components", func() {

		When("starting with Devfile with Deploy commands", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-autobuild-deploybydefault.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
			})

			When("running odo deploy with some components not referenced in the Devfile", func() {
				var stdout string

				BeforeEach(func() {
					stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
				})

				It("should create the appropriate resources", func() {
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

					By("automatically applying image components with autoBuild=true", func() {
						for _, tag := range []string{
							"autobuild-true-and-referenced",
							"autobuild-true-and-not-referenced",
						} {
							Expect(stdout).Should(ContainSubstring("Building & Pushing Image: localhost:5000/odo-dev/node:%s", tag))
						}
					})
					By("automatically applying non-referenced Image components with autoBuild not set", func() {
						Expect(stdout).Should(ContainSubstring("Building & Pushing Image: localhost:5000/odo-dev/node:autobuild-not-set-and-not-referenced"))
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

			When("running odo deploy with some components referenced in the Devfile", func() {
				var stdout string

				BeforeEach(func() {
					//TODO (rm3l): we do not support passing a custom deploy command yet. That's why we are manually updating the Devfile to set the default deploy command.
					helper.UpdateDevfileContent(filepath.Join(commonVar.Context, "devfile.yaml"), []helper.DevfileUpdater{
						helper.DevfileCommandGroupUpdater("deploy", v1alpha2.CompositeCommandType, &v1alpha2.CommandGroup{
							Kind:      v1alpha2.DeployCommandGroupKind,
							IsDefault: pointer.Bool(false),
						}),
						helper.DevfileCommandGroupUpdater("deploy-with-referenced-components", v1alpha2.CompositeCommandType, &v1alpha2.CommandGroup{
							Kind:      v1alpha2.DeployCommandGroupKind,
							IsDefault: pointer.Bool(true),
						}),
					})

					stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
				})

				It("should create the appropriate resources", func() {
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

					By("applying referenced image components", func() {
						for _, tag := range []string{
							"autobuild-true-and-referenced",
							"autobuild-false-and-referenced",
							"autobuild-not-set-and-referenced",
						} {
							Expect(stdout).Should(ContainSubstring("Building & Pushing Image: localhost:5000/odo-dev/node:%s", tag))
						}
					})
					By("automatically applying image components with autoBuild=true", func() {
						for _, tag := range []string{
							"autobuild-true-and-referenced",
							"autobuild-true-and-not-referenced",
						} {
							Expect(stdout).Should(ContainSubstring("Building & Pushing Image: localhost:5000/odo-dev/node:%s", tag))
						}
					})
					By("automatically applying non-referenced Image components with autoBuild not set", func() {
						Expect(stdout).Should(ContainSubstring("Building & Pushing Image: localhost:5000/odo-dev/node:autobuild-not-set-and-not-referenced"))
					})
					By("not applying non-referenced image components with autoBuild=false", func() {
						Expect(stdout).ShouldNot(ContainSubstring("localhost:5000/odo-dev/node:autobuild-false-and-not-referenced"))
					})
				})
			})

		})

		When("starting with Devfile with no Deploy command", func() {
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-autobuild-deploybydefault-no-deploy-cmd.yaml"),
					filepath.Join(commonVar.Context, "devfile.yaml"),
					cmpName)
			})

			It("should fail to run odo deploy", func() {
				stdout, stderr := helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldFail().OutAndErr()
				By("not automatically applying Kubernetes/OpenShift components ", func() {
					for _, s := range []string{stdout, stderr} {
						Expect(s).ShouldNot(ContainSubstring("Creating resource Pod/"))
					}
				})
				By("not automatically applying Image components ", func() {
					for _, s := range []string{stdout, stderr} {
						Expect(s).ShouldNot(ContainSubstring("Building & Pushing Image: localhost:5000/odo-dev/node:"))
					}
				})
				By("displaying an error message", func() {
					Expect(stderr).Should(ContainSubstring("no deploy command found in devfile"))
				})
			})
		})
	})
})
