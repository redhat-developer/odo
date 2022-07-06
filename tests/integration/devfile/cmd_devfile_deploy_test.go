package devfile

import (
	"fmt"
	"path"
	"path/filepath"

	segment "github.com/redhat-developer/odo/pkg/segment/context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/tests/helper"
)

var _ = Describe("odo devfile deploy command tests", func() {

	var commonVar helper.CommonVar

	var _ = BeforeEach(func() {
		commonVar = helper.CommonBeforeEach()
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
			output := helper.Cmd("odo", "deploy").ShouldFail().Err()
			Expect(output).To(ContainSubstring("this command cannot run in an empty directory"))

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
			cmpName := "nodejs-prj1-api-abhz"
			deploymentName := "my-component"
			BeforeEach(func() {
				helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
				helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", ctx.devfileName),
					path.Join(commonVar.Context, "devfile.yaml"))
				if ctx.setupFunc != nil {
					ctx.setupFunc()
				}
			})

			When("running odo deploy", func() {
				var stdout string
				BeforeEach(func() {
					stdout = helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass().Out()
					// An ENV file should have been created indicating current namespace
					Expect(helper.VerifyFileExists(".odo/env/env.yaml")).To(BeTrue())
					helper.FileShouldContainSubstring(".odo/env/env.yaml", "Project: "+commonVar.Project)
				})
				It("should succeed", func() {
					By("building and pushing image to registry", func() {
						Expect(stdout).To(ContainSubstring("build -t quay.io/unknown-account/myimage -f " +
							filepath.Join(commonVar.Context, "Dockerfile ") + commonVar.Context))
						Expect(stdout).To(ContainSubstring("push quay.io/unknown-account/myimage"))
					})
					By("deploying a deployment with the built image", func() {
						out := commonVar.CliRunner.Run("get", "deployment", deploymentName, "-n",
							commonVar.Project, "-o", `jsonpath="{.spec.template.spec.containers[0].image}"`).Wait().Out.Contents()
						Expect(out).To(ContainSubstring("quay.io/unknown-account/myimage"))
					})
				})

				It("should run odo dev successfully", func() {
					session, _, _, _, err := helper.StartDevMode()
					Expect(err).ToNot(HaveOccurred())
					session.Kill()
					session.WaitEnd()
				})

				When("deleting previous deployment and switching kubeconfig to another namespace", func() {
					var otherNS string
					BeforeEach(func() {
						helper.Cmd("odo", "delete", "component", "--name", cmpName, "-f").ShouldPass()
						output := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Err.Contents()
						Expect(string(output)).To(
							ContainSubstring("No resources found in " + commonVar.Project + " namespace."))

						otherNS = commonVar.CliRunner.CreateAndSetRandNamespaceProject()
					})

					AfterEach(func() {
						commonVar.CliRunner.DeleteNamespaceProject(otherNS, false)
					})

					It("should run odo deploy on initial namespace", func() {
						helper.Cmd("odo", "deploy").AddEnv("PODMAN_CMD=echo").ShouldPass()

						output := commonVar.CliRunner.Run("get", "deployment").Err.Contents()
						Expect(string(output)).To(
							ContainSubstring("No resources found in " + otherNS + " namespace."))

						output = commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
						Expect(string(output)).To(ContainSubstring(deploymentName))
					})

				})

				When("running and stopping odo dev", func() {
					BeforeEach(func() {
						session, _, _, _, err := helper.StartDevMode()
						Expect(err).ShouldNot(HaveOccurred())
						session.Stop()
						session.WaitEnd()
					})

					It("should not delete the resources created with odo deploy", func() {
						output := commonVar.CliRunner.Run("get", "deployment", "-n", commonVar.Project).Out.Contents()
						Expect(string(output)).To(ContainSubstring(deploymentName))
					})
				})
			})
		})
	}

	When("using a devfile.yaml containing two deploy commands", func() {
		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-with-two-deploy-commands.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
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
			helper.CopyExampleDevFile(filepath.Join("source", "devfiles", "nodejs", "devfile-deploy.yaml"), path.Join(commonVar.Context, "devfile.yaml"))
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

	When("using a devfile.yaml containing an Image component with no build context", func() {

		BeforeEach(func() {
			helper.CopyExample(filepath.Join("source", "nodejs"), commonVar.Context)
			helper.CopyExampleDevFile(
				filepath.Join("source", "devfiles", "nodejs",
					"issue-5600-devfile-with-image-component-and-no-buildContext.yaml"),
				filepath.Join(commonVar.Context, "devfile.yaml"))
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
})
